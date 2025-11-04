package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	tradingStates = make(map[string]*TradingState)
	statesMutex   sync.RWMutex
)

func handleAPIRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api")

	switch {
	case strings.HasPrefix(path, "/status"):
		handleStatus(w, r)
	case strings.HasPrefix(path, "/pairs"):
		handlePairs(w, r)
	case strings.HasPrefix(path, "/balance-history"):
		handleBalanceHistory(w, r)
	case strings.HasPrefix(path, "/balance"):
		handleBalance(w, r)
	case strings.HasPrefix(path, "/assets"):
		handleAssets(w, r)
	case strings.HasPrefix(path, "/decisions"):
		handleDecisions(w, r)
	case strings.HasPrefix(path, "/decision-detail"):
		handleDecisionDetail(w, r)
	case strings.HasPrefix(path, "/position-snapshots-history"):
		handlePositionSnapshotsHistory(w, r)
	case strings.HasPrefix(path, "/position-decisions"):
		handlePositionDecisions(w, r)
	case strings.HasPrefix(path, "/positions"):
		handlePositions(w, r)
	case strings.HasPrefix(path, "/position-snapshots"):
		handlePositionSnapshots(w, r)
	case strings.HasPrefix(path, "/fud-attacks"):
		handleFudAttacks(w, r)
	case strings.HasPrefix(path, "/pnl-history"):
		handlePnLHistory(w, r)
	case strings.HasPrefix(path, "/ai-validations"):
		handleAIValidations(w, r)
	case strings.HasPrefix(path, "/recent-ai-validations"):
		handleRecentAIValidations(w, r)
	case strings.HasPrefix(path, "/ai-close-analyses"):
		handleAICloseAnalyses(w, r)
	case strings.HasPrefix(path, "/ai-close-analyses-by-position"):
		handleAICloseAnalysesByPosition(w, r)
	default:
		http.NotFound(w, r)
	}
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status": "running",
		"pairs":  len(TradingPairs),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handlePairs(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pairs := make([]map[string]interface{}, len(TradingPairs))
	for i, pair := range TradingPairs {
		pairs[i] = map[string]interface{}{
			"symbol":       pair.Symbol,
			"community_id": pair.CommunityID,
			"leverage":     pair.Leverage,
			"quantity":     pair.Quantity,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pairs": pairs,
	})
}

func UpdateTradingState(symbol string, state *TradingState) {
	statesMutex.Lock()
	defer statesMutex.Unlock()
	tradingStates[symbol] = state
}

func handleBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	balance, err := CalculateCurrentBalance()
	if err != nil {
		http.Error(w, "Failed to calculate balance", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"balance": balance,
	})
}

func handleBalanceHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	asset := r.URL.Query().Get("asset")
	if asset == "" {
		asset = "USDT"
	}

	records, err := GetBalanceHistory(asset, 168)
	if err != nil {
		http.Error(w, "Failed to get balance history", http.StatusInternalServerError)
		return
	}

	type BalanceHistoryPoint struct {
		Timestamp          int64   `json:"timestamp"`
		Balance            float64 `json:"balance"`
		AvailableBalance   float64 `json:"available_balance"`
		MaxWithdrawAmount  float64 `json:"max_withdraw_amount"`
		CrossWalletBalance float64 `json:"cross_wallet_balance"`
		CrossUnPnl         float64 `json:"cross_un_pnl"`
	}

	history := make([]BalanceHistoryPoint, len(records))
	for i, record := range records {
		history[i] = BalanceHistoryPoint{
			Timestamp:          record.Timestamp.UnixMilli(),
			Balance:            record.Balance,
			AvailableBalance:   record.AvailableBalance,
			MaxWithdrawAmount:  record.MaxWithdrawAmount,
			CrossWalletBalance: record.CrossWalletBalance,
			CrossUnPnl:         record.CrossUnPnl,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"history": history,
	})
}

func handleAssets(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	assets, err := GetAllAssets()
	if err != nil {
		http.Error(w, "Failed to get assets", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"assets": assets,
	})
}

