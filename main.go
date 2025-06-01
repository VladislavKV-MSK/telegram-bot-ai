package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

// Message представляет одно сообщение в диалоге с ИИ
type Message struct {
	Role    string `json:"role"`    // "user" или "assistant"
	Content string `json:"content"` // Текст сообщения
}

// UserContext хранит контекст диалога для конкретного пользователя
type UserContext struct {
	Messages []Message // История сообщений пользователя и бота
}

// BotState хранит состояние бота для всех пользователей
type BotState struct {
	UserContexts map[int64]*UserContext // Контексты по ID пользователей
	mu           sync.Mutex             // Мьютекс для безопасного доступа из горутин
}

// NewBotState создает новое состояние бота
func NewBotState() *BotState {
	return &BotState{
		UserContexts: make(map[int64]*UserContext),
	}
}

// AddMessage добавляет сообщение в контекст пользователя, сохраняя только последние 10
func (bs *BotState) AddMessage(userID int64, role, content string) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	// Если у пользователя еще нет контекста, создаем новый
	if _, ok := bs.UserContexts[userID]; !ok {
		bs.UserContexts[userID] = &UserContext{
			Messages: make([]Message, 0),
		}
	}

	// Добавляем новое сообщение
	bs.UserContexts[userID].Messages = append(bs.UserContexts[userID].Messages, Message{
		Role:    role,
		Content: content,
	})

	// Ограничиваем историю последними 10 сообщениями
	if len(bs.UserContexts[userID].Messages) > 10 {
		bs.UserContexts[userID].Messages = bs.UserContexts[userID].Messages[len(bs.UserContexts[userID].Messages)-10:]
	}
}

// GetMessages возвращает текущий контекст пользователя
func (bs *BotState) GetMessages(userID int64) []Message {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if ctx, ok := bs.UserContexts[userID]; ok {
		return ctx.Messages
	}
	return []Message{}
}

// OpenRouterRequest представляет запрос к OpenRouter API
type OpenRouterRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Provider struct {
		AllowFallbacks bool     `json:"allow_fallbacks"`
		Order          []string `json:"order"` // Указываем порядок провайдеров
		MaxPrice       struct {
			Prompt     float64 `json:"prompt"`
			Completion float64 `json:"completion"`
		} `json:"max_price"`
		RequireParameters bool `json:"require_parameters"`
	} `json:"provider"`
}

// OpenRouterResponse представляет ответ от OpenRouter API
type OpenRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// queryDeepSeek отправляет запрос к DeepSeek через OpenRouter и возвращает ответ
func queryDeepSeek(messages []Message, apiKey string) (string, error) {
	requestBody := OpenRouterRequest{
		Model:    "deepseek/deepseek-r1-0528-qwen3-8b",
		Messages: messages,
	}

	// Устанавливаем параметры провайдера
	requestBody.Provider.AllowFallbacks = true
	requestBody.Provider.RequireParameters = true
	requestBody.Provider.Order = []string{"DeepSeek"} // Явно указываем провайдера
	requestBody.Provider.MaxPrice.Prompt = 0
	requestBody.Provider.MaxPrice.Completion = 0

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://github.com")
	req.Header.Set("X-Title", "Telegram Bot")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	var response OpenRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	if response.Error.Message != "" {
		return "", fmt.Errorf("API error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return response.Choices[0].Message.Content, nil
}

func main() {
	// Загрузка переменных окружения
	_ = godotenv.Load()

	// Получение токенов
	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	openRouterToken := os.Getenv("OPENROUTER_API_KEY")
	if telegramToken == "" || openRouterToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN and OPENROUTER_API_KEY must be set")
	}

	// Инициализация бота
	bot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = false
	botName := "@" + bot.Self.UserName
	log.Printf("Authorized as %s", bot.Self.UserName)

	// Список администраторов
	adminIDs := make(map[int64]bool)
	adminList := strings.Split(os.Getenv("ADMIN_IDS"), ",")
	for _, idStr := range adminList {
		if idStr == "" {
			continue
		}
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Printf("Invalid admin ID: %s", idStr)
			continue
		}
		adminIDs[id] = true
	}

	// Состояние чатов
	activeChats := sync.Map{}
	botState := NewBotState()

	// Настройка обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	// Обработка сообщений
	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		userID := update.Message.From.ID
		isPrivate := update.Message.Chat.IsPrivate()
		isAdmin := adminIDs[userID]

		// Обработка команд
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				if !isAdmin {
					msg := tgbotapi.NewMessage(chatID, "❌ Только администратор может активировать бота")
					bot.Send(msg)
					continue
				}
				activeChats.Store(chatID, true)
				msg := tgbotapi.NewMessage(chatID, "✅ Бот активирован администратором")
				if !isPrivate {
					msg.Text += "\nОтвечаю только на упоминания " + botName
				}
				bot.Send(msg)
				continue

			case "stop":
				if !isAdmin {
					msg := tgbotapi.NewMessage(chatID, "❌ Только администратор может деактивировать бота")
					bot.Send(msg)
					continue
				}
				activeChats.Delete(chatID)
				botState.mu.Lock()
				delete(botState.UserContexts, userID)
				botState.mu.Unlock()
				msg := tgbotapi.NewMessage(chatID, "✅ Бот деактивирован администратором")
				bot.Send(msg)
				continue
			}
		}

		// Проверка доступа
		if isPrivate {
			if !isAdmin {
				msg := tgbotapi.NewMessage(chatID, "❌ У вас нет доступа к этому боту")
				bot.Send(msg)
				continue
			}
			// Проверка активации в личном чате
			if _, active := activeChats.Load(chatID); !active {
				msg := tgbotapi.NewMessage(chatID, "ℹ️ Для начала работы отправьте /start")
				bot.Send(msg)
				continue
			}
		} else {
			// Проверка активации в группе
			if _, active := activeChats.Load(chatID); !active {
				continue
			}
			// Проверка упоминания в группе
			if !strings.Contains(strings.ToLower(update.Message.Text), strings.ToLower(botName)) {
				continue
			}
		}

		// Очистка текста от упоминания
		cleanedText := strings.ReplaceAll(update.Message.Text, botName, "")
		cleanedText = strings.TrimSpace(cleanedText)
		if cleanedText == "" {
			continue
		}

		// Добавление сообщения в контекст
		botState.AddMessage(userID, "user", cleanedText)
		messages := botState.GetMessages(userID)

		// Запрос к DeepSeek
		response, err := queryDeepSeek(messages, openRouterToken)
		if err != nil {
			log.Printf("DeepSeek error: %v", err)
			msg := tgbotapi.NewMessage(chatID, "⚠️ Произошла ошибка. Попробуйте позже.")
			bot.Send(msg)
			continue
		}

		// Сохранение ответа и отправка
		botState.AddMessage(userID, "assistant", response)
		msg := tgbotapi.NewMessage(chatID, response)
		msg.ReplyToMessageID = update.Message.MessageID
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Send message error: %v", err)
		}
	}
}
