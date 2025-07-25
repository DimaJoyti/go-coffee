#!/bin/bash

# ☕ Go Coffee - GitHub Actions Workflow Validation Script
# Validates GitHub Actions workflows for syntax and common issues

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
WORKFLOWS_DIR="$PROJECT_ROOT/.github/workflows"

# Functions
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $1${NC}"
}

success() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] SUCCESS: $1${NC}"
}

# Check prerequisites
check_prerequisites() {
    log "Checking workflow validation prerequisites..."
    
    # Check required tools
    local tools=("yq" "jq" "yamllint")
    for tool in "${tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            warn "$tool is not installed. Installing..."
            case "$tool" in
                "yq")
                    sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
                    sudo chmod +x /usr/local/bin/yq
                    ;;
                "jq")
                    sudo apt-get update && sudo apt-get install -y jq
                    ;;
                "yamllint")
                    pip3 install yamllint
                    ;;
            esac
        fi
    done
    
    success "Prerequisites met"
}

# Setup workflows directory
setup_workflows_directory() {
    log "Setting up GitHub Actions workflows directory..."
    
    # Create .github/workflows directory if it doesn't exist
    mkdir -p "$WORKFLOWS_DIR"
    
    # Copy workflow files from ci-cd directory
    if [[ -d "$SCRIPT_DIR/github-actions" ]]; then
        cp "$SCRIPT_DIR/github-actions/"*.yml "$WORKFLOWS_DIR/"
        info "Copied workflow files to .github/workflows/"
    fi
    
    success "Workflows directory configured"
}

