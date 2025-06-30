#!/bin/bash

# Generate Go code from Protocol Buffer definitions
# This script generates Go code from .proto files for the AI agents

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
PROTO_DIR="proto"
OUTPUT_DIR="proto/generated"
GO_MODULE="go-coffee-ai-agents"

echo -e "${GREEN}🚀 Generating Go code from Protocol Buffer definitions...${NC}"

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo -e "${RED}❌ protoc is not installed. Please install Protocol Buffers compiler.${NC}"
    echo "Installation instructions:"
    echo "  macOS: brew install protobuf"
    echo "  Ubuntu: sudo apt-get install protobuf-compiler"
    echo "  Or download from: https://github.com/protocolbuffers/protobuf/releases"
    exit 1
fi

# Check if protoc-gen-go is installed
if ! command -v protoc-gen-go &> /dev/null; then
    echo -e "${YELLOW}⚠️  protoc-gen-go is not installed. Installing...${NC}"
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

# Check if protoc-gen-go-grpc is installed
if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo -e "${YELLOW}⚠️  protoc-gen-go-grpc is not installed. Installing...${NC}"
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# Create output directory
mkdir -p "$OUTPUT_DIR"

echo -e "${GREEN}📁 Created output directory: $OUTPUT_DIR${NC}"

# Function to generate Go code from proto files
generate_proto() {
    local proto_file=$1
    local relative_path=${proto_file#$PROTO_DIR/}
    local output_path="$OUTPUT_DIR/${relative_path%/*}"
    
    echo -e "${GREEN}🔄 Generating: $proto_file${NC}"
    
    # Create output directory for this proto file
    mkdir -p "$output_path"
    
    # Generate Go code
    protoc \
        --proto_path="$PROTO_DIR" \
        --go_out="$OUTPUT_DIR" \
        --go_opt=paths=source_relative \
        --go_opt=module="$GO_MODULE/$OUTPUT_DIR" \
        "$proto_file"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Successfully generated: ${relative_path%.proto}.pb.go${NC}"
    else
        echo -e "${RED}❌ Failed to generate: $proto_file${NC}"
        exit 1
    fi
}

# Find all .proto files and generate Go code
echo -e "${GREEN}🔍 Finding Protocol Buffer files...${NC}"

proto_files=$(find "$PROTO_DIR" -name "*.proto" -type f)

if [ -z "$proto_files" ]; then
    echo -e "${RED}❌ No .proto files found in $PROTO_DIR${NC}"
    exit 1
fi

echo -e "${GREEN}📋 Found $(echo "$proto_files" | wc -l) Protocol Buffer files:${NC}"
echo "$proto_files" | sed 's/^/  /'

echo ""
echo -e "${GREEN}🔨 Generating Go code...${NC}"

# Generate Go code for each proto file
for proto_file in $proto_files; do
    generate_proto "$proto_file"
done

echo ""
echo -e "${GREEN}🎉 Protocol Buffer code generation completed successfully!${NC}"

# Generate a summary
echo ""
echo -e "${GREEN}📊 Generation Summary:${NC}"
echo -e "  📁 Proto files processed: $(echo "$proto_files" | wc -l)"
echo -e "  📁 Output directory: $OUTPUT_DIR"
echo -e "  📁 Go module: $GO_MODULE"

# List generated files
echo ""
echo -e "${GREEN}📋 Generated files:${NC}"
find "$OUTPUT_DIR" -name "*.pb.go" -type f | sed 's/^/  ✅ /'

echo ""
echo -e "${GREEN}🚀 Next steps:${NC}"
echo -e "  1. Run 'go mod tidy' to update dependencies"
echo -e "  2. Import the generated packages in your Go code"
echo -e "  3. Use the generated types for type-safe event handling"

echo ""
echo -e "${GREEN}💡 Example usage:${NC}"
echo -e "  import \"$GO_MODULE/$OUTPUT_DIR/events\""
echo -e "  event := &events.BeverageCreatedEvent{...}"

echo ""
echo -e "${GREEN}✨ Happy coding!${NC}"
