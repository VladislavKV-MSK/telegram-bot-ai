package bot

import (
	"fmt"
	"strings"

	"github.com/VladislavKV-MSK/telegram-bot-ai/internal/metrics"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleMessage(msg *tgbotapi.Message) error {
	metrics.MessagesProcessed.Inc()

	chatID := msg.Chat.ID
	userID := msg.From.ID
	isPrivate := msg.Chat.IsPrivate()
	isAdmin := b.config.AdminIDs[userID]

	// Обработка команд
	if msg.IsCommand() {
		return b.handleCommand(chatID, userID, isPrivate, isAdmin, msg.Command())
	}

	// Проверка доступа
	if isPrivate {
		if !isAdmin {
			return b.sendMessage(chatID, "❌ У вас нет доступа к этому боту", 0)
		}
		if _, active := b.activeChats.Load(chatID); !active {
			return b.sendMessage(chatID, "ℹ️ Для начала работы отправьте /start", 0)
		}
	} else {
		if _, active := b.activeChats.Load(chatID); !active {
			return nil
		}
		if !strings.Contains(strings.ToLower(msg.Text), strings.ToLower(b.botName())) {
			return nil
		}
	}

	// Очистка текста от упоминания
	cleanedText := strings.ReplaceAll(msg.Text, b.botName(), "")
	cleanedText = strings.TrimSpace(cleanedText)
	if cleanedText == "" {
		return nil
	}

	// Обработка сообщения
	return b.handleUserMessage(chatID, userID, msg.MessageID, cleanedText)
}

func (b *Bot) handleCommand(chatID, userID int64, isPrivate, isAdmin bool, command string) error {
	switch command {
	case "start":
		if !isAdmin {
			return b.sendMessage(chatID, "❌ Только администратор может активировать бота", 0)
		}
		b.activeChats.Store(chatID, true)
		metrics.ActiveUsers.Inc()
		msg := "✅ Бот активирован администратором"
		if !isPrivate {
			msg += "\nОтвечаю только на упоминания " + b.botName()
		}
		return b.sendMessage(chatID, msg, 0)

	case "stop":
		if !isAdmin {
			return b.sendMessage(chatID, "❌ Только администратор может деактивировать бота", 0)
		}
		b.activeChats.Delete(chatID)
		b.state.ResetUser(userID)
		metrics.ActiveUsers.Dec()
		return b.sendMessage(chatID, "✅ Бот деактивирован администратором", 0)

	default:
		return nil
	}
}

func (b *Bot) handleUserMessage(chatID, userID int64, messageID int, text string) error {
	// Добавление сообщения в контекст
	b.state.AddMessage(userID, "user", text)
	messages := b.state.GetMessages(userID)

	// Запрос к API
	response, err := b.orClient.Query(messages)
	if err != nil {
		metrics.APIErrors.WithLabelValues("openrouter").Inc()
		return fmt.Errorf("API query error: %w", err)
	}

	// Сохранение ответа и отправка
	b.state.AddMessage(userID, "assistant", response)
	return b.sendMessage(chatID, response, messageID)
}

func (b *Bot) sendMessage(chatID int64, text string, replyTo int) error {
	msg := tgbotapi.NewMessage(chatID, text)
	if replyTo > 0 {
		msg.ReplyToMessageID = replyTo
	}
	_, err := b.api.Send(msg)
	return err
}
