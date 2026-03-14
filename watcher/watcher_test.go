package watcher

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"tinvest/tinkoff"
)

func TestWatcher_New(t *testing.T) {
	client := &tinkoff.TinkoffClient{}

	watcher := New(client, nil, nil, 30*time.Second, "test-account")

	assert.NotNil(t, watcher)
	assert.Equal(t, "test-account", watcher.accountID)
	assert.Equal(t, 30*time.Second, watcher.interval)
}

func TestWatcher_GetAccountID(t *testing.T) {
	client := &tinkoff.TinkoffClient{}
	watcher := New(client, nil, nil, 30*time.Second, "test-account")

	assert.Equal(t, "test-account", watcher.GetAccountID())
}

func TestWatcher_GetAccountIDEmpty(t *testing.T) {
	client := &tinkoff.TinkoffClient{}
	watcher := New(client, nil, nil, 30*time.Second, "")

	assert.Equal(t, "", watcher.GetAccountID())
}
