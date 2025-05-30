#!/bin/bash

# Generate protobuf code for AI Order Management Services
# This script generates Go code from .proto files

set -e

echo "🔄 Generating protobuf code for AI Order Management Services..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Directories
PROTO_DIR="api/proto"
GO_OUT_DIR="api/proto"

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo -e "${RED}❌ protoc is not installed. Please install Protocol Buffers compiler.${NC}"
    echo "Installation instructions:"
    echo "  macOS: brew install protobuf"
    echo "  Ubuntu: sudo apt install protobuf-compiler"
    echo "  Or download from: https://github.com/protocolbuffers/protobuf/releases"
    exit 1
fi

# Check if Go protobuf plugins are installed
if ! command -v protoc-gen-go &> /dev/null; then
    echo -e "${YELLOW}⚠️  protoc-gen-go not found. Installing...${NC}"
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo -e "${YELLOW}⚠️  protoc-gen-go-grpc not found. Installing...${NC}"
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# Create output directory if it doesn't exist
mkdir -p "$GO_OUT_DIR"

echo -e "${BLUE}📁 Proto files directory: $PROTO_DIR${NC}"
echo -e "${BLUE}📁 Output directory: $GO_OUT_DIR${NC}"

# Generate Go code for each proto file
for proto_file in "$PROTO_DIR"/*.proto; do
    if [ -f "$proto_file" ]; then
        filename=$(basename "$proto_file")
        echo -e "${YELLOW}🔄 Processing $filename...${NC}"
        
        # Generate Go code
        protoc \
            --go_out="$GO_OUT_DIR" \
            --go_opt=paths=source_relative \
            --go-grpc_out="$GO_OUT_DIR" \
            --go-grpc_opt=paths=source_relative \
            --proto_path="$PROTO_DIR" \
            "$proto_file"
        
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✅ Generated Go code for $filename${NC}"
        else
            echo -e "${RED}❌ Failed to generate Go code for $filename${NC}"
            exit 1
        fi
    fi
done

# Generate gRPC gateway code (optional)
if command -v protoc-gen-grpc-gateway &> /dev/null; then
    echo -e "${YELLOW}🔄 Generating gRPC Gateway code...${NC}"
    
    for proto_file in "$PROTO_DIR"/*.proto; do
        if [ -f "$proto_file" ]; then
            filename=$(basename "$proto_file")
            
            protoc \
                --grpc-gateway_out="$GO_OUT_DIR" \
                --grpc-gateway_opt=paths=source_relative \
                --proto_path="$PROTO_DIR" \
                "$proto_file"
        fi
    done
    
    echo -e "${GREEN}✅ gRPC Gateway code generated${NC}"
else
    echo -e "${YELLOW}⚠️  protoc-gen-grpc-gateway not found. Skipping gateway generation.${NC}"
    echo "To install: go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest"
fi

# Generate OpenAPI/Swagger documentation (optional)
if command -v protoc-gen-openapiv2 &> /dev/null; then
    echo -e "${YELLOW}🔄 Generating OpenAPI documentation...${NC}"
    
    mkdir -p docs/api
    
    for proto_file in "$PROTO_DIR"/*.proto; do
        if [ -f "$proto_file" ]; then
            filename=$(basename "$proto_file" .proto)
            
            protoc \
                --openapiv2_out=docs/api \
                --openapiv2_opt=logtostderr=true \
                --proto_path="$PROTO_DIR" \
                "$proto_file"
        fi
    done
    
    echo -e "${GREEN}✅ OpenAPI documentation generated in docs/api/${NC}"
else
    echo -e "${YELLOW}⚠️  protoc-gen-openapiv2 not found. Skipping OpenAPI generation.${NC}"
    echo "To install: go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest"
fi

# Validate generated files
echo -e "${YELLOW}🔍 Validating generated files...${NC}"

generated_files=0
for go_file in "$GO_OUT_DIR"/*.pb.go; do
    if [ -f "$go_file" ]; then
        ((generated_files++))
        echo -e "${GREEN}✅ $(basename "$go_file")${NC}"
    fi
done

for grpc_file in "$GO_OUT_DIR"/*_grpc.pb.go; do
    if [ -f "$grpc_file" ]; then
        ((generated_files++))
        echo -e "${GREEN}✅ $(basename "$grpc_file")${NC}"
    fi
done

if [ $generated_files -eq 0 ]; then
    echo -e "${RED}❌ No Go files were generated!${NC}"
    exit 1
fi

# Format generated Go code
echo -e "${YELLOW}🎨 Formatting generated Go code...${NC}"
go fmt "$GO_OUT_DIR"/*.go 2>/dev/null || true

# Run go mod tidy to ensure dependencies are correct
echo -e "${YELLOW}📦 Updating Go modules...${NC}"
go mod tidy

echo ""
echo -e "${GREEN}🎉 Protobuf code generation completed successfully!${NC}"
echo -e "${BLUE}📊 Generated $generated_files Go files${NC}"
echo ""
echo "Generated files:"
ls -la "$GO_OUT_DIR"/*.pb.go 2>/dev/null || true
ls -la "$GO_OUT_DIR"/*_grpc.pb.go 2>/dev/null || true

echo ""
echo "Next steps:"
echo "1. Build the services: make -f Makefile.ai-order build"
echo "2. Run the services: make -f Makefile.ai-order run-all"
echo "3. Test the API: make -f Makefile.ai-order test-api"