func handleDecisions(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 100
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	decisions, err := GetRecentDecisionsWithPagination(limit, offset)
	if err != nil {
		http.Error(w, "Failed to get decisions", http.StatusInternalServerError)
		return
	}

	grouped := make(map[string][]map[string]interface{})

	for _, decision := range decisions {
		item := map[string]interface{}{
			"id":            decision.ID,
			"symbol":        decision.Symbol,
			"position_uuid": decision.PositionUUID,
			"decision":      decision.FinalDecision,
			"btc_ichimoku":  decision.BTCIchimoku,
			"coin_ichimoku": decision.CoinIchimoku,
			"activity":      decision.Activity,
			"fud_activity":  decision.FudActivity,
			"sentiment":     decision.Sentiment,
			"fud_attack":    decision.FudAttack,
			"explanation":   decision.DecisionExplanation,
			"created_at":    decision.CreatedAt,
		}
		grouped[decision.Symbol] = append(grouped[decision.Symbol], item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"decisions": grouped,
	})
}

func handleDecisionDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Decision ID required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid decision ID", http.StatusBadRequest)
		return
	}

	decision, err := GetDecisionByID(uint(id))
	if err != nil {
		http.Error(w, "Decision not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(decision)
}

func handlePositionSnapshotsHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	snapshots, err := GetClosedPositionSnapshots(168)
	if err != nil {
		http.Error(w, "Failed to get position history", http.StatusInternalServerError)
		return
	}

	type PositionHistoryItem struct {
		PositionUUID string    `json:"position_uuid"`
		Symbol       string    `json:"symbol"`
		Side         string    `json:"side"`
		OpenedAt     time.Time `json:"opened_at"`
		ClosedAt     time.Time `json:"closed_at"`
		EntryPrice   float64   `json:"entry_price"`
		ExitPrice    float64   `json:"exit_price"`
		FinalPL      float64   `json:"final_pl"`
		Reason       string    `json:"reason"`
	}

	history := make([]PositionHistoryItem, len(snapshots))
	for i, snap := range snapshots {
		history[i] = PositionHistoryItem{
			PositionUUID: snap.PositionUUID,
			Symbol:       snap.Symbol,
			Side:         snap.Side,
			OpenedAt:     snap.PositionOpenedAt,
			ClosedAt:     snap.CreatedAt,
			EntryPrice:   snap.EntryPrice,
			ExitPrice:    snap.MarkPrice,
			FinalPL:      snap.UnrealizedPL,
			Reason:       "",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"positions": history,
	})
}

func handlePositionDecisions(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	positionUUID := r.URL.Query().Get("position_uuid")
	if positionUUID == "" {
		http.Error(w, "Position UUID required", http.StatusBadRequest)
		return
	}

	decisions, err := GetDecisionsByPositionUUID(positionUUID)
	if err != nil {
		log.Printf("Failed to get decisions for position %s: %v", positionUUID, err)
		decisions = []TradingDecisionRecord{}
	}

	type DecisionItem struct {
		ID                  uint      `json:"id"`
		Symbol              string    `json:"symbol"`
		BTCIchimoku         string    `json:"btc_ichimoku"`
		CoinIchimoku        string    `json:"coin_ichimoku"`
		Activity            string    `json:"activity"`
		FudActivity         string    `json:"fud_activity"`
		Sentiment           string    `json:"sentiment"`
		FudAttack           string    `json:"fud_attack"`
		FinalDecision       string    `json:"final_decision"`
		DecisionExplanation string    `json:"decision_explanation"`
		CreatedAt           time.Time `json:"created_at"`
	}

	decisionItems := make([]DecisionItem, len(decisions))
	for i, d := range decisions {
		decisionItems[i] = DecisionItem{
			ID:                  d.ID,
			Symbol:              d.Symbol,
			BTCIchimoku:         d.BTCIchimoku,
			CoinIchimoku:        d.CoinIchimoku,
			Activity:            d.Activity,
			FudActivity:         d.FudActivity,
			Sentiment:           d.Sentiment,
			FudAttack:           d.FudAttack,
			FinalDecision:       d.FinalDecision,
			DecisionExplanation: d.DecisionExplanation,
			CreatedAt:           d.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"decisions": decisionItems,
	})
}

