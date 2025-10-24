package main

import (
	"context"
	"log"
	"time"
)

type MonitoringService struct {
	dbInfo   *DatabaseInfoService
	dbTrade  *DatabaseTradeService
	business *BusinessLogicService
}

func NewMonitoringService(
	dbInfo *DatabaseInfoService,
	dbTrade *DatabaseTradeService,
	business *BusinessLogicService,
) *MonitoringService {
	return &MonitoringService{
		dbInfo:   dbInfo,
		dbTrade:  dbTrade,
		business: business,
	}
}

func (s *MonitoringService) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Println("Monitoring service started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Monitoring service stopped")
			return
		case <-ticker.C:
			s.monitorMarkets(ctx)
		}
	}
}

func (s *MonitoringService) monitorMarkets(ctx context.Context) {
	tokens, err := s.dbInfo.GetActiveTokens()
	if err != nil {
		log.Printf("Error getting active tokens: %v", err)
		return
	}

	for _, token := range tokens {
		s.analyzeToken(ctx, token.Symbol)
	}
}

func (s *MonitoringService) analyzeToken(ctx context.Context, symbol string) {
	analysis, err := s.business.AnalyzeMarketData(ctx, symbol)
	if err != nil {
		log.Printf("Error analyzing %s: %v", symbol, err)
		return
	}

	log.Printf("Analysis for %s: %s", symbol, analysis.Analysis[:min(100, len(analysis.Analysis))])
}

func (s *MonitoringService) MonitorPositions(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	log.Println("Position monitoring started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Position monitoring stopped")
			return
		case <-ticker.C:
			s.updatePositions(ctx)
		}
	}
}

func (s *MonitoringService) updatePositions(ctx context.Context) {
	positions, err := s.dbTrade.GetAllActivePositions()
	if err != nil {
		log.Printf("Error getting positions: %v", err)
		return
	}

	for _, position := range positions {
		log.Printf("Position %s: %.4f (P/L: %.2f%%)", position.TokenSymbol, position.Amount, position.ProfitLossPerc)
	}
}
