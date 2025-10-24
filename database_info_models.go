package main

import (
	"time"

	"gorm.io/gorm"
)

type TokenInfo struct {
	ID          uint `gorm:"primarykey"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	Symbol      string         `gorm:"uniqueIndex;not null"`
	Name        string
	Address     string `gorm:"uniqueIndex"`
	Decimals    int
	Description string
	IsActive    bool `gorm:"default:true"`
}

type PriceHistory struct {
	ID          uint `gorm:"primarykey"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	TokenSymbol string         `gorm:"index;not null"`
	PriceUSD    float64
	Volume24h   float64
	Liquidity   float64
	Timestamp   time.Time `gorm:"index"`
}

type MarketIndicator struct {
	ID          uint `gorm:"primarykey"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	TokenSymbol string         `gorm:"index;not null"`
	RSI         float64
	MACD        float64
	Signal      float64
	Volume      float64
	Timestamp   time.Time `gorm:"index"`
}
