package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeSmartTradingDecision_PlateauClosesPosition(t *testing.T) {
	activityAnalysis := ActivityAnalysis{
		Trend:              ActivityTrendPlateau,
		AverageCount:       100,
		RecentAverageCount: 100,
		ChangePercent:      0,
	}

	fudActivityAnalysis := ActivityAnalysis{
		Trend:              ActivityTrendPlateau,
		AverageCount:       50,
		RecentAverageCount: 50,
		ChangePercent:      0,
	}

	sentimentAnalysis := ClaudeSentimentResponse{
		OverallSentiment: 0,
		SentimentTrend:   "stable",
		FudLevel:         3,
		Confidence:       0.7,
		KeyThemes:        []string{"neutral"},
		Recommendation:   "neutral",
	}

	signal := MakeSmartTradingDecision(
		activityAnalysis,
		fudActivityAnalysis,
		sentimentAnalysis,
		PositionSideLong,
	)

	assert.Equal(t, TradingActionCloseLong, signal.Action)
	assert.Equal(t, SignalStrengthMedium, signal.Strength)
	assert.Contains(t, signal.Reasons, "Plateau detected - closing LONG position")
}

func TestMakeSmartTradingDecision_PlateauClosesShortPosition(t *testing.T) {
	activityAnalysis := ActivityAnalysis{
		Trend:              ActivityTrendPlateau,
		AverageCount:       100,
		RecentAverageCount: 100,
		ChangePercent:      0,
	}

	fudActivityAnalysis := ActivityAnalysis{
		Trend:              ActivityTrendPlateau,
		AverageCount:       50,
		RecentAverageCount: 50,
		ChangePercent:      0,
	}

	sentimentAnalysis := ClaudeSentimentResponse{
		OverallSentiment: 0,
		SentimentTrend:   "stable",
		FudLevel:         3,
		Confidence:       0.7,
		KeyThemes:        []string{"neutral"},
		Recommendation:   "neutral",
	}

	signal := MakeSmartTradingDecision(
		activityAnalysis,
		fudActivityAnalysis,
		sentimentAnalysis,
		PositionSideShort,
	)

	assert.Equal(t, TradingActionCloseShort, signal.Action)
	assert.Equal(t, SignalStrengthMedium, signal.Strength)
	assert.Contains(t, signal.Reasons, "Plateau detected - closing SHORT position")
}

func TestMakeSmartTradingDecision_BullishSignal(t *testing.T) {
	activityAnalysis := ActivityAnalysis{
		Trend:              ActivityTrendSharpRise,
		AverageCount:       100,
		RecentAverageCount: 150,
		ChangePercent:      50,
	}

	fudActivityAnalysis := ActivityAnalysis{
		Trend:              ActivityTrendPlateau,
		AverageCount:       50,
		RecentAverageCount: 50,
		ChangePercent:      0,
	}

	sentimentAnalysis := ClaudeSentimentResponse{
		OverallSentiment: 7,
		SentimentTrend:   "improving",
		FudLevel:         2,
		Confidence:       0.85,
		KeyThemes:        []string{"positive outlook", "growth"},
		Recommendation:   "bullish",
	}

	signal := MakeSmartTradingDecision(
		activityAnalysis,
		fudActivityAnalysis,
		sentimentAnalysis,
		PositionSideBoth,
	)

	assert.Equal(t, TradingActionOpenLong, signal.Action)
	assert.Equal(t, SignalStrengthStrong, signal.Strength)
}

func TestMakeSmartTradingDecision_BearishSignal(t *testing.T) {
	activityAnalysis := ActivityAnalysis{
		Trend:              ActivityTrendSharpDrop,
		AverageCount:       100,
		RecentAverageCount: 50,
		ChangePercent:      -50,
	}

	fudActivityAnalysis := ActivityAnalysis{
		Trend:              ActivityTrendSharpRise,
		AverageCount:       50,
		RecentAverageCount: 100,
		ChangePercent:      100,
	}

	sentimentAnalysis := ClaudeSentimentResponse{
		OverallSentiment: -7,
		SentimentTrend:   "declining",
		FudLevel:         8,
		Confidence:       0.9,
		KeyThemes:        []string{"fear", "uncertainty"},
		Recommendation:   "bearish",
	}

	signal := MakeSmartTradingDecision(
		activityAnalysis,
		fudActivityAnalysis,
		sentimentAnalysis,
		PositionSideBoth,
	)

	assert.Equal(t, TradingActionOpenShort, signal.Action)
	assert.Equal(t, SignalStrengthStrong, signal.Strength)
}

