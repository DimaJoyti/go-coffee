package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// StreamsMessagesProcessedTotal is a counter for the total number of messages processed by the streams processor
	StreamsMessagesProcessedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "streams_messages_processed_total",
		Help: "The total number of messages processed by the streams processor",
	}, []string{"topic"})

	// StreamsMessagesSuccessTotal is a counter for the total number of messages successfully processed by the streams processor
	StreamsMessagesSuccessTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "streams_messages_success_total",
		Help: "The total number of messages successfully processed by the streams processor",
	}, []string{"topic"})

	// StreamsMessagesFailedTotal is a counter for the total number of messages that failed to process by the streams processor
	StreamsMessagesFailedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "streams_messages_failed_total",
		Help: "The total number of messages that failed to process by the streams processor",
	}, []string{"topic"})

	// StreamsProcessingTime is a histogram for the time it takes to process a message by the streams processor
	StreamsProcessingTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "streams_processing_seconds",
		Help:    "The time it takes to process a message by the streams processor",
		Buckets: prometheus.DefBuckets,
	}, []string{"topic"})

	// StreamsInputMessagesTotal is a counter for the total number of input messages received by the streams processor
	StreamsInputMessagesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "streams_input_messages_total",
		Help: "The total number of input messages received by the streams processor",
	})

	// StreamsOutputMessagesTotal is a counter for the total number of output messages sent by the streams processor
	StreamsOutputMessagesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "streams_output_messages_total",
		Help: "The total number of output messages sent by the streams processor",
	})

	// StreamsErrorsTotal is a counter for the total number of errors encountered by the streams processor
	StreamsErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "streams_errors_total",
		Help: "The total number of errors encountered by the streams processor",
	}, []string{"error_type"})

	// StreamsRunning is a gauge for whether the streams processor is running
	StreamsRunning = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "streams_running",
		Help: "Whether the streams processor is running (1 for running, 0 for not running)",
	})
)