func handlePositions(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 100
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	openPositions, err := GetOpenPositions()
	if err != nil {
		http.Error(w, "Failed to get open positions", http.StatusInternalServerError)
		return
	}

	closedPositions, err := GetClosedPositionsWithPagination(limit, offset)
	if err != nil {
		http.Error(w, "Failed to get closed positions", http.StatusInternalServerError)
		return
	}

	allPositions := append(openPositions, closedPositions...)

	type PositionItem struct {
		UUID              string     `json:"uuid"`
		Symbol            string     `json:"symbol"`
		Side              string     `json:"side"`
		Leverage          int        `json:"leverage"`
		Quantity          float64    `json:"quantity"`
		EntryPrice        float64    `json:"entry_price"`
		OpenedAt          time.Time  `json:"opened_at"`
		IsClosed          bool       `json:"is_closed"`
		ClosedAt          *time.Time `json:"closed_at"`
		ClosePrice        float64    `json:"close_price"`
		RealizedPL        float64    `json:"realized_pl"`
		CurrentPnL        float64    `json:"current_pnl"`
		CurrentPnLPercent float64    `json:"current_pnl_percent"`
		CurrentMarkPrice  float64    `json:"current_mark_price"`
		MaxPnL            float64    `json:"max_pnl"`
		MinPnL            float64    `json:"min_pnl"`
		Duration          int64      `json:"duration"`
		OpenReason        string     `json:"open_reason"`
		CloseReason       string     `json:"close_reason"`
	}

	positions := make([]PositionItem, len(allPositions))

	var totalProfitPositions float64
	var totalLossPositions float64
	var totalLongPositions float64
	var totalShortPositions float64
	var totalInitialMargin float64
	var totalLongInitialMargin float64
	var totalShortInitialMargin float64
	var totalProfitInitialMargin float64
	var totalLossInitialMargin float64

	for i, p := range allPositions {
		pnlPercent := 0.0
		initialMargin := 0.0
		if p.Quantity > 0 && p.Leverage > 0 && p.EntryPrice > 0 {
			initialMargin = (p.EntryPrice * p.Quantity) / float64(p.Leverage)
			if initialMargin > 0 {
				pnlPercent = (p.CurrentPnL / initialMargin) * 100
			}
		}

		totalInitialMargin += initialMargin

		positions[i] = PositionItem{
			UUID:              p.UUID,
			Symbol:            p.Symbol,
			Side:              p.Side,
			Leverage:          p.Leverage,
			Quantity:          p.Quantity,
			EntryPrice:        p.EntryPrice,
			OpenedAt:          p.OpenedAt,
			IsClosed:          p.IsClosed,
			ClosedAt:          p.ClosedAt,
			ClosePrice:        p.ClosePrice,
			RealizedPL:        p.RealizedPL,
			CurrentPnL:        p.CurrentPnL,
			CurrentPnLPercent: pnlPercent,
			CurrentMarkPrice:  p.CurrentMarkPrice,
			MaxPnL:            p.MaxPnL,
			MinPnL:            p.MinPnL,
			Duration:          p.Duration,
			OpenReason:        p.OpenReason,
			CloseReason:       p.CloseReason,
		}

		pnl := p.CurrentPnL
		if pnl >= 0 {
			totalProfitPositions += pnl
			totalProfitInitialMargin += initialMargin
		} else {
			totalLossPositions += pnl
			totalLossInitialMargin += initialMargin
		}

		if p.Side == "LONG" {
			totalLongPositions += pnl
			totalLongInitialMargin += initialMargin
		} else if p.Side == "SHORT" {
			totalShortPositions += pnl
			totalShortInitialMargin += initialMargin
		}
	}

	totalPnLPercent := 0.0
	if totalInitialMargin > 0 {
		totalPnLPercent = ((totalProfitPositions + totalLossPositions) / totalInitialMargin) * 100
	}

	totalProfitPercent := 0.0
	if totalProfitInitialMargin > 0 {
		totalProfitPercent = (totalProfitPositions / totalProfitInitialMargin) * 100
	}

	totalLossPercent := 0.0
	if totalLossInitialMargin > 0 {
		totalLossPercent = (totalLossPositions / totalLossInitialMargin) * 100
	}

	totalLongPercent := 0.0
	if totalLongInitialMargin > 0 {
		totalLongPercent = (totalLongPositions / totalLongInitialMargin) * 100
	}

	totalShortPercent := 0.0
	if totalShortInitialMargin > 0 {
		totalShortPercent = (totalShortPositions / totalShortInitialMargin) * 100
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"positions":               positions,
		"total_profit_positions":  totalProfitPositions,
		"total_loss_positions":    totalLossPositions,
		"total_pnl":               totalProfitPositions + totalLossPositions,
		"total_long_pnl":          totalLongPositions,
		"total_short_pnl":         totalShortPositions,
		"total_pnl_percent":       totalPnLPercent,
		"total_profit_percent":    totalProfitPercent,
		"total_loss_percent":      totalLossPercent,
		"total_long_pnl_percent":  totalLongPercent,
		"total_short_pnl_percent": totalShortPercent,
	})
}

