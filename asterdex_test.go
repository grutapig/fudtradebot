package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func TestAsterDexPositions(t *testing.T) {
	if err := godotenv.Load(); err != nil {
		t.Skip("No .env file found, skipping test")
	}

	apiKey := os.Getenv(ENV_DEX_KEY)
	secretKey := os.Getenv(ENV_DEX_SECRET)

	if apiKey == "" || secretKey == "" {
		t.Skip("ASTERDEX_API_KEY or ASTERDEX_SECRET_KEY not set, skipping test")
	}

	// Create exchange instance
	exchange := NewAsterDexExchange(apiKey, secretKey)

	symbol := "SOLUSDT"
	leverage := 10
	quantity := 0.06 // Small amount for testing

	log.Println("=== AsterDEX Position Test ===")

	// 1. Get current balance
	log.Println("\n1. Checking balance...")
	balance, err := exchange.GetBalance()
	if err != nil {
		t.Fatalf("Failed to get balance: %v", err)
	}
	log.Printf("Available balance: %.2f USDT", balance)
	// 2. Get mark price
	log.Println("\n2. Fetching mark price...")
	markPrice, err := exchange.GetMarkPrice(symbol)
	if err != nil {
		t.Fatalf("Failed to get mark price: %v", err)
	}
	log.Printf("Mark price for %s: %.2f USDT", symbol, markPrice)

	// 3. Check existing positions
	log.Println("\n3. Checking existing positions...")
	positions, err := exchange.GetAllPositions()
	if err != nil {
		t.Fatalf("Failed to get positions: %v", err)
	}
	if len(positions) > 0 {
		log.Printf("Found %d open position(s):", len(positions))
		for _, pos := range positions {
			log.Printf("  - %s %s: Amount=%.6f, Entry=%.2f, PL=%.2f, Leverage=%dx",
				pos.Symbol, pos.Side, pos.Amount, pos.EntryPrice, pos.UnrealizedPL, pos.Leverage)
		}
	} else {
		log.Println("No open positions found")
	}

	// 4. Open LONG position
	log.Println("\n4. Opening LONG position...")
	side := PositionSideLong
	position, err := exchange.OpenPosition(symbol, side, leverage, quantity)
	if err != nil {
		t.Fatalf("Failed to open LONG position: %v", err)
	}
	log.Printf("LONG position opened:")
	log.Printf("  Symbol: %s", position.Symbol)
	log.Printf("  Side: %s", position.Side)
	log.Printf("  Amount: %.6f", position.Amount)
	log.Printf("  Entry Price: %.2f", position.EntryPrice)
	log.Printf("  Leverage: %dx", position.Leverage)

	// Wait a bit
	log.Println("\nWaiting 3 seconds...")
	time.Sleep(3 * time.Second)

	// 5. Check position status
	log.Println("\n5. Checking position status...")
	currentPosition, err := exchange.GetPosition(symbol)
	if err != nil {
		t.Fatalf("Failed to get position: %v", err)
	}
	if currentPosition != nil {
		log.Printf("Current position:")
		log.Printf("  Symbol: %s", currentPosition.Symbol)
		log.Printf("  Side: %s", currentPosition.Side)
		log.Printf("  Amount: %.6f", currentPosition.Amount)
		log.Printf("  Entry Price: %.2f", currentPosition.EntryPrice)
		log.Printf("  Unrealized P/L: %.2f USDT", currentPosition.UnrealizedPL)
		log.Printf("  Leverage: %dx", currentPosition.Leverage)
	} else {
		log.Println("No position found")
	}

	// Wait a bit
	log.Println("\nWaiting 2 seconds...")
	time.Sleep(2 * time.Second)

	// 7. Verify position is closed
	log.Println("\n7. Verifying position is closed...")
	finalPosition, err := exchange.GetPosition(symbol)
	if err != nil {
		t.Fatalf("Failed to verify position: %v", err)
	}
	if finalPosition == nil {
		log.Println("✓ Position successfully closed")
	} else {
		log.Printf("⚠ Position still exists: Amount=%.6f", finalPosition.Amount)
	}

	log.Println("\n=== Test completed ===")
}
func TestAsterDexExchange_ClosePosition(t *testing.T) {
	if err := godotenv.Load(); err != nil {
		t.Skip("No .env file found, skipping test")
	}

	apiKey := os.Getenv(ENV_DEX_KEY)
	secretKey := os.Getenv(ENV_DEX_SECRET)

	if apiKey == "" || secretKey == "" {
		t.Skip("ASTERDEX_API_KEY or ASTERDEX_SECRET_KEY not set, skipping test")
	}

	// Create exchange instance
	exchange := NewAsterDexExchange(apiKey, secretKey)
	time.Sleep(time.Second * 10)
	exchange.ClosePosition("SOLUSDT", PositionSideLong)
}

func TestAsterDexExchange_GetKlines(t *testing.T) {
	godotenv.Load()
	exchange := NewAsterDexExchange(os.Getenv(ENV_DEX_KEY), os.Getenv(ENV_DEX_SECRET))
	result, err := exchange.Klines("BTCUSDT", "4h", 0, 0, 200)
	assert.NoError(t, err)
	svg := GenerateCandlestickSVG(result, 800, 600)
	os.WriteFile("chart.svg", []byte(svg), 0655)
	cloud := CalculateIchimoku(result)
	aga := GenerateIchimokuSVG(result, cloud.Data, 800, 600)
	os.WriteFile("chart.svg", []byte(aga), 0655)
	fmt.Printf("analyze %+v", cloud.Analysis)
}
func TestAsterDexExchange_GetAllBalances(t *testing.T) {
	godotenv.Load()
	exchange := NewAsterDexExchange(os.Getenv(ENV_DEX_KEY), os.Getenv(ENV_DEX_SECRET))
	balances, err := exchange.GetAllBalances()
	assert.NoError(t, err)
	fmt.Printf("%+v\n", balances)
}
