package main

func CheckSentimentMatch(action TradingAction, sentiment ClaudeSentimentResponse) bool {
	if sentiment.Confidence < 0.5 {
		return true
	}

	isSentimentGood := sentiment.OverallSentiment > 7
	isSentimentMedium := sentiment.OverallSentiment >= 2 && sentiment.OverallSentiment <= 7
	isSentimentDeclining := sentiment.SentimentTrend == "declining"
	isSentimentImproving := sentiment.SentimentTrend == "improving"

	switch action {
	case TradingActionOpenLong:
		if isSentimentGood || (isSentimentMedium && isSentimentImproving) {
			return true
		}
		return false

	case TradingActionOpenShort:
		if isSentimentDeclining || sentiment.FudLevel >= 6 {
			return true
		}
		return false

	case TradingActionCloseLong:
		if isSentimentDeclining || (!isSentimentGood && !isSentimentImproving) {
			return true
		}
		return false

	case TradingActionCloseShort:
		if isSentimentGood || isSentimentImproving {
			return true
		}
		return false

	case TradingActionHold:
		return true

	default:
		return true
	}
}
