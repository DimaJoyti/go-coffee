#!/bin/bash

# Fix GitHub Actions Workflow Syntax Issues
# This script fixes common syntax issues in GitHub Actions workflow files

set -euo pipefail

echo "🔧 Fixing GitHub Actions workflow syntax issues..."

# Function to fix double quotes in GitHub Actions expressions
fix_github_expressions() {
    local file="$1"
    echo "Checking $file..."
    
    # Fix double quotes around 'development' in GitHub Actions expressions
    sed -i 's/github\.event\.inputs\.environment || "development"/github.event.inputs.environment || '\''development'\''/g' "$file"
    
    # Fix double quotes around 'staging' in GitHub Actions expressions
    sed -i 's/github\.event\.inputs\.environment || "staging"/github.event.inputs.environment || '\''staging'\''/g' "$file"
    
    # Fix double quotes around 'production' in GitHub Actions expressions
    sed -i 's/github\.event\.inputs\.environment || "production"/github.event.inputs.environment || '\''production'\''/g' "$file"
    
    echo "✅ Fixed expressions in $file"
}

# Function to validate YAML syntax
validate_yaml() {
    local file="$1"
    if command -v yamllint &> /dev/null; then
        if yamllint "$file" 2>/dev/null; then
            echo "✅ YAML syntax valid: $file"
        else
            echo "⚠️  YAML syntax issues in: $file"
        fi
    else
        echo "ℹ️  yamllint not available, skipping validation for $file"
    fi
}

# Process all workflow files
for workflow in .github/workflows/*.yml .github/workflows/*.yaml; do
    if [ -f "$workflow" ]; then
        echo "Processing $workflow..."
        
        # Create backup
        cp "$workflow" "$workflow.backup"
        
        # Fix GitHub Actions expressions
        fix_github_expressions "$workflow"
        
        # Validate YAML syntax
        validate_yaml "$workflow"
        
        echo "---"
    fi
done

echo "🎉 Workflow syntax fix completed!"
echo ""
echo "📋 Summary:"
echo "- Fixed GitHub Actions expressions with incorrect quote usage"
echo "- Validated YAML syntax where possible"
echo "- Created backup files (.backup extension)"
echo ""
echo "🔍 To verify the changes:"
echo "  git diff .github/workflows/"
echo ""
echo "🧹 To clean up backup files:"
echo "  rm .github/workflows/*.backup"