func handlePositionSnapshots(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	positionUUID := r.URL.Query().Get("position_uuid")
	if positionUUID == "" {
		http.Error(w, "Position UUID required", http.StatusBadRequest)
		return
	}

	snapshots, err := GetPositionSnapshotsByUUID(positionUUID)
	if err != nil {
		log.Printf("Failed to get snapshots for position %s: %v", positionUUID, err)
		http.Error(w, "Failed to get snapshots", http.StatusInternalServerError)
		return
	}

	type SnapshotItem struct {
		ID           uint      `json:"id"`
		Symbol       string    `json:"symbol"`
		Side         string    `json:"side"`
		Leverage     int       `json:"leverage"`
		EntryPrice   float64   `json:"entry_price"`
		Amount       float64   `json:"amount"`
		UnrealizedPL float64   `json:"unrealized_pl"`
		MarkPrice    float64   `json:"mark_price"`
		CreatedAt    time.Time `json:"created_at"`
	}

	snapshotItems := make([]SnapshotItem, len(snapshots))
	for i, s := range snapshots {
		snapshotItems[i] = SnapshotItem{
			ID:           s.ID,
			Symbol:       s.Symbol,
			Side:         s.Side,
			Leverage:     s.Leverage,
			EntryPrice:   s.EntryPrice,
			Amount:       s.Amount,
			UnrealizedPL: s.UnrealizedPL,
			MarkPrice:    s.MarkPrice,
			CreatedAt:    s.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"snapshots": snapshotItems,
	})
}

func handleFudAttacks(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	positionUUID := r.URL.Query().Get("position_uuid")
	if positionUUID == "" {
		http.Error(w, "Position UUID required", http.StatusBadRequest)
		return
	}

	attacks, err := GetFudAttacksByPositionUUID(positionUUID)
	if err != nil {
		log.Printf("Failed to get FUD attacks for position %s: %v", positionUUID, err)
		attacks = []FudAttackRecord{}
	}

	type FudAttackItem struct {
		ID              uint      `json:"id"`
		Symbol          string    `json:"symbol"`
		HasAttack       bool      `json:"has_attack"`
		Confidence      float64   `json:"confidence"`
		MessageCount    int       `json:"message_count"`
		FudType         string    `json:"fud_type"`
		Theme           string    `json:"theme"`
		StartedHoursAgo int       `json:"started_hours_ago"`
		LastAttackTime  time.Time `json:"last_attack_time"`
		Justification   string    `json:"justification"`
		Participants    string    `json:"participants"`
		CreatedAt       time.Time `json:"created_at"`
	}

	attackItems := make([]FudAttackItem, len(attacks))
	for i, a := range attacks {
		attackItems[i] = FudAttackItem{
			ID:              a.ID,
			Symbol:          a.Symbol,
			HasAttack:       a.HasAttack,
			Confidence:      a.Confidence,
			MessageCount:    a.MessageCount,
			FudType:         a.FudType,
			Theme:           a.Theme,
			StartedHoursAgo: a.StartedHoursAgo,
			LastAttackTime:  a.LastAttackTime,
			Justification:   a.Justification,
			Participants:    a.Participants,
			CreatedAt:       a.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"fud_attacks": attackItems,
	})
}

func handlePnLHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	positions, err := GetAllClosedPositionsOrdered()
	if err != nil {
		http.Error(w, "Failed to get positions", http.StatusInternalServerError)
		return
	}

	hourlyMap := make(map[string]float64)
	for _, pos := range positions {
		if pos.ClosedAt == nil {
			continue
		}
		hourKey := pos.ClosedAt.Truncate(time.Hour).Format(time.RFC3339)
		hourlyMap[hourKey] += pos.CurrentPnL
	}

	keys := make([]string, 0, len(hourlyMap))
	for k := range hourlyMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	type PnLPoint struct {
		Timestamp     string  `json:"timestamp"`
		PnL           float64 `json:"pnl"`
		CumulativePnL float64 `json:"cumulative_pnl"`
	}

	result := make([]PnLPoint, 0)

	baseDate := time.Date(2024, 11, 3, 0, 0, 0, 0, time.UTC)
	for i := 6; i >= 0; i-- {
		dayTime := baseDate.AddDate(0, 0, -i)
		for hour := 0; hour < 24; hour++ {
			timestamp := dayTime.Add(time.Duration(hour) * time.Hour).Format(time.RFC3339)
			result = append(result, PnLPoint{
				Timestamp:     timestamp,
				PnL:           0,
				CumulativePnL: INITIAL_BALANCE,
			})
		}
	}

	cumulative := INITIAL_BALANCE
	for _, key := range keys {
		cumulative += hourlyMap[key]
		result = append(result, PnLPoint{
			Timestamp:     key,
			PnL:           hourlyMap[key],
			CumulativePnL: cumulative,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"history": result,
	})
}

func handleAIValidations(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	positionUUID := r.URL.Query().Get("position_uuid")
	decisionIDStr := r.URL.Query().Get("decision_id")

	var validations []AIOrderValidationRecord
	var err error

	if idStr != "" {
		id, parseErr := strconv.ParseUint(idStr, 10, 32)
		if parseErr != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		validation, getErr := GetAIValidationByID(uint(id))
		if getErr != nil {
			http.Error(w, "AI validation not found", http.StatusNotFound)
			return
		}
		validations = []AIOrderValidationRecord{*validation}
	} else if decisionIDStr != "" {
		decisionID, parseErr := strconv.ParseUint(decisionIDStr, 10, 32)
		if parseErr != nil {
			http.Error(w, "Invalid Decision ID", http.StatusBadRequest)
			return
		}
		validation, getErr := GetAIValidationByDecisionID(uint(decisionID))
		if getErr != nil {
			log.Printf("Failed to get AI validation for decision %d: %v", decisionID, getErr)
			validations = []AIOrderValidationRecord{}
		} else if validation != nil {
			validations = []AIOrderValidationRecord{*validation}
		} else {
			validations = []AIOrderValidationRecord{}
		}
	} else if positionUUID != "" {
		validations, err = GetAIValidationsByPositionUUID(positionUUID)
		if err != nil {
			log.Printf("Failed to get AI validations for position %s: %v", positionUUID, err)
			validations = []AIOrderValidationRecord{}
		}
	} else {
		http.Error(w, "ID, Decision ID or Position UUID required", http.StatusBadRequest)
		return
	}

	type AIValidationItem struct {
		ID                uint      `json:"id"`
		Symbol            string    `json:"symbol"`
		ShouldOpenOrder   bool      `json:"should_open_order"`
		ConfidencePercent float64   `json:"confidence_percent"`
		Justification     string    `json:"justification"`
		RequestData       string    `json:"request_data"`
		ResponseData      string    `json:"response_data"`
		CreatedAt         time.Time `json:"created_at"`
	}

	validationItems := make([]AIValidationItem, len(validations))
	for i, v := range validations {
		validationItems[i] = AIValidationItem{
			ID:                v.ID,
			Symbol:            v.Symbol,
			ShouldOpenOrder:   v.ShouldOpenOrder,
			ConfidencePercent: v.ConfidencePercent,
			Justification:     v.Justification,
			RequestData:       v.RequestData,
			ResponseData:      v.ResponseData,
			CreatedAt:         v.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ai_validations": validationItems,
	})
}

func handleRecentAIValidations(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	validations, err := GetRecentAIValidations(168)
	if err != nil {
		http.Error(w, "Failed to get AI validations", http.StatusInternalServerError)
		return
	}

	type AIValidationListItem struct {
		ID                uint      `json:"id"`
		Symbol            string    `json:"symbol"`
		PositionUUID      string    `json:"position_uuid"`
		ShouldOpenOrder   bool      `json:"should_open_order"`
		ConfidencePercent float64   `json:"confidence_percent"`
		CreatedAt         time.Time `json:"created_at"`
	}

	items := make([]AIValidationListItem, len(validations))
	for i, v := range validations {
		items[i] = AIValidationListItem{
			ID:                v.ID,
			Symbol:            v.Symbol,
			PositionUUID:      v.PositionUUID,
			ShouldOpenOrder:   v.ShouldOpenOrder,
			ConfidencePercent: v.ConfidencePercent,
			CreatedAt:         v.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"validations": items,
	})
}

