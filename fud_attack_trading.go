package main

import (
	"log"
	"time"
)

func processFudAttackTradingCycle(
	exchange AsterDexExchange,
	pair TradingPair,
	state *TradingState,
	lastFudAttack ClaudeFudAttackResponse,
	coinIchimoku IchimokuAnalysis,
) (bool, error) {

	if !state.FudAttackMode {
		return false, nil
	}

	log.Printf("[%s] === FUD ATTACK MODE ACTIVE ===", pair.Symbol)

	if lastFudAttack.LastAttackTime == nil {
		log.Printf("[%s] FUD attack has no timestamp, skipping FUD mode", pair.Symbol)
		return false, nil
	}

	timeSinceAttack := time.Since(*lastFudAttack.LastAttackTime)

	if timeSinceAttack > 12*time.Hour {
		log.Printf("[%s] More than 12 hours since FUD attack, checking exit conditions...", pair.Symbol)

		coinSignal := convertIchimokuToSignal(coinIchimoku)
		if coinSignal == SignalLong || coinSignal == SignalEmpty {
			log.Printf("[%s] Exit FUD mode: coin signal is %s and 12+ hours passed", pair.Symbol, coinSignal)

			if state.CurrentPosition == PositionSideShort {
				if err := exchange.ClosePosition(pair.Symbol, PositionSideShort); err != nil {
					log.Printf("[%s] Failed to close SHORT: %v", pair.Symbol, err)
					return true, err
				}

				if state.PositionUUID != "" {
					markPrice, _ := exchange.GetMarkPrice(pair.Symbol)
					closedPosition, _ := exchange.GetPosition(pair.Symbol)
					realizedPL := 0.0
					if closedPosition != nil {
						realizedPL = closedPosition.UnrealizedPL
					}
					if err := UpdatePositionClose(state.PositionUUID, markPrice, realizedPL, "fud_mode_exit"); err != nil {
						log.Printf("[%s] Failed to update position close: %v", pair.Symbol, err)
					}
				}

				state.CurrentPosition = PositionSideBoth
				state.PositionUUID = ""
				state.OpenReason = ""
			}

			state.FudAttackMode = false
			state.FudAttackShortStarted = false
			state.FudAttackStartTime = time.Time{}
			log.Printf("[%s] === FUD ATTACK MODE DEACTIVATED ===", pair.Symbol)
			return true, nil
		}
	}

	if !state.FudAttackShortStarted {
		coinSignal := convertIchimokuToSignal(coinIchimoku)
		if coinSignal == SignalShort {
			log.Printf("[%s] Coin Ichimoku SHORT detected while in FUD mode - transitioning to real SHORT", pair.Symbol)
			state.FudAttackShortStarted = true
		}
		return true, nil
	}

	coinSignal := convertIchimokuToSignal(coinIchimoku)
	if coinSignal == SignalLong {
		log.Printf("[%s] Exit FUD mode: coin signal switched to LONG after real SHORT started", pair.Symbol)

		if state.CurrentPosition == PositionSideShort {
			if err := exchange.ClosePosition(pair.Symbol, PositionSideShort); err != nil {
				log.Printf("[%s] Failed to close SHORT: %v", pair.Symbol, err)
				return true, err
			}

			if state.PositionUUID != "" {
				markPrice, _ := exchange.GetMarkPrice(pair.Symbol)
				closedPosition, _ := exchange.GetPosition(pair.Symbol)
				realizedPL := 0.0
				if closedPosition != nil {
					realizedPL = closedPosition.UnrealizedPL
				}
				if err := UpdatePositionClose(state.PositionUUID, markPrice, realizedPL, "fud_mode_long_signal"); err != nil {
					log.Printf("[%s] Failed to update position close: %v", pair.Symbol, err)
				}
			}

			state.CurrentPosition = PositionSideBoth
			state.PositionUUID = ""
			state.OpenReason = ""
		}

		state.FudAttackMode = false
		state.FudAttackShortStarted = false
		state.FudAttackStartTime = time.Time{}
		log.Printf("[%s] === FUD ATTACK MODE DEACTIVATED ===", pair.Symbol)
		return true, nil
	}

	log.Printf("[%s] Holding in FUD attack mode, coin signal: %s", pair.Symbol, coinSignal)
	return true, nil
}
