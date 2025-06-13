package main

import (
	"encoding/json"
	"log"

	"github.com/VladislavKV-MSK/telegram-bot-ai/internal/bot"
	"github.com/VladislavKV-MSK/telegram-bot-ai/internal/config"
	"github.com/VladislavKV-MSK/telegram-bot-ai/internal/metrics"
)

func main() {
	metrics.Init()
	metrics.StartMetricsServer()

	cfg := config.Load()
	if cfg.TelegramToken == "" || cfg.OpenRouterKey == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN and OPENROUTER_API_KEY must be set")
	}
	cfgJSON, _ := json.MarshalIndent(cfg, "", "  ")
	log.Printf("User:\n%s", string(cfgJSON))

	botInstance, err := bot.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	if err := botInstance.Start(); err != nil {
		log.Fatalf("Bot stopped with error: %v", err)
	}
}
