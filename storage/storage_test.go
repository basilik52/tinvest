package storage

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB_Init(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "tinvest-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	db, err := New(tmpFile.Name())
	require.NoError(t, err)
	defer db.Close()

	err = db.Init()
	require.NoError(t, err)
}

func TestDB_SaveAndGetOperation(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "tinvest-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	db, err := New(tmpFile.Name())
	require.NoError(t, err)
	defer db.Close()

	err = db.Init()
	require.NoError(t, err)

	op := &OperationRecord{
		ID:            "test-op-1",
		OperationType: "Buy",
		Status:        "Done",
		Date:          time.Now(),
		Name:          "Тестовая акция",
		Ticker:        "TEST",
		FIGI:          "BBG000B12345",
		Quantity:      10,
		Price:         150.50,
		Total:         1505.00,
		Currency:      "RUB",
		Commission:    0.5,
		AccountID:     "account-1",
		CreatedAt:     time.Now(),
	}

	err = db.SaveOperation(op)
	require.NoError(t, err)

	retrieved, err := db.GetOperationByID("test-op-1")
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, op.ID, retrieved.ID)
	assert.Equal(t, op.OperationType, retrieved.OperationType)
	assert.Equal(t, op.Quantity, retrieved.Quantity)
}

func TestDB_GetOperationByID_NotFound(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "tinvest-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	db, err := New(tmpFile.Name())
	require.NoError(t, err)
	defer db.Close()

	err = db.Init()
	require.NoError(t, err)

	retrieved, err := db.GetOperationByID("non-existent")
	require.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestDB_GetLastOperation(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "tinvest-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	db, err := New(tmpFile.Name())
	require.NoError(t, err)
	defer db.Close()

	err = db.Init()
	require.NoError(t, err)

	op1 := &OperationRecord{
		ID:        "op-1",
		Date:      time.Now().Add(-time.Hour),
		AccountID: "account-1",
	}
	op2 := &OperationRecord{
		ID:        "op-2",
		Date:      time.Now(),
		AccountID: "account-1",
	}

	db.SaveOperation(op1)
	db.SaveOperation(op2)

	last, err := db.GetLastOperation("account-1")
	require.NoError(t, err)
	assert.Equal(t, "op-2", last.ID)
}

func TestDB_SaveAndGetState(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "tinvest-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	db, err := New(tmpFile.Name())
	require.NoError(t, err)
	defer db.Close()

	err = db.Init()
	require.NoError(t, err)

	err = db.SaveState("last_operation_id", "op-123")
	require.NoError(t, err)

	value, err := db.GetState("last_operation_id")
	require.NoError(t, err)
	assert.Equal(t, "op-123", value)
}

func TestDB_GetState_NotFound(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "tinvest-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	db, err := New(tmpFile.Name())
	require.NoError(t, err)
	defer db.Close()

	err = db.Init()
	require.NoError(t, err)

	value, err := db.GetState("non-existent")
	require.NoError(t, err)
	assert.Equal(t, "", value)
}

func TestDB_GetAllOperations(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "tinvest-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	db, err := New(tmpFile.Name())
	require.NoError(t, err)
	defer db.Close()

	err = db.Init()
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		db.SaveOperation(&OperationRecord{
			ID:        "op-" + string(rune('a'+i)),
			Date:      time.Now().Add(-time.Duration(i) * time.Hour),
			AccountID: "account-1",
		})
	}

	ops, err := db.GetAllOperations(3)
	require.NoError(t, err)
	assert.Len(t, ops, 3)
}

func TestOperationRecordStruct(t *testing.T) {
	now := time.Now()
	op := OperationRecord{
		ID:            "id-1",
		OperationType: "Buy",
		Status:        "Done",
		Date:          now,
		Name:          "Test",
		Ticker:        "TEST",
		FIGI:          "FIGI123",
		Quantity:      10,
		Price:         100.0,
		Total:         1000.0,
		Currency:      "RUB",
		Commission:    1.0,
		AccountID:     "acc-1",
		CreatedAt:     now,
	}

	assert.Equal(t, "id-1", op.ID)
	assert.Equal(t, "Buy", op.OperationType)
	assert.Equal(t, int64(10), op.Quantity)
}
