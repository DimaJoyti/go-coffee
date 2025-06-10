module github.com/DimaJoyti/go-coffee/streams

go 1.22

require (
	github.com/IBM/sarama v1.43.2
	github.com/segmentio/kafka-go v0.4.48
	github.com/google/uuid v1.6.0
	github.com/prometheus/client_golang v1.19.0
	github.com/DimaJoyti/go-coffee/pkg v0.0.0-00010101000000-000000000000
)

replace github.com/DimaJoyti/go-coffee/pkg => ../pkg
