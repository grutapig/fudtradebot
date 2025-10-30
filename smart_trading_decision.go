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

	if sentimentAnalysis.OverallSentiment >= 5 {
		longSignals += 3
		reasons = append(reasons, "Sentiment: Very positive (+5 to +10)")
	} else if sentimentAnalysis.OverallSentiment >= 2 {
		longSignals += 2
		reasons = append(reasons, "Sentiment: Moderately positive (+2 to +4)")
	} else if sentimentAnalysis.OverallSentiment >= 0 {
		longSignals += 1
		reasons = append(reasons, "Sentiment: Slightly positive (0 to +1)")
	} else if sentimentAnalysis.OverallSentiment <= -5 {
		shortSignals += 3
		reasons = append(reasons, "Sentiment: Very negative (-5 to -10)")
	} else if sentimentAnalysis.OverallSentiment <= -2 {
		shortSignals += 2
		reasons = append(reasons, "Sentiment: Moderately negative (-2 to -4)")
	} else {
		shortSignals += 1
		reasons = append(reasons, "Sentiment: Slightly negative (-1)")
	}

	switch sentimentAnalysis.SentimentTrend {
	case "improving":
		longSignals += 2
		reasons = append(reasons, "Sentiment trend: Improving")
	case "declining":
		shortSignals += 2
		reasons = append(reasons, "Sentiment trend: Declining")
	case "stable":
		reasons = append(reasons, "Sentiment trend: Stable")
	}

	if sentimentAnalysis.FudLevel >= 7 {
		shortSignals += 3
		reasons = append(reasons, "FUD level: Very high (7-10)")
	} else if sentimentAnalysis.FudLevel >= 4 {
		shortSignals += 2
		reasons = append(reasons, "FUD level: Moderate (4-6)")
	}

	switch activityAnalysis.Trend {
	case ActivityTrendSharpRise:
		longSignals++
		reasons = append(reasons, "Activity: Sharp rise detected")
	case ActivityTrendSharpDrop:
		shortSignals++
		reasons = append(reasons, "Activity: Sharp drop detected")
	}

	switch fudActivityAnalysis.Trend {
	case ActivityTrendSharpRise:
		shortSignals++
		reasons = append(reasons, "FUD Activity: Sharp rise detected")
	}

	switch sentimentAnalysis.Recommendation {
	case "bullish":
		longSignals += 2
		reasons = append(reasons, "Claude recommendation: Bullish")
	case "bearish":
		shortSignals += 2
		reasons = append(reasons, "Claude recommendation: Bearish")
	case "neutral":
		reasons = append(reasons, "Claude recommendation: Neutral")
	}

	if sentimentAnalysis.Confidence < 0.5 {
		reasons = append(reasons, "Low confidence in sentiment analysis (<0.5)")
		if longSignals > 0 {
			longSignals = (longSignals + 1) / 2
		}
		if shortSignals > 0 {
			shortSignals = (shortSignals + 1) / 2
		}
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
