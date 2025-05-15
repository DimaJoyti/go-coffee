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

	"github.com/yourusername/coffee-order-system/pkg/config"
	"github.com/yourusername/coffee-order-system/pkg/kafka"
	"github.com/yourusername/coffee-order-system/pkg/logger"

	"kafka_producer/internal/handler"
	"kafka_producer/internal/service"
	"kafka_producer/internal/repository"
)

func main() {
	// Ініціалізація логера
	logConfig := logger.DefaultConfig()
	log := logger.NewLogger(logConfig)
	log.Info("Starting producer service...")

	// Завантаження конфігурації
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration: %v", err)
	}

	// Створення Kafka producer
	kafkaProducer, err := createKafkaProducer(cfg)
	if err != nil {
		log.Fatal("Failed to create Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	// Створення сховища замовлень
	orderRepo := repository.NewOrderRepository()

	// Створення сервісу
	orderService := service.NewOrderService(kafkaProducer, orderRepo)

	// Створення HTTP обробників
	h := handler.NewHandler(orderService, log)

	// Створення HTTP сервера
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: createHTTPHandler(h),
	}

	// Запуск HTTP сервера в горутині
	go func() {
		log.Info("Starting HTTP server on port %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start HTTP server: %v", err)
		}
	}()

	// Запуск gRPC сервера в горутині
	go func() {
		log.Info("Starting gRPC server on port %d", cfg.GRPC.Port)
		if err := startGRPCServer(orderService, cfg.GRPC.Port); err != nil {
			log.Fatal("Failed to start gRPC server: %v", err)
		}
	}()

	// Очікування сигналу завершення
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down servers...")

	// Створення контексту з таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Зупинка HTTP сервера
	if err := server.Shutdown(ctx); err != nil {
		log.Error("HTTP server shutdown error: %v", err)
	}

	// Зупинка gRPC сервера (реалізація в startGRPCServer)

	log.Info("Servers stopped")
}

// loadConfig завантажує конфігурацію
func loadConfig() (*Config, error) {
	// Реалізація завантаження конфігурації
	// ...
	return &Config{}, nil
}

// createKafkaProducer створює Kafka producer
func createKafkaProducer(cfg *Config) (kafka.Producer, error) {
	// Реалізація створення Kafka producer
	// ...
	return nil, nil
}

// createHTTPHandler створює HTTP обробник з middleware
func createHTTPHandler(h *handler.Handler) http.Handler {
	// Реалізація створення HTTP обробника з middleware
	// ...
	return nil
}

// startGRPCServer запускає gRPC сервер
func startGRPCServer(orderService *service.OrderService, port int) error {
	// Реалізація запуску gRPC сервера
	// ...
	return nil
}

// Config представляє конфігурацію сервісу
type Config struct {
	Server struct {
		Port int
	}
	GRPC struct {
		Port int
	}
	Kafka struct {
		Brokers      []string
		Topic        string
		RetryMax     int
		RequiredAcks string
	}
}
