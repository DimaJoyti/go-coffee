#!/bin/bash

# Exit on error
set -e

# Create output directories
mkdir -p proto/wallet
mkdir -p proto/transaction
mkdir -p proto/contract
mkdir -p proto/security

# Generate Go code from proto files
protoc -I=proto \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  proto/wallet.proto

protoc -I=proto \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  proto/transaction.proto

protoc -I=proto \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  proto/contract.proto

protoc -I=proto \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  proto/security.proto

echo "Code generation completed successfully!"
