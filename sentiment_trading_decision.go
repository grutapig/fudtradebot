package main

func MakeSentimentTradingDecision(
	sentimentAnalysis ClaudeSentimentResponse,
	currentPosition PositionSide,
) TradingSignal {
	reasons := []string{}
	var action TradingAction
	var strength SignalStrength

	isSentimentGood := sentimentAnalysis.OverallSentiment > 7
	isSentimentMedium := sentimentAnalysis.OverallSentiment >= 2 && sentimentAnalysis.OverallSentiment <= 7
	isSentimentDeclining := sentimentAnalysis.SentimentTrend == "declining"
	isSentimentImproving := sentimentAnalysis.SentimentTrend == "improving"
	highFud := sentimentAnalysis.FudLevel >= 6

	reasons = append(reasons, "Sentiment-based analysis:")
	if isSentimentGood {
		reasons = append(reasons, "Sentiment: Excellent (>7)")
	} else if isSentimentMedium {
		reasons = append(reasons, "Sentiment: Medium (2-7)")
	} else {
		reasons = append(reasons, "Sentiment: Poor (<2)")
	}

	if isSentimentImproving {
		reasons = append(reasons, "Trend: Improving")
	} else if isSentimentDeclining {
		reasons = append(reasons, "Trend: Declining")
	} else {
		reasons = append(reasons, "Trend: Stable")
	}

	if highFud {
		reasons = append(reasons, "High FUD detected (≥6)")
	}

	if isSentimentGood || (isSentimentMedium && isSentimentImproving) {
		if currentPosition == PositionSideShort {
			action = TradingActionCloseShort
			strength = SignalStrengthMedium
			reasons = append(reasons, "Closing SHORT: Positive sentiment")
		} else if currentPosition == PositionSideBoth {
			action = TradingActionOpenLong
			if isSentimentGood {
				strength = SignalStrengthStrong
			} else {
				strength = SignalStrengthMedium
			}
			reasons = append(reasons, "Opening LONG: Positive sentiment")
		} else {
			action = TradingActionHold
			strength = SignalStrengthNone
			reasons = append(reasons, "Already in LONG")
		}
	} else if isSentimentDeclining || highFud {
		if currentPosition == PositionSideLong {
			action = TradingActionCloseLong
			strength = SignalStrengthMedium
			reasons = append(reasons, "Closing LONG: Negative sentiment/FUD")
		} else if currentPosition == PositionSideBoth {
			action = TradingActionOpenShort
			if highFud {
				strength = SignalStrengthStrong
			} else {
				strength = SignalStrengthMedium
			}
			reasons = append(reasons, "Opening SHORT: Negative sentiment/FUD")
		} else {
			action = TradingActionHold
			strength = SignalStrengthNone
			reasons = append(reasons, "Already in SHORT")
		}
	} else {
		action = TradingActionHold
		strength = SignalStrengthNone
		reasons = append(reasons, "No clear sentiment signal")
	}

	if sentimentAnalysis.Confidence < 0.5 {
		strength = SignalStrengthWeak
		reasons = append(reasons, "Low confidence - weakening signal")
	}

	return TradingSignal{
		Action:   action,
		Strength: strength,
		Reasons:  reasons,
	}
}
