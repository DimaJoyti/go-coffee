package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/DimaJoyti/go-coffee/streams/config"
	"github.com/DimaJoyti/go-coffee/streams/kafka"
)

func mainWithMetrics() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("Starting Kafka Streams processor...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create stream processor
	processor, err := kafka.NewStreamProcessor(cfg)
	if err != nil {
		log.Fatalf("Failed to create stream processor: %v", err)
	}

	// Start metrics server
	metricsPort := 9092 // Use a different port than the producer and consumer
	go func() {
		metricsMux := http.NewServeMux()
		metricsMux.Handle("/metrics", promhttp.Handler())
		metricsAddr := fmt.Sprintf(":%d", metricsPort)
		fmt.Printf("Starting metrics server on %s\n", metricsAddr)
		if err := http.ListenAndServe(metricsAddr, metricsMux); err != nil {
			log.Printf("Metrics server failed: %v", err)
		}
	}()

	// Start the processor
	if err := processor.Start(); err != nil {
		log.Fatalf("Failed to start stream processor: %v", err)
	}
	log.Println("Stream processor started successfully")

	// Set up signal handler for graceful shutdown
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	sig := <-sigchan
	log.Printf("Received signal %v, shutting down...", sig)

	// Stop the processor
	processor.Stop()
	log.Println("Stream processor stopped")
}
