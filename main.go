package main

import (
	"flag"
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
	webOnly := flag.Bool("web-only", false, "Start only web server without trading")
	flag.Parse()

	log.Println("Starting trading bot...")
	godotenv.Load()

	go StartWebServer()

	if *webOnly {
		log.Println("Running in WEB-ONLY mode - trading disabled")
		select {}
	}

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
	timestampFrom := now.Add(-7 * 24 * time.Hour).UnixMilli()

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

	btcKlines, err := exchange.Klines("BTCUSDT", "1h", 0, 0, 52)
	if err != nil {
		log.Printf("[%s] Failed to get BTC price data: %v", pair.Symbol, err)
		return err
	}
	coinKlines, err := exchange.Klines(pair.Symbol, "1h", 0, 0, 52)
	if err != nil {
		log.Printf("[%s] Failed to get coin price data: %v", pair.Symbol, err)
		return err
	}
	log.Printf("[%s] Market data collected successfully", pair.Symbol)

	log.Printf("\n[%s] ===== ANALYSIS RESULTS =====", pair.Symbol)

	btcIchimoku := CalculateIchimoku(btcKlines)
	log.Printf("[%s] BTC Ichimoku: %s", pair.Symbol, btcIchimoku.Analysis.Signal)

	coinIchimoku := CalculateIchimoku(coinKlines)
	log.Printf("[%s] Coin Ichimoku: %s", pair.Symbol, coinIchimoku.Analysis.Signal)

	activityAnalysis := AnalyzeActivityTrend(activityData)
	log.Printf("[%s] Community activity trend: %v", pair.Symbol, activityAnalysis.Trend)

	fudActivityAnalysis := AnalyzeFudActivityTrend(fudActivityData)
	log.Printf("[%s] FUD activity trend: %v", pair.Symbol, fudActivityAnalysis.Trend)

	sentiment := ClaudeSentimentResponse{}
	if claudeClient != nil {
		tweets, err := activityClient.GetRecentTweets(pair.CommunityID, 50)
		if err != nil {
			log.Printf("[%s] Failed to fetch tweets: %v", pair.Symbol, err)
		} else if len(tweets) > 0 {
			hasNewTweets := state.LastAnalyzedTweetID == "" || tweets[0].ID != state.LastAnalyzedTweetID
			if hasNewTweets {
				log.Printf("[%s] Analyzing %d tweets with Claude...", pair.Symbol, len(tweets))
				sentimentAnalysis, err := AnalyzeCommunitysentiment(*claudeClient, tweets, "Analyze community sentiment")
				if err != nil {
					log.Printf("[%s] Claude analysis failed: %v", pair.Symbol, err)
					if state.LastSentimentAnalysis.Confidence != 0 {
						sentiment = state.LastSentimentAnalysis
					}
				} else {
					log.Printf("[%s] Sentiment: %d/10, Trend: %s", pair.Symbol, sentimentAnalysis.OverallSentiment, sentimentAnalysis.SentimentTrend)
					sentiment = sentimentAnalysis
					state.LastSentimentAnalysis = sentimentAnalysis
					state.LastAnalyzedTweetID = tweets[0].ID
				}
			} else {
				sentiment = state.LastSentimentAnalysis
			}
		}
	}

	signal := MakeTradingDecision(btcIchimoku.Analysis, coinIchimoku.Analysis, activityAnalysis, fudActivityAnalysis, sentiment)
	log.Printf("\n[%s] ===== DECISION: %s =====", pair.Symbol, signal)

	if signal == SignalEmpty {
		if state.CurrentPosition != PositionSideBoth {
			log.Printf("[%s] No signal - closing existing position", pair.Symbol)
			if state.CurrentPosition == PositionSideLong {
				if err := exchange.ClosePosition(pair.Symbol, PositionSideLong); err != nil {
					log.Printf("[%s] Failed to close LONG: %v", pair.Symbol, err)
					return err
				}
			} else if state.CurrentPosition == PositionSideShort {
				if err := exchange.ClosePosition(pair.Symbol, PositionSideShort); err != nil {
					log.Printf("[%s] Failed to close SHORT: %v", pair.Symbol, err)
					return err
				}
			}
			state.CurrentPosition = PositionSideBoth
			log.Printf("[%s] Position closed", pair.Symbol)
		} else {
			log.Printf("[%s] No signal - no action", pair.Symbol)
		}
		return nil
	}

	desiredPosition := PositionSideBoth
	if signal == SignalLong {
		desiredPosition = PositionSideLong
	} else if signal == SignalShort {
		desiredPosition = PositionSideShort
	}

	if state.CurrentPosition == desiredPosition {
		log.Printf("[%s] Position already matches signal - holding", pair.Symbol)
		return nil
	}

	if state.CurrentPosition != PositionSideBoth {
		log.Printf("[%s] Closing existing %s position", pair.Symbol, state.CurrentPosition)
		if err := exchange.ClosePosition(pair.Symbol, state.CurrentPosition); err != nil {
			log.Printf("[%s] Failed to close position: %v", pair.Symbol, err)
			return err
		}
		state.CurrentPosition = PositionSideBoth
	}

	log.Printf("[%s] Opening %s position", pair.Symbol, desiredPosition)
	position, err := exchange.OpenPosition(pair.Symbol, desiredPosition, pair.Leverage, pair.Quantity)
	if err != nil {
		log.Printf("[%s] Failed to open %s: %v", pair.Symbol, desiredPosition, err)
		return err
	}

	state.CurrentPosition = desiredPosition
	state.OpenedAt = time.Now()
	log.Printf("[%s] Position opened: %s (entry: %.6f, amount: %.6f)", pair.Symbol, desiredPosition, position.EntryPrice, position.Amount)

	return nil
}
