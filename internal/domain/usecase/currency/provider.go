package currency

import (
	"context"
	"github.com/avalance-rl/cryptobot/services/api-service/internal/domain/entity"
)

type ExchangeProvider struct {
	// здесь могут быть клиенты для различных бирж
}

func NewExchangeProvider() *ExchangeProvider {
	return &ExchangeProvider{}
}

func (p *ExchangeProvider) FetchPrice(ctx context.Context, name string) (*entity.Currency, error) {
	return &entity.Currency{
		Name:  name,
		Price: 50000.0,
	}, nil
}
