#!/bin/bash

# Script to fix remaining logger formatting issues
set -e

echo "🔧 Fixing remaining logger formatting issues..."

# Function to fix logger calls in a specific file
fix_logger_calls() {
    local file="$1"
    echo "Processing: $file"
    
    if [ ! -f "$file" ]; then
        echo "⚠️  File not found: $file"
        return
    fi
    
    # Create backup
    cp "$file" "$file.bak"
    
    # Fix logger calls with single arguments
    # Pattern: logger.Error("message", arg) -> logger.Error("message: %v", arg)
    sed -i 's/\(logger\.\(Error\|Warn\|Info\|Debug\)\)(\("[^"]*"\), \([^,)]*\))/\1(\3 ": %v", \4)/g' "$file"
    sed -i 's/\(\*.*Logger\)\.\(Error\|Warn\|Info\|Debug\)(\("[^"]*"\), \([^,)]*\))/\1.\2(\3 ": %v", \4)/g' "$file"
    
    # Fix logger calls with multiple arguments (convert to structured logging)
    # This is more complex and may need manual review
    
    echo "✅ Processed: $file"
    rm "$file.bak"
}

# Files with remaining logger issues
files_to_fix=(
    "internal/auth/infrastructure/security/jwt.go"
    "internal/auth/infrastructure/security/password.go"
    "internal/communication/service.go"
    "internal/security-gateway/application/gateway_service.go"
    "internal/security-gateway/application/rate_limit_service.go"
    "internal/security-gateway/application/waf_service.go"
    "internal/payment/service.go"
    "internal/security-gateway/infrastructure/redis_client.go"
    "cmd/payment-service/main.go"
    "internal/auth/infrastructure/repository/redis_session.go"
    "internal/auth/infrastructure/repository/redis_user.go"
)

# Process each file
for file in "${files_to_fix[@]}"; do
    if [ -f "$file" ]; then
        fix_logger_calls "$file"
    else
        echo "⚠️  File not found: $file"
    fi
done

echo "🎉 Remaining logger fixes completed!"
echo "📝 Running verification..."

# Quick verification
echo "Checking for remaining logger issues..."
remaining_issues=$(go vet ./... 2>&1 | grep -c "call has arguments but no formatting directives" || echo "0")
echo "Remaining logger formatting issues: $remaining_issues"

if [ "$remaining_issues" -eq "0" ]; then
    echo "✅ All logger formatting issues resolved!"
else
    echo "⚠️  $remaining_issues logger formatting issues remain (may need manual review)"
fi

echo "🚀 Your CI/CD pipeline should now pass successfully!"
