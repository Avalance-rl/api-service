package currency

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	"github.com/avalance-rl/cryptobot/pkg/logger"
	"github.com/avalance-rl/cryptobot/services/api-service/internal/domain/entity"
	"go.uber.org/zap"
)

type Service interface {
	GetPrice(ctx context.Context, name string) (*entity.Currency, error)
	SavePrice(ctx context.Context, currency entity.Currency) error
}

type PriceProvider interface {
	FetchPrice(ctx context.Context, name string) (*entity.Currency, error)
}

type config struct {
	updateInterval time.Duration
	currencies     []string
	kafkaTopic     string
	retryAttempts  int
	retryDelay     time.Duration
}

type usecase struct {
	//service  Service
	producer sarama.SyncProducer
	provider PriceProvider
	log      *logger.Logger
	config   config
	metrics  *Metrics
}

func New(
	//service Service,
	producer sarama.SyncProducer,
	logger *logger.Logger,
	updateInterval time.Duration,
	currencies []string,
	kafkaTopic string,
	retryAttempts int,
	retryDelay time.Duration,
	provider PriceProvider,
) *usecase {
	return &usecase{
		//service:  service,
		producer: producer,
		log:      logger,
		config: config{
			updateInterval: updateInterval,
			currencies:     currencies,
			kafkaTopic:     kafkaTopic,
			retryAttempts:  retryAttempts,
			retryDelay:     retryDelay,
		},
		provider: provider,
		metrics:  newMetrics(),
	}
}

func (u *usecase) Run(ctx context.Context) error {
	u.log.Info("starting currency service")
	ticker := time.NewTicker(u.config.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			u.log.Info("context canceled", zap.Error(ctx.Err()))
			return ctx.Err()
		case <-ticker.C:

			if err := u.updatePrices(ctx); err != nil {
				u.log.Error("error updating prices", zap.Error(err))
			}
		}
	}
}

func (u *usecase) updatePrices(ctx context.Context) error {
	start := time.Now()
	defer func() {
		u.metrics.updateDuration.Observe(time.Since(start).Seconds())
	}()

	for _, symbol := range u.config.currencies {
		if err := u.updateCurrencyPrice(ctx, symbol); err != nil {
			u.log.Error("error updating currency price", zap.Error(err))
			continue
		}
	}

	return nil
}

func (u *usecase) updateCurrencyPrice(ctx context.Context, name string) error {
	currency, err := u.provider.FetchPrice(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to fetch currency: %w", err)
	}

	//if err := u.service.SavePrice(ctx, *currency); err != nil {
	//	return fmt.Errorf("failed to save currency to cache: %w", err)
	//}

	if err := u.publishPriceWithRetry(ctx, *currency); err != nil {
		return fmt.Errorf("failed to publish currency: %w", err)
	}

	u.metrics.priceCounter.Inc()
	u.log.Debug("price updated", zap.String("name", name))

	return nil
}

func (u *usecase) publishPriceWithRetry(ctx context.Context, currency entity.Currency) error {
	var lastErr error
	for attempt := 0; attempt < u.config.retryAttempts; attempt++ {
		if err := u.publishCurrency(ctx, currency); err != nil {
			lastErr = err
			u.log.Warn("error publishing currency price", zap.Error(err))

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(u.config.retryDelay):
				continue
			}
		}
		return nil
	}
	return fmt.Errorf("failed to publish price after %d attempts: %w",
		u.config.retryAttempts, lastErr)
}

func (u *usecase) publishCurrency(ctx context.Context, currency entity.Currency) error {
	data, err := json.Marshal(currency)
	if err != nil {
		return fmt.Errorf("failed to marshal price: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: u.config.kafkaTopic,
		Key:   sarama.StringEncoder(currency.Name),
		Value: sarama.ByteEncoder(data),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("timestamp"),
				Value: []byte(time.Now().UTC().Format(time.RFC3339)),
			},
			{
				Key:   []byte("version"),
				Value: []byte("1.0"),
			},
		},
		Timestamp: time.Now(),
	}

	partition, offset, err := u.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	u.log.Debug(
		"price published to kafka",
		zap.String("name", currency.Name),
		zap.Int64("offset", offset),
		zap.String("partition", strconv.FormatInt(int64(partition), 10)),
		zap.String("topic", u.config.kafkaTopic),
	)

	return nil
}

func (u *usecase) Shutdown(ctx context.Context) error {
	u.log.Info("shutting down service")

	if err := u.producer.Close(); err != nil {
		return fmt.Errorf("failed to close kafka producer: %w", err)
	}

	return nil
}
