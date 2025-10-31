package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	tradingStates    = make(map[string]*TradingState)
	statesMutex      sync.RWMutex
	balanceHistory   []BalancePoint
	positionsHistory []PositionRecord
	positionDetails  map[string]PositionDetail
)

func handleAPIRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api")

	switch {
	case strings.HasPrefix(path, "/status"):
		handleStatus(w, r)
	case strings.HasPrefix(path, "/positions"):
		handlePositions(w, r)
	case strings.HasPrefix(path, "/pairs"):
		handlePairs(w, r)
	case strings.HasPrefix(path, "/balance-history"):
		handleBalanceHistory(w, r)
	case strings.HasPrefix(path, "/positions-list"):
		handlePositionsList(w, r)
	case strings.HasPrefix(path, "/position-detail"):
		handlePositionDetail(w, r)
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

func handlePositions(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	statesMutex.RLock()
	defer statesMutex.RUnlock()

	positions := make([]map[string]interface{}, 0)
	for symbol, state := range tradingStates {
		positions = append(positions, map[string]interface{}{
			"symbol":   symbol,
			"position": state.CurrentPosition,
			"opened_at": func() interface{} {
				if state.CurrentPosition != PositionSideBoth {
					return state.OpenedAt
				}
				return nil
			}(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"positions": positions,
	})
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

type BalancePoint struct {
	Timestamp int64   `json:"timestamp"`
	Balance   float64 `json:"balance"`
}

type PositionRecord struct {
	ID        string    `json:"id"`
	Date      time.Time `json:"date"`
	Symbol    string    `json:"symbol"`
	Amount    float64   `json:"amount"`
	Direction string    `json:"direction"`
	Reason    string    `json:"reason"`
	Status    string    `json:"status"`
	Result    float64   `json:"result"`
}

type PositionDetail struct {
	ID          string                `json:"id"`
	Symbol      string                `json:"symbol"`
	Amount      float64               `json:"amount"`
	CoinsAmount float64               `json:"coins_amount"`
	Direction   string                `json:"direction"`
	OpenedAt    time.Time             `json:"opened_at"`
	ClosedAt    *time.Time            `json:"closed_at"`
	Result      float64               `json:"result"`
	History     []PositionHistoryItem `json:"history"`
}

type PositionHistoryItem struct {
	Timestamp     time.Time     `json:"timestamp"`
	Action        string        `json:"action"`
	Reason        string        `json:"reason"`
	Amount        float64       `json:"amount"`
	CoinsAmount   float64       `json:"coins_amount"`
	TrendAnalysis TrendAnalysis `json:"trend_analysis"`
}

type TrendAnalysis struct {
	BitcoinTrend   string `json:"bitcoin_trend"`
	CoinTrend      string `json:"coin_trend"`
	ActivityTrend  string `json:"activity_trend"`
	FudSignal      bool   `json:"fud_signal"`
	SentimentTrend string `json:"sentiment_trend"`
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

	records, err := GetBalanceHistory("USDT", 168)
	if err != nil {
		http.Error(w, "Failed to get balance history", http.StatusInternalServerError)
		return
	}

	history := make([]BalancePoint, len(records))
	for i, record := range records {
		history[i] = BalancePoint{
			Timestamp: record.Timestamp.UnixMilli(),
			Balance:   record.Balance,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"history": history,
	})
}

func handlePositionsList(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	statusFilter := r.URL.Query().Get("status")

	statesMutex.RLock()
	defer statesMutex.RUnlock()

	positions := []PositionRecord{}

	for symbol, state := range tradingStates {
		if state.CurrentPosition != PositionSideBoth {
			snapshot, err := GetLatestPositionSnapshot(symbol)
			if err != nil || snapshot == nil {
				continue
			}

			if statusFilter == "" || statusFilter == "active" {
				positions = append(positions, PositionRecord{
					ID:        symbol + "_active",
					Date:      state.OpenedAt,
					Symbol:    symbol,
					Amount:    snapshot.UnrealizedPL,
					Direction: string(state.CurrentPosition),
					Reason:    state.OpenReason,
					Status:    "active",
					Result:    snapshot.UnrealizedPL,
				})
			}
		}
	}

	if statusFilter == "" || statusFilter == "closed" {
		for _, pos := range positionsHistory {
			if pos.Status == "closed" {
				positions = append(positions, pos)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"positions": positions,
	})
}

func handlePositionDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	positionID := r.URL.Query().Get("id")
	if positionID == "" {
		http.Error(w, "Position ID required", http.StatusBadRequest)
		return
	}

	detail, exists := positionDetails[positionID]
	if !exists {
		http.Error(w, "Position not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}

func init() {
	generateBalanceHistory()
	generatePositionsHistory()
}

func generateBalanceHistory() {
	startBalance := 65.0
	currentBalance := 55.0
	now := time.Now()

	for i := 0; i < 168; i++ {
		timestamp := now.Add(-time.Duration(168-i) * time.Hour)
		progress := float64(i) / 167.0
		balance := startBalance + (currentBalance-startBalance)*progress + (rand.Float64()*2-1)*0.5

		balanceHistory = append(balanceHistory, BalancePoint{
			Timestamp: timestamp.UnixMilli(),
			Balance:   balance,
		})
	}
}

func generatePositionsHistory() {
	positionDetails = make(map[string]PositionDetail)
	reasons := []string{"ichimoku", "community", "fud", "sentiment"}
	directions := []string{"LONG", "SHORT"}
	trends := []string{"up", "down", "flat"}

	now := time.Now()
	positionCounter := 1

	for pairIndex, pair := range TradingPairs {
		activePositionAmount := 20.0
		activeDirection := directions[rand.Intn(len(directions))]
		activeReason := reasons[rand.Intn(len(reasons))]
		activeDate := now.Add(-time.Duration(rand.Intn(12)+1) * time.Hour)
		activeID := "pos_" + strconv.Itoa(positionCounter)
		positionCounter++

		positionsHistory = append(positionsHistory, PositionRecord{
			ID:        activeID,
			Date:      activeDate,
			Symbol:    pair.Symbol,
			Amount:    activePositionAmount,
			Direction: activeDirection,
			Reason:    activeReason,
			Status:    "active",
			Result:    rand.Float64()*2 - 1,
		})

		historyItems := make([]PositionHistoryItem, 0)
		activeCoinTrend := "up"
		if activeDirection == "SHORT" {
			activeCoinTrend = "down"
		}
		activeSentimentTrend := "up"
		if activeDirection == "SHORT" {
			activeSentimentTrend = "down"
		}
		activeFudSignal := false
		if activeDirection == "SHORT" && rand.Float64() > 0.5 {
			activeFudSignal = true
		}

		historyItems = append(historyItems, PositionHistoryItem{
			Timestamp:   activeDate,
			Action:      "open",
			Reason:      "Position opened based on " + activeReason + " signal. Strong indicators detected across multiple timeframes.",
			Amount:      activePositionAmount,
			CoinsAmount: pair.Quantity,
			TrendAnalysis: TrendAnalysis{
				BitcoinTrend:   trends[rand.Intn(len(trends))],
				CoinTrend:      activeCoinTrend,
				ActivityTrend:  trends[rand.Intn(len(trends))],
				FudSignal:      activeFudSignal,
				SentimentTrend: activeSentimentTrend,
			},
		})

		numUpdates := rand.Intn(3) + 1
		for j := 0; j < numUpdates; j++ {
			updateTime := activeDate.Add(time.Duration(j+1) * time.Hour * 2)
			updateReasons := []string{
				"Market conditions remain favorable, continuing to hold position",
				"Trend analysis shows strengthening momentum",
				"Community sentiment remains positive, maintaining exposure",
				"Technical indicators confirm original thesis",
			}
			historyItems = append(historyItems, PositionHistoryItem{
				Timestamp:   updateTime,
				Action:      "update",
				Reason:      updateReasons[rand.Intn(len(updateReasons))],
				Amount:      activePositionAmount,
				CoinsAmount: pair.Quantity,
				TrendAnalysis: TrendAnalysis{
					BitcoinTrend:   trends[rand.Intn(len(trends))],
					CoinTrend:      trends[rand.Intn(len(trends))],
					ActivityTrend:  trends[rand.Intn(len(trends))],
					FudSignal:      rand.Float64() > 0.7,
					SentimentTrend: trends[rand.Intn(len(trends))],
				},
			})
		}

		positionDetails[activeID] = PositionDetail{
			ID:          activeID,
			Symbol:      pair.Symbol,
			Amount:      activePositionAmount,
			CoinsAmount: pair.Quantity,
			Direction:   activeDirection,
			OpenedAt:    activeDate,
			ClosedAt:    nil,
			Result:      rand.Float64()*2 - 1,
			History:     historyItems,
		}

		closedPositionsCount := 30 + pairIndex*3
		for i := 0; i < closedPositionsCount; i++ {
			id := "pos_" + strconv.Itoa(positionCounter)
			positionCounter++
			direction := directions[rand.Intn(len(directions))]
			reason := reasons[rand.Intn(len(reasons))]

			date := now.Add(-time.Duration(rand.Intn(720)+24) * time.Hour)
			amount := 15.0 + rand.Float64()*15.0
			result := rand.Float64()*4.0 - 2.0

			positionsHistory = append(positionsHistory, PositionRecord{
				ID:        id,
				Date:      date,
				Symbol:    pair.Symbol,
				Amount:    amount,
				Direction: direction,
				Reason:    reason,
				Status:    "closed",
				Result:    result,
			})

			closedHistoryItems := make([]PositionHistoryItem, 0)
			openReasons := []string{
				"Position opened based on " + reason + " signal. Market analysis indicated favorable conditions.",
				"Strong " + reason + " signal detected. Opening " + direction + " position with confidence.",
				"Multiple indicators aligned with " + reason + " strategy. Entry point confirmed.",
			}
			coinTrend := "up"
			if direction == "SHORT" {
				coinTrend = "down"
			}
			sentimentTrend := "up"
			if direction == "SHORT" {
				sentimentTrend = "down"
			}
			fudSignal := false
			if direction == "SHORT" && rand.Float64() > 0.5 {
				fudSignal = true
			}

			closedHistoryItems = append(closedHistoryItems, PositionHistoryItem{
				Timestamp:   date,
				Action:      "open",
				Reason:      openReasons[rand.Intn(len(openReasons))],
				Amount:      amount,
				CoinsAmount: pair.Quantity,
				TrendAnalysis: TrendAnalysis{
					BitcoinTrend:   trends[rand.Intn(len(trends))],
					CoinTrend:      coinTrend,
					ActivityTrend:  trends[rand.Intn(len(trends))],
					FudSignal:      fudSignal,
					SentimentTrend: sentimentTrend,
				},
			})

			numUpdates := rand.Intn(4) + 1
			for j := 0; j < numUpdates; j++ {
				updateTime := date.Add(time.Duration(j+1) * time.Hour * 3)
				updateReasons := []string{
					"Market conditions evolving, monitoring position closely",
					"Price action showing expected behavior, holding position",
					"Volatility increase detected, adjusting risk parameters",
					"Trend continuation confirmed by technical analysis",
				}
				closedHistoryItems = append(closedHistoryItems, PositionHistoryItem{
					Timestamp:   updateTime,
					Action:      "update",
					Reason:      updateReasons[rand.Intn(len(updateReasons))],
					Amount:      amount,
					CoinsAmount: pair.Quantity,
					TrendAnalysis: TrendAnalysis{
						BitcoinTrend:   trends[rand.Intn(len(trends))],
						CoinTrend:      trends[rand.Intn(len(trends))],
						ActivityTrend:  trends[rand.Intn(len(trends))],
						FudSignal:      rand.Float64() > 0.7,
						SentimentTrend: trends[rand.Intn(len(trends))],
					},
				})
			}

			closeTime := date.Add(time.Duration(numUpdates+2) * time.Hour * 3)
			closeReasons := []string{
				"Target reached, closing position with profit of " + strconv.FormatFloat(result, 'f', 2, 64) + " USDT",
				"Stop loss triggered, closing position with result: " + strconv.FormatFloat(result, 'f', 2, 64) + " USDT",
				"Market conditions changed, exiting position with " + strconv.FormatFloat(result, 'f', 2, 64) + " USDT",
				"Position duration limit reached, closing with result: " + strconv.FormatFloat(result, 'f', 2, 64) + " USDT",
			}
			closedHistoryItems = append(closedHistoryItems, PositionHistoryItem{
				Timestamp:   closeTime,
				Action:      "close",
				Reason:      closeReasons[rand.Intn(len(closeReasons))],
				Amount:      amount,
				CoinsAmount: pair.Quantity,
				TrendAnalysis: TrendAnalysis{
					BitcoinTrend:   trends[rand.Intn(len(trends))],
					CoinTrend:      trends[rand.Intn(len(trends))],
					ActivityTrend:  trends[rand.Intn(len(trends))],
					FudSignal:      rand.Float64() > 0.7,
					SentimentTrend: trends[rand.Intn(len(trends))],
				},
			})

			positionDetails[id] = PositionDetail{
				ID:          id,
				Symbol:      pair.Symbol,
				Amount:      amount,
				CoinsAmount: pair.Quantity,
				Direction:   direction,
				OpenedAt:    date,
				ClosedAt:    &closeTime,
				Result:      result,
				History:     closedHistoryItems,
			}
		}
	}
}
