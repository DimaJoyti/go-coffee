#!/bin/bash

# Exit on error
set -e

# Build the services
echo "Building the services..."

echo "Building the API Gateway..."
cd cmd/api-gateway
go build -o api-gateway
cd ../..

echo "Building the Wallet Service..."
cd cmd/wallet-service
go build -o wallet-service
cd ../..

echo "Building the Transaction Service..."
cd cmd/transaction-service
go build -o transaction-service
cd ../..

echo "Building the Smart Contract Service..."
cd cmd/smart-contract-service
go build -o smart-contract-service
cd ../..

echo "Building the Security Service..."
cd cmd/security-service
go build -o security-service
cd ../..

# Start the services
echo "Starting the services..."

# Start the API Gateway
echo "Starting the API Gateway..."
cd cmd/api-gateway
./api-gateway &
API_GATEWAY_PID=$!
cd ../..

# Start the Wallet Service
echo "Starting the Wallet Service..."
cd cmd/wallet-service
./wallet-service &
WALLET_SERVICE_PID=$!
cd ../..

# Start the Transaction Service
echo "Starting the Transaction Service..."
cd cmd/transaction-service
./transaction-service &
TRANSACTION_SERVICE_PID=$!
cd ../..

# Start the Smart Contract Service
echo "Starting the Smart Contract Service..."
cd cmd/smart-contract-service
./smart-contract-service &
SMART_CONTRACT_SERVICE_PID=$!
cd ../..

# Start the Security Service
echo "Starting the Security Service..."
cd cmd/security-service
./security-service &
SECURITY_SERVICE_PID=$!
cd ../..

echo "All services started successfully!"

# Handle graceful shutdown
function cleanup {
  echo "Stopping services..."
  kill $API_GATEWAY_PID $WALLET_SERVICE_PID $TRANSACTION_SERVICE_PID $SMART_CONTRACT_SERVICE_PID $SECURITY_SERVICE_PID
  wait
  echo "All services stopped."
}

trap cleanup SIGINT SIGTERM

# Wait for user to press Ctrl+C
echo "Press Ctrl+C to stop all services"
wait
