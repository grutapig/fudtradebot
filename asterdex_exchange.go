package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	AsterDexBaseURL = "https://fapi.asterdex.com"
)

type AsterDexExchange struct {
	apiKey    string
	secretKey string
	client    *http.Client
}

// AsterDex API Response structures
type AsterDexPosition struct {
	Symbol           string `json:"symbol"`
	PositionSide     string `json:"positionSide"`
	PositionAmt      string `json:"positionAmt"`
	EntryPrice       string `json:"entryPrice"`
	UnrealizedProfit string `json:"unRealizedProfit"`
	Leverage         string `json:"leverage"`
}

type AsterDexBalance struct {
	AccountAlias       string `json:"accountAlias"`
	Asset              string `json:"asset"`
	Balance            string `json:"balance"`
	CrossWalletBalance string `json:"crossWalletBalance"`
	CrossUnPnl         string `json:"crossUnPnl"`
	AvailableBalance   string `json:"availableBalance"`
	MaxWithdrawAmount  string `json:"maxWithdrawAmount"`
	MarginAvailable    bool   `json:"marginAvailable"`
	UpdateTime         int64  `json:"updateTime"`
}

type AsterDexMarkPrice struct {
	Symbol    string `json:"symbol"`
	MarkPrice string `json:"markPrice"`
	Time      int64  `json:"time"`
}

type AsterDexOrderResponse struct {
	OrderID    int64  `json:"orderId"`
	Symbol     string `json:"symbol"`
	Status     string `json:"status"`
	Side       string `json:"side"`
	Type       string `json:"type"`
	OrigQty    string `json:"origQty"`
	Price      string `json:"price"`
	AvgPrice   string `json:"avgPrice"`
	UpdateTime int64  `json:"updateTime"`
}

type AsterDexKline struct {
	OpenTime       int64
	Open           string
	High           string
	Low            string
	Close          string
	Volume         string
	CloseTime      int64
	QuoteVolume    string
	NumberOfTrades int
	TakerBuyBase   string
	TakerBuyQuote  string
}

type AsterDexIndexKline struct {
	OpenTime            int64  `json:"openTime"`
	Open                string `json:"open"`
	High                string `json:"high"`
	Low                 string `json:"low"`
	Close               string `json:"close"`
	Volume              string `json:"volume"`
	CloseTime           int64  `json:"closeTime"`
	QuoteAssetVolume    string `json:"quoteAssetVolume"`
	NumberOfTrades      int    `json:"numberOfTrades"`
	TakerBuyBaseVolume  string `json:"takerBuyBaseVolume"`
	TakerBuyQuoteVolume string `json:"takerBuyQuoteVolume"`
	Ignore              string `json:"ignore"`
}

