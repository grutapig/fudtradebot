package main

import (
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"time"
)

type BalanceRecord struct {
	ID                 uint   `gorm:"primarykey"`
	AccountAlias       string `gorm:"index"`
	Asset              string `gorm:"index"`
	Balance            float64
	CrossWalletBalance float64
	CrossUnPnl         float64
	AvailableBalance   float64
	MaxWithdrawAmount  float64
	MarginAvailable    bool
	UpdateTime         int64
	Timestamp          time.Time `gorm:"index"`
}

type PositionSnapshot struct {
	ID               uint   `gorm:"primarykey"`
	PositionUUID     string `gorm:"index"`
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

type PositionRecord struct {
	ID               uint      `gorm:"primarykey"`
	UUID             string    `gorm:"uniqueIndex;not null"`
	Symbol           string    `gorm:"index;not null"`
	Side             string    `gorm:"not null"`
	Leverage         int       `gorm:"not null"`
	Quantity         float64   `gorm:"not null"`
	EntryPrice       float64   `gorm:"not null"`
	OpenedAt         time.Time `gorm:"index;not null"`
	IsClosed         bool      `gorm:"index;default:false"`
	ClosedAt         *time.Time
	ClosePrice       float64
	RealizedPL       float64
	CurrentPnL       float64
	CurrentMarkPrice float64
	MaxPnL           float64 `gorm:"column:max_pnl"`
	MinPnL           float64 `gorm:"column:min_pnl"`
	Duration         int64
	OpenReason       string
	CloseReason      string
	CreatedAt        time.Time `gorm:"index"`
	UpdatedAt        time.Time
}

type FudAttackRecord struct {
	ID              uint   `gorm:"primarykey"`
	PositionUUID    string `gorm:"index"`
	Symbol          string `gorm:"index;not null"`
	HasAttack       bool   `gorm:"not null"`
	Confidence      float64
	MessageCount    int
	FudType         string
	Theme           string
	StartedHoursAgo int
	LastAttackTime  time.Time `gorm:"index"`
	Justification   string
	Participants    string
	CreatedAt       time.Time `gorm:"index"`
}

type AIOrderValidationRecord struct {
	ID                uint   `gorm:"primarykey"`
	PositionUUID      string `gorm:"index"`
	DecisionRecordID  uint   `gorm:"index"`
	Symbol            string `gorm:"index;not null"`
	RequestData       string `gorm:"type:text"`
	ResponseData      string `gorm:"type:text"`
	ShouldOpenOrder   bool   `gorm:"not null"`
	ConfidencePercent float64
	Justification     string    `gorm:"type:text"`
	CreatedAt         time.Time `gorm:"index"`
}

type AiPositionCloseRecord struct {
	ID                uint      `gorm:"primarykey"`
	PositionUUID      string    `gorm:"index;not null" json:"position_uuid"`
	Symbol            string    `gorm:"index;not null"`
	SnapshotCount     int       `gorm:"not null"`
	RequestData       string    `gorm:"type:text"`
	ResponseData      string    `gorm:"type:text"`
	ShouldClose       bool      `gorm:"not null"`
	ConfidencePercent float64   `json:"confidence_percent"`
	Justification     string    `gorm:"type:text"`
	ExpectedPnL       float64   `gorm:"column:expected_pnl" json:"expected_pnl"`
	RiskAssessment    string    `gorm:"type:text"`
	CreatedAt         time.Time `gorm:"index"`
}

var DB *gorm.DB

func InitDatabase() error {
	var err error
	DB, err = gorm.Open(sqlite.Open("trading_bot.db"), &gorm.Config{})
	if err != nil {
		return err
	}

	return DB.AutoMigrate(&BalanceRecord{}, &PositionSnapshot{}, &TradingDecisionRecord{}, &PositionRecord{}, &FudAttackRecord{}, &AIOrderValidationRecord{}, &AiPositionCloseRecord{})
}

func SaveBalance(asset string, totalBalance float64, availableBalance float64) error {
	record := BalanceRecord{
		Asset:            asset,
		Balance:          totalBalance,
		AvailableBalance: availableBalance,
		Timestamp:        time.Now(),
	}
	return DB.Create(&record).Error
}

func SaveBalanceInfo(info AccountBalanceInfo) error {
	record := BalanceRecord{
		AccountAlias:       info.AccountAlias,
		Asset:              info.Asset,
		Balance:            info.Balance,
		CrossWalletBalance: info.CrossWalletBalance,
		CrossUnPnl:         info.CrossUnPnl,
		AvailableBalance:   info.AvailableBalance,
		MaxWithdrawAmount:  info.MaxWithdrawAmount,
		MarginAvailable:    info.MarginAvailable,
		UpdateTime:         info.UpdateTime,
		Timestamp:          time.Now(),
	}
	return DB.Create(&record).Error
}

func SaveAllBalances(infos []AccountBalanceInfo) error {
	for _, info := range infos {
		if err := SaveBalanceInfo(info); err != nil {
			return err
		}
	}
	return nil
}

func GetBalanceHistory(asset string, hoursBack int) ([]BalanceRecord, error) {
	var records []BalanceRecord
	startTime := time.Now().Add(-time.Duration(hoursBack) * time.Hour)

	err := DB.Where("asset = ? AND timestamp >= ?", asset, startTime).
		Order("timestamp ASC").
		Find(&records).Error

	return records, err
}

func GetAllAssets() ([]string, error) {
	var assets []string
	err := DB.Model(&BalanceRecord{}).
		Distinct("asset").
		Pluck("asset", &assets).Error
	return assets, err
}

func SavePositionSnapshot(position Position, markPrice float64, positionUUID string) error {
	snapshot := PositionSnapshot{
		PositionUUID:     positionUUID,
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
	if err := DB.Create(&snapshot).Error; err != nil {
		return err
	}

	if err := UpdatePositionMaxMinPnL(positionUUID, position.UnrealizedPL); err != nil {
		return err
	}

	return DB.Model(&PositionRecord{}).
		Where("uuid = ? AND is_closed = ?", positionUUID, false).
		Updates(map[string]interface{}{
			"current_pn_l":       position.UnrealizedPL,
			"current_mark_price": markPrice,
		}).Error
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

func GetRecentDecisions(hoursBack int) ([]TradingDecisionRecord, error) {
	var decisions []TradingDecisionRecord
	startTime := time.Now().Add(-time.Duration(hoursBack) * time.Hour)

	err := DB.Where("created_at >= ?", startTime).
		Order("created_at DESC").
		Find(&decisions).Error

	return decisions, err
}

func GetRecentDecisionsWithPagination(limit int, offset int) ([]TradingDecisionRecord, error) {
	var decisions []TradingDecisionRecord

	err := DB.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&decisions).Error

	return decisions, err
}

func GetDecisionByID(id uint) (*TradingDecisionRecord, error) {
	var decision TradingDecisionRecord

	err := DB.First(&decision, id).Error
	if err != nil {
		return nil, err
	}

	return &decision, nil
}

func UpdateDecisionPositionUUID(symbol string, positionUUID string) error {
	return DB.Model(&TradingDecisionRecord{}).
		Where("symbol = ? AND position_uuid = ''", symbol).
		Order("created_at DESC").
		Limit(1).
		Update("position_uuid", positionUUID).Error
}

func UpdateDecisionPositionUUIDByID(decisionID uint, positionUUID string) error {
	return DB.Model(&TradingDecisionRecord{}).
		Where("id = ?", decisionID).
		Update("position_uuid", positionUUID).Error
}

func GetLatestPositionSnapshotByUUID(positionUUID string) (*PositionSnapshot, error) {
	var snapshot PositionSnapshot

	err := DB.Where("position_uuid = ?", positionUUID).
		Order("created_at DESC").
		First(&snapshot).Error

	if err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func GetClosedPositionSnapshots(hoursBack int) ([]PositionSnapshot, error) {
	var snapshots []PositionSnapshot
	startTime := time.Now().Add(-time.Duration(hoursBack) * time.Hour)

	subQuery := DB.Table("position_snapshots").
		Select("position_uuid, MAX(created_at) as max_created_at").
		Where("created_at >= ?", startTime).
		Group("position_uuid")

	err := DB.Table("position_snapshots as ps").
		Joins("INNER JOIN (?) as latest ON ps.position_uuid = latest.position_uuid AND ps.created_at = latest.max_created_at", subQuery).
		Where("ps.created_at >= ?", startTime).
		Order("ps.created_at DESC").
		Find(&snapshots).Error

	return snapshots, err
}

func GetPositionSnapshotsByUUID(positionUUID string) ([]PositionSnapshot, error) {
	var snapshots []PositionSnapshot
	err := DB.Where("position_uuid = ?", positionUUID).
		Order("created_at ASC").
		Find(&snapshots).Error

	return snapshots, err
}

func GetDecisionsByPositionUUID(positionUUID string) ([]TradingDecisionRecord, error) {
	var decisions []TradingDecisionRecord

	err := DB.Where("position_uuid = ?", positionUUID).
		Order("created_at ASC").
		Find(&decisions).Error

	return decisions, err
}

func SavePositionOpen(position PositionRecord) error {
	return DB.Create(&position).Error
}

func UpdatePositionClose(uuid string, closePrice float64, realizedPL float64, closeReason string) error {
	closedAt := time.Now()
	var position PositionRecord

	if err := DB.Where("uuid = ?", uuid).First(&position).Error; err != nil {
		return err
	}

	duration := closedAt.Sub(position.OpenedAt).Milliseconds()

	maxPnL, minPnL := GetMaxMinPnLFromSnapshots(uuid)

	return DB.Model(&PositionRecord{}).
		Where("uuid = ?", uuid).
		Updates(map[string]interface{}{
			"is_closed":    true,
			"closed_at":    closedAt,
			"close_price":  closePrice,
			"realized_pl":  realizedPL,
			"duration":     duration,
			"close_reason": closeReason,
			"max_pnl":      maxPnL,
			"min_pnl":      minPnL,
		}).Error
}

func GetMaxMinPnLFromSnapshots(uuid string) (float64, float64) {
	var maxPnL, minPnL float64

	DB.Model(&PositionSnapshot{}).
		Where("position_uuid = ?", uuid).
		Select("COALESCE(MAX(unrealized_pl), 0)").
		Row().Scan(&maxPnL)

	DB.Model(&PositionSnapshot{}).
		Where("position_uuid = ?", uuid).
		Select("COALESCE(MIN(unrealized_pl), 0)").
		Row().Scan(&minPnL)

	return maxPnL, minPnL
}

func UpdatePositionMaxMinPnL(uuid string, currentPnL float64) error {
	var position PositionRecord
	if err := DB.Where("uuid = ?", uuid).First(&position).Error; err != nil {
		return err
	}

	updates := make(map[string]interface{})

	if currentPnL > position.MaxPnL {
		updates["max_pnl"] = currentPnL
	}

	if currentPnL < position.MinPnL || position.MinPnL == 0 {
		updates["min_pnl"] = currentPnL
	}

	if len(updates) > 0 {
		return DB.Model(&PositionRecord{}).Where("uuid = ?", uuid).Updates(updates).Error
	}

	return nil
}

func BackfillMaxMinPnL() error {
	var positions []PositionRecord
	if err := DB.Find(&positions).Error; err != nil {
		return err
	}
	filled := 0
	for _, position := range positions {
		var snapshotCount int64
		DB.Model(&PositionSnapshot{}).Where("position_uuid = ?", position.UUID).Count(&snapshotCount)
		if snapshotCount > 0 && position.MaxPnL == 0 && position.MinPnL == 0 {
			maxPnL, minPnL := GetMaxMinPnLFromSnapshots(position.UUID)
			err := DB.Model(&PositionRecord{}).
				Where("uuid = ?", position.UUID).
				Updates(map[string]interface{}{
					"max_pnl": maxPnL,
					"min_pnl": minPnL,
				}).Error
			if err != nil {
				log.Println("cannot update max pnl ont start app for", position.Symbol, position, maxPnL, minPnL)
			}
			filled++
		}

	}
	return nil
}

func GetPositionByUUID(uuid string) (PositionRecord, error) {
	var position PositionRecord
	err := DB.Where("uuid = ?", uuid).First(&position).Error
	return position, err
}

func GetOpenPositionBySymbolAndSide(symbol string, side string) (PositionRecord, error) {
	var position PositionRecord
	err := DB.Where("symbol = ? AND side = ? AND is_closed = ?", symbol, side, false).First(&position).Error
	return position, err
}

func CloseOpenPositionsBySymbol(symbol string) error {
	return DB.Model(&PositionRecord{}).
		Where("symbol = ? AND is_closed = ?", symbol, false).
		Updates(map[string]interface{}{
			"is_closed":    true,
			"closed_at":    time.Now(),
			"close_reason": "no_exchange_position_on_init",
		}).Error
}

func DeleteOpenPositionBySymbolAndSide(symbol string, side string) error {
	return DB.Unscoped().Where("symbol = ? AND side = ? AND is_closed = ?", symbol, side, false).Delete(&PositionRecord{}).Error
}

func GetOpenPositions() ([]PositionRecord, error) {
	var positions []PositionRecord
	err := DB.Where("is_closed = ?", false).Find(&positions).Error
	return positions, err
}

func GetClosedPositions(hoursBack int) ([]PositionRecord, error) {
	var positions []PositionRecord
	startTime := time.Now().Add(-time.Duration(hoursBack) * time.Hour)
	err := DB.Where("is_closed = ? AND closed_at >= ?", true, startTime).
		Order("closed_at DESC").
		Find(&positions).Error
	return positions, err
}

func GetClosedPositionsWithPagination(limit int, offset int) ([]PositionRecord, error) {
	var positions []PositionRecord
	err := DB.Where("is_closed = ?", true).
		Order("closed_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&positions).Error
	return positions, err
}

func GetAllClosedPositionsOrdered() ([]PositionRecord, error) {
	var positions []PositionRecord
	err := DB.Where("is_closed = ?", true).
		Order("closed_at ASC").
		Find(&positions).Error
	return positions, err
}

func GetPositionsBySymbol(symbol string, hoursBack int) ([]PositionRecord, error) {
	var positions []PositionRecord
	startTime := time.Now().Add(-time.Duration(hoursBack) * time.Hour)
	err := DB.Where("symbol = ? AND opened_at >= ?", symbol, startTime).
		Order("opened_at DESC").
		Find(&positions).Error
	return positions, err
}

func SaveFudAttack(attack ClaudeFudAttackResponse, symbol string, positionUUID string) error {
	participantsJSON := ""
	if len(attack.Participants) > 0 {
		for _, p := range attack.Participants {
			if participantsJSON != "" {
				participantsJSON += ", "
			}
			participantsJSON += p.Username
		}
	}

	lastAttackTime := time.Time{}
	if attack.LastAttackTime != nil {
		lastAttackTime = *attack.LastAttackTime
	}

	record := FudAttackRecord{
		PositionUUID:    positionUUID,
		Symbol:          symbol,
		HasAttack:       attack.HasAttack,
		Confidence:      attack.Confidence,
		MessageCount:    attack.MessageCount,
		FudType:         attack.FudType,
		Theme:           attack.Theme,
		StartedHoursAgo: attack.StartedHoursAgo,
		LastAttackTime:  lastAttackTime,
		Justification:   attack.Justification,
		Participants:    participantsJSON,
		CreatedAt:       time.Now(),
	}
	return DB.Create(&record).Error
}

func GetRecentFudAttacks(symbol string, hoursBack int) ([]FudAttackRecord, error) {
	var attacks []FudAttackRecord
	startTime := time.Now().Add(-time.Duration(hoursBack) * time.Hour)
	err := DB.Where("symbol = ? AND created_at >= ?", symbol, startTime).
		Order("created_at DESC").
		Find(&attacks).Error
	return attacks, err
}

func GetLatestFudAttack(symbol string) (*FudAttackRecord, error) {
	var attack FudAttackRecord
	err := DB.Where("symbol = ?", symbol).
		Order("created_at DESC").
		First(&attack).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &attack, nil
}

func GetFudAttacksByPositionUUID(positionUUID string) ([]FudAttackRecord, error) {
	var attacks []FudAttackRecord
	err := DB.Where("position_uuid = ?", positionUUID).
		Order("created_at DESC").
		Find(&attacks).Error
	return attacks, err
}

func SaveAIOrderValidation(validation *AIOrderValidationRecord) error {
	return DB.Create(validation).Error
}

func GetAIValidationsByPositionUUID(positionUUID string) ([]AIOrderValidationRecord, error) {
	var validations []AIOrderValidationRecord
	err := DB.Where("position_uuid = ?", positionUUID).
		Order("created_at DESC").
		Find(&validations).Error
	return validations, err
}

func GetAIValidationByDecisionID(decisionID uint) (*AIOrderValidationRecord, error) {
	var validation AIOrderValidationRecord
	err := DB.Where("decision_record_id = ?", decisionID).
		Order("created_at DESC").
		First(&validation).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &validation, nil
}

func GetRecentAIValidations(hoursBack int) ([]AIOrderValidationRecord, error) {
	var validations []AIOrderValidationRecord
	cutoff := time.Now().Add(-time.Duration(hoursBack) * time.Hour)
	err := DB.Where("created_at > ?", cutoff).
		Order("created_at DESC").
		Limit(100).
		Find(&validations).Error
	return validations, err
}

func GetAIValidationByID(id uint) (*AIOrderValidationRecord, error) {
	var validation AIOrderValidationRecord
	err := DB.Where("id = ?", id).First(&validation).Error
	if err != nil {
		return nil, err
	}
	return &validation, nil
}

func SaveAIPositionClose(record *AiPositionCloseRecord) error {
	return DB.Create(record).Error
}

func GetAIPositionClosesByUUID(positionUUID string) ([]AiPositionCloseRecord, error) {
	var records []AiPositionCloseRecord
	err := DB.Where("position_uuid = ?", positionUUID).
		Order("created_at DESC").
		Find(&records).Error
	return records, err
}
func GetAIPositionCloses() ([]AiPositionCloseRecord, error) {
	var records []AiPositionCloseRecord
	err := DB.Order("created_at DESC").
		Find(&records).Error
	return records, err
}

func GetAIPositionClosesWithPagination(limit int, offset int) ([]AiPositionCloseRecord, error) {
	var records []AiPositionCloseRecord
	err := DB.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error
	return records, err
}

func GetAIPositionCloseByID(id uint) (*AiPositionCloseRecord, error) {
	var record AiPositionCloseRecord
	err := DB.Where("id = ?", id).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func CountPositionSnapshots(positionUUID string) (int64, error) {
	var count int64
	err := DB.Model(&PositionSnapshot{}).
		Where("position_uuid = ?", positionUUID).
		Count(&count).Error
	return count, err
}

func CalculateCurrentBalance() (float64, error) {
	var totalPnL float64
	err := DB.Model(&PositionRecord{}).
		Where("is_closed = ?", true).
		Select("COALESCE(SUM(current_pn_l), 0)").
		Scan(&totalPnL).Error

	if err != nil {
		return 0, err
	}

	return INITIAL_BALANCE + totalPnL, nil
}
