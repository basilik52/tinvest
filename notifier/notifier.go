package notifier

import (
	"errors"
	"fmt"

	"tinvest/bot"
	"tinvest/tinkoff"
)

type Notifier struct {
	bot       *bot.Client
	formatter *Formatter
	tinkoff   *tinkoff.TinkoffClient
}

func New(botClient *bot.Client, tinkoffClient *tinkoff.TinkoffClient) *Notifier {
	return &Notifier{
		bot:       botClient,
		formatter: NewFormatter(),
		tinkoff:   tinkoffClient,
	}
}

func (n *Notifier) NotifyTrade(op *tinkoff.Operation) error {
	if n == nil || n.bot == nil {
		return errors.New("bot client is nil")
	}
	if op == nil {
		return errors.New("operation is nil")
	}

	if (op.Ticker == "" || op.Name == "") && op.FIGI != "" && n.tinkoff != nil {
		ticker, name, _, _, err := n.tinkoff.GetInstrumentInfo(op.FIGI)
		if err == nil {
			if op.Ticker == "" {
				op.Ticker = ticker
			}
			if op.Name == "" {
				op.Name = name
			}
		}
	}

	text := n.formatter.FormatTrade(op)
	_, err := n.bot.SendMessage(text)
	return err
}

func (n *Notifier) NotifyError(err error) error {
	if n == nil || n.bot == nil {
		return errors.New("bot client is nil")
	}
	if err == nil {
		return errors.New("error is nil")
	}
	text := n.formatter.FormatError(err)
	_, err = n.bot.SendMessage(text)
	return err
}

func (n *Notifier) NotifyStart(accountName string, balance float64, currency string) error {
	if n == nil || n.bot == nil {
		return errors.New("bot client is nil")
	}
	text := n.formatter.FormatStartBot(accountName, balance, currency)
	_, err := n.bot.SendMessage(text)
	return err
}

func (n *Notifier) NotifyConnectionError(err error) error {
	if n == nil || n.bot == nil {
		return errors.New("bot client is nil")
	}
	if err == nil {
		return errors.New("error is nil")
	}
	//_ := n.formatter.FormatConnectionError(err)
	//_, err = n.bot.SendMessage(text)
	return err
}

func (n *Notifier) NotifyMultipleTrades(count int) error {
	if n == nil || n.bot == nil {
		return errors.New("bot client is nil")
	}
	text := fmt.Sprintf("📢 Найдено %d новых сделок", count)
	_, err := n.bot.SendMessage(text)
	return err
}
