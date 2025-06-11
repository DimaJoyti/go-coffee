package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/DimaJoyti/go-coffee/streams/config"
	"github.com/DimaJoyti/go-coffee/streams/kafka"
)

func main() {
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
