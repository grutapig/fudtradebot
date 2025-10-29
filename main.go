package main

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"sync"
	"time"
)

var TradingPairs = []TradingPair{
	{
		CommunityID: "1914102634241577036",
		Symbol:      "BTCUSDT",
		Leverage:    10,
	},
}

func main() {
	log.Println("Starting trading bot...")
	godotenv.Load()
	apiKey := os.Getenv(ENV_DEX_KEY)
	secretKey := os.Getenv(ENV_DEX_SECRET)
	externalAPIURL := os.Getenv("EXTERNAL_API_URL")

	if apiKey == "" || secretKey == "" {
		log.Fatalf("%s and %s environment variables must be set", ENV_DEX_KEY, ENV_DEX_SECRET)
	}

	if externalAPIURL == "" {
		externalAPIURL = "http://localhost:3333/grufender"
	}

	exchange := NewAsterDexExchange(apiKey, secretKey)
	activityClient := NewExternalActivityClient(externalAPIURL)

	var wg sync.WaitGroup

	for _, pair := range TradingPairs {
		wg.Add(1)
		go func(pair TradingPair) {
			defer wg.Done()
			runTradingLoop(exchange, activityClient, pair)
		}(pair)
	}

	wg.Wait()
}

func runTradingLoop(exchange AsterDexExchange, activityClient ExternalActivityClient, pair TradingPair) {
	state := TradingState{
		CurrentPosition: PositionSideBoth,
	}

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Printf("[%s] Starting trading loop for community %s", pair.Symbol, pair.CommunityID)

	for range ticker.C {
		if err := processTradingCycle(exchange, activityClient, pair, &state); err != nil {
			log.Printf("[%s] Error in trading cycle: %v", pair.Symbol, err)
		}
	}
}

func processTradingCycle(exchange AsterDexExchange, activityClient ExternalActivityClient, pair TradingPair, state *TradingState) error {
	now := time.Now()
	timestampTo := now.UnixMilli()
	timestampFrom := now.Add(-30 * 24 * time.Hour).UnixMilli()

	activityData, err := activityClient.GetCommunityActivity(pair.CommunityID, timestampFrom, timestampTo, "4hour")
	if err != nil {
		return err
	}

	fudActivityData, err := activityClient.GetCommunityFudActivity(pair.CommunityID, timestampFrom, timestampTo, "4hour")
	if err != nil {
		return err
	}

	klines, err := exchange.Klines(pair.Symbol, "4h", 0, 0, 52)
	if err != nil {
		return err
	}

	ichimokuAnalysis := CalculateIchimoku(klines)
	activityAnalysis := AnalyzeActivityTrend(activityData)
	fudActivityAnalysis := AnalyzeFudActivityTrend(fudActivityData)

	signal := MakeTradingDecision(ichimokuAnalysis.Analysis, activityAnalysis, fudActivityAnalysis, state.CurrentPosition)

	log.Printf("[%s] Signal: %s (Strength: %d)", pair.Symbol, signal.Action, signal.Strength)
	for _, reason := range signal.Reasons {
		log.Printf("[%s]   - %s", pair.Symbol, reason)
	}

	if signal.Strength < SignalStrengthMedium {
		log.Printf("[%s] Signal too weak, skipping action", pair.Symbol)
		return nil
	}

	if state.CurrentPosition != PositionSideBoth {
		elapsed := time.Since(state.OpenedAt)
		if elapsed > 24*time.Hour {
			log.Printf("[%s] Position open for more than 24 hours, closing", pair.Symbol)
			if state.CurrentPosition == PositionSideLong {
				if err := exchange.ClosePosition(pair.Symbol, PositionSideLong); err != nil {
					return err
				}
			} else if state.CurrentPosition == PositionSideShort {
				if err := exchange.ClosePosition(pair.Symbol, PositionSideShort); err != nil {
					return err
				}
			}
			state.CurrentPosition = PositionSideBoth
			log.Printf("[%s] Position closed due to time limit", pair.Symbol)
			return nil
		}
	}

	switch signal.Action {
	case TradingActionOpenLong:
		balance, err := exchange.GetBalance()
		if err != nil {
			return err
		}

		price, err := exchange.GetMarkPrice(pair.Symbol)
		if err != nil {
			return err
		}

		quantity := (balance * float64(pair.Leverage)) / price

		position, err := exchange.OpenPosition(pair.Symbol, PositionSideLong, pair.Leverage, quantity)
		if err != nil {
			return err
		}

		state.CurrentPosition = PositionSideLong
		state.OpenedAt = time.Now()
		log.Printf("[%s] Opened LONG position: %+v", pair.Symbol, position)

	case TradingActionOpenShort:
		balance, err := exchange.GetBalance()
		if err != nil {
			return err
		}

		price, err := exchange.GetMarkPrice(pair.Symbol)
		if err != nil {
			return err
		}

		quantity := (balance * float64(pair.Leverage)) / price

		position, err := exchange.OpenPosition(pair.Symbol, PositionSideShort, pair.Leverage, quantity)
		if err != nil {
			return err
		}

		state.CurrentPosition = PositionSideShort
		state.OpenedAt = time.Now()
		log.Printf("[%s] Opened SHORT position: %+v", pair.Symbol, position)

	case TradingActionCloseLong:
		if err := exchange.ClosePosition(pair.Symbol, PositionSideLong); err != nil {
			return err
		}
		state.CurrentPosition = PositionSideBoth
		log.Printf("[%s] Closed LONG position", pair.Symbol)

	case TradingActionCloseShort:
		if err := exchange.ClosePosition(pair.Symbol, PositionSideShort); err != nil {
			return err
		}
		state.CurrentPosition = PositionSideBoth
		log.Printf("[%s] Closed SHORT position", pair.Symbol)

	case TradingActionHold:
		log.Printf("[%s] HOLD - no action", pair.Symbol)
	}

	return nil
}
