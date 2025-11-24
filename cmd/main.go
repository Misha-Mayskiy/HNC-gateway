package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/shvdev1/HackNeChange/api-gateway/config"
	redisstore "github.com/shvdev1/HackNeChange/api-gateway/internal/storage/redis"
	rest "github.com/shvdev1/HackNeChange/api-gateway/internal/transport/rest"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	addr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// quick ping to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect to redis at %s: %v", addr, err)
	}

	store := redisstore.NewSettingsStorage(rdb)
	handler := rest.NewSettingsHandler(store)

	router := gin.Default()
	api := router.Group("/api/v1")
	handler.RegisterRoutes(api)

	serverAddr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("starting server on %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
