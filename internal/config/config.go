package config

import (
	"fmt"
	baseConfig "github.com/avalance-rl/cryptobot/pkg/config"
	"github.com/spf13/viper"
)

type ApiConfig struct {
	*baseConfig.Config
	Redis struct {
		Host string
		Port int
	}
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

	if err := viper.UnmarshalKey("jwt", &config.Redis); err != nil {
		return nil, fmt.Errorf("failed to unmarshal redis config: %w", err)
	}
	return config, nil
}
