#!/bin/bash

# Script to fix logger calls in the project

echo "🔧 Fixing logger calls in the project..."

# Fix DeFi handler
echo "Fixing internal/defi/handler.go..."
sed -i 's/"error", err/zap.Error(err)/g' internal/defi/handler.go
sed -i 's/"method", method/zap.String("method", method)/g' internal/defi/handler.go
sed -i 's/"request", string(reqJSON)/zap.String("request", string(reqJSON))/g' internal/defi/handler.go
sed -i 's/"response", string(respJSON)/zap.String("response", string(respJSON))/g' internal/defi/handler.go

# Fix failover service
echo "Fixing pkg/failover/service.go..."
sed -i 's/"region", /zap.String("region", /g' pkg/failover/service.go
sed -i 's/"error", err/zap.Error(err)/g' pkg/failover/service.go

# Fix Kafka producer
echo "Fixing pkg/kafka/producer.go..."
sed -i 's/map\[string\]interface{}{.*}/zap.Any("metadata", metadata)/g' pkg/kafka/producer.go
sed -i 's/nil/zap.String("status", "success")/g' pkg/kafka/producer.go

echo "✅ Logger calls fixed!"
