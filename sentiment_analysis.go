package main

import (
	"encoding/json"
	"fmt"
	"github.com/grutapig/fudtradebot/claude"
)

func AnalyzeCommunitysentiment(claudeClient claude.ClaudeApi, tweets []CommunityTweet, prompt string) (ClaudeSentimentResponse, error) {
	systemPrompt := `You are analyzing community sentiment based on tweets. Your task is to determine:
1. Overall sentiment from -10 to +10
2. Sentiment trend: "improving", "declining", or "stable"
3. FUD level (fear, uncertainty, doubt) from 0 to 10
4. Confidence in analysis from 0.0 to 1.0
5. Key discussion themes
6. Recommendation: "bullish", "bearish", or "neutral"

Response must be STRICTLY in JSON format:
{
  "overall_sentiment": -5,
  "sentiment_trend": "declining",
  "fud_level": 8,
  "confidence": 0.85,
  "key_themes": ["price drop", "concern about project"],
  "recommendation": "bearish"
}`

	tweetsJSON, err := json.Marshal(tweets)
	if err != nil {
		return ClaudeSentimentResponse{}, fmt.Errorf("failed to marshal tweets: %w", err)
	}

	userMessage := fmt.Sprintf("%s\n\nTweets to analyze:\n%s", prompt, string(tweetsJSON))

	messages := claude.ClaudeMessages{
		{
			Role:    claude.ROLE_USER,
			Content: userMessage,
		},
		{
			Role:    claude.ROLE_ASSISTANT,
			Content: "{",
		},
	}

	response, err := claudeClient.SendMessage(messages, systemPrompt)
	if err != nil {
		return ClaudeSentimentResponse{}, fmt.Errorf("failed to send message to Claude: %w", err)
	}

	if len(response.Content) == 0 {
		return ClaudeSentimentResponse{}, fmt.Errorf("empty response from Claude")
	}

	var sentimentResponse ClaudeSentimentResponse
	if err := json.Unmarshal([]byte("{"+response.Content[0].Text), &sentimentResponse); err != nil {
		return ClaudeSentimentResponse{}, fmt.Errorf("failed to parse Claude response: %w", err)
	}

	return sentimentResponse, nil
}
