package main

type Signal string

const (
	SignalLong  Signal = "LONG"
	SignalShort Signal = "SHORT"
	SignalEmpty Signal = "EMPTY"
)

type TradingDecisionResult struct {
	Signal             Signal
	Reason             string
	Explanation        string
	BTCIchimokuSignal  string
	CoinIchimokuSignal string
	ActivitySignal     string
	FudActivitySignal  string
	SentimentSignal    string
}

func MakeTradingDecision(
	btcIchimoku IchimokuAnalysis,
	coinIchimoku IchimokuAnalysis,
	activityAnalysis ActivityAnalysis,
	fudActivityAnalysis ActivityAnalysis,
	sentimentAnalysis ClaudeSentimentResponse,
) TradingDecisionResult {

	btcSignal := convertIchimokuToSignal(btcIchimoku)
	coinSignal := convertIchimokuToSignal(coinIchimoku)

	result := TradingDecisionResult{
		BTCIchimokuSignal:  string(btcSignal),
		CoinIchimokuSignal: string(coinSignal),
	}

	signal := SignalEmpty
	reason := ""
	explanation := ""

	if btcSignal == SignalEmpty && coinSignal != SignalEmpty {
		signal = coinSignal
		reason = "ichimoku"
		explanation = "BTC Ichimoku neutral, Coin Ichimoku " + string(coinSignal)
	} else if btcSignal != SignalEmpty && coinSignal != SignalEmpty {
		if btcSignal == coinSignal {
			signal = coinSignal
			reason = "ichimoku"
			explanation = "BTC and Coin Ichimoku aligned: " + string(coinSignal)
		} else {
			signal = SignalEmpty
			explanation = "BTC Ichimoku " + string(btcSignal) + " contradicts Coin Ichimoku " + string(coinSignal)
		}
	} else {
		explanation = "Both BTC and Coin Ichimoku neutral"
	}

	if signal == SignalEmpty {
		result.Signal = SignalEmpty
		result.Reason = ""
		result.Explanation = explanation
		result.ActivitySignal = string(convertActivityToSignal(activityAnalysis))
		result.FudActivitySignal = string(convertFudActivityToSignal(fudActivityAnalysis))
		result.SentimentSignal = string(convertSentimentToSignal(sentimentAnalysis))
		return result
	}

	activitySignal := convertActivityToSignal(activityAnalysis)
	result.ActivitySignal = string(activitySignal)

	if activitySignal == SignalEmpty {
		explanation += ". Activity neutral"
	} else if signal == activitySignal {
		signal = activitySignal
		reason = "community"
		explanation += ". Activity confirms: " + string(activitySignal)
	} else {
		signal = SignalEmpty
		reason = ""
		explanation += ". Activity " + string(activitySignal) + " contradicts signal"
	}

	if signal == SignalEmpty {
		result.Signal = SignalEmpty
		result.Reason = ""
		result.Explanation = explanation
		result.FudActivitySignal = string(convertFudActivityToSignal(fudActivityAnalysis))
		result.SentimentSignal = string(convertSentimentToSignal(sentimentAnalysis))
		return result
	}

	fudSignal := convertFudActivityToSignal(fudActivityAnalysis)
	result.FudActivitySignal = string(fudSignal)

	if fudSignal == SignalEmpty {
		explanation += ". FUD activity neutral"
	} else if signal == fudSignal {
		signal = fudSignal
		reason = "fud"
		explanation += ". FUD activity confirms: " + string(fudSignal)
	} else {
		signal = SignalEmpty
		reason = ""
		explanation += ". FUD activity " + string(fudSignal) + " contradicts signal"
	}

	if signal == SignalEmpty {
		result.Signal = SignalEmpty
		result.Reason = ""
		result.Explanation = explanation
		result.SentimentSignal = string(convertSentimentToSignal(sentimentAnalysis))
		return result
	}

	sentimentSignal := convertSentimentToSignal(sentimentAnalysis)
	result.SentimentSignal = string(sentimentSignal)

	if sentimentSignal == SignalEmpty {
		explanation += ". Sentiment neutral"
	} else if signal == sentimentSignal {
		signal = sentimentSignal
		reason = "sentiment"
		explanation += ". Sentiment confirms: " + string(sentimentSignal)
	} else {
		signal = SignalEmpty
		reason = ""
		explanation += ". Sentiment " + string(sentimentSignal) + " contradicts signal"
	}

	result.Signal = signal
	result.Reason = reason
	result.Explanation = explanation
	return result
}

func convertIchimokuToSignal(analysis IchimokuAnalysis) Signal {
	switch analysis.Signal {
	case IchimokuSignalStrongLong, IchimokuSignalLong:
		return SignalLong
	case IchimokuSignalStrongShort, IchimokuSignalShort:
		return SignalShort
	default:
		return SignalEmpty
	}
}

func convertActivityToSignal(analysis ActivityAnalysis) Signal {
	switch analysis.Trend {
	case ActivityTrendSharpRise:
		return SignalLong
	case ActivityTrendSharpDrop:
		return SignalShort
	default:
		return SignalEmpty
	}
}

func convertFudActivityToSignal(analysis ActivityAnalysis) Signal {
	if analysis.Trend == ActivityTrendSharpRise {
		return SignalShort
	}
	return SignalEmpty
}

func convertSentimentToSignal(analysis ClaudeSentimentResponse) Signal {
	if analysis.SentimentTrend == "declining" && analysis.OverallSentiment < 3 {
		return SignalShort
	}
	return SignalEmpty
}
