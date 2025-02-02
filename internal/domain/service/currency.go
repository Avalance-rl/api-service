package service

import (
	"context"
	"github.com/avalance-rl/cryptobot/services/api-service/internal/domain/entity"
)

type CurrencyFinder interface {
	FindPrice(ctx context.Context, name string) (*entity.Currency, error)
}

type CurrencySetter interface {
	SetPrice(ctx context.Context, currency entity.Currency) error
}

type currencyService struct {
	currencyFinder CurrencyFinder
	currencySetter CurrencySetter
}

func NewCurrencyService(f CurrencyFinder, s CurrencySetter) *currencyService {
	return &currencyService{
		currencyFinder: f,
		currencySetter: s,
	}
}

func (s *currencyService) GetPrice(ctx context.Context, name string) (*entity.Currency, error) {
	return s.currencyFinder.FindPrice(ctx, name)
}

func (s *currencyService) SavePrice(ctx context.Context, currency entity.Currency) error {
	return s.currencySetter.SetPrice(ctx, currency)
}
