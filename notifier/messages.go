package notifier

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"tinvest/tinkoff"
)

type Formatter struct{}

var kemerovoLocation *time.Location

func init() {
	loc, err := time.LoadLocation("Asia/Krasnoyarsk")
	if err != nil {
		kemerovoLocation = time.FixedZone("Kemerovo", 7*3600)
	} else {
		kemerovoLocation = loc
	}
}

func NewFormatter() *Formatter {
	return &Formatter{}
}

func (f *Formatter) FormatTrade(op *tinkoff.Operation) string {
	var emoji string
	var opType string
	switch op.OperationType {
	case "OPERATION_TYPE_BUY", "Buy":
		emoji = "📈"
		opType = "Покупка"
	case "OPERATION_TYPE_SELL", "Sell":
		emoji = "📉"
		opType = "Продажа"
	default:
		emoji = "💱"
		opType = op.OperationType
	}

	var status string
	switch op.Status {
	case "Done":
		status = "✅ Исполнена"
	case "Declined":
		status = "❌ Отклонена"
	case "Pending":
		status = "⏳ В ожидании"
	case "":
		status = "✅ Исполнена"
	default:
		status = "❓ " + op.Status
	}

	text := fmt.Sprintf(`%s <b>Новая сделка</b>

🏷 <b>Тип:</b> %s %s
📊 <b>Статус:</b> %s
🕐 <b>Время:</b> %s

🎫 <b>Тикер:</b> %s
💰 <b>Название:</b> %s
📁 <b>FIGI:</b> %s

💵 <b>Количество:</b> %s
💎 <b>Цена:</b> %s
💵 <b>Сумма:</b> %s %s
	`,
		emoji,
		opType, emoji,
		status,
		formatTimeKemerovo(op.Date),
		formatTickerWithLink(op.Ticker, op.FIGI),
		op.Name,
		op.FIGI,
		formatQuantity(op.Int64()),
		formatPrice(op.Price),
		formatMoney(calculateTotal(op), op.Currency),
		op.Currency,
	)

	if !op.Commission.IsZero() {
		text += fmt.Sprintf(`💸 <b>Комиссия:</b> %s %s\n`,
			formatMoney(op.Commission.Float(), op.Currency),
			op.Currency,
		)
	}

	return text
}

func (f *Formatter) FormatError(err error) string {
	return fmt.Sprintf(`❌ <b>Ошибка</b>

%s`, err.Error())
}

func (f *Formatter) FormatStartBot(accountName string, balance float64, currency string) string {
	return fmt.Sprintf(`🤖 <b>Бот запущен</b>

💼 <b>Счёт:</b> %s
💰 <b>Баланс:</b> %s %s
`,
		accountName,
		formatMoney(balance, currency),
		currency,
	)
}

func (f *Formatter) FormatConnectionError(err error) string {
	return fmt.Sprintf(`🔌 <b>Ошибка подключения</b>

%s

Бот попытается переподключиться автоматически...`, err.Error())
}

func formatQuantity(q int64) string {
	return strconv.FormatInt(q, 10)
}

func formatPrice(price tinkoff.MoneyValue) string {
	return strconv.FormatFloat(price.Float(), 'f', 2, 64)
}

func formatMoney(amount float64, currency string) string {
	return strconv.FormatFloat(amount, 'f', 2, 64)
}

func calculateTotal(op *tinkoff.Operation) float64 {
	total := op.Float()
	if total > 0 {
		return total
	}
	return op.Price.Float() * float64(op.Int64())
}

func formatTimeKemerovo(t time.Time) string {
	return t.In(kemerovoLocation).Format("02.01.2006 15:04:05")
}

func getTicker(ticker, figi string) string {
	if ticker != "" {
		return ticker
	}
	return figiToTicker(figi)
}

func formatTickerWithLink(ticker, figi string) string {
	tickerStr := getTicker(ticker, figi)
	cleanTicker := cleanTicker(tickerStr)
	url := fmt.Sprintf("https://www.tbank.ru/invest/stocks/%s?utm_source=security_share", cleanTicker)
	return fmt.Sprintf(`<a href="%s">%s</a>`, url, tickerStr)
}

func cleanTicker(ticker string) string {
	if idx := strings.Index(ticker, " ("); idx > 0 {
		return ticker[:idx]
	}
	return ticker
}

func getInstrumentName(op *tinkoff.Operation) string {
	if op.Ticker != "" && op.Name != "" {
		return fmt.Sprintf("%s (%s)", op.Name, op.Ticker)
	}
	if op.Ticker != "" {
		return op.Ticker
	}
	if op.Name != "" {
		return op.Name
	}
	if op.FIGI != "" {
		return figiToTicker(op.FIGI)
	}
	return "Неизвестно"
}

func figiToTicker(figi string) string {
	mappings := map[string]string{
		"BBG004730N88": "SBER (Сбербанк)",
		"BBG004GZ4WN1": "SBERP (Сбербанк-п)",
		"BBG002B2MQR3": "GAZP (Газпром)",
		"BBG004S68B31": "YNDX (Яндекс)",
		"BBG00475JY02": "LKOH (Лукойл)",
		"BBG004RVFCQ2": "MGNT (Магнит)",
		"BBG006L5STG1": "MTSS (МТС)",
		"BBG003Z3JCS3": "TATN (Татнефть)",
		"BBG005HKJ9V5": "SNGSP (Сургутнефтегаз)",
		"BBG0047Z6NC0": "NVTK (Новатэк)",
		"BBG004730ZJ9": "VTBR (ВТБ)",
	}
	if ticker, ok := mappings[figi]; ok {
		return ticker
	}
	return figi
}

func ParseTime(s string) (time.Time, error) {
	layouts := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05.999Z07:00",
		"2006-01-02T15:04:05",
		"02.01.2006 15:04:05",
	}

	for _, layout := range layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}
