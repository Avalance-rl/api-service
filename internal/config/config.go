package config

import (
	"fmt"
	"time"

	baseConfig "github.com/avalance-rl/cryptobot/pkg/config"
	"github.com/spf13/viper"
)

type ApiConfig struct {
	*baseConfig.Config
	Redis struct {
		Host string
		Port string
		Pass string
		DB   int
	}
	Kafka struct {
		Brokers        []string
		Topic          string
		UpdateInterval time.Duration
		Currencies     []string
		RetryAttempts  int
		RetryDelay     time.Duration
	}
	APIKey string
}

func Load(path string) (*ApiConfig, error) {
	baseCFG, err := baseConfig.Load(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load base config: %w", err)
	}

	config := &ApiConfig{
		Config: baseCFG,
	}

	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read api config: %w", err)
	}

	if err := viper.UnmarshalKey("Redis", &config.Redis); err != nil {
		return nil, fmt.Errorf("failed to unmarshal redis config: %w", err)
	}

	if err := viper.UnmarshalKey("Kafka", &config.Kafka); err != nil {
		return nil, fmt.Errorf("failed to unmarshal kafka config: %w", err)
	}

	if err := viper.UnmarshalKey("APIKey", &config.APIKey); err != nil {
		return nil, fmt.Errorf("failed to unmarshal api_key config: %w", err)
	}

	return config, nil
}
