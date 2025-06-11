#!/bin/bash

# Validate CI/CD Setup
# This script checks if all necessary files are in place for the CI/CD pipeline

echo "🔍 Validating CI/CD Setup for Go Coffee"
echo "======================================="

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

check_file() {
    if [ -f "$1" ]; then
        echo -e "${GREEN}✅ $1${NC}"
        return 0
    else
        echo -e "${RED}❌ $1 (missing)${NC}"
        return 1
    fi
}

check_dir() {
    if [ -d "$1" ]; then
        echo -e "${GREEN}✅ $1/ (directory)${NC}"
        return 0
    else
        echo -e "${RED}❌ $1/ (missing directory)${NC}"
        return 1
    fi
}

echo "Checking CI/CD workflow files..."
check_file ".github/workflows/ci-cd.yaml"

echo ""
echo "Checking configuration files..."
check_file ".golangci.yml"
check_file "go.mod"
check_file "go.sum"

echo ""
echo "Checking root module files..."
check_file "main.go"
check_file "main_test.go"

echo ""
echo "Checking service directories and files..."
services=("accounts-service" "producer" "consumer" "streams" "api-gateway")
for service in "${services[@]}"; do
    if check_dir "$service"; then
        check_file "$service/go.mod"
        if [ "$service" != "api-gateway" ]; then
            check_file "$service/Dockerfile"
        fi
    fi
done

echo ""
echo "Checking cmd services..."
cmd_services=("user-gateway" "security-gateway" "kitchen-service")
for service in "${cmd_services[@]}"; do
    if check_dir "cmd/$service"; then
        check_file "cmd/$service/main.go"
        check_file "cmd/$service/Dockerfile"
    fi
done

echo ""
echo "Checking scripts..."
check_file "scripts/test-ci-pipeline.sh"
check_file "scripts/validate-ci-setup.sh"

echo ""
echo "Checking documentation..."
check_file "docs/CI-CD-PIPELINE-FIXES-COMPLETE.md"

echo ""
echo "Checking test directories..."
check_dir "tests"
if [ ! -d "tests/integration" ]; then
    echo -e "${YELLOW}⚠️  tests/integration/ (will be created by CI)${NC}"
fi
if [ ! -d "tests/performance" ]; then
    echo -e "${YELLOW}⚠️  tests/performance/ (will be created by CI)${NC}"
fi

echo ""
echo "🎯 Validation Summary"
echo "===================="

# Count files
total_files=0
missing_files=0

# Essential files for CI/CD
essential_files=(
    ".github/workflows/ci-cd.yaml"
    ".golangci.yml"
    "go.mod"
    "main.go"
    "main_test.go"
    "accounts-service/go.mod"
    "producer/go.mod"
    "consumer/go.mod"
    "streams/go.mod"
    "producer/Dockerfile"
    "consumer/Dockerfile"
    "streams/Dockerfile"
    "cmd/user-gateway/Dockerfile"
    "cmd/security-gateway/Dockerfile"
    "cmd/kitchen-service/Dockerfile"
)

for file in "${essential_files[@]}"; do
    total_files=$((total_files + 1))
    if [ ! -f "$file" ]; then
        missing_files=$((missing_files + 1))
    fi
done

present_files=$((total_files - missing_files))
echo "Essential files: $present_files/$total_files present"

if [ $missing_files -eq 0 ]; then
    echo -e "${GREEN}🎉 All essential files are present!${NC}"
    echo -e "${GREEN}✅ CI/CD pipeline is ready to run${NC}"
    exit 0
else
    echo -e "${RED}❌ $missing_files essential files are missing${NC}"
    echo -e "${YELLOW}⚠️  Please create missing files before running CI/CD${NC}"
    exit 1
fi
