package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"tinvest/bot"
	"tinvest/config"
	"tinvest/notifier"
	"tinvest/storage"
	"tinvest/tinkoff"
	"tinvest/watcher"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if cfg.TelegramBotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}
	if cfg.TinkoffToken == "" {
		log.Fatal("TINKOFF_TOKEN is not set")
	}
	if cfg.TelegramGroupID == 0 {
		log.Fatal("TELEGRAM_GROUP_ID is not set")
	}

	botClient, err := bot.NewClient(cfg.TelegramBotToken, cfg.TelegramGroupID)
	if err != nil {
		log.Fatalf("Failed to create bot client: %v", err)
	}

	db, err := storage.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if err := db.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	tinkoffClient, err := tinkoff.NewClient(cfg.TinkoffToken, cfg.UseSandbox)
	if err != nil {
		log.Fatalf("Failed to create Tinkoff client: %v", err)
	}
	defer tinkoffClient.Close()

	accountID := cfg.TinkoffAccountID
	if accountID == "" {
		accounts, err := tinkoffClient.GetAccounts()
		if err != nil {
			log.Fatalf("Failed to get accounts: %v", err)
		}
		if len(accounts) == 0 {
			log.Fatal("No accounts found")
		}
		accountID = accounts[0].ID
		log.Printf("Using account: %s", accountID)
	}

	notifierClient := notifier.New(botClient, tinkoffClient)

	// Start watcher first (so it can detect the test trade as new operation)
	w := watcher.New(tinkoffClient, db, notifierClient, cfg.PollInterval, accountID)

	if err := w.Start(); err != nil {
		log.Fatalf("Failed to start watcher: %v", err)
	}

	log.Println("Bot started successfully")

	// For sandbox: create test trade after watcher is running (so it will be detected as new operation)
	if cfg.UseSandbox {
		log.Println("Attempting to create test trade in sandbox...")

		// Deposit money to sandbox account first
		log.Println("Depositing money to sandbox...")
		err = tinkoffClient.DepositSandboxMoney(accountID, 10000)
		if err != nil {
			log.Printf("Note: Could not deposit money: %v", err)
		} else {
			log.Println("Money deposited successfully!")
		}

		// Try placing order directly with known FIGI
		// Using SBER as test instrument
		testFIGI := "BBG004730N88" // SBER

		orderID, err := tinkoffClient.PlaceSandboxOrder(accountID, testFIGI, 1, 100.0)
		if err != nil {
			log.Printf("Note: Could not place test order: %v", err)
		} else {
			log.Printf("Test order placed successfully! Order ID: %s", orderID)
		}
	} else {
		log.Println("Production mode - monitoring for new trades")
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh

	log.Println("Shutting down...")
	w.Stop()

	fmt.Println("Bot stopped")
}
