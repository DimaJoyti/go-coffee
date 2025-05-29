#!/bin/bash

# Script to fix logger calls in bulk

echo "🔧 Fixing logger calls in bulk..."

# Add zap import to files that need it
echo "Adding zap imports..."
files_to_fix=(
    "internal/order/service.go"
    "internal/defi/handler.go"
    "pkg/kafka/producer.go"
    "pkg/failover/service.go"
)

for file in "${files_to_fix[@]}"; do
    if [ -f "$file" ]; then
        echo "Processing $file..."
        # Add zap import if not present
        if ! grep -q "go.uber.org/zap" "$file"; then
            sed -i '/import (/a\\t"go.uber.org/zap"' "$file"
        fi
        
        # Fix logger calls
        sed -i 's/"error", err/zap.Error(err)/g' "$file"
        sed -i 's/"id", id/zap.String("id", id)/g' "$file"
        sed -i 's/"orderID", orderID/zap.String("orderID", orderID)/g' "$file"
        sed -i 's/"key", cacheKey/zap.String("key", cacheKey)/g' "$file"
        sed -i 's/"orderID", order\.ID/zap.String("orderID", order.ID)/g' "$file"
        sed -i 's/"id", order\.ID/zap.String("id", order.ID)/g' "$file"
        
        echo "Fixed $file"
    else
        echo "File $file not found"
    fi
done

echo "✅ Logger calls fixed!"
