#!/bin/bash

# Script to fix logger formatting issues across the codebase
# This script fixes logger calls that have arguments but no formatting directives

set -e

echo "🔧 Fixing logger formatting issues..."

# Function to fix logger calls in a file
fix_logger_calls() {
    local file="$1"
    echo "Processing: $file"
    
    # Create a backup
    cp "$file" "$file.bak"
    
    # Fix logger calls with arguments but no format strings
    # Pattern: logger.Info("message", arg) -> logger.Info("message: %v", arg)
    # Pattern: logger.Info("message", key, value) -> logger.Info("message", "key", value) (structured logging)
    
    # For calls like logger.Info("message", arg) where arg is a variable
    sed -i 's/\(logger\.\(Info\|Error\|Warn\|Debug\)\)(\("[^"]*"\), \([^)]*\))/\1(\3 ": %v", \4)/g' "$file"
    sed -i 's/\(\*.*Logger\)\.\(Info\|Error\|Warn\|Debug\)(\("[^"]*"\), \([^)]*\))/\1.\2(\3 ": %v", \4)/g' "$file"
    
    # For calls like l.logger.Info("message", arg)
    sed -i 's/\(l\.logger\.\(Info\|Error\|Warn\|Debug\)\)(\("[^"]*"\), \([^)]*\))/\1(\3 ": %v", \4)/g' "$file"
    
    # Remove backup if changes were successful
    if [ $? -eq 0 ]; then
        rm "$file.bak"
        echo "✅ Fixed: $file"
    else
        mv "$file.bak" "$file"
        echo "❌ Failed to fix: $file"
    fi
}

# Find all Go files with logger issues
echo "🔍 Finding files with logger issues..."

# Get list of files from the test output
files_with_issues=(
    "internal/gateway/handlers.go"
    "internal/gateway/service.go"
    "internal/ai/service.go"
    "cmd/api-gateway/main.go"
    "cmd/ai-service/main.go"
    "internal/auth/infrastructure/security/jwt.go"
    "internal/auth/infrastructure/security/password.go"
    "internal/communication/service.go"
    "internal/ai-order/ai_processor.go"
    "internal/ai-order/repository.go"
    "internal/ai-order/service.go"
    "cmd/redis-mcp-demo/main.go"
    "cmd/redis-mcp-server/main.go"
    "internal/security-gateway/application/gateway_service.go"
    "internal/security-gateway/application/rate_limit_service.go"
    "internal/security-gateway/application/waf_service.go"
    "internal/payment/service.go"
    "internal/security-gateway/infrastructure/redis_client.go"
    "cmd/payment-service/main.go"
    "internal/user/handlers.go"
    "internal/auth/infrastructure/repository/redis_session.go"
    "internal/auth/infrastructure/repository/redis_user.go"
)

# Process each file
for file in "${files_with_issues[@]}"; do
    if [ -f "$file" ]; then
        fix_logger_calls "$file"
    else
        echo "⚠️  File not found: $file"
    fi
done

echo "🎉 Logger formatting fixes completed!"
echo "📝 Note: Some fixes may need manual review for complex cases"
