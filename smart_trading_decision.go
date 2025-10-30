package main

func MakeSmartTradingDecision(
	activityAnalysis ActivityAnalysis,
	fudActivityAnalysis ActivityAnalysis,
	sentimentAnalysis ClaudeSentimentResponse,
	currentPosition PositionSide,
) TradingSignal {
	longSignals := 0
	shortSignals := 0
	reasons := []string{}

	sentimentConfirmation := false

	switch activityAnalysis.Trend {
	case ActivityTrendSharpRise:
		longSignals += 2
		reasons = append(reasons, "Activity: Sharp rise detected")
	case ActivityTrendSharpDrop:
		shortSignals += 2
		reasons = append(reasons, "Activity: Sharp drop detected")
	}

	switch fudActivityAnalysis.Trend {
	case ActivityTrendSharpRise:
		shortSignals += 2
		reasons = append(reasons, "FUD Activity: Sharp rise detected")
	}

	if sentimentAnalysis.OverallSentiment >= 5 {
		longSignals += 1
		reasons = append(reasons, "Sentiment confirmation: Very positive (+5 to +10)")
		if activityAnalysis.Trend == ActivityTrendSharpRise {
			sentimentConfirmation = true
		}
	} else if sentimentAnalysis.OverallSentiment >= 2 {
		longSignals += 1
		reasons = append(reasons, "Sentiment confirmation: Moderately positive (+2 to +4)")
		if activityAnalysis.Trend == ActivityTrendSharpRise {
			sentimentConfirmation = true
		}
	} else if sentimentAnalysis.OverallSentiment >= 0 {
		reasons = append(reasons, "Sentiment confirmation: Slightly positive (0 to +1)")
		if activityAnalysis.Trend == ActivityTrendSharpRise {
			sentimentConfirmation = true
		}
	} else if sentimentAnalysis.OverallSentiment <= -5 {
		shortSignals += 1
		reasons = append(reasons, "Sentiment confirmation: Very negative (-5 to -10)")
		if activityAnalysis.Trend == ActivityTrendSharpDrop || fudActivityAnalysis.Trend == ActivityTrendSharpRise {
			sentimentConfirmation = true
		}
	} else if sentimentAnalysis.OverallSentiment <= -2 {
		shortSignals += 1
		reasons = append(reasons, "Sentiment confirmation: Moderately negative (-2 to -4)")
		if activityAnalysis.Trend == ActivityTrendSharpDrop || fudActivityAnalysis.Trend == ActivityTrendSharpRise {
			sentimentConfirmation = true
		}
	} else {
		reasons = append(reasons, "Sentiment confirmation: Slightly negative (-1)")
		if activityAnalysis.Trend == ActivityTrendSharpDrop || fudActivityAnalysis.Trend == ActivityTrendSharpRise {
			sentimentConfirmation = true
		}
	}

	switch sentimentAnalysis.SentimentTrend {
	case "improving":
		reasons = append(reasons, "Sentiment trend: Improving")
	case "declining":
		reasons = append(reasons, "Sentiment trend: Declining")
	case "stable":
		reasons = append(reasons, "Sentiment trend: Stable")
	}

	if sentimentAnalysis.FudLevel >= 7 {
		reasons = append(reasons, "FUD level: Very high (7-10)")
	} else if sentimentAnalysis.FudLevel >= 4 {
		reasons = append(reasons, "FUD level: Moderate (4-6)")
	}

	switch sentimentAnalysis.Recommendation {
	case "bullish":
		reasons = append(reasons, "Claude recommendation: Bullish")
	case "bearish":
		reasons = append(reasons, "Claude recommendation: Bearish")
	case "neutral":
		reasons = append(reasons, "Claude recommendation: Neutral")
	}

	if sentimentAnalysis.Confidence < 0.5 {
		reasons = append(reasons, "Low confidence in sentiment analysis (<0.5)")
		sentimentConfirmation = false
	}

	if !sentimentConfirmation {
		reasons = append(reasons, "⚠️ Sentiment does NOT confirm activity signals - weakening signal strength")
		if longSignals > 0 {
			longSignals = (longSignals + 1) / 2
		}
		if shortSignals > 0 {
			shortSignals = (shortSignals + 1) / 2
		}
	} else {
		reasons = append(reasons, "✓ Sentiment CONFIRMS activity signals")
	}

	if longSignals > shortSignals && shortSignals > 0 {
		reasons = append(reasons, "Mixed signals detected - proceeding with caution")
	} else if shortSignals > longSignals && longSignals > 0 {
		reasons = append(reasons, "Mixed signals detected - proceeding with caution")
	}

	var action TradingAction
	var strength SignalStrength

	isPlateauSituation := activityAnalysis.Trend == ActivityTrendPlateau &&
		fudActivityAnalysis.Trend == ActivityTrendPlateau &&
		sentimentAnalysis.SentimentTrend == "stable" &&
		sentimentAnalysis.OverallSentiment >= -2 && sentimentAnalysis.OverallSentiment <= 2

	if isPlateauSituation && currentPosition != PositionSideBoth {
		if currentPosition == PositionSideLong {
			action = TradingActionCloseLong
			strength = SignalStrengthMedium
			reasons = append(reasons, "Plateau detected - closing LONG position")
		} else if currentPosition == PositionSideShort {
			action = TradingActionCloseShort
			strength = SignalStrengthMedium
			reasons = append(reasons, "Plateau detected - closing SHORT position")
		} else {
			action = TradingActionHold
			strength = SignalStrengthNone
			reasons = append(reasons, "No clear direction - HOLD")
		}
	} else if longSignals > shortSignals {
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
			reasons = append(reasons, "Already in LONG position")
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
			reasons = append(reasons, "Already in SHORT position")
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
