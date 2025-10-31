package main

type Signal string

const (
	SignalLong  Signal = "LONG"
	SignalShort Signal = "SHORT"
	SignalEmpty Signal = "EMPTY"
)

func MakeTradingDecision(
	btcIchimoku IchimokuAnalysis,
	coinIchimoku IchimokuAnalysis,
	activityAnalysis ActivityAnalysis,
	fudActivityAnalysis ActivityAnalysis,
	sentimentAnalysis ClaudeSentimentResponse,
) Signal {

	btcSignal := convertIchimokuToSignal(btcIchimoku)
	coinSignal := convertIchimokuToSignal(coinIchimoku)

	signal := SignalEmpty

	if btcSignal == SignalEmpty && coinSignal != SignalEmpty {
		signal = coinSignal
	} else if btcSignal != SignalEmpty && coinSignal != SignalEmpty {
		if btcSignal == coinSignal {
			signal = coinSignal
		} else {
			signal = SignalEmpty
		}
	}

	if signal == SignalEmpty {
		return SignalEmpty
	}

	activitySignal := convertActivityToSignal(activityAnalysis)

	if activitySignal == SignalEmpty {
	} else if signal == activitySignal {
		signal = activitySignal
	} else {
		signal = SignalEmpty
	}

	if signal == SignalEmpty {
		return SignalEmpty
	}

	fudSignal := convertFudActivityToSignal(fudActivityAnalysis)

	if fudSignal == SignalEmpty {
	} else if signal == fudSignal {
		signal = fudSignal
	} else {
		signal = SignalEmpty
	}

	if signal == SignalEmpty {
		return SignalEmpty
	}

	sentimentSignal := convertSentimentToSignal(sentimentAnalysis)

	if sentimentSignal == SignalEmpty {
	} else if signal == sentimentSignal {
		signal = sentimentSignal
	} else {
		signal = SignalEmpty
	}

	return signal
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
