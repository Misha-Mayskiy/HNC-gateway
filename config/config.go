package config

import (
	"os"
	"strconv"
)

// Config holds service configuration values
type Config struct {
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
	ServerPort    string
}

// LoadConfig reads configuration from environment variables with sensible defaults.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		RedisHost:     getEnv("REDIS_HOST", "127.0.0.1"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		ServerPort:    getEnv("SERVER_PORT", "8080"),
	}

	dbStr := getEnv("REDIS_DB", "0")
	db, err := strconv.Atoi(dbStr)
	if err != nil {
		return nil, err
	}
	cfg.RedisDB = db

	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
