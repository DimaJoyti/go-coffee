# Kafka Integration

This document describes how the Coffee Order System integrates with Apache Kafka.

## Overview

The Coffee Order System uses Apache Kafka as a message broker to decouple the Producer and Consumer services. The Producer service publishes coffee orders to a Kafka topic, and the Consumer service consumes these orders and processes them.

## Kafka Configuration

### Broker Configuration

By default, the system connects to a Kafka broker at `localhost:9092`. This can be configured using the `KAFKA_BROKERS` environment variable or the `kafka.brokers` configuration option.

### Topic Configuration

By default, the system uses a Kafka topic named `coffee_orders`. This can be configured using the `KAFKA_TOPIC` environment variable or the `kafka.topic` configuration option.

## Producer Integration

The Producer service uses the [Sarama](https://github.com/IBM/sarama) library to publish messages to Kafka.

### Message Format

Messages are published in JSON format with the following structure:

```json
{
  "customer_name": "John Doe",
  "coffee_type": "Latte"
}
```

### Producer Configuration

The Producer service can be configured with the following options:

- `KAFKA_RETRY_MAX`: Maximum number of retries for publishing messages (default: 5)
- `KAFKA_REQUIRED_ACKS`: Required acknowledgments for published messages (default: "all")

### Publishing Flow

1. The Producer service receives an order via the API.
2. The order is converted to JSON.
3. The JSON is published to the Kafka topic.
4. The Producer service waits for acknowledgment from Kafka.
5. The Producer service returns a response to the client.

## Consumer Integration

The Consumer service also uses the [Sarama](https://github.com/IBM/sarama) library to consume messages from Kafka.

### Consumer Configuration

The Consumer service is configured to consume messages from the beginning of the topic (`sarama.OffsetOldest`).

### Consumption Flow

1. The Consumer service connects to Kafka and subscribes to the topic.
2. When a message is received, it is parsed from JSON.
3. The Consumer service processes the order (in this case, by logging it).
4. The Consumer service continues to listen for new messages.

## Error Handling

### Producer Errors

If the Producer service fails to publish a message to Kafka, it returns a 500 Internal Server Error to the client.

### Consumer Errors

If the Consumer service encounters an error while consuming messages, it logs the error and continues to listen for new messages.

## Scalability

### Producer Scalability

Multiple instances of the Producer service can be deployed to handle high load. Each instance will publish messages to the same Kafka topic.

### Consumer Scalability

Multiple instances of the Consumer service can be deployed to process messages in parallel. Kafka will distribute messages among the consumers.

## Monitoring

The system does not currently include built-in monitoring for Kafka. However, Kafka provides its own monitoring tools, such as JMX metrics and the Kafka Manager UI.

## Next Steps

- [Development Guide](development-guide.md): Learn how to develop the Kafka integration.
- [Testing](testing.md): Learn how to test the Kafka integration.
- [Troubleshooting](troubleshooting.md): Troubleshoot Kafka integration issues.
