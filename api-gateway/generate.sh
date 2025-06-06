#!/bin/bash

# Створення директорій для згенерованого коду
mkdir -p proto/coffee

# Генерація Go коду з .proto файлів
protoc -I=proto \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=. \
  --grpc-gateway_opt=paths=source_relative \
  proto/coffee_service.proto

echo "Code generation completed successfully!"
