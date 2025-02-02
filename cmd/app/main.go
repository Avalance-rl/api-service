package main

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/avalance-rl/cryptobot-pkg/logger"
	"github.com/avalance-rl/cryptobot/services/api-service/internal/adapter/repository"
	"github.com/avalance-rl/cryptobot/services/api-service/internal/config"
	"github.com/avalance-rl/cryptobot/services/api-service/internal/domain/service"
	"github.com/avalance-rl/cryptobot/services/api-service/internal/domain/usecase/currency"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	// Инициализируем логгер
	log := logger.New()
	defer log.Sync()

	// Загружаем конфигурацию
	cfg, err := config.Load(os.Getenv("CONFIG_FILE"))
	if err != nil {
		log.Fatal(err.Error())
	}

	// Настраиваем Kafka producer
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Retry.Max = 5
	kafkaConfig.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(cfg.Kafka.Brokers, kafkaConfig)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer producer.Close()

	// Создаем провайдер цен
	priceProvider := currency.NewExchangeProvider()

	rdb := redis.Client{}
	repo := repository.NewCurrencyRepository(&rdb)
	svc := service.NewCurrencyService(repo, repo)

	usc := currency.New(
		svc,
		producer,
		log,
		cfg.Kafka.UpdateInterval,
		cfg.Kafka.Currencies,
		cfg.Kafka.Topic,
		cfg.Kafka.RetryAttempts,
		cfg.Kafka.RetryDelay,
		priceProvider,
	)

	// Запускаем административный сервер для метрик
	go func() {
		if err := runAdminServer(":8081"); err != nil {
			log.Fatal(err.Error())
		}
	}()

	// Запускаем основной сервис
	go func() {
		if err := usc.Run(ctx); err != nil {
			log.Fatal(err.Error())
		}
	}()

	log.Info(
		"service started",
		zap.Strings("currencies", cfg.Kafka.Currencies),
		zap.Duration("interval", cfg.Kafka.UpdateInterval),
	)

	<-ctx.Done()
	log.Info("shutdown signal received")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := usc.Shutdown(shutdownCtx); err != nil {
		log.Fatal("error shutting down service", zap.Error(err))
	}

	log.Info("service shutdown completed")
}

func runAdminServer(addr string) error {
	mux := http.NewServeMux()

	// Хелсчек эндпоинт
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Метрики Prometheus
	mux.Handle("/metrics", promhttp.Handler())

	return http.ListenAndServe(addr, mux)
}
