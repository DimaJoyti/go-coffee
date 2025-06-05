package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api_gateway/client"
	"api_gateway/config"
	"api_gateway/server"
)

func initProducerClient(cfg *config.Config) (*client.CoffeeClient, error) {
	var lastErr error
	for retries := 0; retries <= cfg.GRPC.MaxRetries; retries++ {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.GRPC.ConnectionTimeout)
		client := client.NewCoffeeClient(cfg.GRPC.ProducerAddress)

		if err := client.Connect(ctx); err == nil {
			cancel()
			return client, nil
		} else {
			cancel()
			lastErr = err
			if retries < cfg.GRPC.MaxRetries {
				log.Printf("Failed to connect to Producer service, retrying in %v... (%d/%d)",
					cfg.GRPC.RetryDelay, retries+1, cfg.GRPC.MaxRetries)
				time.Sleep(cfg.GRPC.RetryDelay)
			}
		}
	}
	return nil, fmt.Errorf("failed to connect after %d retries: %w", cfg.GRPC.MaxRetries, lastErr)
}

func main() {
	// Завантаження конфігурації
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Створення gRPC клієнта для Producer сервісу
	producerClient, err := initProducerClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize producer client: %v", err)
	}
	defer producerClient.Close()

	// Створення HTTP сервера
	httpServer := server.NewHTTPServer(cfg, producerClient)

	// Запуск HTTP сервера в окремій горутині
	go func() {
		serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
		log.Printf("Starting HTTP server on %s", serverAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Налаштування обробки сигналів для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
