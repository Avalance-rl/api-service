package cache

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

const ttl time.Duration = 180 * time.Second

type cache struct {
	client *redis.Client
}

func New(client *redis.Client) *cache {
	return &cache{
		client: client,
	}
}

func (c *cache) GetPrice(ctx context.Context, symbol string) (float64, error) {
	val, err := c.client.Get(ctx, symbol).Result()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(val, 64)
}
func (c *cache) SetPrice(ctx context.Context, symbol string, price float64) error {
	return c.client.Set(ctx, symbol, price, ttl).Err()
}
