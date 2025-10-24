package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/grutapig/fudtradebot/claude"
)

type BusinessLogicService struct {
	dbInfo       *DatabaseInfoService
	dbTrade      *DatabaseTradeService
	exchange     ExchangeInterface
	telegram     *TelegramService
	claudeClient *claude.ClaudeApi
}

func NewBusinessLogicService(
	dbInfo *DatabaseInfoService,
	dbTrade *DatabaseTradeService,
	exchange ExchangeInterface,
	telegram *TelegramService,
	claudeClient *claude.ClaudeApi,
) *BusinessLogicService {
	return &BusinessLogicService{
		dbInfo:       dbInfo,
		dbTrade:      dbTrade,
		exchange:     exchange,
		telegram:     telegram,
		claudeClient: claudeClient,
	}
}

func (s *BusinessLogicService) BuyToken(ctx context.Context, symbol string, amount float64) error {
	token, err := s.dbInfo.GetTokenInfo(symbol)
	if err != nil {
		return fmt.Errorf("get token info: %w", err)
	}

	trade := &Trade{
		TokenSymbol: symbol,
		Side:        "buy",
		Amount:      amount,
		Status:      "pending",
		Timestamp:   time.Now(),
	}

	if err := s.dbTrade.CreateTrade(trade); err != nil {
		return fmt.Errorf("create trade record: %w", err)
	}

	txHash, err := s.exchange.Buy(ctx, token.Address, amount)
	if err != nil {
		s.dbTrade.UpdateTradeStatus(trade.ID, "failed", "", err.Error())
		return fmt.Errorf("execute buy: %w", err)
	}

	s.dbTrade.UpdateTradeStatus(trade.ID, "completed", txHash, "")

	price, _ := s.exchange.GetTokenPrice(ctx, token.Address)
	s.updatePosition(symbol, amount, price, "buy")

	s.telegram.SendNotification(fmt.Sprintf("✅ Bought %.4f %s", amount, symbol))

	return nil
}

func (s *BusinessLogicService) SellToken(ctx context.Context, symbol string, amount float64) error {
	token, err := s.dbInfo.GetTokenInfo(symbol)
	if err != nil {
		return fmt.Errorf("get token info: %w", err)
	}

	trade := &Trade{
		TokenSymbol: symbol,
		Side:        "sell",
		Amount:      amount,
		Status:      "pending",
		Timestamp:   time.Now(),
	}

	if err := s.dbTrade.CreateTrade(trade); err != nil {
		return fmt.Errorf("create trade record: %w", err)
	}

	txHash, err := s.exchange.Sell(ctx, token.Address, amount)
	if err != nil {
		s.dbTrade.UpdateTradeStatus(trade.ID, "failed", "", err.Error())
		return fmt.Errorf("execute sell: %w", err)
	}

	s.dbTrade.UpdateTradeStatus(trade.ID, "completed", txHash, "")

	price, _ := s.exchange.GetTokenPrice(ctx, token.Address)
	s.updatePosition(symbol, -amount, price, "sell")

	s.telegram.SendNotification(fmt.Sprintf("✅ Sold %.4f %s", amount, symbol))

	return nil
}

func (s *BusinessLogicService) GetBalance(ctx context.Context, symbol string) (*BalanceInfo, error) {
	token, err := s.dbInfo.GetTokenInfo(symbol)
	if err != nil {
		return nil, fmt.Errorf("get token info: %w", err)
	}

	balance, err := s.exchange.GetBalance(ctx, token.Address)
	if err != nil {
		return nil, fmt.Errorf("get balance: %w", err)
	}

	price, _ := s.exchange.GetTokenPrice(ctx, token.Address)

	return &BalanceInfo{
		TokenSymbol: symbol,
		Amount:      balance,
		ValueUSD:    balance * price,
	}, nil
}

func (s *BusinessLogicService) CalculateBuySize(ctx context.Context, symbol string, confidence float64) (float64, error) {
	nativeBalance, err := s.exchange.GetNativeBalance(ctx)
	if err != nil {
		return 0, err
	}

	maxSize := nativeBalance * 0.1
	size := maxSize * confidence

	if size < nativeBalance*0.01 {
		size = nativeBalance * 0.01
	}

	return size, nil
}

func (s *BusinessLogicService) CalculateSellSize(ctx context.Context, symbol string, confidence float64) (float64, error) {
	position, err := s.dbTrade.GetPosition(symbol)
	if err != nil {
		return 0, err
	}

	maxSize := position.Amount
	size := maxSize * confidence

	if size < position.Amount*0.1 {
		size = position.Amount * 0.1
	}

	return size, nil
}

