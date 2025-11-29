package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"api-gateway/config"
	customerclient "api-gateway/internal/clients/customer"
	grpcserver "api-gateway/internal/grpc/server"

	"api-gateway/internal/infrastructure/kafka"
	"api-gateway/internal/service"
	redisstorage "api-gateway/internal/storage/redis"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Redis
	store, err := redisstorage.New(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("failed to init redis storage: %v", err)
	}

	// Kafka Producer
	producer, err := kafka.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopic)
	if err != nil {
		log.Fatalf("failed to init kafka producer: %v", err)
	}
	defer producer.Close()
	log.Println("✅ Kafka producer initialized")

	// Customer Client
	client, err := customerclient.New(cfg.CustomerServiceAddr)
	if err != nil {
		log.Fatalf("failed to init customer client: %v", cfg.CustomerServiceAddr)
	}
	defer client.Close()

	// Service
	svc := service.New(store, client, producer)

	// Server
	srv := grpcserver.New(svc, client)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("shutting down gRPC server")
		os.Exit(0)
	}()

	if err := grpcserver.Run(cfg.GRPCPort, srv); err != nil {
		log.Fatalf("failed to run gRPC server: %v", err)
	}
}
