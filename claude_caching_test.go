package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClaudeCachingLogic_FirstRun(t *testing.T) {
	state := TradingState{
		CurrentPosition: PositionSideBoth,
	}

	tweets := []CommunityTweet{
		{ID: "tweet1", Text: "Test tweet", Date: time.Now()},
	}

	minIntervalMinutes := 10

	timeSinceLastAnalysis := time.Since(state.LastClaudeAnalysis)
	minInterval := time.Duration(minIntervalMinutes) * time.Minute
	hasNewTweets := len(tweets) > 0 && (state.LastAnalyzedTweetID == "" || tweets[0].ID != state.LastAnalyzedTweetID)
	shouldAnalyze := state.LastClaudeAnalysis.IsZero() || (timeSinceLastAnalysis >= minInterval && hasNewTweets)

	assert.True(t, shouldAnalyze, "Should analyze on first run")
}

func TestClaudeCachingLogic_IntervalNotReached(t *testing.T) {
	state := TradingState{
		CurrentPosition:    PositionSideBoth,
		LastClaudeAnalysis: time.Now().Add(-5 * time.Minute),
		LastSentimentAnalysis: ClaudeSentimentResponse{
			OverallSentiment: 5,
			Confidence:       0.8,
		},
		LastAnalyzedTweetID: "tweet1",
	}

	tweets := []CommunityTweet{
		{ID: "tweet2", Text: "New tweet", Date: time.Now()},
	}

	minIntervalMinutes := 10

	timeSinceLastAnalysis := time.Since(state.LastClaudeAnalysis)
	minInterval := time.Duration(minIntervalMinutes) * time.Minute
	hasNewTweets := len(tweets) > 0 && (state.LastAnalyzedTweetID == "" || tweets[0].ID != state.LastAnalyzedTweetID)
	shouldAnalyze := state.LastClaudeAnalysis.IsZero() || (timeSinceLastAnalysis >= minInterval && hasNewTweets)

	assert.False(t, shouldAnalyze, "Should NOT analyze when interval not reached")
	assert.True(t, timeSinceLastAnalysis < minInterval, "Time should be less than min interval")
}

func TestClaudeCachingLogic_IntervalReachedWithNewTweets(t *testing.T) {
	state := TradingState{
		CurrentPosition:    PositionSideBoth,
		LastClaudeAnalysis: time.Now().Add(-15 * time.Minute),
		LastSentimentAnalysis: ClaudeSentimentResponse{
			OverallSentiment: 5,
			Confidence:       0.8,
		},
		LastAnalyzedTweetID: "tweet1",
	}

	tweets := []CommunityTweet{
		{ID: "tweet2", Text: "New tweet", Date: time.Now()},
	}

	minIntervalMinutes := 10

	timeSinceLastAnalysis := time.Since(state.LastClaudeAnalysis)
	minInterval := time.Duration(minIntervalMinutes) * time.Minute
	hasNewTweets := len(tweets) > 0 && (state.LastAnalyzedTweetID == "" || tweets[0].ID != state.LastAnalyzedTweetID)
	shouldAnalyze := state.LastClaudeAnalysis.IsZero() || (timeSinceLastAnalysis >= minInterval && hasNewTweets)

	assert.True(t, shouldAnalyze, "Should analyze when interval reached AND new tweets exist")
	assert.True(t, timeSinceLastAnalysis >= minInterval, "Time should be >= min interval")
	assert.True(t, hasNewTweets, "Should have new tweets")
}

func TestClaudeCachingLogic_IntervalReachedButNoNewTweets(t *testing.T) {
	state := TradingState{
		CurrentPosition:    PositionSideBoth,
		LastClaudeAnalysis: time.Now().Add(-15 * time.Minute),
		LastSentimentAnalysis: ClaudeSentimentResponse{
			OverallSentiment: 5,
			Confidence:       0.8,
		},
		LastAnalyzedTweetID: "tweet1",
	}

	tweets := []CommunityTweet{
		{ID: "tweet1", Text: "Same tweet", Date: time.Now()},
	}

	minIntervalMinutes := 10

	timeSinceLastAnalysis := time.Since(state.LastClaudeAnalysis)
	minInterval := time.Duration(minIntervalMinutes) * time.Minute
	hasNewTweets := len(tweets) > 0 && (state.LastAnalyzedTweetID == "" || tweets[0].ID != state.LastAnalyzedTweetID)
	shouldAnalyze := state.LastClaudeAnalysis.IsZero() || (timeSinceLastAnalysis >= minInterval && hasNewTweets)

	assert.False(t, shouldAnalyze, "Should NOT analyze when no new tweets even if interval reached")
	assert.True(t, timeSinceLastAnalysis >= minInterval, "Time should be >= min interval")
	assert.False(t, hasNewTweets, "Should NOT have new tweets")
}

func TestClaudeCachingLogic_EmptyTweets(t *testing.T) {
	state := TradingState{
		CurrentPosition:    PositionSideBoth,
		LastClaudeAnalysis: time.Now().Add(-15 * time.Minute),
		LastSentimentAnalysis: ClaudeSentimentResponse{
			OverallSentiment: 5,
			Confidence:       0.8,
		},
		LastAnalyzedTweetID: "tweet1",
	}

	tweets := []CommunityTweet{}

	minIntervalMinutes := 10

	timeSinceLastAnalysis := time.Since(state.LastClaudeAnalysis)
	minInterval := time.Duration(minIntervalMinutes) * time.Minute
	hasNewTweets := len(tweets) > 0 && (state.LastAnalyzedTweetID == "" || tweets[0].ID != state.LastAnalyzedTweetID)
	shouldAnalyze := state.LastClaudeAnalysis.IsZero() || (timeSinceLastAnalysis >= minInterval && hasNewTweets)

	assert.False(t, shouldAnalyze, "Should NOT analyze when tweets are empty")
	assert.False(t, hasNewTweets, "Should NOT have new tweets when array is empty")
}

