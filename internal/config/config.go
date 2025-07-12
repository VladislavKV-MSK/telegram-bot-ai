package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken  string
	OpenRouterKey  string
	AdminIDs       map[int64]bool
	BotDebug       bool
	ModelName      string
	MaxHistorySize int
}

func Load() *Config {
	_ = godotenv.Load()

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

	return &Config{
		TelegramToken:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		OpenRouterKey:  os.Getenv("OPENROUTER_API_KEY"),
		AdminIDs:       adminIDs,
		BotDebug:       os.Getenv("BOT_DEBUG") == "false",
		ModelName:      getEnv("MODEL_NAME", "deepseek/deepseek-r1-0528-qwen3-8b:free"),
		MaxHistorySize: getEnvAsInt("MAX_HISTORY_SIZE", 10),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