func handleAICloseAnalyses(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	positionUUID := r.URL.Query().Get("position_uuid")

	var analyses []AiPositionCloseRecord
	var err error

	if idStr != "" {
		id, parseErr := strconv.ParseUint(idStr, 10, 32)
		if parseErr != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		analysis, getErr := GetAIPositionCloseByID(uint(id))
		if getErr != nil {
			http.Error(w, "AI close analysis not found", http.StatusNotFound)
			return
		}
		analyses = []AiPositionCloseRecord{*analysis}
	} else if positionUUID != "" {
		analyses, err = GetAIPositionClosesByUUID(positionUUID)
		if err != nil {
			log.Printf("Failed to get AI close analyses for position %s: %v", positionUUID, err)
			analyses = []AiPositionCloseRecord{}
		}
	} else {
		limit := 20
		offset := 0
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}
		if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
			if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
				offset = parsedOffset
			}
		}

		analyses, err = GetAIPositionClosesWithPagination(limit, offset)
		if err != nil {
			log.Printf("Failed to get AI close analyses: %s", err)
			http.Error(w, "Failed to get AI close analyses", http.StatusMethodNotAllowed)
			return
		}
	}

	type AICloseAnalysisItem struct {
		ID                uint      `json:"id"`
		Symbol            string    `json:"symbol"`
		SnapshotCount     int       `json:"snapshot_count"`
		ShouldClose       bool      `json:"should_close"`
		ConfidencePercent float64   `json:"confidence_percent"`
		Justification     string    `json:"justification"`
		ExpectedPnL       float64   `json:"expected_pnl"`
		RiskAssessment    string    `json:"risk_assessment"`
		RequestData       string    `json:"request_data"`
		ResponseData      string    `json:"response_data"`
		CreatedAt         time.Time `json:"created_at"`
	}

	items := make([]AICloseAnalysisItem, len(analyses))
	for i, a := range analyses {
		items[i] = AICloseAnalysisItem{
			ID:                a.ID,
			Symbol:            a.Symbol,
			SnapshotCount:     a.SnapshotCount,
			ShouldClose:       a.ShouldClose,
			ConfidencePercent: a.ConfidencePercent,
			Justification:     a.Justification,
			ExpectedPnL:       a.ExpectedPnL,
			RiskAssessment:    a.RiskAssessment,
			RequestData:       a.RequestData,
			ResponseData:      a.ResponseData,
			CreatedAt:         a.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ai_close_analyses": items,
	})
}

func handleAICloseAnalysesByPosition(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	positionUUID := r.URL.Query().Get("position_uuid")
	if positionUUID == "" {
		http.Error(w, "Position UUID required", http.StatusBadRequest)
		return
	}

	analyses, err := GetAIPositionClosesByUUID(positionUUID)
	if err != nil {
		log.Printf("Failed to get AI close analyses for position %s: %v", positionUUID, err)
		analyses = []AiPositionCloseRecord{}
	}

	type AICloseAnalysisItem struct {
		ID                uint      `json:"id"`
		PositionUUID      string    `json:"position_uuid"`
		Symbol            string    `json:"symbol"`
		SnapshotCount     int       `json:"snapshot_count"`
		ShouldClose       bool      `json:"should_close"`
		ConfidencePercent float64   `json:"confidence_percent"`
		Justification     string    `json:"justification"`
		ExpectedPnL       float64   `json:"expected_pnl"`
		RiskAssessment    string    `json:"risk_assessment"`
		RequestData       string    `json:"request_data"`
		ResponseData      string    `json:"response_data"`
		CreatedAt         time.Time `json:"created_at"`
	}

	items := make([]AICloseAnalysisItem, len(analyses))
	for i, a := range analyses {
		items[i] = AICloseAnalysisItem{
			ID:                a.ID,
			PositionUUID:      a.PositionUUID,
			Symbol:            a.Symbol,
			SnapshotCount:     a.SnapshotCount,
			ShouldClose:       a.ShouldClose,
			ConfidencePercent: a.ConfidencePercent,
			Justification:     a.Justification,
			ExpectedPnL:       a.ExpectedPnL,
			RiskAssessment:    a.RiskAssessment,
			RequestData:       a.RequestData,
			ResponseData:      a.ResponseData,
			CreatedAt:         a.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ai_close_analyses": items,
	})
}
