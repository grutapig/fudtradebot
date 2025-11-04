package main

import (
	"encoding/json"
	"fmt"
	"github.com/grutapig/fudtradebot/claude"
	"math"
	"sort"
	"time"
)

type SnapshotStatistics struct {
	TotalCount    int     `json:"total_count"`
	MinPnL        float64 `json:"min_pnl"`
	MaxPnL        float64 `json:"max_pnl"`
	MeanPnL       float64 `json:"mean_pnl"`
	MedianPnL     float64 `json:"median_pnl"`
	Quantile99PnL float64 `json:"quantile_99_pnl"`
	Quantile01PnL float64 `json:"quantile_01_pnl"`
}

func CalculateSnapshotStatistics(snapshots []PositionSnapshot) SnapshotStatistics {
	if len(snapshots) == 0 {
		return SnapshotStatistics{}
	}

	pnls := make([]float64, len(snapshots))
	var sum float64
	minPnL := math.MaxFloat64
	maxPnL := -math.MaxFloat64

	for i, snap := range snapshots {
		pnl := snap.UnrealizedPL
		pnls[i] = pnl
		sum += pnl

		if pnl < minPnL {
			minPnL = pnl
		}
		if pnl > maxPnL {
			maxPnL = pnl
		}
	}

	sort.Float64s(pnls)

	mean := sum / float64(len(pnls))

	var median float64
	if len(pnls)%2 == 0 {
		median = (pnls[len(pnls)/2-1] + pnls[len(pnls)/2]) / 2
	} else {
		median = pnls[len(pnls)/2]
	}

	quantile99Index := int(math.Ceil(float64(len(pnls)) * 0.99))
	if quantile99Index >= len(pnls) {
		quantile99Index = len(pnls) - 1
	}
	quantile99 := pnls[quantile99Index]

	quantile01Index := int(math.Floor(float64(len(pnls)) * 0.01))
	quantile01 := pnls[quantile01Index]

	return SnapshotStatistics{
		TotalCount:    len(snapshots),
		MinPnL:        minPnL,
		MaxPnL:        maxPnL,
		MeanPnL:       mean,
		MedianPnL:     median,
		Quantile99PnL: quantile99,
		Quantile01PnL: quantile01,
	}
}

func AnalyzePositionClose(claudeClient claude.ClaudeApi, position PositionRecord, snapshots []PositionSnapshot, recentTweets []CommunityTweet, btcIchimoku IchimokuAnalysis, coinIchimoku IchimokuAnalysis, ichimoku ClosePositionReason, maSignal MovingAveragePnLSignal) (ClaudePositionCloseResponse, error) {
	systemPrompt := `You are a cryptocurrency trading assistant analyzing whether to close an open position.

You will receive:
1. Current open position details (side, entry price, current P/L, max/min P/L)
2. Position snapshots statistics (count, min/max/mean/median PnL, quantiles 0.01/0.99)
3. Last 50 community messages
4. BTC Ichimoku analysis
5. Coin Ichimoku analysis
6. Moving Average PnL Exit Signal - THIS IS CRITICAL!

Your task is to analyze:
- Historical P/L statistics and distribution
- Community sentiment trends from messages
- How BTC signals might affect this coin
- Current position performance vs historical patterns (use quantiles to understand volatility)
- Risk of reversal vs potential for more profit

🚨 CRITICAL RULE: Moving Average Exit Signal
If the "moving_average_signal.should_close" field is TRUE, you MUST recommend closing the position.
This is a mandatory exit signal that cannot be overridden. The system will force close the position regardless of your recommendation.
When this signal is active:
- Set should_close: true
- Include the MA signal reason in your justification
- Your analysis is informative only - the position WILL be closed

Be careful more attention what exactly side for our current position, if it is short and price go down and all indicators for it,  we should hold this position open.
Same for LONG position, is its long, and price grow up, and all indicators for it, we should continue hold position don't close.
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
		Position                   PositionRecord         `json:"position"`
		SnapshotStats              SnapshotStatistics     `json:"snapshot_statistics"`
		Tweets                     []CommunityTweet       `json:"recent_tweets"`
		BTCIchimoku                IchimokuAnalysis       `json:"btc_ichimoku"`
		CoinIchimoku               IchimokuAnalysis       `json:"coin_ichimoku"`
		CurrentPositionShortOrLong string                 `json:"current_position_short_or_long"`
		ShouldCloseByIchimoku      ClosePositionReason    `json:"should_close_by_ichimoku"`
		MovingAverageSignal        MovingAveragePnLSignal `json:"moving_average_signal"`
		CurrentDate                string                 `json:"current_date"`
		PositionOpenDate           string                 `json:"position_open_date"`
	}

	snapshotStats := CalculateSnapshotStatistics(snapshots)

	requestData := CloseAnalysisRequest{
		CurrentPositionShortOrLong: position.Side,
		Position:                   position,
		SnapshotStats:              snapshotStats,
		Tweets:                     recentTweets,
		BTCIchimoku:                btcIchimoku,
		CoinIchimoku:               coinIchimoku,
		ShouldCloseByIchimoku:      ichimoku,
		MovingAverageSignal:        maSignal,
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
