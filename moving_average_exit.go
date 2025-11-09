package main

import (
	"log"
	"math"
)

func CalculateMovingAveragePnLSignal(snapshots []PositionSnapshot, currentPnL float64) MovingAveragePnLSignal {
	signal := MovingAveragePnLSignal{
		ShouldClose:    false,
		CurrentPnL:     currentPnL,
		SnapshotsCount: len(snapshots),
	}

	if len(snapshots) < 10 {
		signal.TriggerReason = "Not enough snapshots for analysis (minimum 10 required)"
		return signal
	}

	var sum float64
	for _, snap := range snapshots {
		sum += snap.UnrealizedPL
	}
	movingAverage := sum / float64(len(snapshots))
	signal.MovingAverage = movingAverage

	if movingAverage <= 0 {
		signal.TriggerReason = "Moving average is negative or zero - no exit signal"
		return signal
	}

	threshold := movingAverage * 0.7
	signal.Threshold = threshold

	percentBelowMA := ((movingAverage - currentPnL) / math.Abs(movingAverage)) * 100
	signal.PercentBelowMA = percentBelowMA

	if currentPnL < threshold && currentPnL > 0 {
		signal.ShouldClose = true
		signal.TriggerReason = "Current PnL dropped below 70% of moving average - exit signal triggered"
		log.Printf("⚠️ MA EXIT SIGNAL: Current PnL ($%.2f) is %.1f%% below MA ($%.2f), threshold: $%.2f",
			currentPnL, percentBelowMA, movingAverage, threshold)
	} else if currentPnL <= 0 && movingAverage > 0 {
		signal.ShouldClose = true
		signal.TriggerReason = "Current PnL turned negative while MA is positive - exit signal triggered"
		log.Printf("⚠️ MA EXIT SIGNAL: PnL turned negative ($%.2f) while MA is positive ($%.2f)",
			currentPnL, movingAverage)
	} else {
		signal.TriggerReason = "No exit signal - position performing within acceptable range"
	}

	return signal
}
