package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"kafka_worker/config"
	"kafka_worker/kafka"
	"kafka_worker/worker"
)

func main() {
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

	// Create a goroutine to run the consumer group
	go func() {
		topics := []string{cfg.Kafka.Topic, cfg.Kafka.ProcessedTopic}
		log.Printf("Starting to consume from topics: %v", topics)

		// Consume messages
		for {
			// Consume messages from both topics
			if err := groupConsumer.Consume(ctx, topics, handler); err != nil {
				log.Printf("Error from consumer: %v", err)
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

	// Process messages
	go func() {
		for msg := range handler.Messages() {
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
