package main

import (
	"encoding/json"
	"log"
	"net/http"
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

	decisions, err := GetRecentDecisions(168)
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

	openPositions, err := GetOpenPositions()
	if err != nil {
		http.Error(w, "Failed to get open positions", http.StatusInternalServerError)
		return
	}

	closedPositions, err := GetClosedPositions(168)
	if err != nil {
		http.Error(w, "Failed to get closed positions", http.StatusInternalServerError)
		return
	}

	allPositions := append(openPositions, closedPositions...)

	type PositionItem struct {
		UUID             string     `json:"uuid"`
		Symbol           string     `json:"symbol"`
		Side             string     `json:"side"`
		Leverage         int        `json:"leverage"`
		Quantity         float64    `json:"quantity"`
		EntryPrice       float64    `json:"entry_price"`
		OpenedAt         time.Time  `json:"opened_at"`
		IsClosed         bool       `json:"is_closed"`
		ClosedAt         *time.Time `json:"closed_at"`
		ClosePrice       float64    `json:"close_price"`
		RealizedPL       float64    `json:"realized_pl"`
		CurrentPnL       float64    `json:"current_pnl"`
		CurrentMarkPrice float64    `json:"current_mark_price"`
		Duration         int64      `json:"duration"`
		OpenReason       string     `json:"open_reason"`
		CloseReason      string     `json:"close_reason"`
	}

	positions := make([]PositionItem, len(allPositions))
	for i, p := range allPositions {
		positions[i] = PositionItem{
			UUID:             p.UUID,
			Symbol:           p.Symbol,
			Side:             p.Side,
			Leverage:         p.Leverage,
			Quantity:         p.Quantity,
			EntryPrice:       p.EntryPrice,
			OpenedAt:         p.OpenedAt,
			IsClosed:         p.IsClosed,
			ClosedAt:         p.ClosedAt,
			ClosePrice:       p.ClosePrice,
			RealizedPL:       p.RealizedPL,
			CurrentPnL:       p.CurrentPnL,
			CurrentMarkPrice: p.CurrentMarkPrice,
			Duration:         p.Duration,
			OpenReason:       p.OpenReason,
			CloseReason:      p.CloseReason,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"positions": positions,
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
