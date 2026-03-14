package notifier

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tinvest/tinkoff"
)

func TestNewFormatter(t *testing.T) {
	f := NewFormatter()
	assert.NotNil(t, f)
}

func TestFormatTrade(t *testing.T) {
	f := NewFormatter()

	op := &tinkoff.Operation{
		ID:            "op-1",
		OperationType: "Buy",
		Status:        "Done",
		Date:          time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Name:          "Test Stock",
		Ticker:        "TEST",
		FIGI:          "BBG000B12345",
		Quantity:      10,
		Price:         150.50,
		Total:         1505.00,
		Currency:      "RUB",
	}

	result := f.FormatTrade(op)

	require.NotEmpty(t, result)
	assert.Contains(t, result, "Buy")
	assert.Contains(t, result, "TEST")
	assert.Contains(t, result, "10")
}

func TestFormatTradeSell(t *testing.T) {
	f := NewFormatter()

	op := &tinkoff.Operation{
		OperationType: "Sell",
		Status:        "Done",
		Date:          time.Now(),
		Name:          "Test",
		Ticker:        "TEST",
		FIGI:          "FIGI123",
		Quantity:      5,
		Price:         200.00,
		Total:         1000.00,
		Currency:      "USD",
	}

	result := f.FormatTrade(op)
	assert.Contains(t, result, "📉")
	assert.Contains(t, result, "Sell")
}

func TestFormatError(t *testing.T) {
	f := NewFormatter()

	result := f.FormatError(assert.AnError)

	assert.Contains(t, result, "Ошибка")
}

func TestFormatStartBot(t *testing.T) {
	f := NewFormatter()

	result := f.FormatStartBot("Test Account", 10000.50, "RUB")

	assert.Contains(t, result, "Бот запущен")
	assert.Contains(t, result, "Test Account")
	assert.Contains(t, result, "10000.50")
}

func TestFormatConnectionError(t *testing.T) {
	f := NewFormatter()

	result := f.FormatConnectionError(assert.AnError)

	assert.Contains(t, result, "Ошибка подключения")
	assert.Contains(t, result, "переподключиться")
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"2024-01-15T10:30:00Z", false},
		{"2024-01-15T10:30:00+03:00", false},
		{"2024-01-15T10:30:00.123Z", false},
		{"15.01.2024 10:30:00", false},
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		_, err := ParseTime(tt.input)
		if tt.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestFormatQuantity(t *testing.T) {
	assert.Equal(t, "0", formatQuantity(0))
	assert.Equal(t, "10", formatQuantity(10))
	assert.Equal(t, "-5", formatQuantity(-5))
}

func TestFormatPrice(t *testing.T) {
	assert.Equal(t, "0.00", formatPrice(0))
	assert.Equal(t, "100.00", formatPrice(100))
	assert.Equal(t, "150.50", formatPrice(150.5))
}

func TestFormatMoney(t *testing.T) {
	assert.Equal(t, "0.00", formatMoney(0, "RUB"))
	assert.Equal(t, "100.50", formatMoney(100.50, "RUB"))
}
