package repository

import (
	"context"
	"cryptobot/api-service/internal/adapter/repository/cache"
	"github.com/redis/go-redis/v9"
)

type CurrencyRepository interface {
	GetPrice(ctx context.Context, symbol string) (float64, error)
	SetPrice(ctx context.Context, symbol string, price float64) error
}

func NewCurrencyRepository(client *redis.Client) CurrencyRepository {
	return cache.New(client)
}
