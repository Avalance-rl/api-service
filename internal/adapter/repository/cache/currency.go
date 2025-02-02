package cache

import (
	"context"
	"errors"
	"github.com/avalance-rl/cryptobot/services/api-service/internal/domain/entity"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
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

func (c *currencyRepository) FindPrice(ctx context.Context, name string) (float64, error) {
	val, err := c.client.Get(ctx, name).Result()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(val, 64)
}

func (c *currencyRepository) SetPrice(ctx context.Context, currency entity.Currency) error {
	return c.client.Set(ctx, currency.Name, currency.Price, ttl).Err()
}
