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

	"github.com/golang-migrate/migrate/v4"
	pgmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/config"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/kafka"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/repository/postgres"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/server"
	"github.com/yourusername/coffee-order-system/accounts-service/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to the database
	dbConfig := postgres.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}

	db, err := postgres.NewDatabase(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run database migrations
	if err := runMigrations(db, cfg.Database); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create Kafka producer
	kafkaProducer, err := kafka.NewKafkaProducer(cfg)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	// Create repositories
	accountRepo := postgres.NewAccountRepository(db)
	vendorRepo := postgres.NewVendorRepository(db)
	productRepo := postgres.NewProductRepository(db)
	orderRepo := postgres.NewOrderRepository(db)
	orderItemRepo := postgres.NewOrderItemRepository(db)

	// Create services
	accountService := service.NewAccountService(accountRepo)
	vendorService := service.NewVendorService(vendorRepo)
	productService := service.NewProductService(productRepo, vendorRepo)
	orderService := service.NewOrderService(orderRepo, orderItemRepo, accountRepo, productRepo)

	// Use services in the resolver
	resolver := &server.Resolver{
		AccountService: accountService,
		VendorService:  vendorService,
		ProductService: productService,
		OrderService:   orderService,
	}

	// Create HTTP server
	httpServer := server.NewHTTPServer(cfg, resolver)

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting server on port %d", cfg.Server.Port)
		if err := httpServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shut down the server
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

// runMigrations runs the database migrations
func runMigrations(db *postgres.Database, dbConfig config.DatabaseConfig) error {
	driver, err := pgmigrate.WithInstance(db.GetDB().DB, &pgmigrate.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		dbConfig.DBName,
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
