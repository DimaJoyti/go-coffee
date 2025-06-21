#!/bin/bash

# Script to fix logger formatting directive issues across the project
# This converts zap-style logger calls to use the custom logger's field-based methods

echo "🔧 Fixing logger formatting directive issues..."

# Function to fix a single file
fix_file() {
    local file="$1"
    echo "  📝 Fixing $file"
    
    # Create a backup
    cp "$file" "$file.bak"
    
    # Remove zap import if it exists alongside our custom logger
    if grep -q 'github.com/DimaJoyti/go-coffee/pkg/logger' "$file" && grep -q 'go.uber.org/zap' "$file"; then
        sed -i '/^\s*"go\.uber\.org\/zap"/d' "$file"
        # Also remove empty lines left by import removal
        sed -i '/^$/N;/^\n$/d' "$file"
    fi
    
    # Fix single field logger calls
    # Pattern: logger.Info("message", zap.String("key", value))
    sed -i 's/\(logger\.\(Info\|Warn\|Error\)\)("\([^"]*\)", zap\.\w\+("\([^"]*\)", \([^)]*\)))/logger.WithField("\4", \5).\2("\3")/g' "$file"
    
    # Fix multi-field logger calls (more complex - we'll handle these manually if needed)
    
    echo "  ✅ Fixed $file"
}

# Find all Go files with logger formatting issues
echo "🔍 Finding files with logger issues..."

# Get list of files from go vet output
go vet ./... 2>&1 | grep "formatting directive" | cut -d: -f1 | sort -u > /tmp/logger_issues.txt

if [ ! -s /tmp/logger_issues.txt ]; then
    echo "✅ No logger formatting issues found!"
    exit 0
fi

echo "📋 Files to fix:"
cat /tmp/logger_issues.txt

# Fix each file
while IFS= read -r file; do
    if [ -f "$file" ]; then
        fix_file "$file"
    fi
done < /tmp/logger_issues.txt

echo ""
echo "🧪 Running go vet to check if issues are resolved..."
remaining_issues=$(go vet ./... 2>&1 | grep "formatting directive" | wc -l)

if [ "$remaining_issues" -eq 0 ]; then
    echo "✅ All logger formatting issues resolved!"
    # Clean up backup files
    find . -name "*.go.bak" -delete
else
    echo "⚠️  Still have $remaining_issues logger formatting issues"
    echo "📋 Remaining issues:"
    go vet ./... 2>&1 | grep "formatting directive"
    echo ""
    echo "💡 Some complex multi-field logger calls may need manual fixing"
    echo "🔄 Backup files saved with .bak extension"
fi

# Clean up
rm -f /tmp/logger_issues.txt

echo "🏁 Logger fixing script completed!"