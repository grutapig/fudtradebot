package main

import (
	"time"

	"gorm.io/gorm"
)

type Trade struct {
	ID          uint `gorm:"primarykey"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	TokenSymbol string         `gorm:"index;not null"`
	Side        string         `gorm:"not null"`
	Amount      float64        `gorm:"not null"`
	PriceUSD    float64
	TxHash      string `gorm:"uniqueIndex"`
	Status      string `gorm:"default:'pending'"`
	ErrorMsg    string
	Timestamp   time.Time `gorm:"index"`
}

type Position struct {
	ID             uint `gorm:"primarykey"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
	TokenSymbol    string         `gorm:"uniqueIndex;not null"`
	Amount         float64        `gorm:"not null"`
	AvgBuyPrice    float64
	CurrentPrice   float64
	ProfitLossUSD  float64
	ProfitLossPerc float64
	IsActive       bool `gorm:"default:true"`
}

type BotState struct {
	ID           uint `gorm:"primarykey"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	Key          string         `gorm:"uniqueIndex;not null"`
	Value        string
	LastModified time.Time
}

type TradingSignal struct {
	ID          uint `gorm:"primarykey"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	TokenSymbol string         `gorm:"index;not null"`
	Signal      string         `gorm:"not null"`
	Confidence  float64
	Source      string
	IsProcessed bool      `gorm:"default:false"`
	Timestamp   time.Time `gorm:"index"`
}