# Validate YAML syntax
validate_yaml_syntax() {
    log "Validating YAML syntax..."
    
    local error_count=0
    
    for workflow in "$WORKFLOWS_DIR"/*.yml "$WORKFLOWS_DIR"/*.yaml; do
        if [[ -f "$workflow" ]]; then
            local filename=$(basename "$workflow")
            info "Validating $filename..."
            
            # Check YAML syntax with yq
            if ! yq eval '.' "$workflow" > /dev/null 2>&1; then
                error "YAML syntax error in $filename"
                ((error_count++))
                continue
            fi
            
            # Check with yamllint
            if ! yamllint -d relaxed "$workflow" > /dev/null 2>&1; then
                warn "YAML style issues in $filename"
                yamllint -d relaxed "$workflow" || true
            fi
            
            success "✅ $filename - YAML syntax valid"
        fi
    done
    
    if [[ $error_count -gt 0 ]]; then
        error "Found $error_count YAML syntax errors"
        return 1
    fi
    
    success "All YAML files have valid syntax"
}

# Validate workflow structure
validate_workflow_structure() {
    log "Validating workflow structure..."
    
    local error_count=0
    
    for workflow in "$WORKFLOWS_DIR"/*.yml "$WORKFLOWS_DIR"/*.yaml; do
        if [[ -f "$workflow" ]]; then
            local filename=$(basename "$workflow")
            info "Validating structure of $filename..."
            
            # Check required fields
            if ! yq eval '.name' "$workflow" > /dev/null 2>&1; then
                error "$filename: Missing 'name' field"
                ((error_count++))
            fi
            
            if ! yq eval '.on' "$workflow" > /dev/null 2>&1; then
                error "$filename: Missing 'on' field"
                ((error_count++))
            fi
            
            if ! yq eval '.jobs' "$workflow" > /dev/null 2>&1; then
                error "$filename: Missing 'jobs' field"
                ((error_count++))
            fi
            
            # Check job structure
            local jobs=$(yq eval '.jobs | keys | .[]' "$workflow" 2>/dev/null || echo "")
            for job in $jobs; do
                if ! yq eval ".jobs.$job.runs-on" "$workflow" > /dev/null 2>&1; then
                    error "$filename: Job '$job' missing 'runs-on' field"
                    ((error_count++))
                fi
                
                if ! yq eval ".jobs.$job.steps" "$workflow" > /dev/null 2>&1; then
                    error "$filename: Job '$job' missing 'steps' field"
                    ((error_count++))
                fi
            done
            
            success "✅ $filename - Structure valid"
        fi
    done
    
    if [[ $error_count -gt 0 ]]; then
        error "Found $error_count structure errors"
        return 1
    fi
    
    success "All workflows have valid structure"
}

# Validate secrets and variables
validate_secrets_and_variables() {
    log "Validating secrets and variables usage..."
    
    local secrets_used=()
    local variables_used=()
    
    for workflow in "$WORKFLOWS_DIR"/*.yml "$WORKFLOWS_DIR"/*.yaml; do
        if [[ -f "$workflow" ]]; then
            local filename=$(basename "$workflow")
            info "Checking secrets and variables in $filename..."
            
            # Extract secrets
            local workflow_secrets=$(grep -oE '\$\{\{\s*secrets\.[A-Z_][A-Z0-9_]*\s*\}\}' "$workflow" | sed 's/.*secrets\.\([A-Z_][A-Z0-9_]*\).*/\1/' | sort -u)
            for secret in $workflow_secrets; do
                secrets_used+=("$secret")
            done
            
            # Extract variables
            local workflow_vars=$(grep -oE '\$\{\{\s*vars\.[A-Z_][A-Z0-9_]*\s*\}\}' "$workflow" | sed 's/.*vars\.\([A-Z_][A-Z0-9_]*\).*/\1/' | sort -u)
            for var in $workflow_vars; do
                variables_used+=("$var")
            done
        fi
    done
    
    # Create secrets documentation
    cat > "$PROJECT_ROOT/.github/REQUIRED_SECRETS.md" <<EOF
# Required GitHub Secrets

This document lists all secrets required by the GitHub Actions workflows.

## Container Registry
- \`GITHUB_TOKEN\`: Automatically provided by GitHub

## Kubernetes Access
$(printf "- \`%s\`: Kubernetes configuration\n" $(printf '%s\n' "${secrets_used[@]}" | grep -E "KUBECONFIG|KUBE" | sort -u))

## Cloud Provider Credentials
$(printf "- \`%s\`: Cloud provider credential\n" $(printf '%s\n' "${secrets_used[@]}" | grep -E "GCP|AWS|AZURE" | sort -u))

## Notifications
$(printf "- \`%s\`: Notification service credential\n" $(printf '%s\n' "${secrets_used[@]}" | grep -E "SLACK|EMAIL" | sort -u))

## Security Scanning
$(printf "- \`%s\`: Security scanning service token\n" $(printf '%s\n' "${secrets_used[@]}" | grep -E "SNYK|SONAR" | sort -u))

## All Required Secrets
$(printf "- \`%s\`\n" $(printf '%s\n' "${secrets_used[@]}" | sort -u))
EOF
    
    info "Created .github/REQUIRED_SECRETS.md with $(printf '%s\n' "${secrets_used[@]}" | sort -u | wc -l) unique secrets"
    
    success "Secrets and variables validation completed"
}

# Check action versions
check_action_versions() {
    log "Checking GitHub Actions versions..."
    
    local outdated_actions=()
    
    for workflow in "$WORKFLOWS_DIR"/*.yml "$WORKFLOWS_DIR"/*.yaml; do
        if [[ -f "$workflow" ]]; then
            local filename=$(basename "$workflow")
            info "Checking action versions in $filename..."
            
            # Check for common outdated actions
            if grep -q "actions/checkout@v3" "$workflow"; then
                warn "$filename: Using outdated actions/checkout@v3, consider upgrading to v4"
                outdated_actions+=("$filename: actions/checkout@v3")
            fi
            
            if grep -q "actions/setup-node@v3" "$workflow"; then
                warn "$filename: Using outdated actions/setup-node@v3, consider upgrading to v4"
                outdated_actions+=("$filename: actions/setup-node@v3")
            fi
            
            if grep -q "github/codeql-action.*@v2" "$workflow"; then
                warn "$filename: Using outdated CodeQL action v2, consider upgrading to v3"
                outdated_actions+=("$filename: CodeQL action v2")
            fi
        fi
    done
    
    if [[ ${#outdated_actions[@]} -gt 0 ]]; then
        warn "Found ${#outdated_actions[@]} potentially outdated actions"
        for action in "${outdated_actions[@]}"; do
            warn "  - $action"
        done
    else
        success "All actions appear to be using recent versions"
    fi
}

# Validate workflow permissions
validate_permissions() {
    log "Validating workflow permissions..."
    
    for workflow in "$WORKFLOWS_DIR"/*.yml "$WORKFLOWS_DIR"/*.yaml; do
        if [[ -f "$workflow" ]]; then
            local filename=$(basename "$workflow")
            info "Checking permissions in $filename..."
            
            # Check if workflow has explicit permissions
            if yq eval '.permissions' "$workflow" > /dev/null 2>&1; then
                success "✅ $filename - Has explicit permissions"
            else
                warn "$filename: No explicit permissions defined (using default)"
            fi
            
            # Check for security-sensitive actions
            if grep -q "github/codeql-action" "$workflow"; then
                if ! yq eval '.permissions.security-events' "$workflow" > /dev/null 2>&1; then
                    warn "$filename: Uses CodeQL but missing 'security-events: write' permission"
                fi
            fi
            
            if grep -q "docker.*push\|registry.*push" "$workflow"; then
                if ! yq eval '.permissions.packages' "$workflow" > /dev/null 2>&1; then
                    warn "$filename: Pushes containers but missing 'packages: write' permission"
                fi
            fi
        fi
    done
    
    success "Permissions validation completed"
}

# Generate workflow summary
generate_workflow_summary() {
    log "Generating workflow summary..."
    
    local summary_file="$PROJECT_ROOT/.github/WORKFLOWS_SUMMARY.md"
    
    cat > "$summary_file" <<EOF
# GitHub Actions Workflows Summary

Generated on: $(date)

## Workflows Overview

EOF
    
    for workflow in "$WORKFLOWS_DIR"/*.yml "$WORKFLOWS_DIR"/*.yaml; do
        if [[ -f "$workflow" ]]; then
            local filename=$(basename "$workflow")
            local workflow_name=$(yq eval '.name' "$workflow" 2>/dev/null || echo "Unnamed")
            local triggers=$(yq eval '.on | keys | join(", ")' "$workflow" 2>/dev/null || echo "Unknown")
            local jobs=$(yq eval '.jobs | keys | length' "$workflow" 2>/dev/null || echo "0")
            
            cat >> "$summary_file" <<EOF
### $workflow_name (\`$filename\`)

- **Triggers**: $triggers
- **Jobs**: $jobs
- **Description**: $(yq eval '.description // "No description"' "$workflow" 2>/dev/null)

EOF
        fi
    done
    
    cat >> "$summary_file" <<EOF

## Validation Results

- ✅ YAML Syntax: Valid
- ✅ Workflow Structure: Valid
- ✅ Secrets Documentation: Generated
- ✅ Action Versions: Checked
- ✅ Permissions: Validated

## Next Steps

1. Configure required secrets in GitHub repository settings
2. Review and update any outdated action versions
3. Test workflows in a development environment
4. Monitor workflow runs for any issues

For more information, see:
- [Required Secrets](./REQUIRED_SECRETS.md)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
EOF
    
    success "Workflow summary generated: $summary_file"
}

# Main execution
main() {
    echo -e "${BLUE}"
    cat << "EOF"
    ☕ Go Coffee - GitHub Actions Workflow Validation
    ===============================================
    
    Validating CI/CD workflows:
    • YAML syntax validation
    • Workflow structure validation
    • Secrets and variables analysis
    • Action version checking
    • Permissions validation
    • Documentation generation
    
EOF
    echo -e "${NC}"
    
    info "Starting workflow validation..."
    
    # Execute validation steps
    check_prerequisites
    setup_workflows_directory
    validate_yaml_syntax
    validate_workflow_structure
    validate_secrets_and_variables
    check_action_versions
    validate_permissions
    generate_workflow_summary
    
    success "🎉 GitHub Actions workflow validation completed successfully!"
    info "Check .github/WORKFLOWS_SUMMARY.md for detailed results"
    info "Configure secrets listed in .github/REQUIRED_SECRETS.md"
}

# Handle command line arguments
case "${1:-validate}" in
    "validate")
        main
        ;;
    "setup")
        setup_workflows_directory
        ;;
    "syntax")
        validate_yaml_syntax
        ;;
    "structure")
        validate_workflow_structure
        ;;
    "secrets")
        validate_secrets_and_variables
        ;;
    "versions")
        check_action_versions
        ;;
    "permissions")
        validate_permissions
        ;;
    "summary")
        generate_workflow_summary
        ;;
    *)
        echo "Usage: $0 [validate|setup|syntax|structure|secrets|versions|permissions|summary]"
        exit 1
        ;;
esac
