package config

import baseConfig "github.com/avalance-rl/cryptobot/pkg/config"

type ApiConfig struct {
	*baseConfig.Config
	Redis struct {
		Host string
		Port int
	}
}
