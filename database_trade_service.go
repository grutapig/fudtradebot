package main

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DatabaseTradeService struct {
	db *gorm.DB
}

func NewDatabaseTradeService(dsn string) (*DatabaseTradeService, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&Trade{}, &Position{}, &BotState{}, &TradingSignal{})
	if err != nil {
		return nil, err
	}

	return &DatabaseTradeService{db: db}, nil
}

func (s *DatabaseTradeService) CreateTrade(trade *Trade) error {
	return s.db.Create(trade).Error
}

func (s *DatabaseTradeService) UpdateTradeStatus(id uint, status string, txHash string, errorMsg string) error {
	return s.db.Model(&Trade{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":    status,
		"tx_hash":   txHash,
		"error_msg": errorMsg,
	}).Error
}

func (s *DatabaseTradeService) GetTrade(id uint) (*Trade, error) {
	var trade Trade
	err := s.db.First(&trade, id).Error
	return &trade, err
}

func (s *DatabaseTradeService) GetTradeHistory(symbol string, limit int) ([]Trade, error) {
	var trades []Trade
	query := s.db.Order("created_at DESC")
	if symbol != "" {
		query = query.Where("token_symbol = ?", symbol)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&trades).Error
	return trades, err
}

func (s *DatabaseTradeService) UpsertPosition(position *Position) error {
	var existing Position
	err := s.db.Where("token_symbol = ?", position.TokenSymbol).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return s.db.Create(position).Error
	}
	return s.db.Model(&existing).Updates(position).Error
}

func (s *DatabaseTradeService) GetPosition(symbol string) (*Position, error) {
	var position Position
	err := s.db.Where("token_symbol = ? AND is_active = ?", symbol, true).First(&position).Error
	return &position, err
}

func (s *DatabaseTradeService) GetAllActivePositions() ([]Position, error) {
	var positions []Position
	err := s.db.Where("is_active = ?", true).Find(&positions).Error
	return positions, err
}

func (s *DatabaseTradeService) ClosePosition(symbol string) error {
	return s.db.Model(&Position{}).Where("token_symbol = ?", symbol).Update("is_active", false).Error
}

func (s *DatabaseTradeService) SetState(key string, value string) error {
	var state BotState
	err := s.db.Where("key = ?", key).First(&state).Error
	if err == gorm.ErrRecordNotFound {
		state = BotState{Key: key, Value: value, LastModified: time.Now()}
		return s.db.Create(&state).Error
	}
	return s.db.Model(&state).Updates(map[string]interface{}{
		"value":         value,
		"last_modified": time.Now(),
	}).Error
}

func (s *DatabaseTradeService) GetState(key string) (string, error) {
	var state BotState
	err := s.db.Where("key = ?", key).First(&state).Error
	if err != nil {
		return "", err
	}
	return state.Value, nil
}

func (s *DatabaseTradeService) CreateSignal(signal *TradingSignal) error {
	return s.db.Create(signal).Error
}

func (s *DatabaseTradeService) GetUnprocessedSignals() ([]TradingSignal, error) {
	var signals []TradingSignal
	err := s.db.Where("is_processed = ?", false).Order("timestamp ASC").Find(&signals).Error
	return signals, err
}

func (s *DatabaseTradeService) MarkSignalProcessed(id uint) error {
	return s.db.Model(&TradingSignal{}).Where("id = ?", id).Update("is_processed", true).Error
}
