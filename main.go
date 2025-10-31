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

	if err := InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database initialized successfully")

	go StartWebServer()

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

	go runBalanceCollector(exchange)

	if *webOnly {
		log.Println("Running in WEB-ONLY mode - trading disabled")
		select {}
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

func runBalanceCollector(exchange AsterDexExchange) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		balance, err := exchange.GetBalance()
		if err != nil {
			log.Printf("Failed to get balance: %v", err)
		} else {
			if err := SaveBalance("USDT", balance); err != nil {
				log.Printf("Failed to save balance: %v", err)
			} else {
				log.Printf("Balance saved: %.2f USDT", balance)
			}
		}
		<-ticker.C
	}
}

func runTradingLoop(exchange AsterDexExchange, activityClient ExternalActivityClient, claudeClient *claude.ClaudeApi, pair TradingPair, claudeMinIntervalMinutes int) {
	state := TradingState{
		CurrentPosition: PositionSideBoth,
	}

	UpdateTradingState(pair.Symbol, &state)

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

	var lastFudAttack ClaudeFudAttackResponse

	currentPosition, err := exchange.GetPosition(pair.Symbol)
	if err != nil {
		log.Printf("[%s] Failed to get current position: %v", pair.Symbol, err)
	} else if currentPosition != nil {
		markPrice, err := exchange.GetMarkPrice(pair.Symbol)
		if err != nil {
			log.Printf("[%s] Failed to get mark price: %v", pair.Symbol, err)
			markPrice = 0
		}

		if err := SavePositionSnapshot(*currentPosition, markPrice); err != nil {
			log.Printf("[%s] Failed to save position snapshot: %v", pair.Symbol, err)
		} else {
			log.Printf("[%s] Position snapshot saved: P/L %.2f USDT, Mark Price %.6f",
				pair.Symbol, currentPosition.UnrealizedPL, markPrice)
		}
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

	if claudeClient != nil {
		now := time.Now()
		timeSinceLastCheck := now.Sub(state.LastFudCheckTime)
		if timeSinceLastCheck >= 10*time.Minute || state.LastFudCheckTime.IsZero() {
			tweets, err := activityClient.GetRecentTweets(pair.CommunityID, 200)
			if err != nil {
				log.Printf("[%s] Failed to fetch tweets for FUD check: %v", pair.Symbol, err)
			} else if len(tweets) >= 3 {
				newMessagesCount := 0
				if state.LastFudCheckTweetID != "" && len(tweets) > 0 {
					for i, tweet := range tweets {
						if tweet.ID == state.LastFudCheckTweetID {
							newMessagesCount = i
							break
						}
					}
					if newMessagesCount == 0 && len(tweets) > 0 && tweets[0].ID != state.LastFudCheckTweetID {
						newMessagesCount = len(tweets)
					}
				} else {
					newMessagesCount = len(tweets)
				}

				if newMessagesCount >= 3 {
					log.Printf("[%s] Checking for coordinated FUD attack (%d new messages)...", pair.Symbol, newMessagesCount)
					fudAttack, err := AnalyzeFudAttack(*claudeClient, tweets)
					if err != nil {
						log.Printf("[%s] FUD attack analysis failed: %v", pair.Symbol, err)
					} else {
						lastFudAttack = fudAttack
						log.Printf("\n[%s] ===== FUD ATTACK ANALYSIS =====", pair.Symbol)
						if fudAttack.HasAttack {
							log.Printf("[%s] ⚠️  COORDINATED FUD ATTACK DETECTED!", pair.Symbol)
							log.Printf("[%s]   Confidence: %.0f%%", pair.Symbol, fudAttack.Confidence*100)
							log.Printf("[%s]   Messages: %d", pair.Symbol, fudAttack.MessageCount)
							log.Printf("[%s]   FUD Type: %s", pair.Symbol, fudAttack.FudType)
							log.Printf("[%s]   Theme: %s", pair.Symbol, fudAttack.Theme)
							log.Printf("[%s]   Started: %d hours ago", pair.Symbol, fudAttack.StartedHoursAgo)
							log.Printf("[%s]   Participants:", pair.Symbol)
							for _, p := range fudAttack.Participants {
								log.Printf("[%s]     - %s (%d messages)", pair.Symbol, p.Username, p.MessageCount)
							}
							log.Printf("[%s]   Justification: %s", pair.Symbol, fudAttack.Justification)
						} else {
							log.Printf("[%s] ✓ No coordinated FUD attack detected", pair.Symbol)
							log.Printf("[%s]   Confidence: %.0f%%", pair.Symbol, fudAttack.Confidence*100)
							log.Printf("[%s]   %s", pair.Symbol, fudAttack.Justification)
						}
						log.Printf("[%s] ================================\n", pair.Symbol)
					}
					state.LastFudCheckTime = now
					if len(tweets) > 0 {
						state.LastFudCheckTweetID = tweets[0].ID
					}
				} else {
					log.Printf("[%s] Skipping FUD check: only %d new messages (need 3+)", pair.Symbol, newMessagesCount)
				}
			}
		}
	}

	decision := MakeTradingDecision(btcIchimoku.Analysis, coinIchimoku.Analysis, activityAnalysis, fudActivityAnalysis, sentiment)
	log.Printf("\n[%s] ===== DECISION: %s (reason: %s) =====", pair.Symbol, decision.Signal, decision.Reason)
	log.Printf("[%s] Explanation: %s", pair.Symbol, decision.Explanation)

	fudAttackInfo := "no"
	if lastFudAttack.HasAttack {
		fudAttackInfo = "yes"
	}

	decisionRecord := TradingDecisionRecord{
		PositionUUID:        state.PositionUUID,
		Symbol:              pair.Symbol,
		BTCIchimoku:         decision.BTCIchimokuSignal,
		CoinIchimoku:        decision.CoinIchimokuSignal,
		Activity:            decision.ActivitySignal,
		FudActivity:         decision.FudActivitySignal,
		Sentiment:           decision.SentimentSignal,
		FudAttack:           fudAttackInfo,
		FinalDecision:       string(decision.Signal),
		DecisionExplanation: decision.Explanation,
		CreatedAt:           time.Now(),
	}

	shouldSave := false
	lastDecision, err := GetLatestTradingDecision(pair.Symbol)
	if err != nil {
		log.Printf("[%s] Failed to get last decision: %v", pair.Symbol, err)
		shouldSave = true
	} else if lastDecision == nil {
		shouldSave = true
	} else if lastDecision.BTCIchimoku != decisionRecord.BTCIchimoku ||
		lastDecision.CoinIchimoku != decisionRecord.CoinIchimoku ||
		lastDecision.Activity != decisionRecord.Activity ||
		lastDecision.FudActivity != decisionRecord.FudActivity ||
		lastDecision.Sentiment != decisionRecord.Sentiment ||
		lastDecision.FudAttack != decisionRecord.FudAttack ||
		lastDecision.FinalDecision != decisionRecord.FinalDecision {
		shouldSave = true
	}

	if shouldSave {
		if err := SaveTradingDecision(decisionRecord); err != nil {
			log.Printf("[%s] Failed to save trading decision: %v", pair.Symbol, err)
		} else {
			log.Printf("[%s] Trading decision saved to database", pair.Symbol)
		}
	}

	if decision.Signal == SignalEmpty {
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
			state.PositionUUID = ""
			log.Printf("[%s] Position closed", pair.Symbol)
		} else {
			log.Printf("[%s] No signal - no action", pair.Symbol)
		}
		state.OpenReason = ""
		return nil
	}

	desiredPosition := PositionSideBoth
	if decision.Signal == SignalLong {
		desiredPosition = PositionSideLong
	} else if decision.Signal == SignalShort {
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
		state.OpenReason = ""
		state.PositionUUID = ""
	}

	log.Printf("[%s] Opening %s position", pair.Symbol, desiredPosition)
	position, err := exchange.OpenPosition(pair.Symbol, desiredPosition, pair.Leverage, pair.Quantity)
	if err != nil {
		log.Printf("[%s] Failed to open %s: %v", pair.Symbol, desiredPosition, err)
		return err
	}

	state.CurrentPosition = desiredPosition
	state.OpenedAt = time.Now()
	state.OpenReason = decision.Reason
	state.PositionUUID = GeneratePositionUUID()

	if err := SaveTradingDecision(TradingDecisionRecord{
		PositionUUID:        state.PositionUUID,
		Symbol:              pair.Symbol,
		BTCIchimoku:         decision.BTCIchimokuSignal,
		CoinIchimoku:        decision.CoinIchimokuSignal,
		Activity:            decision.ActivitySignal,
		FudActivity:         decision.FudActivitySignal,
		Sentiment:           decision.SentimentSignal,
		FudAttack:           fudAttackInfo,
		FinalDecision:       string(decision.Signal),
		DecisionExplanation: decision.Explanation,
		CreatedAt:           time.Now(),
	}); err != nil {
		log.Printf("[%s] Failed to save opening decision: %v", pair.Symbol, err)
	}

	log.Printf("[%s] Position opened: %s (entry: %.6f, amount: %.6f, reason: %s, UUID: %s)", pair.Symbol, desiredPosition, position.EntryPrice, position.Amount, decision.Reason, state.PositionUUID)

	return nil
}
