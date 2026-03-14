package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "test_token")
	t.Setenv("TELEGRAM_GROUP_ID", "-1001234567890")
	t.Setenv("TINKOFF_TOKEN", "test_tinkoff_token")
	t.Setenv("TINKOFF_ACCOUNT_ID", "test_account")
	t.Setenv("DB_PATH", "/tmp/test.db")
	t.Setenv("POLL_INTERVAL", "60s")
	t.Setenv("USE_SANDBOX", "true")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "test_token", cfg.TelegramBotToken)
	assert.Equal(t, int64(-1001234567890), cfg.TelegramGroupID)
	assert.Equal(t, "test_tinkoff_token", cfg.TinkoffToken)
	assert.Equal(t, "test_account", cfg.TinkoffAccountID)
	assert.Equal(t, "/tmp/test.db", cfg.DBPath)
	assert.Equal(t, 60*time.Second, cfg.PollInterval)
	assert.True(t, cfg.UseSandbox)
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "token")
	t.Setenv("TELEGRAM_GROUP_ID", "-100")
	t.Setenv("POLL_INTERVAL", "")
	t.Setenv("USE_SANDBOX", "")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, 30*time.Second, cfg.PollInterval)
	assert.False(t, cfg.UseSandbox)
}
