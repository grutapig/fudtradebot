package main

import (
	"context"
	"errors"
)

type ExchangeInterface interface {
	Buy(ctx context.Context, tokenAddress string, amountIn float64) (string, error)
	Sell(ctx context.Context, tokenAddress string, amountOut float64) (string, error)
	GetBalance(ctx context.Context, tokenAddress string) (float64, error)
	GetNativeBalance(ctx context.Context) (float64, error)
	GetTokenPrice(ctx context.Context, tokenAddress string) (float64, error)
}

type ExchangeService struct {
	privateKey string
	rpcURL     string
}

func NewExchangeService(privateKey string, rpcURL string) (*ExchangeService, error) {
	if privateKey == "" || rpcURL == "" {
		return nil, errors.New("private key and RPC URL are required")
	}

	return &ExchangeService{
		privateKey: privateKey,
		rpcURL:     rpcURL,
	}, nil
}

func (s *ExchangeService) Buy(ctx context.Context, tokenAddress string, amountIn float64) (string, error) {
	return "", errors.New("not implemented")
}

func (s *ExchangeService) Sell(ctx context.Context, tokenAddress string, amountOut float64) (string, error) {
	return "", errors.New("not implemented")
}

func (s *ExchangeService) GetBalance(ctx context.Context, tokenAddress string) (float64, error) {
	return 0, errors.New("not implemented")
}

func (s *ExchangeService) GetNativeBalance(ctx context.Context) (float64, error) {
	return 0, errors.New("not implemented")
}

func (s *ExchangeService) GetTokenPrice(ctx context.Context, tokenAddress string) (float64, error) {
	return 0, errors.New("not implemented")
}
