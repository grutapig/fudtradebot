package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/grutapig/fudtradebot/claude"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("Starting trading bot...")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbInfoDSN := os.Getenv(ENV_DATABASE_DSN_INFO)
	dbTradeDSN := os.Getenv(ENV_DATABASE_DSN_TRADE)
	telegramToken := os.Getenv(ENV_TELEGRAM_BOT_TOKEN)
	claudeAPIKey := os.Getenv(ENV_CLAUDE_API_KEY)
	exchangePrivateKey := os.Getenv(ENV_EXCHANGE_PRIVATE_KEY)
	exchangeRPCURL := os.Getenv(ENV_EXCHANGE_RPC_URL)
	notifyChatIDStr := os.Getenv("TELEGRAM_NOTIFY_CHAT_ID")

	notifyChatID, err := strconv.ParseInt(notifyChatIDStr, 10, 64)
	if err != nil {
		log.Fatal("Invalid TELEGRAM_NOTIFY_CHAT_ID")
	}

	dbInfoService, err := NewDatabaseInfoService(dbInfoDSN)
	if err != nil {
		log.Fatalf("Failed to initialize info database: %v", err)
	}
	log.Println("Info database initialized")

	dbTradeService, err := NewDatabaseTradeService(dbTradeDSN)
	if err != nil {
		log.Fatalf("Failed to initialize trade database: %v", err)
	}
	log.Println("Trade database initialized")

	exchangeService, err := NewExchangeService(exchangePrivateKey, exchangeRPCURL)
	if err != nil {
		log.Fatalf("Failed to initialize exchange service: %v", err)
	}
	log.Println("Exchange service initialized")

	telegramService, err := NewTelegramService(telegramToken, notifyChatID)
	if err != nil {
		log.Fatalf("Failed to initialize telegram service: %v", err)
	}
	log.Println("Telegram service initialized")

	claudeClient, err := claude.NewClaudeClient(claudeAPIKey, "", claude.CLAUDE_45_MODEL)
	if err != nil {
		log.Fatalf("Failed to initialize Claude client: %v", err)
	}
	log.Println("Claude client initialized")

	businessLogic := NewBusinessLogicService(
		dbInfoService,
		dbTradeService,
		exchangeService,
		telegramService,
		claudeClient,
	)
	log.Println("Business logic service initialized")

	monitoringService := NewMonitoringService(
		dbInfoService,
		dbTradeService,
		businessLogic,
	)
	log.Println("Monitoring service initialized")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	telegramService.Start()
	log.Println("Telegram service started")

	go businessLogic.ProcessTelegramCommands(ctx)
	log.Println("Telegram command processor started")

	go businessLogic.ProcessSignals(ctx)
	log.Println("Signal processor started")

	go monitoringService.Start(ctx)
	log.Println("Market monitoring started")

	go monitoringService.MonitorPositions(ctx)
	log.Println("Position monitoring started")

	telegramService.SendNotification("🤖 Trading bot started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutdown signal received")

	telegramService.SendNotification("🛑 Trading bot stopping")

	cancel()
	telegramService.Stop()

	log.Println("Trading bot stopped")
}