func (s *BusinessLogicService) AnalyzeMarketData(ctx context.Context, symbol string) (*AnalysisResult, error) {
	promptData, err := os.ReadFile(PROMPT_FILE_ANALYZE)
	if err != nil {
		return nil, fmt.Errorf("read prompt file: %w", err)
	}

	priceHistory, err := s.dbInfo.GetPriceHistory(symbol, time.Now().Add(-24*time.Hour), time.Now())
	if err != nil {
		return nil, fmt.Errorf("get price history: %w", err)
	}

	indicators, err := s.dbInfo.GetLatestIndicators(symbol)
	if err != nil {
		return nil, fmt.Errorf("get indicators: %w", err)
	}

	dataContext := fmt.Sprintf("Symbol: %s\nPrice History (24h): %v\nIndicators: RSI=%.2f MACD=%.2f",
		symbol, priceHistory, indicators.RSI, indicators.MACD)

	messages := claude.ClaudeMessages{
		{Role: claude.ROLE_USER, Content: dataContext},
	}

	response, err := s.claudeClient.SendMessage(messages, string(promptData))
	if err != nil {
		return nil, fmt.Errorf("claude analysis: %w", err)
	}

	return &AnalysisResult{
		TokenSymbol: symbol,
		Analysis:    response.Content[0].Text,
		Timestamp:   time.Now(),
	}, nil
}

func (s *BusinessLogicService) MakeDecision(ctx context.Context, symbol string) (*TradeDecision, error) {
	promptData, err := os.ReadFile(PROMPT_FILE_DECISION)
	if err != nil {
		return nil, fmt.Errorf("read prompt file: %w", err)
	}

	analysis, err := s.AnalyzeMarketData(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("analyze market: %w", err)
	}

	messages := claude.ClaudeMessages{
		{Role: claude.ROLE_USER, Content: analysis.Analysis},
	}

	response, err := s.claudeClient.SendMessage(messages, string(promptData))
	if err != nil {
		return nil, fmt.Errorf("claude decision: %w", err)
	}

	return &TradeDecision{
		TokenSymbol: symbol,
		Action:      response.Content[0].Text,
		Timestamp:   time.Now(),
	}, nil
}

func (s *BusinessLogicService) updatePosition(symbol string, amount float64, price float64, side string) {
	position, err := s.dbTrade.GetPosition(symbol)
	if err != nil {
		position = &Position{
			TokenSymbol:  symbol,
			Amount:       0,
			AvgBuyPrice:  0,
			CurrentPrice: price,
			IsActive:     true,
		}
	}

	if side == "buy" {
		totalValue := (position.Amount * position.AvgBuyPrice) + (amount * price)
		position.Amount += amount
		if position.Amount > 0 {
			position.AvgBuyPrice = totalValue / position.Amount
		}
	} else {
		position.Amount -= amount
	}

	position.CurrentPrice = price
	position.ProfitLossUSD = (price - position.AvgBuyPrice) * position.Amount
	if position.AvgBuyPrice > 0 {
		position.ProfitLossPerc = ((price - position.AvgBuyPrice) / position.AvgBuyPrice) * 100
	}

	if position.Amount <= 0.0001 {
		s.dbTrade.ClosePosition(symbol)
	} else {
		s.dbTrade.UpsertPosition(position)
	}
}

func (s *BusinessLogicService) ProcessTelegramCommands(ctx context.Context) {
	messageCh := s.telegram.GetMessageChannel()

	for msg := range messageCh {
		if !msg.IsReply {
			continue
		}

		switch msg.Text {
		case "balance":
			s.handleBalanceCommand(ctx, msg.ChatID)
		case "positions":
			s.handlePositionsCommand(ctx, msg.ChatID)
		case "status":
			s.handleStatusCommand(ctx, msg.ChatID)
		}
	}
}

func (s *BusinessLogicService) handleBalanceCommand(ctx context.Context, chatID int64) {
	balance, err := s.exchange.GetNativeBalance(ctx)
	if err != nil {
		s.telegram.Reply(chatID, fmt.Sprintf("Error: %v", err))
		return
	}
	s.telegram.Reply(chatID, fmt.Sprintf("Balance: %.4f", balance))
}

func (s *BusinessLogicService) handlePositionsCommand(ctx context.Context, chatID int64) {
	positions, err := s.dbTrade.GetAllActivePositions()
	if err != nil {
		s.telegram.Reply(chatID, fmt.Sprintf("Error: %v", err))
		return
	}

	if len(positions) == 0 {
		s.telegram.Reply(chatID, "No active positions")
		return
	}

	response := "Active Positions:\n"
	for _, pos := range positions {
		response += fmt.Sprintf("%s: %.4f (P/L: %.2f%%)\n", pos.TokenSymbol, pos.Amount, pos.ProfitLossPerc)
	}
	s.telegram.Reply(chatID, response)
}

func (s *BusinessLogicService) handleStatusCommand(ctx context.Context, chatID int64) {
	s.telegram.Reply(chatID, "Bot is running")
}

func (s *BusinessLogicService) ProcessSignals(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			signals, err := s.dbTrade.GetUnprocessedSignals()
			if err != nil {
				log.Printf("Error getting signals: %v", err)
				continue
			}

			for _, signal := range signals {
				s.processSignal(ctx, &signal)
				s.dbTrade.MarkSignalProcessed(signal.ID)
			}
		}
	}
}

func (s *BusinessLogicService) processSignal(ctx context.Context, signal *TradingSignal) {
	decision, err := s.MakeDecision(ctx, signal.TokenSymbol)
	if err != nil {
		log.Printf("Error making decision for %s: %v", signal.TokenSymbol, err)
		return
	}

	log.Printf("Decision for %s: %s", signal.TokenSymbol, decision.Action)
}
