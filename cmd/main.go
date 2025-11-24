package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/shvdev1/HackNeChange/api-gateway/config"
	customerclient "github.com/shvdev1/HackNeChange/api-gateway/internal/clients/customer"
	grpcserver "github.com/shvdev1/HackNeChange/api-gateway/internal/grpc/server"
	"github.com/shvdev1/HackNeChange/api-gateway/internal/service"
	redisstorage "github.com/shvdev1/HackNeChange/api-gateway/internal/storage/redis"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// init redis storage
	store, err := redisstorage.New(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("failed to init redis storage: %v", err)
	}

	client, err := customerclient.New(cfg.CustomerServiceAddr)
	if err != nil {
		log.Fatalf("failed to init customer client: %v", err)
	}
	defer client.Close()

	svc := service.New(store, client)
	srv := grpcserver.New(svc, client)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("shutting down gRPC server")
		// TODO: graceful stop handled by context or grpc server reference
		os.Exit(0)
	}()

	if err := grpcserver.Run(cfg.GRPCPort, srv); err != nil {
		log.Fatalf("failed to run gRPC server: %v", err)
	}
}
