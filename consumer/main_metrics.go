package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/DimaJoyti/go-coffee/consumer/config"
	"github.com/DimaJoyti/go-coffee/consumer/kafka"
	"github.com/DimaJoyti/go-coffee/consumer/metrics"
	"github.com/DimaJoyti/go-coffee/consumer/worker"
)

func mainWithMetrics() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("Starting consumer...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create worker pool
	workerPool := worker.NewWorkerPool(cfg.Kafka.WorkerPoolSize)
	workerPool.Start()
	defer workerPool.Stop()

	// Create Kafka consumer group
	groupConsumer, err := kafka.NewGroupConsumer(cfg)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer group: %v", err)
	}
	defer groupConsumer.Close()

	// Create consumer handler
	handler := kafka.NewOrderConsumerHandler()

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals - used to stop the process
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Start metrics server
	metricsPort := 9091 // Use a different port than the producer
	go func() {
		metricsMux := http.NewServeMux()
		metricsMux.Handle("/metrics", promhttp.Handler())
		metricsAddr := fmt.Sprintf(":%d", metricsPort)
		fmt.Printf("Starting metrics server on %s\n", metricsAddr)
		if err := http.ListenAndServe(metricsAddr, metricsMux); err != nil {
			log.Printf("Metrics server failed: %v", err)
		}
	}()

	// Create a goroutine to run the consumer group
	go func() {
		topics := []string{cfg.Kafka.Topic, cfg.Kafka.ProcessedTopic}
		log.Printf("Starting to consume from topics: %v", topics)

		// Consume messages
		for {
			// Consume messages from both topics
			if err := groupConsumer.Consume(ctx, topics, handler); err != nil {
				log.Printf("Error from consumer: %v", err)
				metrics.StreamsErrorsTotal.WithLabelValues("consume_error").Inc()
			}

			// Check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				log.Println("Context cancelled, stopping consumer")
				return
			}
			log.Println("Consumer restarting...")
		}
	}()

	// Wait for consumer to be ready
	handler.WaitReady()
	fmt.Println("Consumer started")

	// Update metrics for worker pool
	go func() {
		for {
			metrics.WorkerPoolQueueSize.Set(float64(workerPool.QueueSize()))
			metrics.WorkerPoolActiveWorkers.Set(float64(workerPool.ActiveWorkers()))
			select {
			case <-ctx.Done():
				return
			default:
				// Continue monitoring
			}
		}
	}()

	// Process messages
	go func() {
		for msg := range handler.Messages() {
			// Update metrics
			metrics.KafkaMessagesReceivedTotal.WithLabelValues(msg.Topic).Inc()

			// Submit message to worker pool
			workerPool.Submit(msg)
		}
	}()

	// Wait for termination signal
	<-sigchan
	log.Println("Interrupt detected, shutting down...")
	cancel() // Cancel the context to stop the consumer

	log.Println("Consumer stopped")
}