func NewAsterDexExchange(apiKey, secretKey string) AsterDexExchange {
	return AsterDexExchange{
		apiKey:    apiKey,
		secretKey: secretKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func NewAsterDexExchangeWithProxy(apiKey, secretKey, proxyDSN string) (AsterDexExchange, error) {
	transport := &http.Transport{}
	if proxyDSN != "" {
		proxyURL, err := url.Parse(proxyDSN)
		if err != nil {
			return AsterDexExchange{}, fmt.Errorf("new asterdex exchange proxy dsn error: %s", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	return AsterDexExchange{
		apiKey:    apiKey,
		secretKey: secretKey,
		client: &http.Client{
			Transport: transport,
			Timeout:   10 * time.Second,
		},
	}, nil
}

// generateSignature creates HMAC SHA256 signature for signed endpoints
func (e *AsterDexExchange) generateSignature(params string) string {
	mac := hmac.New(sha256.New, []byte(e.secretKey))
	mac.Write([]byte(params))
	return hex.EncodeToString(mac.Sum(nil))
}

// doRequest performs HTTP request with authentication
func (e *AsterDexExchange) doRequest(method, endpoint, params string, signed bool) ([]byte, error) {
	url := AsterDexBaseURL + endpoint

	if signed {
		timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
		if params != "" {
			params += "&timestamp=" + timestamp
		} else {
			params = "timestamp=" + timestamp
		}
		signature := e.generateSignature(params)
		params += "&signature=" + signature
	}

	if method == "GET" && params != "" {
		url += "?" + params
	}

	var req *http.Request
	var err error

	if method == "POST" || method == "PUT" || method == "DELETE" {
		if params != "" {
			req, err = http.NewRequest(method, url, bytes.NewBufferString(params))
		} else {
			req, err = http.NewRequest(method, url, nil)
		}
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return nil, err
		}
	}

	if signed {
		req.Header.Set("X-MBX-APIKEY", e.apiKey)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error [%d]: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GetPositionMode gets current position mode (Hedge or One-way)
func (e *AsterDexExchange) GetPositionMode() (bool, error) {
	body, err := e.doRequest("GET", "/fapi/v1/positionSide/dual", "", true)
	if err != nil {
		return false, err
	}

	var result struct {
		DualSidePosition bool `json:"dualSidePosition"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, err
	}

	return result.DualSidePosition, nil
}

// SetPositionMode sets position mode (Hedge or One-way)
func (e *AsterDexExchange) SetPositionMode(hedgeMode bool) error {
	params := fmt.Sprintf("dualSidePosition=%t", hedgeMode)
	_, err := e.doRequest("POST", "/fapi/v1/positionSide/dual", params, true)
	return err
}

// SetLeverage sets leverage for a symbol
func (e *AsterDexExchange) SetLeverage(symbol string, leverage int) error {
	params := fmt.Sprintf("symbol=%s&leverage=%d", symbol, leverage)
	_, err := e.doRequest("POST", "/fapi/v1/leverage", params, true)
	return err
}

func (e *AsterDexExchange) OpenPosition(symbol string, side PositionSide, leverage int, quantity float64) (*Position, error) {
	// Check and set position mode if needed
	isHedgeMode, err := e.GetPositionMode()
	if err != nil {
		return nil, fmt.Errorf("failed to get position mode: %w", err)
	}

	// If not in Hedge Mode and we're using LONG/SHORT, enable it
	if !isHedgeMode && (side == PositionSideLong || side == PositionSideShort) {
		if err := e.SetPositionMode(true); err != nil {
			return nil, fmt.Errorf("failed to enable hedge mode: %w", err)
		}
	}

	// Set leverage
	if err := e.SetLeverage(symbol, leverage); err != nil {
		return nil, fmt.Errorf("failed to set leverage: %w", err)
	}

	// Determine order side based on position side
	var orderSide string
	if side == PositionSideLong {
		orderSide = "BUY"
	} else {
		orderSide = "SELL"
	}

	// Place market order with positionSide parameter
	params := fmt.Sprintf("symbol=%s&side=%s&type=MARKET&quantity=%.2f&positionSide=%s",
		symbol, orderSide, quantity, side)

	body, err := e.doRequest("POST", "/fapi/v1/order", params, true)
	if err != nil {
		return nil, fmt.Errorf("failed to open position: %w", err)
	}

	var orderResp AsterDexOrderResponse
	if err := json.Unmarshal(body, &orderResp); err != nil {
		return nil, fmt.Errorf("failed to parse order response: %w", err)
	}

	// Get the position info
	position, err := e.GetPosition(symbol)
	if err != nil {
		return nil, fmt.Errorf("position opened but failed to fetch details: %w", err)
	}

	return position, nil
}

func (e *AsterDexExchange) ClosePosition(symbol string, side PositionSide) error {
	// Get current position to know the amount
	position, err := e.GetPosition(symbol)
	if err != nil {
		return fmt.Errorf("failed to get position: %w", err)
	}

	if position == nil || position.Amount == 0 {
		return fmt.Errorf("no open position found for %s", symbol)
	}

	// Determine order side (opposite of position side)
	var orderSide string
	if side == PositionSideLong {
		orderSide = "SELL"
	} else {
		orderSide = "BUY"
	}

	// Close position with market order
	params := fmt.Sprintf("symbol=%s&side=%s&type=MARKET&positionSide=%s&quantity=%.8f",
		symbol, orderSide, side, math.Abs(position.Amount))
	fmt.Println(params)
	_, err = e.doRequest("POST", "/fapi/v1/order", params, true)
	if err != nil {
		return fmt.Errorf("failed to close position: %w", err)
	}

	return nil
}

// GetPosition retrieves position information for a specific symbol
func (e *AsterDexExchange) GetPosition(symbol string) (*Position, error) {
	params := fmt.Sprintf("symbol=%s", symbol)
	body, err := e.doRequest("GET", "/fapi/v2/positionRisk", params, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get position: %w", err)
	}

	var positions []AsterDexPosition
	if err := json.Unmarshal(body, &positions); err != nil {
		return nil, fmt.Errorf("failed to parse positions: %w", err)
	}

	// Find the position with non-zero amount
	for _, pos := range positions {
		amount, _ := strconv.ParseFloat(pos.PositionAmt, 64)
		if amount != 0 {
			entryPrice, _ := strconv.ParseFloat(pos.EntryPrice, 64)
			unrealizedPL, _ := strconv.ParseFloat(pos.UnrealizedProfit, 64)
			leverage, _ := strconv.Atoi(pos.Leverage)

			return &Position{
				Symbol:       pos.Symbol,
				Side:         PositionSide(pos.PositionSide),
				Leverage:     leverage,
				EntryPrice:   entryPrice,
				Amount:       amount,
				UnrealizedPL: unrealizedPL,
				Timestamp:    time.Now(),
			}, nil
		}
	}

	return nil, nil // No open position
}

// GetAllPositions retrieves all open positions
func (e *AsterDexExchange) GetAllPositions() ([]*Position, error) {
	body, err := e.doRequest("GET", "/fapi/v2/positionRisk", "", true)
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	var adePositions []AsterDexPosition
	if err := json.Unmarshal(body, &adePositions); err != nil {
		return nil, fmt.Errorf("failed to parse positions: %w", err)
	}

	var positions []*Position
	for _, pos := range adePositions {
		amount, _ := strconv.ParseFloat(pos.PositionAmt, 64)
		if amount != 0 {
			entryPrice, _ := strconv.ParseFloat(pos.EntryPrice, 64)
			unrealizedPL, _ := strconv.ParseFloat(pos.UnrealizedProfit, 64)
			leverage, _ := strconv.Atoi(pos.Leverage)

			positions = append(positions, &Position{
				Symbol:       pos.Symbol,
				Side:         PositionSide(pos.PositionSide),
				Leverage:     leverage,
				EntryPrice:   entryPrice,
				Amount:       amount,
				UnrealizedPL: unrealizedPL,
				Timestamp:    time.Now(),
			})
		}
	}

	return positions, nil
}

func (e *AsterDexExchange) GetMarkPrice(symbol string) (float64, error) {
	params := fmt.Sprintf("symbol=%s", symbol)
	body, err := e.doRequest("GET", "/fapi/v1/premiumIndex", params, false)
	if err != nil {
		return 0, fmt.Errorf("failed to get mark price: %w", err)
	}

	var markPrice AsterDexMarkPrice
	if err := json.Unmarshal(body, &markPrice); err != nil {
		return 0, fmt.Errorf("failed to parse mark price: %w", err)
	}

	price, err := strconv.ParseFloat(markPrice.MarkPrice, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse price value: %w", err)
	}

	return price, nil
}

type AccountBalanceInfo struct {
	AccountAlias       string
	Asset              string
	Balance            float64
	CrossWalletBalance float64
	CrossUnPnl         float64
	AvailableBalance   float64
	MaxWithdrawAmount  float64
	MarginAvailable    bool
	UpdateTime         int64
}

func (e *AsterDexExchange) GetBalance() (float64, error) {
	infos, err := e.GetAllBalances()
	if err != nil {
		return 0, err
	}
	for _, info := range infos {
		if info.Asset == "USDT" {
			return info.Balance, nil
		}
	}
	return 0, fmt.Errorf("USDT balance not found")
}

func (e *AsterDexExchange) GetBalanceInfo() (AccountBalanceInfo, error) {
	infos, err := e.GetAllBalances()
	if err != nil {
		return AccountBalanceInfo{}, err
	}
	for _, info := range infos {
		if info.Asset == "USDT" {
			return info, nil
		}
	}
	return AccountBalanceInfo{}, fmt.Errorf("USDT balance not found")
}

func (e *AsterDexExchange) GetAllBalances() ([]AccountBalanceInfo, error) {
	body, err := e.doRequest("GET", "/fapi/v2/balance", "", true)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	var balances []AsterDexBalance
	if err := json.Unmarshal(body, &balances); err != nil {
		return nil, fmt.Errorf("failed to parse balance: %w", err)
	}

	var result []AccountBalanceInfo
	for _, bal := range balances {
		balance, err := strconv.ParseFloat(bal.Balance, 64)
		if err != nil {
			continue
		}

		crossWalletBalance, _ := strconv.ParseFloat(bal.CrossWalletBalance, 64)
		crossUnPnl, _ := strconv.ParseFloat(bal.CrossUnPnl, 64)
		availableBalance, _ := strconv.ParseFloat(bal.AvailableBalance, 64)
		maxWithdrawAmount, _ := strconv.ParseFloat(bal.MaxWithdrawAmount, 64)

		result = append(result, AccountBalanceInfo{
			AccountAlias:       bal.AccountAlias,
			Asset:              bal.Asset,
			Balance:            balance,
			CrossWalletBalance: crossWalletBalance,
			CrossUnPnl:         crossUnPnl,
			AvailableBalance:   availableBalance,
			MaxWithdrawAmount:  maxWithdrawAmount,
			MarginAvailable:    bal.MarginAvailable,
			UpdateTime:         bal.UpdateTime,
		})
	}

	return result, nil
}

func (e *AsterDexExchange) Klines(pair string, interval string, startTime, endTime int64, limit int) ([]AsterDexKline, error) {
	params := fmt.Sprintf("symbol=%s&interval=%s", pair, interval)

	if startTime > 0 {
		params += fmt.Sprintf("&startTime=%d", startTime)
	}
	if endTime > 0 {
		params += fmt.Sprintf("&endTime=%d", endTime)
	}
	if limit > 0 {
		params += fmt.Sprintf("&limit=%d", limit)
	}

	body, err := e.doRequest("GET", "/fapi/v1/klines", params, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get klines: %w", err)
	}

	var rawKlines [][]interface{}
	if err := json.Unmarshal(body, &rawKlines); err != nil {
		return nil, fmt.Errorf("failed to parse klines: %w", err)
	}

	klines := make([]AsterDexKline, len(rawKlines))
	for i, raw := range rawKlines {
		klines[i] = AsterDexKline{
			OpenTime:       int64(raw[0].(float64)),
			Open:           raw[1].(string),
			High:           raw[2].(string),
			Low:            raw[3].(string),
			Close:          raw[4].(string),
			Volume:         raw[5].(string),
			CloseTime:      int64(raw[6].(float64)),
			QuoteVolume:    raw[7].(string),
			NumberOfTrades: int(raw[8].(float64)),
			TakerBuyBase:   raw[9].(string),
			TakerBuyQuote:  raw[10].(string),
		}
	}

	return klines, nil
}
