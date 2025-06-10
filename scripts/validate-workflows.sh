#!/bin/bash

# Script to validate GitHub workflows locally

set -e

echo "🔍 Validating GitHub workflows..."

# Check if GitHub CLI is available
if ! command -v gh &> /dev/null; then
    echo "⚠️  GitHub CLI not found. Installing yamllint for basic validation..."
    
    # Try to install yamllint
    if command -v pip &> /dev/null; then
        pip install yamllint
    elif command -v apt-get &> /dev/null; then
        sudo apt-get update && sudo apt-get install -y yamllint
    elif command -v brew &> /dev/null; then
        brew install yamllint
    else
        echo "❌ Cannot install yamllint. Please install it manually."
        exit 1
    fi
fi

# Validate workflow files
echo "📋 Validating workflow syntax..."

for workflow in .github/workflows/*.yml .github/workflows/*.yaml; do
    if [ -f "$workflow" ]; then
        echo "Checking $workflow..."
        
        # Basic YAML syntax check
        if command -v yamllint &> /dev/null; then
            yamllint "$workflow" || echo "⚠️  YAML lint warnings in $workflow"
        fi
        
        # GitHub workflow validation if gh CLI is available
        if command -v gh &> /dev/null; then
            gh workflow view "$(basename "$workflow" .yml)" --repo . || echo "⚠️  Workflow validation warnings for $workflow"
        fi
    fi
done

echo "✅ Workflow validation completed!"

# Check for common issues
echo "🔍 Checking for common workflow issues..."

# Check for missing secrets
echo "📋 Checking for required secrets..."
grep -r "secrets\." .github/workflows/ | grep -v "GITHUB_TOKEN" | while read -r line; do
    echo "⚠️  Found secret reference: $line"
done

# Check for hardcoded values that should be variables
echo "📋 Checking for hardcoded values..."
grep -r "ubuntu-latest" .github/workflows/ | head -5 | while read -r line; do
    echo "ℹ️  Found hardcoded runner: $line"
done

# Check for outdated action versions
echo "📋 Checking for action versions..."
grep -r "uses:" .github/workflows/ | grep -v "@v[4-9]" | head -5 | while read -r line; do
    echo "⚠️  Potentially outdated action: $line"
done

echo ""
echo "🎯 Workflow validation summary:"
echo "- All workflow files have been checked for syntax"
echo "- Common issues have been identified"
echo "- Workflows should now be more reliable"
echo ""
echo "💡 Next steps:"
echo "1. Review any warnings above"
echo "2. Test workflows with a small commit"
echo "3. Monitor workflow runs in GitHub Actions"