func TestMakeSmartTradingDecision_CloseLongOnBearish(t *testing.T) {
	activityAnalysis := ActivityAnalysis{
		Trend:              ActivityTrendSharpDrop,
		AverageCount:       100,
		RecentAverageCount: 50,
		ChangePercent:      -50,
	}

	fudActivityAnalysis := ActivityAnalysis{
		Trend:              ActivityTrendSharpRise,
		AverageCount:       50,
		RecentAverageCount: 100,
		ChangePercent:      100,
	}

	sentimentAnalysis := ClaudeSentimentResponse{
		OverallSentiment: -6,
		SentimentTrend:   "declining",
		FudLevel:         8,
		Confidence:       0.85,
		KeyThemes:        []string{"panic", "sell-off"},
		Recommendation:   "bearish",
	}

	signal := MakeSmartTradingDecision(
		activityAnalysis,
		fudActivityAnalysis,
		sentimentAnalysis,
		PositionSideLong,
	)

	assert.Equal(t, TradingActionCloseLong, signal.Action)
	assert.Equal(t, SignalStrengthStrong, signal.Strength)
}

func TestMakeSmartTradingDecision_CloseShortOnBullish(t *testing.T) {
	activityAnalysis := ActivityAnalysis{
		Trend:              ActivityTrendSharpRise,
		AverageCount:       100,
		RecentAverageCount: 200,
		ChangePercent:      100,
	}

	fudActivityAnalysis := ActivityAnalysis{
		Trend:              ActivityTrendPlateau,
		AverageCount:       50,
		RecentAverageCount: 50,
		ChangePercent:      0,
	}

	sentimentAnalysis := ClaudeSentimentResponse{
		OverallSentiment: 8,
		SentimentTrend:   "improving",
		FudLevel:         1,
		Confidence:       0.9,
		KeyThemes:        []string{"bullish", "rally"},
		Recommendation:   "bullish",
	}

	signal := MakeSmartTradingDecision(
		activityAnalysis,
		fudActivityAnalysis,
		sentimentAnalysis,
		PositionSideShort,
	)

	assert.Equal(t, TradingActionCloseShort, signal.Action)
	assert.Equal(t, SignalStrengthStrong, signal.Strength)
}

func TestMakeSmartTradingDecision_LowConfidence(t *testing.T) {
	activityAnalysis := ActivityAnalysis{
		Trend:              ActivityTrendSharpRise,
		AverageCount:       100,
		RecentAverageCount: 150,
		ChangePercent:      50,
	}

	fudActivityAnalysis := ActivityAnalysis{
		Trend:              ActivityTrendPlateau,
		AverageCount:       50,
		RecentAverageCount: 50,
		ChangePercent:      0,
	}

	sentimentAnalysis := ClaudeSentimentResponse{
		OverallSentiment: 6,
		SentimentTrend:   "improving",
		FudLevel:         2,
		Confidence:       0.3,
		KeyThemes:        []string{"uncertain"},
		Recommendation:   "bullish",
	}

	signal := MakeSmartTradingDecision(
		activityAnalysis,
		fudActivityAnalysis,
		sentimentAnalysis,
		PositionSideBoth,
	)

	assert.Contains(t, signal.Reasons, "Low confidence in sentiment analysis (<0.5)")
}

func TestMakeSmartTradingDecision_HoldWhenAlreadyInPosition(t *testing.T) {
	activityAnalysis := ActivityAnalysis{
		Trend:              ActivityTrendSharpRise,
		AverageCount:       100,
		RecentAverageCount: 150,
		ChangePercent:      50,
	}

	fudActivityAnalysis := ActivityAnalysis{
		Trend:              ActivityTrendPlateau,
		AverageCount:       50,
		RecentAverageCount: 50,
		ChangePercent:      0,
	}

	sentimentAnalysis := ClaudeSentimentResponse{
		OverallSentiment: 6,
		SentimentTrend:   "improving",
		FudLevel:         2,
		Confidence:       0.8,
		KeyThemes:        []string{"positive"},
		Recommendation:   "bullish",
	}

	signal := MakeSmartTradingDecision(
		activityAnalysis,
		fudActivityAnalysis,
		sentimentAnalysis,
		PositionSideLong,
	)

	assert.Equal(t, TradingActionHold, signal.Action)
	assert.Equal(t, SignalStrengthNone, signal.Strength)
	assert.Contains(t, signal.Reasons, "Already in LONG position")
}
