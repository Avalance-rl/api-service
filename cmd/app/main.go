package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/avalance-rl/cryptobot/pkg/logger"
	"github.com/avalance-rl/cryptobot/services/api-service/internal/config"
	"github.com/avalance-rl/cryptobot/services/api-service/internal/domain/usecase/currency"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	log := logger.New()
	defer log.Sync()

	cfg, err := config.Load(os.Getenv("CONFIG_FILE"))
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Info(cfg.APIKey)

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Retry.Max = 5
	kafkaConfig.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(cfg.Kafka.Brokers, kafkaConfig)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer producer.Close()
	fmt.Println(cfg.APIKey)
	priceProvider := currency.NewExchangeProvider(cfg.APIKey)

	//rdb := redis.Client{}
	//repo := repository.NewCurrencyRepository(&rdb)
	//svc := service.NewCurrencyService(repo, repo)

	usc := currency.New(
		//svc,
		producer,
		log,
		cfg.Kafka.UpdateInterval,
		cfg.Kafka.Currencies,
		cfg.Kafka.Topic,
		cfg.Kafka.RetryAttempts,
		cfg.Kafka.RetryDelay,
		priceProvider,
	)

	go func() {
		if err := runAdminServer(":8081"); err != nil {
			log.Fatal(err.Error())
		}
	}()

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

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := usc.Shutdown(shutdownCtx); err != nil {
		log.Fatal("error shutting down service", zap.Error(err))
	}

	log.Info("service shutdown completed")
}

func runAdminServer(addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.Handle("/metrics", promhttp.Handler())

	return http.ListenAndServe(addr, mux)
}
