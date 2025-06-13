package bot

import (
	"log"
	"sync"
	"time"

	"github.com/VladislavKV-MSK/telegram-bot-ai/internal/config"
	"github.com/VladislavKV-MSK/telegram-bot-ai/internal/metrics"
	"github.com/VladislavKV-MSK/telegram-bot-ai/internal/openrouter"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api         *tgbotapi.BotAPI
	config      *config.Config
	orClient    *openrouter.Client
	state       *State
	activeChats sync.Map
}

func New(cfg *config.Config) (*Bot, error) {
	botAPI, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		return nil, err
	}

	botAPI.Debug = cfg.BotDebug

	return &Bot{
		api:         botAPI,
		config:      cfg,
		orClient:    openrouter.NewClient(cfg.OpenRouterKey, cfg.ModelName),
		state:       NewState(cfg.MaxHistorySize),
		activeChats: sync.Map{},
	}, nil
}

func (b *Bot) Start() error {
	log.Printf("Authorized as @%s", b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		startTime := time.Now()
		if err := b.handleMessage(update.Message); err != nil {
			log.Printf("Error handling message: %v", err)
		}
		metrics.TrackResponseTime(startTime)
	}

	return nil
}

func (b *Bot) botName() string {
	return "@" + b.api.Self.UserName
}
