package bot

import (
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Client struct {
	api     *tgbotapi.BotAPI
	groupID int64
}

func NewClient(token string, groupID int64) (*Client, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	return &Client{
		api:     api,
		groupID: groupID,
	}, nil
}

func (c *Client) SendMessage(text string) (int, error) {
	msg := tgbotapi.NewMessage(c.groupID, text)
	msg.ParseMode = "HTML"

	sent, err := c.api.Send(msg)
	if err != nil {
		return 0, fmt.Errorf("failed to send message: %w", err)
	}

	return sent.MessageID, nil
}

func (c *Client) GetGroupID() int64 {
	return c.groupID
}

func (c *Client) IsGroupMessage(message *tgbotapi.Message) bool {
	if message == nil || message.Chat == nil {
		return false
	}
	return message.Chat.ID == c.groupID
}

func (c *Client) ParseGroupID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
