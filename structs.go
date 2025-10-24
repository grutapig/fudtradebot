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
