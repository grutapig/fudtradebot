package main

import (
	"github.com/grutapig/fudtradebot/claude"
	"github.com/joho/godotenv"
	"log"
	"os"
	"sync"
	"time"
)

var TradingPairs = []TradingPair{
	{
		CommunityID: "1969807538154811438",
		Symbol:      "GIGGLEUSDT",
		Leverage:    1,
		Quantity:    0.2,
	},
	{
		CommunityID: "1786006467847368871",
		Symbol:      "TOSHIUSDT",
		Leverage:    1,
		Quantity:    22000,
	},
	{
		CommunityID: "1938175945476555178",
		Symbol:      "TURTLEUSDT",
		Leverage:    1,
		Quantity:    150,
	},
}

func main() {
	log.Println("Starting trading bot...")
	godotenv.Load()
	apiKey := os.Getenv(ENV_DEX_KEY)
	secretKey := os.Getenv(ENV_DEX_SECRET)
	grufenderApiURL := os.Getenv(ENV_GRUFENDER_API_URL)
	proxyDSN := os.Getenv(ENV_PROXY_DSN)
	claudeAPIKey := os.Getenv(ENV_CLAUDE_API_KEY)
	claudeMinIntervalMinutes := getEnvAsInt(ENV_CLAUDE_MIN_INTERVAL_MINUTES, 10)

	if apiKey == "" || secretKey == "" {
		log.Fatalf("%s and %s environment variables must be set", ENV_DEX_KEY, ENV_DEX_SECRET)
	}

	if grufenderApiURL == "" {
		log.Fatalf("ENV_GRUFENDER_API_URL is not set.")
	}

	if claudeAPIKey == "" {
		log.Println("Warning: CLAUDE_API_KEY not set, sentiment analysis will be disabled")
	}

	var exchange AsterDexExchange
	var activityClient ExternalActivityClient
	var err error

	if proxyDSN != "" {
		log.Printf("Initializing clients with proxy")

		exchange, err = NewAsterDexExchangeWithProxy(apiKey, secretKey, proxyDSN)
		if err != nil {
			log.Fatalf("Failed to create exchange with proxy: %v", err)
		}

		activityClient, err = NewExternalActivityClientWithProxy(grufenderApiURL, proxyDSN)
		if err != nil {
			log.Fatalf("Failed to create activity client with proxy: %v", err)
		}
	} else {
		exchange = NewAsterDexExchange(apiKey, secretKey)
		activityClient = NewExternalActivityClient(grufenderApiURL)
	}

	var claudeClient *claude.ClaudeApi
	if claudeAPIKey != "" {
		claudeClient, err = claude.NewClaudeClient(claudeAPIKey, proxyDSN, claude.CLAUDE_45_MODEL)
		if err != nil {
			log.Fatalf("Failed to create Claude client: %v", err)
		}
		claudeClient.SetMaxTokens(4000)
	}

	var wg sync.WaitGroup

	for _, pair := range TradingPairs {
		wg.Add(1)
		go func(pair TradingPair) {
			defer wg.Done()
			runTradingLoop(exchange, activityClient, claudeClient, pair, claudeMinIntervalMinutes)
		}(pair)
	}

	wg.Wait()
}

func runTradingLoop(exchange AsterDexExchange, activityClient ExternalActivityClient, claudeClient *claude.ClaudeApi, pair TradingPair, claudeMinIntervalMinutes int) {
	state := TradingState{
		CurrentPosition: PositionSideBoth,
	}

	log.Printf("[%s] Starting trading loop for community %s", pair.Symbol, pair.CommunityID)

	log.Printf("[%s] Restoring position state from exchange...", pair.Symbol)
	position, err := exchange.GetPosition(pair.Symbol)
	if err != nil {
		log.Printf("[%s] Warning: Failed to restore position state: %v", pair.Symbol, err)
		log.Printf("[%s] Starting with default state (no position)", pair.Symbol)
	} else if position != nil {
		state.CurrentPosition = position.Side
		state.OpenedAt = position.Timestamp
		log.Printf("[%s] ✓ Restored existing position: %s (opened at %s)",
			pair.Symbol, position.Side, position.Timestamp.Format("2006-01-02 15:04:05"))
		log.Printf("[%s]   Entry price: %.6f, Amount: %.6f, P/L: %.2f USDT",
			pair.Symbol, position.EntryPrice, position.Amount, position.UnrealizedPL)
	} else {
		log.Printf("[%s] ✓ No existing position found, starting fresh", pair.Symbol)
	}

	for {
		if err := processTradingCycle(exchange, activityClient, claudeClient, pair, &state, claudeMinIntervalMinutes); err != nil {
			log.Printf("[%s] Error in trading cycle: %v", pair.Symbol, err)
		}
		time.Sleep(time.Second * 60)
	}
}

