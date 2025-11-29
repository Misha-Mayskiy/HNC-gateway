package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

// Config holds application configuration loaded from environment variables
type Config struct {
	GRPCPort            string   `env:"GRPC_PORT" env-default:":50052" yaml:"grpc_port"`
	RedisAddr           string   `env:"REDIS_ADDR" env-default:"localhost:6379" yaml:"redis_addr"`
	CustomerServiceAddr string   `env:"CUSTOMER_SERVICE_ADDR" env-default:"localhost:50051" yaml:"customer_service_addr"`
	KafkaBrokers        []string `env:"KAFKA_BROKERS" env-default:"localhost:9092" yaml:"kafka_brokers"`
	KafkaTopic          string   `env:"KAFKA_TOPIC" env-default:"reviews.raw" yaml:"kafka_topic"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}
	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
