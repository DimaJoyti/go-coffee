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

	"github.com/DimaJoyti/go-coffee/internal/web3"
	"github.com/DimaJoyti/go-coffee/internal/web3/payment"
	httpTransport "github.com/DimaJoyti/go-coffee/internal/web3/transport/http"
	"github.com/DimaJoyti/go-coffee/pkg/blockchain"
	"github.com/DimaJoyti/go-coffee/pkg/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Initialize structured logger
	logger := initLogger()
	defer logger.Sync()

	logger.Info("Starting Web3 Payment Service...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize blockchain clients
	ethClient, err := blockchain.NewEthereumClient(cfg.Web3.Blockchain.Ethereum)
	if err != nil {
		logger.Fatal("Failed to create Ethereum client", zap.Error(err))
	}

	bscClient, err := blockchain.NewBSCClient(cfg.Web3.Blockchain.BSC)
	if err != nil {
		logger.Fatal("Failed to create BSC client", zap.Error(err))
	}

	polygonClient, err := blockchain.NewPolygonClient(cfg.Web3.Blockchain.Polygon)
	if err != nil {
		logger.Fatal("Failed to create Polygon client", zap.Error(err))
	}

	solanaClient, err := blockchain.NewSolanaClient(cfg.Web3.Blockchain.Solana)
	if err != nil {
		logger.Fatal("Failed to create Solana client", zap.Error(err))
	}

	// Create payment processor
	paymentProcessor := payment.NewProcessor(
		ethClient,
		bscClient,
		polygonClient,
		solanaClient,
		logger,
		cfg.Web3.Payment,
	)

	// Create Web3 service
	web3Service := web3.NewService(
		paymentProcessor,
		logger,
		cfg.Web3,
	)

	// Start Web3 service
	ctx := context.Background()
	if err := web3Service.Start(ctx); err != nil {
		logger.Fatal("Failed to start Web3 service", zap.Error(err))
	}

	// Start health check server
	go startHealthServer(logger, cfg)

	// Create HTTP handler
	handler := httpTransport.NewHandler(web3Service, logger)

	// Setup HTTP routes
	mux := http.NewServeMux()
	
	// Payment endpoints
	mux.HandleFunc("/payment/create", handler.CreatePayment)
	mux.HandleFunc("/payment/status/", handler.GetPaymentStatus)
	mux.HandleFunc("/payment/confirm", handler.ConfirmPayment)
	mux.HandleFunc("/payment/cancel", handler.CancelPayment)
	
	// Wallet endpoints
	mux.HandleFunc("/wallet/balance/", handler.GetWalletBalance)
	mux.HandleFunc("/wallet/transactions/", handler.GetWalletTransactions)
	
	// Token endpoints
	mux.HandleFunc("/token/price/", handler.GetTokenPrice)
	mux.HandleFunc("/token/swap", handler.SwapTokens)
	
	// DeFi endpoints
	mux.HandleFunc("/defi/yield", handler.GetYieldOpportunities)
	mux.HandleFunc("/defi/stake", handler.StakeTokens)
	mux.HandleFunc("/defi/unstake", handler.UnstakeTokens)
	
	// Observability endpoints
	mux.HandleFunc("/health", handler.HealthCheck)
	mux.HandleFunc("/ready", handler.ReadinessCheck)
	mux.Handle("/metrics", promhttp.Handler())

	// Create HTTP server with enhanced configuration
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Web3.Payment.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server in goroutine
	go func() {
		logger.Info("Starting Web3 Payment HTTP server", zap.Int("port", cfg.Web3.Payment.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down Web3 Payment service...")

	// Create context with timeout for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error", zap.Error(err))
	}

	// Stop Web3 service
	web3Service.Stop()

	logger.Info("Web3 Payment service stopped gracefully")
}

// initLogger initializes a structured logger with appropriate configuration
func initLogger() *zap.Logger {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.StacktraceKey = "stacktrace"

	logger, err := config.Build()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	return logger
}

// startHealthServer starts a health check server for the Web3 payment service
func startHealthServer(logger *zap.Logger, cfg *config.Config) {
	mux := http.NewServeMux()
	
	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"status":"ok",
			"service":"web3-payment",
			"timestamp":"%s",
			"version":"1.0.0",
			"chains":["ethereum","bsc","polygon","solana"]
		}`, time.Now().UTC().Format(time.RFC3339))
	})
	
	// Readiness check endpoint
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"status":"ready",
			"service":"web3-payment",
			"timestamp":"%s",
			"checks":{
				"ethereum":"ok",
				"bsc":"ok",
				"polygon":"ok",
				"solana":"ok",
				"payment_processor":"ok"
			}
		}`, time.Now().UTC().Format(time.RFC3339))
	})
	
	// Metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Start health server on a different port
	healthPort := 8083
	if cfg.Web3.Payment.HealthPort != 0 {
		healthPort = cfg.Web3.Payment.HealthPort
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", healthPort),
		Handler: mux,
	}

	logger.Info("Starting Web3 Payment health check server", zap.Int("port", healthPort))
	if err := server.ListenAndServe(); err != nil {
		logger.Error("Health server error", zap.Error(err))
	}
}
