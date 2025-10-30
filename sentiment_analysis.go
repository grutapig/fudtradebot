package main

import (
	"encoding/json"
	"fmt"
	"github.com/grutapig/fudtradebot/claude"
)

func AnalyzeCommunitysentiment(claudeClient claude.ClaudeApi, tweets []CommunityTweet, prompt string) (ClaudeSentimentResponse, error) {
	systemPrompt := `You are analyzing cryptocurrency community sentiment based on recent tweets. 

IMPORTANT CONTEXT:
- Negative messages are rare and get deleted quickly by moderators
- Most sentiment will be positive (2-10 range)
- Focus on SENTIMENT DYNAMICS rather than absolute values
- Even small drops in positivity can signal trouble

SENTIMENT SCALE (-10 to +10):
- Above 7: Excellent sentiment, strong bullish signal
- 2 to 7: Medium/neutral sentiment, need to check trend direction
- Below 2: Rare, very concerning (bad news or project issues)

Your task:
1. Overall sentiment from -10 to +10 (expect mostly 2-10 range)
2. Sentiment trend: "improving", "declining", or "stable" - THIS IS CRITICAL
   - "declining" = enthusiasm is fading, even if still positive (7→5 is declining)
   - "improving" = excitement is building
3. FUD level from 0 to 10 (fear, uncertainty, doubt)
4. Confidence in analysis from 0.0 to 1.0
5. Key discussion themes
6. Recommendation: "bullish", "bearish", or "neutral"
   - "bullish": sentiment > 7 OR (2-7 AND improving)
   - "bearish": declining trend OR high FUD (≥6)
   - "neutral": stable medium sentiment

Response must be STRICTLY in JSON format:
{
  "overall_sentiment": 6,
  "sentiment_trend": "declining",
  "fud_level": 3,
  "confidence": 0.85,
  "key_themes": ["price discussion", "less excitement than before"],
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
