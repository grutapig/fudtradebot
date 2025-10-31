package main

type Signal string

const (
	SignalLong  Signal = "LONG"
	SignalShort Signal = "SHORT"
	SignalEmpty Signal = "EMPTY"
)

type TradingDecisionResult struct {
	Signal Signal
	Reason string
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

	signal := SignalEmpty
	reason := ""

	if btcSignal == SignalEmpty && coinSignal != SignalEmpty {
		signal = coinSignal
		reason = "ichimoku"
	} else if btcSignal != SignalEmpty && coinSignal != SignalEmpty {
		if btcSignal == coinSignal {
			signal = coinSignal
			reason = "ichimoku"
		} else {
			signal = SignalEmpty
		}
	}

	if signal == SignalEmpty {
		return TradingDecisionResult{SignalEmpty, ""}
	}

	activitySignal := convertActivityToSignal(activityAnalysis)

	if activitySignal == SignalEmpty {
	} else if signal == activitySignal {
		signal = activitySignal
		reason = "community"
	} else {
		signal = SignalEmpty
		reason = ""
	}

	if signal == SignalEmpty {
		return TradingDecisionResult{SignalEmpty, ""}
	}

	fudSignal := convertFudActivityToSignal(fudActivityAnalysis)

	if fudSignal == SignalEmpty {
	} else if signal == fudSignal {
		signal = fudSignal
		reason = "fud"
	} else {
		signal = SignalEmpty
		reason = ""
	}

	if signal == SignalEmpty {
		return TradingDecisionResult{SignalEmpty, ""}
	}

	sentimentSignal := convertSentimentToSignal(sentimentAnalysis)

	if sentimentSignal == SignalEmpty {
	} else if signal == sentimentSignal {
		signal = sentimentSignal
		reason = "sentiment"
	} else {
		signal = SignalEmpty
		reason = ""
	}

	return TradingDecisionResult{signal, reason}
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
