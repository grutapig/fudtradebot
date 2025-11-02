package main

import (
	"encoding/json"
	"fmt"
	"github.com/grutapig/fudtradebot/claude"
)

func ValidateOrderWithAI(claudeClient claude.ClaudeApi, decision TradingDecisionResult, btcIchimoku IchimokuAnalysis, coinIchimoku IchimokuAnalysis, activityAnalysis ActivityAnalysis, fudActivityAnalysis ActivityAnalysis, sentimentAnalysis ClaudeSentimentResponse) (ClaudeOrderValidationResponse, error) {
	systemPrompt := `You are a cryptocurrency trading assistant. Your task is to validate whether a trading decision should be executed based on the provided market data and technical analysis.

You will receive:
1. Trading decision from the automated system (LONG/SHORT)
2. BTC Ichimoku analysis
3. Coin Ichimoku analysis
4. Community activity trend
5. FUD activity trend
6. Sentiment analysis

Your task is to evaluate all this data and decide:
- Should we open the order? (true/false)
- Confidence percentage (0-100)
- Justification for your decision

Consider:
- Are all indicators aligned?
- Is there conflicting data?
- Are market conditions favorable?
- Are there any red flags in sentiment or FUD activity?
- Is the timing appropriate?

Response must be STRICTLY in JSON format:
{
  "should_open_order": true,
  "confidence_percent": 75.5,
  "justification": "All indicators are aligned for a LONG position. BTC and coin Ichimoku both show bullish signals, community activity is rising, and sentiment is positive."
}`

	type ValidationRequest struct {
		Decision     TradingDecisionResult   `json:"decision"`
		BTCIchimoku  IchimokuAnalysis        `json:"btc_ichimoku"`
		CoinIchimoku IchimokuAnalysis        `json:"coin_ichimoku"`
		Activity     ActivityAnalysis        `json:"activity"`
		FudActivity  ActivityAnalysis        `json:"fud_activity"`
		Sentiment    ClaudeSentimentResponse `json:"sentiment"`
	}

	requestData := ValidationRequest{
		Decision:     decision,
		BTCIchimoku:  btcIchimoku,
		CoinIchimoku: coinIchimoku,
		Activity:     activityAnalysis,
		FudActivity:  fudActivityAnalysis,
		Sentiment:    sentimentAnalysis,
	}

	requestJSON, err := json.Marshal(requestData)
	if err != nil {
		return ClaudeOrderValidationResponse{}, fmt.Errorf("failed to marshal request data: %w", err)
	}

	userMessage := fmt.Sprintf("Please validate this trading decision:\n\n%s", string(requestJSON))

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
		return ClaudeOrderValidationResponse{}, fmt.Errorf("failed to send message to Claude: %w", err)
	}

	if len(response.Content) == 0 {
		return ClaudeOrderValidationResponse{}, fmt.Errorf("empty response from Claude")
	}

	var validationResponse ClaudeOrderValidationResponse
	if err := json.Unmarshal([]byte("{"+response.Content[0].Text), &validationResponse); err != nil {
		return ClaudeOrderValidationResponse{}, fmt.Errorf("failed to parse Claude response: %w", err)
	}

	return validationResponse, nil
}
