package watcher

import (
	"fmt"
	"log"
	"sync"
	"time"

	"tinvest/notifier"
	"tinvest/storage"
	"tinvest/tinkoff"
)

type Watcher struct {
	tinkoff   *tinkoff.TinkoffClient
	storage   *storage.DB
	notifier  *notifier.Notifier
	interval  time.Duration
	accountID string
	stopCh    chan struct{}
	wg        sync.WaitGroup
}

func New(tinkoffClient *tinkoff.TinkoffClient, db *storage.DB, notifierClient *notifier.Notifier, interval time.Duration, accountID string) *Watcher {
	return &Watcher{
		tinkoff:   tinkoffClient,
		storage:   db,
		notifier:  notifierClient,
		interval:  interval,
		accountID: accountID,
		stopCh:    make(chan struct{}),
	}
}

func (w *Watcher) Start() error {
	log.Printf("Starting watcher for account: %s", w.accountID)

	if err := w.syncInitialOperations(); err != nil {
		return fmt.Errorf("failed to sync initial operations: %w", err)
	}

	w.wg.Add(1)
	go w.run()

	return nil
}

func (w *Watcher) Stop() {
	log.Println("Stopping watcher...")
	close(w.stopCh)
	w.wg.Wait()
	log.Println("Watcher stopped")
}

func (w *Watcher) run() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			if err := w.checkNewOperations(); err != nil {
				log.Printf("Error checking operations: %v", err)
			}
		}
	}
}

func (w *Watcher) checkNewOperations() error {
	now := time.Now()
	from := now.Add(-30 * time.Second)

	operations, err := w.tinkoff.GetOperations(w.accountID, from, now)
	if err != nil {
		return fmt.Errorf("failed to get operations: %w", err)
	}

	newOps := make([]tinkoff.Operation, 0)
	for i := len(operations) - 1; i >= 0; i-- {
		op := operations[i]
		existing, err := w.storage.GetOperationByID(op.ID)
		if err != nil {
			return fmt.Errorf("failed to check operation existence: %w", err)
		}
		if existing == nil {
			newOps = append(newOps, op)
		}
	}

	if len(newOps) > 0 {
		for _, op := range newOps {
			log.Printf("New operation detected: %s (%s)", op.ID, op.OperationType)
			if err := w.notifier.NotifyTrade(&op); err != nil {
				log.Printf("Failed to notify about operation %s: %v", op.ID, err)
			} else {
				log.Printf("Notification sent for operation %s", op.ID)
			}

			record := &storage.OperationRecord{
				ID:            op.ID,
				OperationType: op.OperationType,
				Status:        op.Status,
				Date:          op.Date,
				Name:          op.Name,
				Ticker:        op.Ticker,
				FIGI:          op.FIGI,
				Quantity:      op.Int64(),
				Price:         op.Price.Float(),
				Total:         op.Float(),
				Currency:      op.Currency,
				Commission:    op.Commission.Float(),
				AccountID:     op.AccountID,
				CreatedAt:     time.Now(),
			}

			if err := w.storage.SaveOperation(record); err != nil {
				log.Printf("Failed to save operation %s: %v", op.ID, err)
			}
		}

		log.Printf("Processed %d new operations", len(newOps))
	}

	return nil
}

func (w *Watcher) syncInitialOperations() error {
	log.Println("Syncing initial operations...")

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	operations, err := w.tinkoff.GetOperations(w.accountID, todayStart, now)
	if err != nil {
		return fmt.Errorf("failed to get operations: %w", err)
	}

	if len(operations) == 0 {
		log.Println("No operations found")
		return nil
	}

	newNotifications := 0

	for i := len(operations) - 1; i >= 0; i-- {
		op := operations[i]

		existing, err := w.storage.GetOperationByID(op.ID)
		if err != nil {
			return fmt.Errorf("failed to check operation: %w", err)
		}

		if existing == nil {
			log.Printf("New operation from today: %s (%s)", op.ID, op.OperationType)
			if err := w.notifier.NotifyTrade(&op); err != nil {
				log.Printf("Failed to notify about operation %s: %v", op.ID, err)
			} else {
				newNotifications++
			}

			record := &storage.OperationRecord{
				ID:            op.ID,
				OperationType: op.OperationType,
				Status:        op.Status,
				Date:          op.Date,
				Name:          op.Name,
				Ticker:        op.Ticker,
				FIGI:          op.FIGI,
				Quantity:      op.Int64(),
				Price:         op.Price.Float(),
				Total:         op.Float(),
				Currency:      op.Currency,
				Commission:    op.Commission.Float(),
				AccountID:     op.AccountID,
				CreatedAt:     time.Now(),
			}

			if err := w.storage.SaveOperation(record); err != nil {
				log.Printf("Failed to save operation %s: %v", op.ID, err)
			}
		}
	}

	log.Printf("Synced %d operations, sent %d notifications for today's trades", len(operations), newNotifications)

	return nil
}

func (w *Watcher) GetAccountID() string {
	return w.accountID
}
