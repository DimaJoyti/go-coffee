#!/bin/bash

# Comprehensive script to fix all logger formatting issues
set -e

echo "🔧 Fixing all logger formatting issues..."

# Function to fix logger calls in a file
fix_logger_in_file() {
    local file="$1"
    echo "Processing: $file"
    
    if [ ! -f "$file" ]; then
        echo "⚠️  File not found: $file"
        return
    fi
    
    # Create backup
    cp "$file" "$file.bak"
    
    # Fix structured logging calls that use maps
    # Pattern: logger.Info("message", map[string]interface{}{...}) -> logger.WithFields(map[string]interface{}{...}).Info("message")
    sed -i 's/\(logger\.\(Info\|Error\|Warn\|Debug\)\)(\("[^"]*"\), \(map\[string\]interface{}{[^}]*}\))/logger.WithFields(\4).\2(\3)/g' "$file"
    sed -i 's/\(\*.*Logger\)\.\(Info\|Error\|Warn\|Debug\)(\("[^"]*"\), \(map\[string\]interface{}{[^}]*}\))/\1.WithFields(\4).\2(\3)/g' "$file"
    
    # Fix simple logger calls with single arguments
    # Pattern: logger.Info("message", arg) -> logger.Info("message: %v", arg)
    sed -i 's/\(logger\.\(Info\|Error\|Warn\|Debug\)\)(\("[^"]*"\), \([^,)]*\))/\1(\3 ": %v", \4)/g' "$file"
    sed -i 's/\(\*.*Logger\)\.\(Info\|Error\|Warn\|Debug\)(\("[^"]*"\), \([^,)]*\))/\1.\2(\3 ": %v", \4)/g' "$file"
    
    # Check if changes were made successfully
    if [ $? -eq 0 ]; then
        rm "$file.bak"
        echo "✅ Fixed: $file"
    else
        mv "$file.bak" "$file"
        echo "❌ Failed to fix: $file"
    fi
}

# List of files with logger issues (from test output)
files_to_fix=(
    "internal/ai/service.go"
    "cmd/api-gateway/main.go"
    "cmd/ai-service/main.go"
    "internal/ai-order/ai_processor.go"
    "internal/ai-order/repository.go"
    "internal/ai-order/service.go"
    "internal/auth/infrastructure/security/jwt.go"
    "internal/auth/infrastructure/security/password.go"
    "internal/communication/service.go"
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
for file in "${files_to_fix[@]}"; do
    fix_logger_in_file "$file"
done

echo "🎉 Logger formatting fixes completed!"
echo "📝 Running quick test to verify fixes..."

# Quick test to see if we reduced the number of errors
go build ./... 2>&1 | grep -c "call has arguments but no formatting directives" || echo "0 logger formatting errors remaining!"

echo "✅ Script completed successfully!"
