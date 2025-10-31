package main

import (
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

type BalanceRecord struct {
	ID        uint   `gorm:"primarykey"`
	Asset     string `gorm:"index"`
	Balance   float64
	Timestamp time.Time `gorm:"index"`
}

type PositionSnapshot struct {
	ID               uint   `gorm:"primarykey"`
	Symbol           string `gorm:"index"`
	Side             string
	Leverage         int
	EntryPrice       float64
	Amount           float64
	UnrealizedPL     float64
	MarkPrice        float64
	PositionOpenedAt time.Time
	CreatedAt        time.Time `gorm:"index"`
}

type TradingDecisionRecord struct {
	ID                  uint   `gorm:"primarykey"`
	PositionUUID        string `gorm:"index"`
	Symbol              string `gorm:"index"`
	BTCIchimoku         string
	CoinIchimoku        string
	Activity            string
	FudActivity         string
	Sentiment           string
	FudAttack           string
	FinalDecision       string
	DecisionExplanation string
	CreatedAt           time.Time `gorm:"index"`
}

var DB *gorm.DB

func InitDatabase() error {
	var err error
	DB, err = gorm.Open(sqlite.Open("trading_bot.db"), &gorm.Config{})
	if err != nil {
		return err
	}

	return DB.AutoMigrate(&BalanceRecord{}, &PositionSnapshot{}, &TradingDecisionRecord{})
}

func SaveBalance(asset string, balance float64) error {
	record := BalanceRecord{
		Asset:     asset,
		Balance:   balance,
		Timestamp: time.Now(),
	}
	return DB.Create(&record).Error
}

func GetBalanceHistory(asset string, hoursBack int) ([]BalanceRecord, error) {
	var records []BalanceRecord
	startTime := time.Now().Add(-time.Duration(hoursBack) * time.Hour)

	err := DB.Where("asset = ? AND timestamp >= ?", asset, startTime).
		Order("timestamp ASC").
		Find(&records).Error

	return records, err
}

func SavePositionSnapshot(position Position, markPrice float64) error {
	snapshot := PositionSnapshot{
		Symbol:           position.Symbol,
		Side:             string(position.Side),
		Leverage:         position.Leverage,
		EntryPrice:       position.EntryPrice,
		Amount:           position.Amount,
		UnrealizedPL:     position.UnrealizedPL,
		MarkPrice:        markPrice,
		PositionOpenedAt: position.Timestamp,
		CreatedAt:        time.Now(),
	}
	return DB.Create(&snapshot).Error
}

func GetPositionHistory(symbol string, hoursBack int) ([]PositionSnapshot, error) {
	var snapshots []PositionSnapshot
	startTime := time.Now().Add(-time.Duration(hoursBack) * time.Hour)

	err := DB.Where("symbol = ? AND created_at >= ?", symbol, startTime).
		Order("created_at ASC").
		Find(&snapshots).Error

	return snapshots, err
}

func GetAllPositionSnapshots(hoursBack int) ([]PositionSnapshot, error) {
	var snapshots []PositionSnapshot
	startTime := time.Now().Add(-time.Duration(hoursBack) * time.Hour)

	err := DB.Where("created_at >= ?", startTime).
		Order("created_at DESC").
		Find(&snapshots).Error

	return snapshots, err
}

func GetLatestPositionSnapshot(symbol string) (*PositionSnapshot, error) {
	var snapshot PositionSnapshot

	err := DB.Where("symbol = ?", symbol).
		Order("created_at DESC").
		First(&snapshot).Error

	if err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func SaveTradingDecision(decision TradingDecisionRecord) error {
	return DB.Create(&decision).Error
}

func GetLatestTradingDecision(symbol string) (*TradingDecisionRecord, error) {
	var record TradingDecisionRecord

	err := DB.Where("symbol = ?", symbol).
		Order("created_at DESC").
		First(&record).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &record, nil
}

func GeneratePositionUUID() string {
	return uuid.New().String()
}
