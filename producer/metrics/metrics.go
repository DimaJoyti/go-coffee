package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// OrdersTotal is a counter for the total number of orders received
	OrdersTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "coffee_orders_total",
		Help: "The total number of coffee orders received",
	})

	// OrdersSuccessTotal is a counter for the total number of orders successfully processed
	OrdersSuccessTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "coffee_orders_success_total",
		Help: "The total number of coffee orders successfully processed",
	})

	// OrdersFailedTotal is a counter for the total number of orders that failed to process
	OrdersFailedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "coffee_orders_failed_total",
		Help: "The total number of coffee orders that failed to process",
	})

	// OrderProcessingTime is a histogram for the time it takes to process an order
	OrderProcessingTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "coffee_order_processing_seconds",
		Help:    "The time it takes to process a coffee order",
		Buckets: prometheus.DefBuckets,
	})

	// KafkaMessagesSentTotal is a counter for the total number of messages sent to Kafka
	KafkaMessagesSentTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kafka_messages_sent_total",
		Help: "The total number of messages sent to Kafka",
	})

	// KafkaMessagesFailedTotal is a counter for the total number of messages that failed to send to Kafka
	KafkaMessagesFailedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kafka_messages_failed_total",
		Help: "The total number of messages that failed to send to Kafka",
	})

	// HttpRequestsTotal is a counter for the total number of HTTP requests
	HttpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "The total number of HTTP requests",
	}, []string{"method", "endpoint", "status"})

	// HttpRequestDuration is a histogram for the duration of HTTP requests
	HttpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "The duration of HTTP requests",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "endpoint"})
)
