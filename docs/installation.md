# Installation Guide

This guide provides instructions for installing and running the Coffee Order System.

## Prerequisites

Before installing the Coffee Order System, ensure you have the following prerequisites:

- **Go**: Version 1.22 or higher. [Download Go](https://golang.org/dl/).
- **Kafka**: A running Kafka instance. [Download Kafka](https://kafka.apache.org/downloads).
- **Git**: For cloning the repository. [Download Git](https://git-scm.com/downloads).

## Setting Up Kafka

If you don't have Kafka running, you can set it up using Docker:

```bash
# Pull the Kafka image
docker pull bitnami/kafka

# Start Kafka
docker run -d --name kafka \
  -p 9092:9092 \
  -e KAFKA_CFG_ZOOKEEPER_CONNECT=zookeeper:2181 \
  -e ALLOW_PLAINTEXT_LISTENER=yes \
  bitnami/kafka
```

## Installing the Coffee Order System

### Clone the Repository

```bash
git clone https://github.com/yourusername/coffee-order-system.git
cd coffee-order-system
```

### Install Dependencies

```bash
# Install Producer dependencies
cd producer
go mod tidy

# Install Consumer dependencies
cd ../consumer
go mod tidy
```

## Running the System

### Start the Producer Service

```bash
cd producer
go run main.go
```

The Producer service will start an HTTP server on port 3000 (by default).

### Start the Consumer Service

```bash
cd consumer
go run main.go
```

The Consumer service will start consuming messages from the Kafka topic.

## Verifying the Installation

To verify that the system is working correctly, you can place a coffee order using the API:

```bash
curl -X POST http://localhost:3000/order \
  -H "Content-Type: application/json" \
  -d '{"customer_name":"John Doe","coffee_type":"Latte"}'
```

You should see a response like:

```json
{
  "success": true,
  "msg": "Order for John Doe placed successfully!"
}
```

And in the Consumer service logs, you should see a message like:

```
Brewing Latte coffee for John Doe
```

## Configuration

By default, the system uses the following configuration:

- Producer service runs on port 3000
- Kafka broker is at localhost:9092
- Kafka topic is "coffee_orders"

To customize the configuration, see the [Configuration](configuration.md) document.

## Troubleshooting

If you encounter any issues during installation or running the system, see the [Troubleshooting](troubleshooting.md) document.

## Next Steps

- [Configuration](configuration.md): Configure the system.
- [API Reference](api-reference.md): Explore the API endpoints.
- [Development Guide](development-guide.md): Learn how to develop the system.
