package main

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DatabaseInfoService struct {
	db *gorm.DB
}

func NewDatabaseInfoService(dsn string) (*DatabaseInfoService, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &DatabaseInfoService{db: db}, nil
}

func (s *DatabaseInfoService) GetTokenInfo(symbol string) (*TokenInfo, error) {
	var token TokenInfo
	err := s.db.Where("symbol = ? AND is_active = ?", symbol, true).First(&token).Error
	return &token, err
}

func (s *DatabaseInfoService) GetLatestPrice(symbol string) (*PriceHistory, error) {
	var price PriceHistory
	err := s.db.Where("token_symbol = ?", symbol).Order("timestamp DESC").First(&price).Error
	return &price, err
}

func (s *DatabaseInfoService) GetPriceHistory(symbol string, from time.Time, to time.Time) ([]PriceHistory, error) {
	var prices []PriceHistory
	err := s.db.Where("token_symbol = ? AND timestamp BETWEEN ? AND ?", symbol, from, to).
		Order("timestamp ASC").Find(&prices).Error
	return prices, err
}

func (s *DatabaseInfoService) GetLatestIndicators(symbol string) (*MarketIndicator, error) {
	var indicator MarketIndicator
	err := s.db.Where("token_symbol = ?", symbol).Order("timestamp DESC").First(&indicator).Error
	return &indicator, err
}

func (s *DatabaseInfoService) GetIndicatorsHistory(symbol string, from time.Time, to time.Time) ([]MarketIndicator, error) {
	var indicators []MarketIndicator
	err := s.db.Where("token_symbol = ? AND timestamp BETWEEN ? AND ?", symbol, from, to).
		Order("timestamp ASC").Find(&indicators).Error
	return indicators, err
}

func (s *DatabaseInfoService) GetActiveTokens() ([]TokenInfo, error) {
	var tokens []TokenInfo
	err := s.db.Where("is_active = ?", true).Find(&tokens).Error
	return tokens, err
}
