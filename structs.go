package main

import (
	"time"
)

type TelegramMessage struct {
	ChatID  int64
	Text    string
	IsReply bool
}

type TradeDecision struct {
	Action      string
	TokenSymbol string
	Amount      float64
	Reason      string
	Timestamp   time.Time
}

type BalanceInfo struct {
	TokenSymbol string
	Amount      float64
	ValueUSD    float64
}

type MarketData struct {
	TokenSymbol string
	PriceUSD    float64
	Volume24h   float64
	Timestamp   time.Time
}

type AnalysisResult struct {
	TokenSymbol   string
	Signal        string
	Confidence    float64
	SuggestedSize float64
	Analysis      string
	Timestamp     time.Time
}

// PositionSide represents the direction of a futures position
type PositionSide string

const (
	PositionSideLong  PositionSide = "LONG"
	PositionSideShort PositionSide = "SHORT"
	PositionSideBoth  PositionSide = "BOTH"
)

// Position represents an open futures position
type Position struct {
	Symbol       string
	Side         PositionSide
	Leverage     int
	EntryPrice   float64
	Amount       float64
	UnrealizedPL float64
	Timestamp    time.Time
}

// OrderSide represents buy or sell
type OrderSide string

const (
	OrderBuy  OrderSide = "BUY"
	OrderSell OrderSide = "SELL"
)

// OrderType represents the type of order
type OrderType string

const (
	OrderTypeLimit  OrderType = "LIMIT"
	OrderTypeMarket OrderType = "MARKET"
)

type TradingAction string

const (
	TradingActionOpenLong   TradingAction = "OPEN_LONG"
	TradingActionOpenShort  TradingAction = "OPEN_SHORT"
	TradingActionCloseLong  TradingAction = "CLOSE_LONG"
	TradingActionCloseShort TradingAction = "CLOSE_SHORT"
	TradingActionHold       TradingAction = "HOLD"
	TradingActionSkip       TradingAction = "SKIP"
)

type SignalStrength int

const (
	SignalStrengthNone   SignalStrength = 0
	SignalStrengthWeak   SignalStrength = 1
	SignalStrengthMedium SignalStrength = 2
	SignalStrengthStrong SignalStrength = 3
)

type TradingSignal struct {
	Action   TradingAction
	Strength SignalStrength
	Reasons  []string
}

type TradingPair struct {
	CommunityID string
	Symbol      string
	Leverage    int
	Quantity    float64
}

type TradingState struct {
	CurrentPosition        PositionSide
	OpenedAt               time.Time
	OpenReason             string
	PositionUUID           string
	LastSentimentAnalysis  ClaudeSentimentResponse
	LastSentimentFetchTime time.Time
	LastFudAttack          ClaudeFudAttackResponse
	LastFudAttackFetchTime time.Time
	LastAnalyzedTweetID    string
	LastFudCheckTime       time.Time
	LastFudCheckTweetID    string
	LastDecisionHash       string
	FudAttackMode          bool
	FudAttackStartTime     time.Time
	FudAttackShortStarted  bool
	LastAIRejectionTime    time.Time
	LastRejectedDecision   string
}

type CommunityTweet struct {
	ID        string    `json:"id"`
	Date      time.Time `json:"date"`
	Text      string    `json:"text"`
	Sentiment int       `json:"sentiment"`
	IsFud     bool      `json:"is_fud"`
}
type TweetsResponse struct {
	Status  string           `json:"status"`
	Data    []CommunityTweet `json:"data,omitempty"`
	Message string           `json:"message,omitempty"`
	Error   string           `json:"error,omitempty"`
}

type ClaudeSentimentResponse struct {
	OverallSentiment int      `json:"overall_sentiment"`
	SentimentTrend   string   `json:"sentiment_trend"`
	FudLevel         int      `json:"fud_level"`
	Confidence       float64  `json:"confidence"`
	KeyThemes        []string `json:"key_themes"`
	Recommendation   string   `json:"recommendation"`
}

type FudAttackParticipant struct {
	Username     string `json:"username"`
	MessageCount int    `json:"message_count"`
}

type ClaudeFudAttackResponse struct {
	HasAttack       bool                   `json:"has_attack"`
	Confidence      float64                `json:"confidence"`
	MessageCount    int                    `json:"message_count"`
	Participants    []FudAttackParticipant `json:"participants"`
	FudType         string                 `json:"fud_type"`
	Theme           string                 `json:"theme"`
	StartedHoursAgo int                    `json:"started_hours_ago"`
	LastAttackTime  time.Time              `json:"last_attack_time"`
	Justification   string                 `json:"justification"`
}

type ClaudeOrderValidationResponse struct {
	ShouldOpenOrder   bool    `json:"should_open_order"`
	ConfidencePercent float64 `json:"confidence_percent"`
	Justification     string  `json:"justification"`
}

type ClaudePositionCloseResponse struct {
	ShouldClose       bool    `json:"should_close"`
	ConfidencePercent float64 `json:"confidence_percent"`
	Justification     string  `json:"justification"`
	ExpectedPnL       float64 `json:"expected_pnl"`
	RiskAssessment    string  `json:"risk_assessment"`
}

type MovingAveragePnLSignal struct {
	ShouldClose    bool    `json:"should_close"`
	CurrentPnL     float64 `json:"current_pnl"`
	MovingAverage  float64 `json:"moving_average"`
	Threshold      float64 `json:"threshold"`
	SnapshotsCount int     `json:"snapshots_count"`
	PercentBelowMA float64 `json:"percent_below_ma"`
	TriggerReason  string  `json:"trigger_reason"`
}
