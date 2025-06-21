#!/bin/bash

echo "🔧 Fixing logger calls throughout the project..."

# Find all Go files and fix common logger issues
find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" | while read -r file; do
    echo "Processing: $file"
    
    # Fix Debug calls to use zap.String
    sed -i 's/\.Debug(\(".*"\), \(".*"\), \([^)]*\))/\.Debug(\1, zap.String(\2, \3))/g' "$file"
    
    # Fix Info calls to use zap.String  
    sed -i 's/\.Info(\(".*"\), \(".*"\), \([^)]*\))/\.Info(\1, zap.String(\2, \3))/g' "$file"
    
    # Fix Error calls to use zap.Error
    sed -i 's/\.Error(\(".*"\), \("error"\), \(err\))/\.Error(\1, zap.Error(\3))/g' "$file"
    
    # Fix Warn calls
    sed -i 's/\.Warn(\(".*"\), \(".*"\), \([^)]*\))/\.Warn(\1, zap.String(\2, \3))/g' "$file"
    
    # Add zap import if logger calls are present and zap import is missing
    if grep -q "zap\." "$file" && ! grep -q "go.uber.org/zap" "$file"; then
        # Insert zap import after package declaration
        sed -i '/^package /a\\nimport "go.uber.org/zap"' "$file"
    fi
done

echo "✅ Logger calls fixed!"

# Run gofmt to clean up formatting
echo "🧹 Running gofmt..."
find . -name "*.go" -not -path "./vendor/*" -exec gofmt -w {} \;

echo "🎉 All done!"
