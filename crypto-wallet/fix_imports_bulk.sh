#!/bin/bash

# Script to fix all import paths in the crypto-wallet project

echo "🔧 Fixing import paths in the crypto-wallet project..."

# Find all Go files and replace the old import path with the new one
find . -name "*.go" -type f -exec sed -i 's|github.com/DimaJoyti/go-coffee/web3-wallet-backend|github.com/DimaJoyti/go-coffee/crypto-wallet|g' {} +

echo "✅ Import paths fixed!"

# Run go mod tidy to clean up
echo "🧹 Running go mod tidy..."
go mod tidy

echo "🎉 All done!"