func TestClaudeCachingLogic_CachedAnalysisAvailable(t *testing.T) {
	state := TradingState{
		CurrentPosition:    PositionSideBoth,
		LastClaudeAnalysis: time.Now().Add(-5 * time.Minute),
		LastSentimentAnalysis: ClaudeSentimentResponse{
			OverallSentiment: 5,
			SentimentTrend:   "stable",
			FudLevel:         3,
			Confidence:       0.8,
			KeyThemes:        []string{"test"},
			Recommendation:   "neutral",
		},
		LastAnalyzedTweetID: "tweet1",
	}

	hasCachedAnalysis := state.LastSentimentAnalysis.Confidence != 0

	assert.True(t, hasCachedAnalysis, "Should have cached analysis available")
	assert.Equal(t, 0.8, state.LastSentimentAnalysis.Confidence)
}

func TestClaudeCachingLogic_NoCachedAnalysis(t *testing.T) {
	state := TradingState{
		CurrentPosition: PositionSideBoth,
	}

	hasCachedAnalysis := state.LastSentimentAnalysis.Confidence != 0

	assert.False(t, hasCachedAnalysis, "Should NOT have cached analysis on fresh state")
}

func TestClaudeCachingLogic_StateUpdateAfterAnalysis(t *testing.T) {
	state := TradingState{
		CurrentPosition: PositionSideBoth,
	}

	tweets := []CommunityTweet{
		{ID: "tweet1", Text: "Test tweet", Date: time.Now()},
		{ID: "tweet2", Text: "Test tweet 2", Date: time.Now()},
	}

	sentimentAnalysis := ClaudeSentimentResponse{
		OverallSentiment: 7,
		SentimentTrend:   "improving",
		FudLevel:         2,
		Confidence:       0.9,
		KeyThemes:        []string{"positive"},
		Recommendation:   "bullish",
	}

	beforeAnalysis := time.Now()
	state.LastClaudeAnalysis = time.Now()
	state.LastSentimentAnalysis = sentimentAnalysis
	if len(tweets) > 0 {
		state.LastAnalyzedTweetID = tweets[0].ID
	}
	state.LastTweetsCount = len(tweets)

	assert.False(t, state.LastClaudeAnalysis.Before(beforeAnalysis), "LastClaudeAnalysis should be updated")
	assert.Equal(t, sentimentAnalysis, state.LastSentimentAnalysis)
	assert.Equal(t, "tweet1", state.LastAnalyzedTweetID)
	assert.Equal(t, 2, state.LastTweetsCount)
}

func TestClaudeCachingLogic_DifferentIntervals(t *testing.T) {
	tests := []struct {
		name                  string
		minIntervalMinutes    int
		timeSinceAnalysis     time.Duration
		hasNewTweets          bool
		expectedShouldAnalyze bool
	}{
		{
			name:                  "5 min interval, 3 min passed, new tweets",
			minIntervalMinutes:    5,
			timeSinceAnalysis:     3 * time.Minute,
			hasNewTweets:          true,
			expectedShouldAnalyze: false,
		},
		{
			name:                  "5 min interval, 6 min passed, new tweets",
			minIntervalMinutes:    5,
			timeSinceAnalysis:     6 * time.Minute,
			hasNewTweets:          true,
			expectedShouldAnalyze: true,
		},
		{
			name:                  "15 min interval, 10 min passed, new tweets",
			minIntervalMinutes:    15,
			timeSinceAnalysis:     10 * time.Minute,
			hasNewTweets:          true,
			expectedShouldAnalyze: false,
		},
		{
			name:                  "15 min interval, 20 min passed, no new tweets",
			minIntervalMinutes:    15,
			timeSinceAnalysis:     20 * time.Minute,
			hasNewTweets:          false,
			expectedShouldAnalyze: false,
		},
		{
			name:                  "10 min interval, 10 min passed exactly, new tweets",
			minIntervalMinutes:    10,
			timeSinceAnalysis:     10 * time.Minute,
			hasNewTweets:          true,
			expectedShouldAnalyze: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := TradingState{
				CurrentPosition:     PositionSideBoth,
				LastClaudeAnalysis:  time.Now().Add(-tt.timeSinceAnalysis),
				LastAnalyzedTweetID: "tweet1",
			}

			minInterval := time.Duration(tt.minIntervalMinutes) * time.Minute
			timeSinceLastAnalysis := time.Since(state.LastClaudeAnalysis)
			shouldAnalyze := timeSinceLastAnalysis >= minInterval && tt.hasNewTweets

			assert.Equal(t, tt.expectedShouldAnalyze, shouldAnalyze,
				"Expected shouldAnalyze=%v for interval=%d min, elapsed=%.1f min, hasNewTweets=%v",
				tt.expectedShouldAnalyze, tt.minIntervalMinutes, timeSinceLastAnalysis.Minutes(), tt.hasNewTweets)
		})
	}
}
