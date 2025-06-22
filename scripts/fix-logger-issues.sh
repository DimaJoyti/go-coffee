#!/bin/bash

# Fix logger formatting issues across the codebase
# This script fixes logger calls that have arguments but no formatting directives

set -e

echo "🔧 Fixing logger formatting issues..."

# Function to fix logger calls in a file
fix_logger_calls() {
    local file="$1"
    echo "Fixing logger calls in: $file"
    
    # Create a backup
    cp "$file" "$file.bak"
    
    # Fix logger calls with arguments but no format strings
    # Pattern: logger.Method(arg1, arg2, ...) -> logger.Method("%v %v ...", arg1, arg2, ...)
    
    # Fix Info calls
    sed -i 's/\.Info(\([^"]*[^,)]\), \([^)]*\))/\.Info("%v", \1, \2)/g' "$file"
    sed -i 's/\.Info(\([^"]*[^,)]\))/\.Info("%v", \1)/g' "$file"
    
    # Fix Error calls
    sed -i 's/\.Error(\([^"]*[^,)]\), \([^)]*\))/\.Error("%v", \1, \2)/g' "$file"
    sed -i 's/\.Error(\([^"]*[^,)]\))/\.Error("%v", \1)/g' "$file"
    
    # Fix Debug calls
    sed -i 's/\.Debug(\([^"]*[^,)]\), \([^)]*\))/\.Debug("%v", \1, \2)/g' "$file"
    sed -i 's/\.Debug(\([^"]*[^,)]\))/\.Debug("%v", \1)/g' "$file"
    
    # Fix Warn calls
    sed -i 's/\.Warn(\([^"]*[^,)]\), \([^)]*\))/\.Warn("%v", \1, \2)/g' "$file"
    sed -i 's/\.Warn(\([^"]*[^,)]\))/\.Warn("%v", \1)/g' "$file"
    
    # Fix Fatal calls
    sed -i 's/\.Fatal(\([^"]*[^,)]\), \([^)]*\))/\.Fatal("%v", \1, \2)/g' "$file"
    sed -i 's/\.Fatal(\([^"]*[^,)]\))/\.Fatal("%v", \1)/g' "$file"
    
    echo "✅ Fixed logger calls in: $file"
}

# Find all Go files with logger issues
echo "🔍 Finding files with logger issues..."

# Files with logger issues based on the test output
FILES_TO_FIX=(
    "internal/auth/infrastructure/security/jwt.go"
    "internal/auth/infrastructure/security/password.go"
    "internal/payment/service.go"
    "cmd/payment-service/main.go"
    "internal/security-gateway/infrastructure/redis_client.go"
    "internal/security-gateway/application/gateway_service.go"
    "internal/security-gateway/application/rate_limit_service.go"
    "internal/security-gateway/application/waf_service.go"
    "cmd/redis-mcp-demo/main.go"
    "cmd/redis-mcp-server/main.go"
)

# Fix each file
for file in "${FILES_TO_FIX[@]}"; do
    if [ -f "$file" ]; then
        fix_logger_calls "$file"
    else
        echo "⚠️  File not found: $file"
    fi
done

echo "🎉 Logger formatting issues fixed!"
echo "📝 Backup files created with .bak extension"
echo "🧪 Run 'go test ./...' to verify fixes"
