package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

type DB struct {
	db *sql.DB
}

type OperationInfo struct {
	ID            string
	OperationType string
	Date          string
	Ticker        string
	FIGI          string
	Quantity      int64
	Price         float64
	Total         float64
}

type OperationRecord struct {
	ID            string
	OperationType string
	Status        string
	Date          time.Time
	Name          string
	Ticker        string
	FIGI          string
	Quantity      int64
	Price         float64
	Total         float64
	Currency      string
	Commission    float64
	AccountID     string
	CreatedAt     time.Time
}

func New(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db: db}, nil
}

func (d *DB) Init() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS operations (
			id TEXT PRIMARY KEY,
			operation_type TEXT NOT NULL,
			status TEXT NOT NULL,
			date TEXT NOT NULL,
			name TEXT,
			ticker TEXT,
			figi TEXT,
			quantity INTEGER,
			price REAL,
			total REAL,
			currency TEXT,
			commission REAL,
			account_id TEXT,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS state (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_operations_date ON operations(date DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_operations_account ON operations(account_id)`,
	}

	for _, query := range queries {
		if _, err := d.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

func (d *DB) SaveOperation(op *OperationRecord) error {
	query := `
		INSERT OR REPLACE INTO operations 
		(id, operation_type, status, date, name, ticker, figi, quantity, price, total, currency, commission, account_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := d.db.Exec(query,
		op.ID,
		op.OperationType,
		op.Status,
		op.Date.Format(time.RFC3339),
		op.Name,
		op.Ticker,
		op.FIGI,
		op.Quantity,
		op.Price,
		op.Total,
		op.Currency,
		op.Commission,
		op.AccountID,
		op.CreatedAt.Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("failed to save operation: %w", err)
	}

	return nil
}

func (d *DB) GetOperationByID(id string) (*OperationRecord, error) {
	query := `SELECT * FROM operations WHERE id = ?`

	row := d.db.QueryRow(query, id)

	var op OperationRecord
	var dateStr, createdAtStr string

	err := row.Scan(
		&op.ID,
		&op.OperationType,
		&op.Status,
		&dateStr,
		&op.Name,
		&op.Ticker,
		&op.FIGI,
		&op.Quantity,
		&op.Price,
		&op.Total,
		&op.Currency,
		&op.Commission,
		&op.AccountID,
		&createdAtStr,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get operation: %w", err)
	}

	op.Date, _ = time.Parse(time.RFC3339, dateStr)
	op.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)

	return &op, nil
}

func (d *DB) GetLastOperation(accountID string) (*OperationRecord, error) {
	query := `
		SELECT * FROM operations 
		WHERE account_id = ? 
		ORDER BY date DESC 
		LIMIT 1
	`

	row := d.db.QueryRow(query, accountID)

	var op OperationRecord
	var dateStr, createdAtStr string

	err := row.Scan(
		&op.ID,
		&op.OperationType,
		&op.Status,
		&dateStr,
		&op.Name,
		&op.Ticker,
		&op.FIGI,
		&op.Quantity,
		&op.Price,
		&op.Total,
		&op.Currency,
		&op.Commission,
		&op.AccountID,
		&createdAtStr,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get last operation: %w", err)
	}

	op.Date, _ = time.Parse(time.RFC3339, dateStr)
	op.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)

	return &op, nil
}

func (d *DB) GetAllOperations(limit int) ([]OperationRecord, error) {
	query := `
		SELECT * FROM operations 
		ORDER BY date DESC 
		LIMIT ?
	`

	rows, err := d.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get operations: %w", err)
	}
	defer rows.Close()

	operations := make([]OperationRecord, 0)
	for rows.Next() {
		var op OperationRecord
		var dateStr, createdAtStr string

		err := rows.Scan(
			&op.ID,
			&op.OperationType,
			&op.Status,
			&dateStr,
			&op.Name,
			&op.Ticker,
			&op.FIGI,
			&op.Quantity,
			&op.Price,
			&op.Total,
			&op.Currency,
			&op.Commission,
			&op.AccountID,
			&createdAtStr,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan operation: %w", err)
		}

		op.Date, _ = time.Parse(time.RFC3339, dateStr)
		op.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)

		operations = append(operations, op)
	}

	return operations, nil
}

func (d *DB) SaveState(key, value string) error {
	query := `
		INSERT OR REPLACE INTO state (key, value, updated_at)
		VALUES (?, ?, ?)
	`

	_, err := d.db.Exec(query, key, value, time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

func (d *DB) GetState(key string) (string, error) {
	query := `SELECT value FROM state WHERE key = ?`

	row := d.db.QueryRow(query, key)

	var value string
	err := row.Scan(&value)

	if err == sql.ErrNoRows {
		return "", nil
	}

	if err != nil {
		return "", fmt.Errorf("failed to get state: %w", err)
	}

	return value, nil
}

func (d *DB) Close() error {
	return d.db.Close()
}
