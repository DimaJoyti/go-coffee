#!/bin/bash

# Script to update Bitcoin package import paths to new structure

echo "Updating Bitcoin package import paths..."

# Find all Go files in the Bitcoin package
find pkg/bitcoin -name "*.go" -type f | while read -r file; do
    echo "Updating imports in: $file"
    
    # Update import paths
    sed -i 's|github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/bitcoin|github.com/DimaJoyti/go-coffee/pkg/bitcoin|g' "$file"
    
    echo "Updated: $file"
done

echo "Bitcoin import paths updated successfully!"
