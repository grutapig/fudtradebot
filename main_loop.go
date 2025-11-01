package main

import (
	"encoding/json"
	"github.com/grutapig/fudtradebot/claude"
	"log"
	"time"
)

func runBalanceCollector(exchange AsterDexExchange) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		balances, err := exchange.GetAllBalances()
		if err != nil {
			log.Printf("Failed to get balances: %v", err)
		} else {
			if err := SaveAllBalances(balances); err != nil {
				log.Printf("Failed to save balances: %v", err)
			} else {
				log.Printf("Balances saved - %d assets", len(balances))
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
	} else if position == nil {
		log.Printf("[%s] ✓ No existing position found on exchange", pair.Symbol)
		if err := CloseOpenPositionsBySymbol(pair.Symbol); err != nil {
			log.Printf("[%s] Failed to close orphaned DB positions: %v", pair.Symbol, err)
		} else {
			log.Printf("[%s] ✓ Closed any orphaned positions in database", pair.Symbol)
		}
	} else if position != nil {
		state.CurrentPosition = position.Side
		state.OpenedAt = position.Timestamp
		log.Printf("[%s] ✓ Restored existing position: %s (opened at %s)",
			pair.Symbol, position.Side, position.Timestamp.Format("2006-01-02 15:04:05"))
		log.Printf("[%s]   Entry price: %.6f, Amount: %.6f, P/L: %.2f USDT",
			pair.Symbol, position.EntryPrice, position.Amount, position.UnrealizedPL)

		oppositeSide := PositionSideLong
		if position.Side == PositionSideLong {
			oppositeSide = PositionSideShort
		}
		if err := DeleteOpenPositionBySymbolAndSide(pair.Symbol, string(oppositeSide)); err == nil {
			log.Printf("[%s] ✓ Deleted opposite %s position from database", pair.Symbol, oppositeSide)
		}

		dbPosition, err := GetOpenPositionBySymbolAndSide(pair.Symbol, string(position.Side))
		if err == nil {
			state.PositionUUID = dbPosition.UUID
			log.Printf("[%s] ✓ Imported position UUID from database: %s", pair.Symbol, state.PositionUUID)
		} else {
			state.PositionUUID = GeneratePositionUUID()
			log.Printf("[%s] ✓ Generated new position UUID: %s", pair.Symbol, state.PositionUUID)

			positionRecord := PositionRecord{
				UUID:       state.PositionUUID,
				Symbol:     pair.Symbol,
				Side:       string(position.Side),
				Leverage:   pair.Leverage,
				Quantity:   pair.Quantity,
				EntryPrice: position.EntryPrice,
				OpenedAt:   position.Timestamp,
				OpenReason: "restored_from_exchange",
				MaxPnL:     position.UnrealizedPL,
				MinPnL:     position.UnrealizedPL,
				CreatedAt:  time.Now(),
			}
			if err := SavePositionOpen(positionRecord); err != nil {
				log.Printf("[%s] Failed to save restored position to database: %v", pair.Symbol, err)
			} else {
				log.Printf("[%s] Restored position saved to database", pair.Symbol)
			}
		}
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

		if err := SavePositionSnapshot(*currentPosition, markPrice, state.PositionUUID); err != nil {
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

	interval := KLINES_INTERVAL
	btcKlines, err := exchange.Klines("BTCUSDT", interval, 0, 0, 52)
	if err != nil {
		log.Printf("[%s] Failed to get BTC price data: %v", pair.Symbol, err)
		return err
	}
	coinKlines, err := exchange.Klines(pair.Symbol, interval, 0, 0, 52)
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

						if err := SaveFudAttack(fudAttack, pair.Symbol, state.PositionUUID); err != nil {
							log.Printf("[%s] Failed to save FUD attack to database: %v", pair.Symbol, err)
						} else {
							log.Printf("[%s] FUD attack analysis saved to database", pair.Symbol)
						}

						log.Printf("\n[%s] ===== FUD ATTACK ANALYSIS =====", pair.Symbol)
						if fudAttack.HasAttack {
							log.Printf("[%s] ⚠️  COORDINATED FUD ATTACK DETECTED!", pair.Symbol)
							log.Printf("[%s]   Confidence: %.0f%%", pair.Symbol, fudAttack.Confidence*100)
							log.Printf("[%s]   Messages: %d", pair.Symbol, fudAttack.MessageCount)
							log.Printf("[%s]   FUD Type: %s", pair.Symbol, fudAttack.FudType)
							log.Printf("[%s]   Theme: %s", pair.Symbol, fudAttack.Theme)
							log.Printf("[%s]   Started: %d hours ago", pair.Symbol, fudAttack.StartedHoursAgo)
							log.Printf("[%s]   Last Attack Time: %s", pair.Symbol, fudAttack.LastAttackTime.Format("2006-01-02 15:04:05"))
							log.Printf("[%s]   Participants:", pair.Symbol)
							for _, p := range fudAttack.Participants {
								log.Printf("[%s]     - %s (%d messages)", pair.Symbol, p.Username, p.MessageCount)
							}
							log.Printf("[%s]   Justification: %s", pair.Symbol, fudAttack.Justification)

							timeSinceAttack := time.Since(fudAttack.LastAttackTime)
							if timeSinceAttack <= 1*time.Hour && !state.FudAttackMode {
								log.Printf("[%s] 🚨 ACTIVATING FUD ATTACK TRADING MODE (attack is fresh: %.0f min ago)", pair.Symbol, timeSinceAttack.Minutes())
								state.FudAttackMode = true
								state.FudAttackStartTime = fudAttack.LastAttackTime
								state.FudAttackShortStarted = false
							}
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

	handledByFudMode, err := processFudAttackTradingCycle(exchange, pair, state, lastFudAttack, coinIchimoku.Analysis)
	if err != nil {
		log.Printf("[%s] Error in FUD attack trading cycle: %v", pair.Symbol, err)
	}
	if handledByFudMode {
		log.Printf("[%s] Cycle handled by FUD attack mode", pair.Symbol)
		return nil
	}

	if state.FudAttackMode {
		log.Printf("[%s] 🚨 FUD ATTACK MODE: Opening forced SHORT position", pair.Symbol)

		if state.CurrentPosition != PositionSideShort {
			if state.CurrentPosition != PositionSideBoth {
				log.Printf("[%s] Closing existing %s position", pair.Symbol, state.CurrentPosition)
				if err := exchange.ClosePosition(pair.Symbol, state.CurrentPosition); err != nil {
					log.Printf("[%s] Failed to close position: %v", pair.Symbol, err)
					return err
				}

				if state.PositionUUID != "" {
					markPrice, _ := exchange.GetMarkPrice(pair.Symbol)
					closedPosition, _ := exchange.GetPosition(pair.Symbol)
					realizedPL := 0.0
					if closedPosition != nil {
						realizedPL = closedPosition.UnrealizedPL
					}
					if err := UpdatePositionClose(state.PositionUUID, markPrice, realizedPL, "fud_mode_switch"); err != nil {
						log.Printf("[%s] Failed to update position close: %v", pair.Symbol, err)
					}
				}

				state.CurrentPosition = PositionSideBoth
				state.PositionUUID = ""
				state.OpenReason = ""
			}

			log.Printf("[%s] Opening forced SHORT position due to FUD attack", pair.Symbol)
			position, err := exchange.OpenPosition(pair.Symbol, PositionSideShort, pair.Leverage, pair.Quantity)
			if err != nil {
				log.Printf("[%s] Failed to open SHORT: %v", pair.Symbol, err)
				return err
			}

			state.CurrentPosition = PositionSideShort
			state.OpenedAt = time.Now()
			state.OpenReason = "fud_attack_forced"
			state.PositionUUID = GeneratePositionUUID()

			log.Printf("[%s] FUD SHORT position opened: entry %.6f, amount %.6f, UUID: %s",
				pair.Symbol, position.EntryPrice, position.Amount, state.PositionUUID)

			positionRecord := PositionRecord{
				UUID:       state.PositionUUID,
				Symbol:     pair.Symbol,
				Side:       string(PositionSideShort),
				Leverage:   pair.Leverage,
				Quantity:   pair.Quantity,
				EntryPrice: position.EntryPrice,
				OpenedAt:   state.OpenedAt,
				OpenReason: "fud_attack_forced",
				MaxPnL:     position.UnrealizedPL,
				MinPnL:     position.UnrealizedPL,
				CreatedAt:  time.Now(),
			}
			if err := SavePositionOpen(positionRecord); err != nil {
				log.Printf("[%s] Failed to save position to database: %v", pair.Symbol, err)
			}
		}

		return nil
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

	var savedDecision *TradingDecisionRecord
	if shouldSave {
		if err := SaveTradingDecision(decisionRecord); err != nil {
			log.Printf("[%s] Failed to save trading decision: %v", pair.Symbol, err)
		} else {
			log.Printf("[%s] Trading decision saved to database", pair.Symbol)
			lastSaved, _ := GetLatestTradingDecision(pair.Symbol)
			savedDecision = lastSaved
		}
	}

	if state.CurrentPosition != PositionSideBoth {
		shouldClose := ShouldClosePosition(state.CurrentPosition, coinIchimoku)
		if shouldClose {
			log.Printf("[%s] Ichimoku signals to close %s position", pair.Symbol, state.CurrentPosition)

			var closedPosition *Position
			if state.CurrentPosition == PositionSideLong {
				if err := exchange.ClosePosition(pair.Symbol, PositionSideLong); err != nil {
					log.Printf("[%s] Failed to close LONG: %v", pair.Symbol, err)
					return err
				}
				closedPosition, _ = exchange.GetPosition(pair.Symbol)
			} else if state.CurrentPosition == PositionSideShort {
				if err := exchange.ClosePosition(pair.Symbol, PositionSideShort); err != nil {
					log.Printf("[%s] Failed to close SHORT: %v", pair.Symbol, err)
					return err
				}
				closedPosition, _ = exchange.GetPosition(pair.Symbol)
			}

			if state.PositionUUID != "" {
				markPrice, _ := exchange.GetMarkPrice(pair.Symbol)
				realizedPL := 0.0
				if closedPosition != nil {
					realizedPL = closedPosition.UnrealizedPL
				}
				if err := UpdatePositionClose(state.PositionUUID, markPrice, realizedPL, "ichimoku_exit"); err != nil {
					log.Printf("[%s] Failed to update position close: %v", pair.Symbol, err)
				} else {
					log.Printf("[%s] Position close recorded in database", pair.Symbol)
				}
			}

			state.CurrentPosition = PositionSideBoth
			state.PositionUUID = ""
			state.OpenReason = ""
			log.Printf("[%s] Position closed by Ichimoku exit signal", pair.Symbol)
		} else {
			log.Printf("[%s] Position held - Ichimoku conditions not met for exit", pair.Symbol)
		}
		return nil
	}

	if decision.Signal == SignalEmpty {
		log.Printf("[%s] No signal - no action", pair.Symbol)
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
	var validationRecord *AIOrderValidationRecord
	if claudeClient != nil {
		log.Printf("[%s] Validating order decision with AI...", pair.Symbol)
		aiValidation, err := ValidateOrderWithAI(*claudeClient, decision, btcIchimoku.Analysis, coinIchimoku.Analysis, activityAnalysis, fudActivityAnalysis, sentiment)
		if err != nil {
			log.Printf("[%s] AI validation failed: %v", pair.Symbol, err)
			log.Printf("[%s] Proceeding without AI validation", pair.Symbol)
		} else {
			log.Printf("[%s] AI Validation Result:", pair.Symbol)
			log.Printf("[%s]   Should Open: %v", pair.Symbol, aiValidation.ShouldOpenOrder)
			log.Printf("[%s]   Confidence: %.1f%%", pair.Symbol, aiValidation.ConfidencePercent)
			log.Printf("[%s]   Justification: %s", pair.Symbol, aiValidation.Justification)

			requestDataJSON, _ := json.Marshal(map[string]interface{}{
				"decision":      decision,
				"btc_ichimoku":  btcIchimoku.Analysis,
				"coin_ichimoku": coinIchimoku.Analysis,
				"activity":      activityAnalysis,
				"fud_activity":  fudActivityAnalysis,
				"sentiment":     sentiment,
			})
			responseDataJSON, _ := json.Marshal(aiValidation)

			validationRecord = &AIOrderValidationRecord{
				PositionUUID:      "",
				DecisionRecordID:  0,
				Symbol:            pair.Symbol,
				RequestData:       string(requestDataJSON),
				ResponseData:      string(responseDataJSON),
				ShouldOpenOrder:   aiValidation.ShouldOpenOrder,
				ConfidencePercent: aiValidation.ConfidencePercent,
				Justification:     aiValidation.Justification,
				CreatedAt:         time.Now(),
			}

			if savedDecision != nil && savedDecision.ID > 0 {
				validationRecord.DecisionRecordID = savedDecision.ID
			}

			if !aiValidation.ShouldOpenOrder {
				log.Printf("[%s] ❌ AI rejected the order - not opening position", pair.Symbol)
				if err := SaveAIOrderValidation(validationRecord); err != nil {
					log.Printf("[%s] Failed to save AI validation: %v", pair.Symbol, err)
				} else {
					log.Printf("[%s] AI validation saved to database", pair.Symbol)
				}
				return nil
			}

			log.Printf("[%s] ✅ AI approved the order - proceeding", pair.Symbol)
		}
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
	if validationRecord != nil {
		validationRecord.PositionUUID = state.PositionUUID
		DB.Save(validationRecord)
	}
	if savedDecision != nil && savedDecision.ID > 0 {
		if err := UpdateDecisionPositionUUIDByID(savedDecision.ID, state.PositionUUID); err != nil {
			log.Printf("[%s] Failed to update decision ID %d with position UUID: %v", pair.Symbol, savedDecision.ID, err)
		} else {
			log.Printf("[%s] Decision ID %d updated with position UUID: %s", pair.Symbol, savedDecision.ID, state.PositionUUID)
		}
	} else {
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
		} else {
			log.Printf("[%s] New decision saved with position UUID: %s", pair.Symbol, state.PositionUUID)
		}
	}

	log.Printf("[%s] Position opened: %s (entry: %.6f, amount: %.6f, reason: %s, UUID: %s)", pair.Symbol, desiredPosition, position.EntryPrice, position.Amount, decision.Reason, state.PositionUUID)

	positionRecord := PositionRecord{
		UUID:       state.PositionUUID,
		Symbol:     pair.Symbol,
		Side:       string(desiredPosition),
		Leverage:   pair.Leverage,
		Quantity:   pair.Quantity,
		EntryPrice: position.EntryPrice,
		OpenedAt:   state.OpenedAt,
		OpenReason: decision.Reason,
		MaxPnL:     position.UnrealizedPL,
		MinPnL:     position.UnrealizedPL,
		CreatedAt:  time.Now(),
	}
	if err := SavePositionOpen(positionRecord); err != nil {
		log.Printf("[%s] Failed to save position to database: %v", pair.Symbol, err)
	} else {
		log.Printf("[%s] Position record saved to database", pair.Symbol)
	}

	if claudeClient != nil {
		if err := DB.Model(&AIOrderValidationRecord{}).
			Where("symbol = ? AND position_uuid = ''", pair.Symbol).
			Order("created_at DESC").
			Limit(1).
			Update("position_uuid", state.PositionUUID).Error; err != nil {
			log.Printf("[%s] Failed to link AI validation to position: %v", pair.Symbol, err)
		} else {
			log.Printf("[%s] AI validation linked to position UUID: %s", pair.Symbol, state.PositionUUID)
		}
	}

	return nil
}
