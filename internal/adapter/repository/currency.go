package repository

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/avalance-rl/cryptobot/services/api-service/internal/domain/entity"
	"github.com/redis/go-redis/v9"
)

const ttl time.Duration = 60 * time.Second

type currencyRepository struct {
	client *redis.Client
}

func NewCurrencyRepository(client *redis.Client) *currencyRepository {
	return &currencyRepository{
		client: client,
	}
}

func (c *currencyRepository) FindPrice(ctx context.Context, name string) (*entity.Currency, error) {
	val, err := c.client.Get(ctx, name).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	price, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return nil, err
	}

	return &entity.Currency{
		Name:  name,
		Price: price,
	}, nil
}

func (c *currencyRepository) SetPrice(ctx context.Context, currency entity.Currency) error {
	return c.client.Set(ctx, currency.Name, currency.Price, ttl).Err()
}
