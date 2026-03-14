package tinkoff

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient("test_token", true)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "https://sandbox-invest-public-api.tbank.ru", client.baseURL)
}

func TestNewClientProduction(t *testing.T) {
	client, err := NewClient("test_token", false)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "https://invest-public-api.tbank.ru", client.baseURL)
}

func TestTinkoffClient_Struct(t *testing.T) {
	client := &TinkoffClient{
		token:      "test_token",
		useSandbox: true,
		baseURL:    "https://sandbox-invest-public-api.tbank.ru",
		httpClient: &http.Client{},
	}

	assert.Equal(t, "test_token", client.token)
	assert.True(t, client.useSandbox)
	assert.NotNil(t, client.httpClient)
}

func TestGetAccounts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := AccountsResponse{
			Accounts: []Account{
				{ID: "acc-1", Name: "Test Account"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &TinkoffClient{
		token:      "test_token",
		useSandbox: false,
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	accounts, err := client.GetAccounts()
	assert.NoError(t, err)
	assert.Len(t, accounts, 1)
	assert.Equal(t, "acc-1", accounts[0].ID)
}

func TestGetOperations_Executed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := OperationsResponse{
			Operations: []Operation{
				{
					ID:            "op-1",
					OperationType: "Buy",
					Status:        "OPERATION_STATE_EXECUTED",
					Date:          time.Now(),
					FIGI:          "BBG000B12345",
					Quantity:      10,
					Payment:       MoneyValue{Units: 1500, Nano: 0, Currency: "RUB"},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &TinkoffClient{
		token:      "test_token",
		useSandbox: false,
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	ops, err := client.GetOperations("acc-1", time.Now().Add(-time.Hour), time.Now())
	assert.NoError(t, err)
	assert.Len(t, ops, 1)
	assert.Equal(t, "op-1", ops[0].ID)
}

func TestGetPortfolio(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := PortfolioResponse{
			TotalAmountPortfolio: MoneyValue{Units: 10000, Nano: 500000000, Currency: "RUB"},
			Positions: []Position{
				{
					FIGI:   "BBG000B12345",
					Ticker: "TEST",
					Name:   "Test Stock",
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &TinkoffClient{
		token:      "test_token",
		useSandbox: false,
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	portfolio, err := client.GetPortfolio("acc-1")
	assert.NoError(t, err)
	assert.Equal(t, 10000.5, portfolio.TotalAmount)
	assert.Len(t, portfolio.Positions, 1)
}

func TestGetLastOperationID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := OperationsResponse{
			Operations: []Operation{
				{ID: "op-latest", Status: "OPERATION_STATE_EXECUTED"},
				{ID: "op-older", Status: "OPERATION_STATE_EXECUTED"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &TinkoffClient{
		token:      "test_token",
		useSandbox: false,
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	id, err := client.GetLastOperationID("acc-1")
	assert.NoError(t, err)
	assert.Equal(t, "op-latest", id)
}

func TestGetLastOperationID_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := OperationsResponse{
			Operations: []Operation{},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &TinkoffClient{
		token:      "test_token",
		useSandbox: false,
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	id, err := client.GetLastOperationID("acc-1")
	assert.NoError(t, err)
	assert.Equal(t, "", id)
}

func TestClose(t *testing.T) {
	client := &TinkoffClient{}
	err := client.Close()
	assert.NoError(t, err)
}

func TestMakeRequestInvalidURL(t *testing.T) {
	client := &TinkoffClient{
		token:      "test_token",
		useSandbox: false,
		baseURL:    "://invalid",
		httpClient: &http.Client{},
	}

	_, err := client.makeRequest(context.Background(), "POST", "/test", []byte("{}"))
	assert.Error(t, err)
}

func TestMoneyValueFloat(t *testing.T) {
	mv := MoneyValue{
		Units:    100,
		Nano:     500000000,
		Currency: "RUB",
	}

	result := mv.Float()
	assert.Equal(t, 100.5, result)
}

func TestMoneyValueIsZero(t *testing.T) {
	tests := []struct {
		mv     MoneyValue
		isZero bool
	}{
		{MoneyValue{Units: 0, Nano: 0}, true},
		{MoneyValue{Units: 1, Nano: 0}, false},
		{MoneyValue{Units: 0, Nano: 1}, false},
		{MoneyValue{Units: 100, Nano: 500000000}, false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.isZero, tt.mv.IsZero())
	}
}

func TestOperationStruct(t *testing.T) {
	op := Operation{
		ID:            "op-1",
		OperationType: "Buy",
		Status:        "Done",
		Date:          time.Now(),
		Name:          "Test Stock",
		Ticker:        "TEST",
		FIGI:          "BBG000B12345",
		Quantity:      10,
		Price:         150.50,
		Total:         1505.00,
		Currency:      "RUB",
		Commission:    MoneyValue{Units: 1, Nano: 0},
		AccountID:     "account-1",
	}

	assert.Equal(t, "op-1", op.ID)
	assert.Equal(t, "Buy", op.OperationType)
	assert.Equal(t, int64(10), op.Quantity)
}

func TestAccountStruct(t *testing.T) {
	acc := Account{
		ID:     "acc-1",
		Name:   "Тестовый счёт",
		Status: "Open",
		Type:   "Tinkoff",
	}

	assert.Equal(t, "acc-1", acc.ID)
	assert.Equal(t, "Тестовый счёт", acc.Name)
}

func TestPortfolioStruct(t *testing.T) {
	p := Portfolio{
		TotalAmount: 100000.50,
		Currency:    "RUB",
		Positions: []Position{
			{
				FIGI:         "BBG000B12345",
				Ticker:       "TEST",
				Name:         "Тест",
				Quantity:     10,
				AveragePrice: MoneyValue{Units: 100, Nano: 0},
				CurrentPrice: MoneyValue{Units: 150, Nano: 0},
				Currency:     "RUB",
			},
		},
	}

	assert.Equal(t, 100000.50, p.TotalAmount)
	assert.Len(t, p.Positions, 1)
	assert.Equal(t, "TEST", p.Positions[0].Ticker)
}

func TestPositionStruct(t *testing.T) {
	pos := Position{
		FIGI:              "BBG000B12345",
		Ticker:            "AAPL",
		Name:              "Apple Inc.",
		Quantity:          100,
		QuantityAvailable: 100,
		AveragePrice:      MoneyValue{Units: 150, Nano: 500000000},
		CurrentPrice:      MoneyValue{Units: 175, Nano: 250000000},
		Currency:          "USD",
	}

	assert.Equal(t, "AAPL", pos.Ticker)
	assert.Equal(t, int64(100), pos.Quantity)
	assert.Equal(t, 150.5, pos.AveragePrice.Float())
}

func TestTradeStruct(t *testing.T) {
	trade := Trade{
		TradeID:  "trade-1",
		Date:     time.Now(),
		Price:    MoneyValue{Units: 100, Nano: 500000000},
		Quantity: 10,
	}

	assert.Equal(t, "trade-1", trade.TradeID)
	assert.Equal(t, int64(10), trade.Quantity)
	assert.Equal(t, 100.5, trade.Price.Float())
}
