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
)

type TradingPair struct {
	CommunityID string
	Symbol      string
	Leverage    int
}

type TradingState struct {
	CurrentPosition PositionSide
	OpenedAt        time.Time
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
