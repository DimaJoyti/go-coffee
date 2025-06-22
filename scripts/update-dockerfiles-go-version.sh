#!/bin/bash

# Update all Dockerfiles to use Go 1.23
set -e

echo "🔧 Updating all Dockerfiles to use Go 1.23..."

# Find all Dockerfiles that contain golang references
DOCKERFILES=$(find . -name "Dockerfile" -exec grep -l "golang:" {} \;)

# Function to update Go version in a Dockerfile
update_dockerfile() {
    local file="$1"
    echo "Processing: $file"
    
    # Create backup
    cp "$file" "$file.bak"
    
    # Update various Go versions to 1.23
    sed -i 's/FROM golang:1\.21[^-]*/FROM golang:1.23/g' "$file"
    sed -i 's/FROM golang:1\.22[^-]*/FROM golang:1.23/g' "$file"
    sed -i 's/FROM golang:1\.24[^-]*/FROM golang:1.23/g' "$file"
    sed -i 's/FROM golang:1\.25[^-]*/FROM golang:1.23/g' "$file"
    
    # Handle alpine variants
    sed -i 's/FROM golang:1\.21-alpine/FROM golang:1.23-alpine/g' "$file"
    sed -i 's/FROM golang:1\.22-alpine/FROM golang:1.23-alpine/g' "$file"
    sed -i 's/FROM golang:1\.24-alpine/FROM golang:1.23-alpine/g' "$file"
    sed -i 's/FROM golang:1\.25-alpine/FROM golang:1.23-alpine/g' "$file"
    
    echo "✅ Updated: $file"
}

# Update each Dockerfile
for dockerfile in $DOCKERFILES; do
    if [ -f "$dockerfile" ]; then
        update_dockerfile "$dockerfile"
    else
        echo "⚠️  File not found: $dockerfile"
    fi
done

echo "🎉 All Dockerfiles updated to use Go 1.23!"
echo "📝 Backup files created with .bak extension"
echo "🧪 Test Docker builds with: docker build -f <dockerfile> ."
