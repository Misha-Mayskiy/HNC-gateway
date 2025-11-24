package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

// Config holds application configuration loaded from environment variables
type Config struct {
	GRPCPort            string `env:"GRPC_PORT" env-default:":50052" yaml:"grpc_port"`
	RedisAddr           string `env:"REDIS_ADDR" env-default:"localhost:6379" yaml:"redis_addr"`
	CustomerServiceAddr string `env:"CUSTOMER_SERVICE_ADDR" env-default:"localhost:50051" yaml:"customer_service_addr"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}
	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
