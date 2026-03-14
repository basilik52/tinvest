package tinkoff

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type TinkoffClient struct {
	token      string
	useSandbox bool
	baseURL    string
	httpClient *http.Client
}

type AccountsResponse struct {
	Accounts []Account `json:"accounts"`
}

type OperationsResponse struct {
	Operations []Operation `json:"operations"`
}

type PortfolioResponse struct {
	TotalAmountPortfolio MoneyValue `json:"totalAmountPortfolio"`
	Positions            []Position `json:"positions"`
}

type MoneyValue struct {
	Units    interface{} `json:"units"` // Can be string or int
	Nano     interface{} `json:"nano"`  // Can be string or int
	Currency string      `json:"currency"`
}

func (m MoneyValue) Float() float64 {
	units := m.toInt64(m.Units)
	nano := m.toInt64(m.Nano)
	return float64(units) + float64(nano)/1e9
}

func (m MoneyValue) toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case float64:
		return int64(val)
	case string:
		n, _ := strconv.ParseInt(val, 10, 64)
		return n
	case int64:
		return val
	case int:
		return int64(val)
	case int32:
		return int64(val)
	}
	return 0
}

func (m MoneyValue) IsZero() bool {
	return m.toInt64(m.Units) == 0 && m.toInt64(m.Nano) == 0
}

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case float64:
		return int64(val)
	case string:
		n, _ := strconv.ParseInt(val, 10, 64)
		return n
	case int64:
		return val
	case int:
		return int64(val)
	case int32:
		return int64(val)
	}
	return 0
}

func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	case int64:
		return float64(val)
	case int:
		return float64(val)
	case int32:
		return float64(val)
	}
	return 0
}

type Operation struct {
	ID            string      `json:"id"`
	OperationType string      `json:"operationType"`
	Status        string      `json:"status"`
	Date          time.Time   `json:"date"`
	Name          string      `json:"name,omitempty"`
	Ticker        string      `json:"ticker,omitempty"`
	FIGI          string      `json:"figi,omitempty"`
	Quantity      interface{} `json:"quantity,omitempty"`
	Payment       MoneyValue  `json:"payment,omitempty"`
	Price         MoneyValue  `json:"price,omitempty"`
	Total         interface{} `json:"total,omitempty"`
	Currency      string      `json:"currency,omitempty"`
	Commission    MoneyValue  `json:"commission,omitempty"`
	Trades        []Trade     `json:"trades,omitempty"`
	AccountID     string      `json:"accountId,omitempty"`
}

func (o Operation) Int64() int64 {
	return toInt64(o.Quantity)
}

func (o Operation) Float() float64 {
	return toFloat(o.Total)
}

