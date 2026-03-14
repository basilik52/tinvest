package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramBotToken string
	TelegramGroupID  int64
	TinkoffToken     string
	TinkoffAccountID string
	DBPath           string
	PollInterval     time.Duration
	UseSandbox       bool
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	pollInterval := os.Getenv("POLL_INTERVAL")
	if pollInterval == "" {
		pollInterval = "30s"
	}

	interval, err := time.ParseDuration(pollInterval)
	if err != nil {
		interval = 30 * time.Second
	}

	useSandbox := os.Getenv("USE_SANDBOX")
	sandboxMode := useSandbox == "1" || useSandbox == "true"

	groupID, _ := strconv.ParseInt(os.Getenv("TELEGRAM_GROUP_ID"), 10, 64)

	return &Config{
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramGroupID:  groupID,
		TinkoffToken:     os.Getenv("TINKOFF_TOKEN"),
		TinkoffAccountID: os.Getenv("TINKOFF_ACCOUNT_ID"),
		DBPath:           os.Getenv("DB_PATH"),
		PollInterval:     interval,
		UseSandbox:       sandboxMode,
	}, nil
}
