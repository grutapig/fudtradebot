package main

import (
	"encoding/json"
	"fmt"
	"github.com/grutapig/fudtradebot/claude"
	"time"
)

func AnalyzePositionClose(claudeClient claude.ClaudeApi, position PositionRecord, snapshots []PositionSnapshot, recentTweets []CommunityTweet, btcIchimoku IchimokuAnalysis, coinIchimoku IchimokuAnalysis, ichimoku ClosePositionReason) (ClaudePositionCloseResponse, error) {
	systemPrompt := `You are a cryptocurrency trading assistant analyzing whether to close an open position.

You will receive:
1. Current open position details (side, entry price, current P/L, max/min P/L)
2. Position snapshots history showing P/L evolution
3. Last 50 community messages
4. BTC Ichimoku analysis
5. Coin Ichimoku analysis

Your task is to analyze:
- Historical P/L movements and their ranges
- Community sentiment trends from messages
- How BTC signals might affect this coin
- Current position performance vs historical patterns
- Risk of reversal vs potential for more profit

Be careful more attention what exactly side for our current position, if it is short and price go down and all indicators for it,  we should hold this position open.
Same for LONG position, is its long, and price grow up, and all indicators for it, we should continue hold position don't close.
Only if contradicts all signals with current side position we should close it. But don't forget  to try maximize our profit you should also use snapshot information of history.
Also consider the position opening date and how much time has passed, we use candles with 1-hour interval for the coin and 4-hour interval for Bitcoin.
Also consider the analysis of whether to close based on the Ichimoku cloud.

Response must be STRICTLY in JSON format:
{
  "should_close": true|false,
  "confidence_percent": 75.5,
  "justification": "Position has reached 85% of historical max P/L. Community sentiment shows early reversal signs. BTC showing bearish divergence.",
  "expected_pnl": 12.45,
  "risk_assessment": "medium-high"
}`

	type CloseAnalysisRequest struct {
		Position                   PositionRecord      `json:"position"`
		Snapshots                  []PositionSnapshot  `json:"snapshots"`
		Tweets                     []CommunityTweet    `json:"recent_tweets"`
		BTCIchimoku                IchimokuAnalysis    `json:"btc_ichimoku"`
		CoinIchimoku               IchimokuAnalysis    `json:"coin_ichimoku"`
		CurrentPositionShortOrLong string              `json:"current_position_short_or_long"`
		ShouldCloseByIchimoku      ClosePositionReason `json:"should_close_by_ichimoku"`
		CurrentDate                string              `json:"current_date"`
		PositionOpenDate           string              `json:"position_open_date"`
	}

	requestData := CloseAnalysisRequest{
		CurrentPositionShortOrLong: position.Side,
		Position:                   position,
		Snapshots:                  snapshots,
		Tweets:                     recentTweets,
		BTCIchimoku:                btcIchimoku,
		CoinIchimoku:               coinIchimoku,
		ShouldCloseByIchimoku:      ichimoku,
		CurrentDate:                time.Now().Format(time.RFC3339),
		PositionOpenDate:           position.OpenedAt.Format(time.RFC3339),
	}

	requestJSON, err := json.Marshal(requestData)
	if err != nil {
		return ClaudePositionCloseResponse{}, fmt.Errorf("failed to marshal request data: %w", err)
	}

	userMessage := fmt.Sprintf("Should we close this position? Analyze the data:\n\n%s", string(requestJSON))

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
		return ClaudePositionCloseResponse{}, fmt.Errorf("failed to send message to Claude: %w", err)
	}

	if len(response.Content) == 0 {
		return ClaudePositionCloseResponse{}, fmt.Errorf("empty response from Claude")
	}

	var closeResponse ClaudePositionCloseResponse
	if err := json.Unmarshal([]byte("{"+response.Content[0].Text), &closeResponse); err != nil {
		return ClaudePositionCloseResponse{}, fmt.Errorf("failed to parse Claude response: %w", err)
	}

	return closeResponse, nil
}