func isInvestKopilka(ticker, figi string) bool {
	investKopilkaPrefixes := []string{"TCS10", "TCS11", "TCS12"}
	for _, prefix := range investKopilkaPrefixes {
		if len(ticker) >= len(prefix) && ticker[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

type Trade struct {
	TradeID  string      `json:"tradeId"`
	Date     time.Time   `json:"date"`
	Price    MoneyValue  `json:"price"`
	Quantity interface{} `json:"quantity"`
}

func (t Trade) Int64() int64 {
	return toInt64(t.Quantity)
}

type Account struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Type       string `json:"type"`
	OpenedDate string `json:"openedDate"`
}

type Position struct {
	FIGI              string      `json:"figi"`
	Ticker            string      `json:"ticker"`
	Name              string      `json:"name"`
	Quantity          interface{} `json:"quantity"`
	QuantityAvailable interface{} `json:"quantityAvailable"`
	AveragePrice      MoneyValue  `json:"averagePositionPrice"`
	CurrentPrice      MoneyValue  `json:"currentPrice"`
	Currency          string      `json:"currency"`
}

func (p Position) Int64() int64 {
	return toInt64(p.Quantity)
}

func (p Position) AvailableInt64() int64 {
	return toInt64(p.QuantityAvailable)
}

type Portfolio struct {
	TotalAmount float64
	Currency    string
	Positions   []Position
}

func NewClient(token string, useSandbox bool) (*TinkoffClient, error) {
	baseURL := "https://invest-public-api.tbank.ru"
	if useSandbox {
		baseURL = "https://sandbox-invest-public-api.tbank.ru"
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &TinkoffClient{
		token:      token,
		useSandbox: useSandbox,
		baseURL:    baseURL,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: tr,
		},
	}, nil
}

func (c *TinkoffClient) makeRequest(ctx context.Context, method, endpoint string, body []byte) ([]byte, error) {
	url := c.baseURL + endpoint

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	if body != nil {
		req.Body = io.NopCloser(bytes.NewReader(body))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	if len(respBody) == 0 {
		return nil, fmt.Errorf("empty response from API")
	}

	return respBody, nil
}

func (c *TinkoffClient) GetAccounts() ([]Account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sandboxEndpoint := "/rest/tinkoff.public.invest.api.contract.v1.SandboxService/GetSandboxAccounts"
	regularEndpoint := "/rest/tinkoff.public.invest.api.contract.v1.UsersService/GetAccounts"

	endpoint := sandboxEndpoint
	if !c.useSandbox {
		endpoint = regularEndpoint
	}

	body, err := c.makeRequest(ctx, "POST", endpoint, []byte("{}"))
	if err != nil {
		if c.useSandbox {
			endpoint = regularEndpoint
			body, err = c.makeRequest(ctx, "POST", endpoint, []byte("{}"))
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	var resp AccountsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// If sandbox and no accounts, try to open a sandbox account
	if c.useSandbox && len(resp.Accounts) == 0 {
		openEndpoint := "/rest/tinkoff.public.invest.api.contract.v1.SandboxService/OpenSandboxAccount"
		_, err := c.makeRequest(ctx, "POST", openEndpoint, []byte("{}"))
		if err != nil {
			return nil, fmt.Errorf("failed to open sandbox account: %w", err)
		}

		// Try getting accounts again
		body, err = c.makeRequest(ctx, "POST", endpoint, []byte("{}"))
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return resp.Accounts, nil
}

func (c *TinkoffClient) GetOperations(accountID string, from, to time.Time) ([]Operation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	type GetOperationsRequest struct {
		AccountID string `json:"accountId"`
		From      string `json:"from"`
		To        string `json:"to"`
	}

	reqBody := GetOperationsRequest{
		AccountID: accountID,
		From:      from.Format(time.RFC3339),
		To:        to.Format(time.RFC3339),
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := "/rest/tinkoff.public.invest.api.contract.v1.OperationsService/GetOperations"

	body, err := c.makeRequest(ctx, "POST", endpoint, reqBytes)
	if err != nil {
		return nil, err
	}

	var resp OperationsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	operations := make([]Operation, 0)
	for _, op := range resp.Operations {
		if op.OperationType == "OPERATION_TYPE_BUY" || op.OperationType == "OPERATION_TYPE_SELL" {
			if isInvestKopilka(op.Ticker, op.FIGI) {
				continue
			}
			operations = append(operations, op)
		}
	}

	return operations, nil
}

func (c *TinkoffClient) GetPortfolio(accountID string) (*Portfolio, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	type PortfolioRequest struct {
		AccountID string `json:"accountId"`
	}

	reqBody := PortfolioRequest{AccountID: accountID}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := "/rest/tinkoff.public.invest.api.contract.v1.OperationsService/GetPortfolio"

	body, err := c.makeRequest(ctx, "POST", endpoint, reqBytes)
	if err != nil {
		return nil, err
	}

	var resp PortfolioResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &Portfolio{
		TotalAmount: resp.TotalAmountPortfolio.Float(),
		Currency:    resp.TotalAmountPortfolio.Currency,
		Positions:   resp.Positions,
	}, nil
}

func (c *TinkoffClient) GetLastOperationID(accountID string) (string, error) {
	now := time.Now()
	from := now.AddDate(0, 0, -7)

	ops, err := c.GetOperations(accountID, from, now)
	if err != nil {
		return "", err
	}

	if len(ops) == 0 {
		return "", nil
	}

	return ops[0].ID, nil
}

func (c *TinkoffClient) GetLastOperations(accountID string, limit int) ([]Operation, error) {
	now := time.Now()
	from := now.AddDate(0, 0, -30)

	ops, err := c.GetOperations(accountID, from, now)
	if err != nil {
		return nil, err
	}

	if len(ops) == 0 {
		return []Operation{}, nil
	}

	if len(ops) > limit {
		ops = ops[:limit]
	}

	return ops, nil
}

func (c *TinkoffClient) Close() error {
	return nil
}

func (c *TinkoffClient) GetInstruments() ([]Instrument, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	endpoint := "/rest/tinkoff.public.invest.api.contract.v1.InstrumentsService/Shares"
	if c.useSandbox {
		endpoint = "/rest/tinkoff.public.invest.api.contract.v1.SandboxService/GetSandboxInstruments"
	}

	body, err := c.makeRequest(ctx, "POST", endpoint, []byte("{}"))
	if err != nil {
		return nil, err
	}

	type InstrumentsResponse struct {
		Instruments []Instrument `json:"instruments"`
	}

	var resp InstrumentsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return resp.Instruments, nil
}

func (c *TinkoffClient) GetInstrumentInfo(figi string) (ticker, name, brand, logoName string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	type InstrumentRequest struct {
		IDType string `json:"id_type"`
		ID     string `json:"id"`
	}

	reqBody := InstrumentRequest{
		IDType: "INSTRUMENT_ID_TYPE_FIGI",
		ID:     figi,
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := "/rest/tinkoff.public.invest.api.contract.v1.InstrumentsService/GetInstrumentBy"

	body, err := c.makeRequest(ctx, "POST", endpoint, reqBytes)
	if err != nil {
		return "", "", "", "", err
	}

	type BrandInfo struct {
		Name     string `json:"name"`
		LogoName string `json:"logoName"`
	}

	type InstrumentResponse struct {
		Instrument struct {
			Ticker string    `json:"ticker"`
			Name   string    `json:"name"`
			FIGI   string    `json:"figi"`
			Brand  BrandInfo `json:"brand"`
		} `json:"instrument"`
	}

	var resp InstrumentResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", "", "", "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return resp.Instrument.Ticker, resp.Instrument.Name, resp.Instrument.Brand.Name, resp.Instrument.Brand.LogoName, nil
}

type Instrument struct {
	FIGI   string `json:"figi"`
	Ticker string `json:"ticker"`
	Name   string `json:"name"`
	UID    string `json:"uid"`
}

func (c *TinkoffClient) PlaceSandboxOrder(accountID, figi string, quantity int64, price float64) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build JSON manually to ensure correct structure
	jsonStr := fmt.Sprintf(`{
		"accountId": "%s",
		"figi": "%s",
		"direction": "ORDER_DIRECTION_BUY",
		"quantity": %d,
		"orderType": "ORDER_TYPE_MARKET"
	}`, accountID, figi, quantity)

	endpoint := "/rest/tinkoff.public.invest.api.contract.v1.SandboxService/PostSandboxOrder"

	body, err := c.makeRequest(ctx, "POST", endpoint, []byte(jsonStr))
	if err != nil {
		return "", err
	}

	log.Printf("Order response: %s", string(body))

	type OrderResponse struct {
		OrderID string `json:"orderId"`
	}

	var resp OrderResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return resp.OrderID, nil
}

func (c *TinkoffClient) DepositSandboxMoney(accountID string, amount float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	amountUnits := int64(amount)
	amountNano := int64((amount - float64(amountUnits)) * 1e9)

	jsonStr := fmt.Sprintf(`{
		"accountId": "%s",
		"amount": {"units": %d, "nano": %d, "currency": "RUB"}
	}`, accountID, amountUnits, amountNano)

	endpoint := "/rest/tinkoff.public.invest.api.contract.v1.SandboxService/SandboxPayIn"

	body, err := c.makeRequest(ctx, "POST", endpoint, []byte(jsonStr))
	if err != nil {
		return err
	}

	log.Printf("Deposit response: %s", string(body))
	return nil
}

func (c *TinkoffClient) GetAccountBalance(accountID string) (float64, string, error) {
	portfolio, err := c.GetPortfolio(accountID)
	if err != nil {
		return 0, "", err
	}
	return portfolio.TotalAmount, portfolio.Currency, nil
}
