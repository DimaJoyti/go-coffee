#!/bin/bash

# Final import fix script for Developer DAO Platform
echo "🔧 Fixing all import paths to use TypeScript path mapping..."

# Fix imports in dao-portal
echo "Fixing dao-portal imports..."
find dao-portal -name "*.tsx" -o -name "*.ts" | while read file; do
    if grep -q "../../../shared/src/" "$file"; then
        echo "Fixing imports in: $file"
        sed -i "s|../../../shared/src/|@/shared/|g" "$file"
    fi
done

# Fix imports in governance-ui
echo "Fixing governance-ui imports..."

# First, add the path mapping to governance-ui tsconfig.json
if ! grep -q "@/shared" governance-ui/tsconfig.json; then
    echo "Adding path mapping to governance-ui tsconfig.json..."
    # Add the shared path mapping
    sed -i '/"@\/hooks\/\*": \["\.\/src\/hooks\/\*"\]/a\      "@/shared/*": ["../shared/src/*"]' governance-ui/tsconfig.json
fi

find governance-ui -name "*.tsx" -o -name "*.ts" | while read file; do
    if grep -q "../../../shared/src/" "$file"; then
        echo "Fixing imports in: $file"
        sed -i "s|../../../shared/src/|@/shared/|g" "$file"
    fi
done

echo "✅ All import paths fixed!"
echo "🔄 Restarting frontend applications..."

# Stop and restart frontend
./build-frontend.sh stop
sleep 2
./build-frontend.sh start