func processTradingCycle(exchange AsterDexExchange, activityClient ExternalActivityClient, claudeClient *claude.ClaudeApi, pair TradingPair, state *TradingState, claudeMinIntervalMinutes int) error {
	log.Printf("\n========== [%s] Starting analysis cycle ==========", pair.Symbol)
	if state.CurrentPosition != PositionSideBoth {
		log.Printf("[%s] Current position: %v (opened %v ago)", pair.Symbol, state.CurrentPosition, time.Since(state.OpenedAt).Round(time.Minute))
	} else {
		log.Printf("[%s] Current position: no active position", pair.Symbol)
	}

	now := time.Now()
	timestampTo := now.UnixMilli()
	timestampFrom := now.Add(-30 * 24 * time.Hour).UnixMilli()

	log.Printf("[%s] Collecting market data...", pair.Symbol)
	activityData, err := activityClient.GetCommunityActivity(pair.CommunityID, timestampFrom, timestampTo, "hour")
	if err != nil {
		log.Printf("[%s] Failed to get community activity: %v", pair.Symbol, err)
		return err
	}

	fudActivityData, err := activityClient.GetCommunityFudActivity(pair.CommunityID, timestampFrom, timestampTo, "hour")
	if err != nil {
		log.Printf("[%s] Failed to get FUD activity: %v", pair.Symbol, err)
		return err
	}

	klines, err := exchange.Klines(pair.Symbol, "1h", 0, 0, 52)
	if err != nil {
		log.Printf("[%s] Failed to get price data: %v", pair.Symbol, err)
		return err
	}
	log.Printf("[%s] Market data collected successfully", pair.Symbol)

	log.Printf("\n[%s] ===== ANALYSIS RESULTS =====", pair.Symbol)

	activityAnalysis := AnalyzeActivityTrend(activityData)
	log.Printf("[%s] Community activity trend: %v", pair.Symbol, activityAnalysis)

	fudActivityAnalysis := AnalyzeFudActivityTrend(fudActivityData)
	log.Printf("[%s] FUD activity trend: %v", pair.Symbol, fudActivityAnalysis)

	var signal TradingSignal
	var sentimentSignal TradingSignal

	ichimokuAnalysis := CalculateIchimoku(klines)
	log.Printf("[%s] Ichimoku analysis: %s", pair.Symbol, ichimokuAnalysis.Analysis.Description)
	signal = MakeTradingDecision(ichimokuAnalysis.Analysis, activityAnalysis, fudActivityAnalysis, state.CurrentPosition)

	log.Printf("\n[%s] ===== TECHNICAL SIGNAL =====", pair.Symbol)
	for _, reason := range signal.Reasons {
		log.Printf("[%s] • %s", pair.Symbol, reason)
	}
	log.Printf("[%s] Technical action: %s (strength: %d)", pair.Symbol, signal.Action, signal.Strength)

	if claudeClient != nil {
		tweets, err := activityClient.GetRecentTweets(pair.CommunityID, 50)
		if err != nil {
			log.Printf("[%s] Failed to fetch tweets: %v", pair.Symbol, err)
		} else if len(tweets) > 0 {
			hasNewTweets := state.LastAnalyzedTweetID == "" || (len(tweets) > 0 && tweets[0].ID != state.LastAnalyzedTweetID)

			if hasNewTweets {
				log.Printf("[%s] Analyzing %d new tweets with Claude...", pair.Symbol, len(tweets))
				sentimentAnalysis, err := AnalyzeCommunitysentiment(*claudeClient, tweets, "Analyze community sentiment based on recent tweets")
				if err != nil {
					log.Printf("[%s] Claude analysis failed: %v", pair.Symbol, err)
					if state.LastSentimentAnalysis.Confidence != 0 {
						log.Printf("[%s] Using cached sentiment analysis", pair.Symbol)
						sentimentSignal = MakeSentimentTradingDecision(state.LastSentimentAnalysis, state.CurrentPosition)
					}
				} else {
					log.Printf("[%s] Claude sentiment:", pair.Symbol)
					log.Printf("[%s]   Overall: %d/10", pair.Symbol, sentimentAnalysis.OverallSentiment)
					log.Printf("[%s]   Trend: %s", pair.Symbol, sentimentAnalysis.SentimentTrend)
					log.Printf("[%s]   FUD level: %d/10", pair.Symbol, sentimentAnalysis.FudLevel)
					log.Printf("[%s]   Confidence: %.2f", pair.Symbol, sentimentAnalysis.Confidence)
					log.Printf("[%s]   Recommendation: %s", pair.Symbol, sentimentAnalysis.Recommendation)

					state.LastSentimentAnalysis = sentimentAnalysis
					state.LastAnalyzedTweetID = tweets[0].ID
					sentimentSignal = MakeSentimentTradingDecision(sentimentAnalysis, state.CurrentPosition)
				}
			} else {
				log.Printf("[%s] No new tweets, using cached sentiment with current position", pair.Symbol)
				if state.LastSentimentAnalysis.Confidence != 0 {
					sentimentSignal = MakeSentimentTradingDecision(state.LastSentimentAnalysis, state.CurrentPosition)
				}
			}
		}
	}

	if sentimentSignal.Action != "" {
		log.Printf("\n[%s] ===== SENTIMENT SIGNAL =====", pair.Symbol)
		for _, reason := range sentimentSignal.Reasons {
			log.Printf("[%s] • %s", pair.Symbol, reason)
		}
		log.Printf("[%s] Sentiment action: %s (strength: %d)", pair.Symbol, sentimentSignal.Action, sentimentSignal.Strength)

		log.Printf("\n[%s] ===== SIGNAL COMPARISON =====", pair.Symbol)
		if signal.Action == sentimentSignal.Action {
			log.Printf("[%s] ✓ Signals MATCH: %s", pair.Symbol, signal.Action)
		} else {
			log.Printf("[%s] ⚠️  Signals MISMATCH: technical=%s, sentiment=%s", pair.Symbol, signal.Action, sentimentSignal.Action)
			if state.CurrentPosition == PositionSideLong {
				log.Printf("[%s] Closing LONG due to signal mismatch", pair.Symbol)
				if err := exchange.ClosePosition(pair.Symbol, PositionSideLong); err != nil {
					return err
				}
				state.CurrentPosition = PositionSideBoth
				return nil
			} else if state.CurrentPosition == PositionSideShort {
				log.Printf("[%s] Closing SHORT due to signal mismatch", pair.Symbol)
				if err := exchange.ClosePosition(pair.Symbol, PositionSideShort); err != nil {
					return err
				}
				state.CurrentPosition = PositionSideBoth
				return nil
			} else {
				log.Printf("[%s] No position open, skipping action due to mismatch", pair.Symbol)
				return nil
			}
		}
	} else {
		log.Printf("[%s] No sentiment signal available (Claude not configured or no tweets)", pair.Symbol)
	}

	if signal.Strength < SignalStrengthMedium {
		log.Printf("[%s] ❌ Signal too weak - closing position if open", pair.Symbol)
		if state.CurrentPosition == PositionSideLong {
			if err := exchange.ClosePosition(pair.Symbol, PositionSideLong); err != nil {
				return err
			}
			return nil
		} else if state.CurrentPosition == PositionSideShort {
			if err := exchange.ClosePosition(pair.Symbol, PositionSideShort); err != nil {
				return err
			}
			return nil
		}
		return nil
	}
	log.Printf("[%s] ✓ Signal strong enough to proceed", pair.Symbol)

	if state.CurrentPosition != PositionSideBoth {
		elapsed := time.Since(state.OpenedAt)
		if elapsed > 24*time.Hour {
			log.Printf("\n[%s] ⚠️  Position has been open for more than 24 hours (%v)", pair.Symbol, elapsed.Round(time.Minute))
			log.Printf("[%s] Decision: Force close %v position due to time limit", pair.Symbol, state.CurrentPosition)
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
			log.Printf("[%s] ✓ Position closed (time limit exceeded)", pair.Symbol)
			return nil
		}
	}

	log.Printf("\n[%s] ===== FINAL DECISION: %s =====", pair.Symbol, signal.Action)

	switch signal.Action {
	case TradingActionOpenLong:
		price, err := exchange.GetMarkPrice(pair.Symbol)
		if err != nil {
			log.Printf("[%s] ❌ Failed to get price: %v", pair.Symbol, err)
			return err
		}

		log.Printf("[%s] Opening LONG position:", pair.Symbol)
		log.Printf("[%s]   Entry price: %.6f", pair.Symbol, price)
		log.Printf("[%s]   Quantity: %.6f (leverage %dx)", pair.Symbol, pair.Quantity, pair.Leverage)
		log.Printf("[%s]   Position value: %.2f USDT", pair.Symbol, pair.Quantity*price)

		position, err := exchange.OpenPosition(pair.Symbol, PositionSideLong, pair.Leverage, pair.Quantity)
		if err != nil {
			log.Printf("[%s] ❌ Failed to open LONG: %v", pair.Symbol, err)
			return err
		}

		state.CurrentPosition = PositionSideLong
		state.OpenedAt = time.Now()
		log.Printf("[%s] ✓ LONG position opened (entry: %.6f, amount: %.6f)", pair.Symbol, position.EntryPrice, position.Amount)
		log.Printf("[%s] State updated to LONG position\n", pair.Symbol)

	case TradingActionOpenShort:
		price, err := exchange.GetMarkPrice(pair.Symbol)
		if err != nil {
			log.Printf("[%s] ❌ Failed to get price: %v", pair.Symbol, err)
			return err
		}

		log.Printf("[%s] Opening SHORT position:", pair.Symbol)
		log.Printf("[%s]   Entry price: %.6f", pair.Symbol, price)
		log.Printf("[%s]   Quantity: %.6f (leverage %dx)", pair.Symbol, pair.Quantity, pair.Leverage)
		log.Printf("[%s]   Position value: %.2f USDT", pair.Symbol, pair.Quantity*price)

		position, err := exchange.OpenPosition(pair.Symbol, PositionSideShort, pair.Leverage, pair.Quantity)
		if err != nil {
			log.Printf("[%s] ❌ Failed to open SHORT: %v", pair.Symbol, err)
			return err
		}

		state.CurrentPosition = PositionSideShort
		state.OpenedAt = time.Now()
		log.Printf("[%s] ✓ SHORT position opened (entry: %.6f, amount: %.6f)", pair.Symbol, position.EntryPrice, position.Amount)
		log.Printf("[%s] State updated to SHORT position\n", pair.Symbol)

	case TradingActionCloseLong:
		log.Printf("[%s] Closing LONG position:", pair.Symbol)
		log.Printf("[%s]   Position held for: %v", pair.Symbol, time.Since(state.OpenedAt).Round(time.Minute))
		log.Printf("[%s]   Reason: Market conditions changed (signal strength: %d)", pair.Symbol, signal.Strength)
		if err := exchange.ClosePosition(pair.Symbol, PositionSideLong); err != nil {
			return err
		}
		state.CurrentPosition = PositionSideBoth
		log.Printf("[%s] ✓ LONG position closed", pair.Symbol)
		log.Printf("[%s] Now in neutral position\n", pair.Symbol)

	case TradingActionCloseShort:
		log.Printf("[%s] Closing SHORT position:", pair.Symbol)
		log.Printf("[%s]   Position held for: %v", pair.Symbol, time.Since(state.OpenedAt).Round(time.Minute))
		log.Printf("[%s]   Reason: Market conditions changed (signal strength: %d)", pair.Symbol, signal.Strength)
		if err := exchange.ClosePosition(pair.Symbol, PositionSideShort); err != nil {
			return err
		}
		state.CurrentPosition = PositionSideBoth
		log.Printf("[%s] ✓ SHORT position closed", pair.Symbol)
		log.Printf("[%s] Now in neutral position\n", pair.Symbol)

	case TradingActionHold:
		log.Printf("[%s] Decision: HOLD position", pair.Symbol)
		if state.CurrentPosition != PositionSideBoth {
			log.Printf("[%s] Maintaining %v position (held for %v)", pair.Symbol, state.CurrentPosition, time.Since(state.OpenedAt).Round(time.Minute))
		} else {
			log.Printf("[%s] No position - waiting for better signal", pair.Symbol)
		}
		log.Printf("[%s] No action required\n", pair.Symbol)
	}

	return nil
}
