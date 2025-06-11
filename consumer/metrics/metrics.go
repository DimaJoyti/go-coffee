package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// OrdersProcessedTotal is a counter for the total number of orders processed
	OrdersProcessedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "coffee_orders_processed_total",
		Help: "The total number of coffee orders processed",
	})

	// OrdersProcessedSuccessTotal is a counter for the total number of orders successfully processed
	OrdersProcessedSuccessTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "coffee_orders_processed_success_total",
		Help: "The total number of coffee orders successfully processed",
	})

	// OrdersProcessedFailedTotal is a counter for the total number of orders that failed to process
	OrdersProcessedFailedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "coffee_orders_processed_failed_total",
		Help: "The total number of coffee orders that failed to process",
	})

	// OrderPreparationTime is a histogram for the time it takes to prepare an order
	OrderPreparationTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "coffee_order_preparation_seconds",
		Help:    "The time it takes to prepare a coffee order",
		Buckets: prometheus.DefBuckets,
	})

	// KafkaMessagesReceivedTotal is a counter for the total number of messages received from Kafka
	KafkaMessagesReceivedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "kafka_messages_received_total",
		Help: "The total number of messages received from Kafka",
	}, []string{"topic"})

	// KafkaMessagesProcessedTotal is a counter for the total number of messages processed from Kafka
	KafkaMessagesProcessedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "kafka_messages_processed_total",
		Help: "The total number of messages processed from Kafka",
	}, []string{"topic"})

	// KafkaMessagesFailedTotal is a counter for the total number of messages that failed to process from Kafka
	KafkaMessagesFailedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "kafka_messages_failed_total",
		Help: "The total number of messages that failed to process from Kafka",
	}, []string{"topic"})

	// WorkerPoolQueueSize is a gauge for the current size of the worker pool queue
	WorkerPoolQueueSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "worker_pool_queue_size",
		Help: "The current size of the worker pool queue",
	})

	// WorkerPoolActiveWorkers is a gauge for the current number of active workers in the worker pool
	WorkerPoolActiveWorkers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "worker_pool_active_workers",
		Help: "The current number of active workers in the worker pool",
	})

	// WorkerProcessingTime is a histogram for the time it takes for a worker to process a message
	WorkerProcessingTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "worker_processing_seconds",
		Help:    "The time it takes for a worker to process a message",
		Buckets: prometheus.DefBuckets,
	}, []string{"worker_id", "topic"})

	// StreamsErrorsTotal is a counter for the total number of stream errors
	StreamsErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "streams_errors_total",
		Help: "The total number of stream errors",
	}, []string{"error_type"})
)
