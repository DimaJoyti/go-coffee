#!/bin/bash

# Function to check if a command exists
command_exists() {
  command -v "$1" >/dev/null 2>&1
}

# Check if Go is installed
if ! command_exists go; then
  echo "Error: Go is not installed. Please install Go and try again."
  exit 1
fi

# Check if Kafka is running
if ! command_exists nc; then
  echo "Warning: 'nc' command not found. Cannot check if Kafka is running."
else
  if ! nc -z localhost 9092 >/dev/null 2>&1; then
    echo "Warning: Kafka does not appear to be running on localhost:9092."
    echo "Please make sure Kafka is running before starting the services."
    read -p "Do you want to continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      exit 1
    fi
  else
    echo "Kafka is running on localhost:9092."
  fi
fi

# Create Kafka topics if they don't exist
if command_exists kafka-topics.sh; then
  echo "Creating Kafka topics if they don't exist..."
  kafka-topics.sh --create --if-not-exists --bootstrap-server localhost:9092 --replication-factor 1 --partitions 1 --topic coffee_orders
  kafka-topics.sh --create --if-not-exists --bootstrap-server localhost:9092 --replication-factor 1 --partitions 1 --topic processed_orders
  echo "Kafka topics created."
else
  echo "Warning: 'kafka-topics.sh' command not found. Cannot create Kafka topics."
  echo "Please make sure the topics 'coffee_orders' and 'processed_orders' exist."
fi

# Build the services
echo "Building the services..."

echo "Building the producer..."
cd producer
go mod tidy
go build -o producer
cd ..

echo "Building the consumer..."
cd consumer
go mod tidy
go build -o consumer
cd ..

echo "Building the streams processor..."
cd streams
go mod tidy
go build -o streams
cd ..

echo "Building the API Gateway..."
cd api-gateway
go mod tidy
go build -o api-gateway
cd ..

# Start the services
echo "Starting the services..."

# Start the streams processor
echo "Starting the streams processor..."
cd streams
./streams &
STREAMS_PID=$!
cd ..

# Start the consumer
echo "Starting the consumer..."
cd consumer
./consumer &
CONSUMER_PID=$!
cd ..

# Start the producer with gRPC support
echo "Starting the producer..."
cd producer
go run main_grpc.go &
PRODUCER_PID=$!
cd ..

# Start the API Gateway
echo "Starting the API Gateway..."
cd api-gateway
./api-gateway &
API_GATEWAY_PID=$!
cd ..

echo "All services started."
echo "Producer PID: $PRODUCER_PID"
echo "Consumer PID: $CONSUMER_PID"
echo "Streams PID: $STREAMS_PID"
echo "API Gateway PID: $API_GATEWAY_PID"

# Function to handle signals
cleanup() {
  echo "Stopping services..."
  kill $PRODUCER_PID
  kill $CONSUMER_PID
  kill $STREAMS_PID
  kill $API_GATEWAY_PID
  echo "All services stopped."
  exit 0
}

# Register the cleanup function for signals
trap cleanup SIGINT SIGTERM

# Wait for all services to finish
wait $PRODUCER_PID $CONSUMER_PID $STREAMS_PID $API_GATEWAY_PID
