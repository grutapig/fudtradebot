package main

type SignalStrength int

const (
	SignalStrengthNone   SignalStrength = 0
	SignalStrengthWeak   SignalStrength = 1
	SignalStrengthMedium SignalStrength = 2
	SignalStrengthStrong SignalStrength = 3
)

type TradingSignal struct {
	Action   TradingAction
	Strength SignalStrength
	Reasons  []string
}

func MakeTradingDecision(
	ichimokuAnalysis IchimokuAnalysis,
	activityAnalysis ActivityAnalysis,
	fudActivityAnalysis ActivityAnalysis,
	currentPosition PositionSide,
) TradingSignal {

	longSignals := 0
	shortSignals := 0
	reasons := []string{}

	ichimokuLong := false
	ichimokuShort := false
	ichimokuNeutral := false

	switch ichimokuAnalysis.Signal {
	case IchimokuSignalStrongLong:
		longSignals += 3
		ichimokuLong = true
		reasons = append(reasons, "Ichimoku: Strong LONG signal")
	case IchimokuSignalLong:
		longSignals += 2
		ichimokuLong = true
		reasons = append(reasons, "Ichimoku: LONG signal")
	case IchimokuSignalStrongShort:
		shortSignals += 3
		ichimokuShort = true
		reasons = append(reasons, "Ichimoku: Strong SHORT signal")
	case IchimokuSignalShort:
		shortSignals += 2
		ichimokuShort = true
		reasons = append(reasons, "Ichimoku: SHORT signal")
	case IchimokuSignalNeutral, IchimokuSignalUncertain:
		ichimokuNeutral = true
		reasons = append(reasons, "Ichimoku: Neutral/Uncertain")
	}

	switch activityAnalysis.Trend {
	case ActivityTrendSharpRise:
		longSignals++
		reasons = append(reasons, "Activity: Sharp rise detected (LONG signal)")
	case ActivityTrendSharpDrop:
		shortSignals++
		reasons = append(reasons, "Activity: Sharp drop detected (SHORT signal)")
	}

	switch fudActivityAnalysis.Trend {
	case ActivityTrendSharpRise:
		shortSignals++
		reasons = append(reasons, "FUD Activity: Sharp rise detected (SHORT signal)")
	}

	if ichimokuLong && activityAnalysis.Trend == ActivityTrendSharpRise {
		reasons = append(reasons, "Confirmation: Ichimoku LONG + Activity rise")
	}

	if ichimokuShort && activityAnalysis.Trend == ActivityTrendSharpDrop {
		reasons = append(reasons, "Confirmation: Ichimoku SHORT + Activity drop")
	}

	if ichimokuShort && fudActivityAnalysis.Trend == ActivityTrendSharpRise {
		reasons = append(reasons, "Confirmation: Ichimoku SHORT + FUD rise")
	}

	if ichimokuLong && (activityAnalysis.Trend == ActivityTrendSharpDrop || fudActivityAnalysis.Trend == ActivityTrendSharpRise) {
		reasons = append(reasons, "Contradiction: Ichimoku LONG conflicts with sentiment")
		if currentPosition == PositionSideLong {
			return TradingSignal{
				Action:   TradingActionCloseLong,
				Strength: SignalStrengthMedium,
				Reasons:  append(reasons, "Closing LONG due to contradiction"),
			}
		}
		return TradingSignal{
			Action:   TradingActionHold,
			Strength: SignalStrengthNone,
			Reasons:  append(reasons, "HOLD: Contradicting signals"),
		}
	}

	if ichimokuShort && activityAnalysis.Trend == ActivityTrendSharpRise {
		reasons = append(reasons, "Contradiction: Ichimoku SHORT conflicts with activity rise")
		if currentPosition == PositionSideShort {
			return TradingSignal{
				Action:   TradingActionCloseShort,
				Strength: SignalStrengthMedium,
				Reasons:  append(reasons, "Closing SHORT due to contradiction"),
			}
		}
		return TradingSignal{
			Action:   TradingActionHold,
			Strength: SignalStrengthNone,
			Reasons:  append(reasons, "HOLD: Contradicting signals"),
		}
	}

	if ichimokuNeutral && (activityAnalysis.Trend != ActivityTrendPlateau || fudActivityAnalysis.Trend != ActivityTrendPlateau) {
		if activityAnalysis.Trend == ActivityTrendSharpRise {
			return TradingSignal{
				Action:   TradingActionOpenLong,
				Strength: SignalStrengthWeak,
				Reasons:  append(reasons, "Ichimoku neutral, but activity rising - weak LONG"),
			}
		}
		if activityAnalysis.Trend == ActivityTrendSharpDrop || fudActivityAnalysis.Trend == ActivityTrendSharpRise {
			return TradingSignal{
				Action:   TradingActionOpenShort,
				Strength: SignalStrengthWeak,
				Reasons:  append(reasons, "Ichimoku neutral, but sentiment bearish - weak SHORT"),
			}
		}
	}

	var action TradingAction
	var strength SignalStrength

	if longSignals > shortSignals {
		if currentPosition == PositionSideShort {
			action = TradingActionCloseShort
			strength = getStrength(longSignals)
			reasons = append(reasons, "Closing SHORT position")
		} else if currentPosition == PositionSideBoth {
			action = TradingActionOpenLong
			strength = getStrength(longSignals)
			reasons = append(reasons, "Opening LONG position")
		} else {
			action = TradingActionHold
			strength = SignalStrengthNone
			reasons = append(reasons, "Already in LONG or signals too weak")
		}
	} else if shortSignals > longSignals {
		if currentPosition == PositionSideLong {
			action = TradingActionCloseLong
			strength = getStrength(shortSignals)
			reasons = append(reasons, "Closing LONG position")
		} else if currentPosition == PositionSideBoth {
			action = TradingActionOpenShort
			strength = getStrength(shortSignals)
			reasons = append(reasons, "Opening SHORT position")
		} else {
			action = TradingActionHold
			strength = SignalStrengthNone
			reasons = append(reasons, "Already in SHORT or signals too weak")
		}
	} else {
		action = TradingActionHold
		strength = SignalStrengthNone
		reasons = append(reasons, "No clear direction - HOLD")
	}

	return TradingSignal{
		Action:   action,
		Strength: strength,
		Reasons:  reasons,
	}
}

func getStrength(signalCount int) SignalStrength {
	if signalCount >= 5 {
		return SignalStrengthStrong
	} else if signalCount >= 3 {
		return SignalStrengthMedium
	} else if signalCount >= 1 {
		return SignalStrengthWeak
	}
	return SignalStrengthNone
}
