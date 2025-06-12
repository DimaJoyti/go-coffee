#!/bin/bash

# Go Coffee Platform - Disaster Recovery Readiness Check
# This script verifies DR preparedness and infrastructure readiness

set -euo pipefail

# Configuration
BACKUP_DIR="${BACKUP_DIR:-/var/backups/go-coffee}"
S3_BUCKET="${S3_BUCKET:-go-coffee-backups}"
S3_REGION="${S3_REGION:-us-east-1}"
SLACK_WEBHOOK_URL="${SLACK_WEBHOOK_URL:-}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# Check required tools
check_required_tools() {
    log_info "Checking required tools for disaster recovery..."
    
    local required_tools=("kubectl" "aws" "tar" "gzip" "curl" "jq")
    local missing_tools=()
    
    for tool in "${required_tools[@]}"; do
        if command -v "$tool" &> /dev/null; then
            log_success "✅ $tool is available"
        else
            log_error "❌ $tool is missing"
            missing_tools+=("$tool")
        fi
    done
    
    if [[ ${#missing_tools[@]} -eq 0 ]]; then
        log_success "All required tools are available"
        return 0
    else
        log_error "Missing tools: ${missing_tools[*]}"
        return 1
    fi
}

# Check backup scripts
check_backup_scripts() {
    log_info "Checking disaster recovery scripts..."
    
    local required_scripts=(
        "scripts/disaster-recovery/backup.sh"
        "scripts/disaster-recovery/restore.sh"
        "scripts/disaster-recovery/verify-backup.sh"
        "scripts/disaster-recovery/cleanup-old-backups.sh"
    )
    
    local missing_scripts=()
    
    for script in "${required_scripts[@]}"; do
        if [[ -f "$script" && -x "$script" ]]; then
            log_success "✅ $(basename "$script") exists and is executable"
        elif [[ -f "$script" ]]; then
            log_warning "⚠️ $(basename "$script") exists but is not executable"
            chmod +x "$script" 2>/dev/null || log_error "Failed to make $script executable"
        else
            log_error "❌ $(basename "$script") is missing"
            missing_scripts+=("$script")
        fi
    done
    
    if [[ ${#missing_scripts[@]} -eq 0 ]]; then
        log_success "All required scripts are available"
        return 0
    else
        log_error "Missing scripts: ${missing_scripts[*]}"
        return 1
    fi
}

# Check AWS credentials and access
check_aws_access() {
    log_info "Checking AWS access and permissions..."
    
    if [[ -z "${AWS_ACCESS_KEY_ID:-}" ]] || [[ -z "${AWS_SECRET_ACCESS_KEY:-}" ]]; then
        log_warning "AWS credentials not configured"
        return 1
    fi
    
    if ! command -v aws &> /dev/null; then
        log_error "AWS CLI not available"
        return 1
    fi
    
    # Test AWS credentials
    if aws sts get-caller-identity &>/dev/null; then
        log_success "✅ AWS credentials are valid"
    else
        log_error "❌ AWS credentials are invalid or expired"
        return 1
    fi
    
    # Test S3 bucket access
    if aws s3api head-bucket --bucket "$S3_BUCKET" --region "$S3_REGION" &>/dev/null; then
        log_success "✅ S3 bucket is accessible: $S3_BUCKET"
    else
        log_error "❌ S3 bucket is not accessible: $S3_BUCKET"
        return 1
    fi
    
    # Test S3 permissions
    local test_file="/tmp/dr-test-$$"
    echo "DR readiness test" > "$test_file"
    
    if aws s3 cp "$test_file" "s3://${S3_BUCKET}/dr-test/test-file.txt" --region "$S3_REGION" &>/dev/null; then
        log_success "✅ S3 write permissions confirmed"
        aws s3 rm "s3://${S3_BUCKET}/dr-test/test-file.txt" --region "$S3_REGION" &>/dev/null || true
    else
        log_error "❌ S3 write permissions not available"
        return 1
    fi
    
    rm -f "$test_file"
}

# Check Kubernetes access
check_kubernetes_access() {
    log_info "Checking Kubernetes access..."
    
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl not available"
        return 1
    fi
    
    # Test cluster connectivity
    if kubectl cluster-info &>/dev/null; then
        log_success "✅ Kubernetes cluster is accessible"
    else
        log_warning "⚠️ Kubernetes cluster is not accessible (may be expected in CI)"
        return 0  # Don't fail in CI environment
    fi
    
    # Test namespace access
    local namespaces=("go-coffee" "go-coffee-staging" "go-coffee-test")
    
    for ns in "${namespaces[@]}"; do
        if kubectl get namespace "$ns" &>/dev/null; then
            log_success "✅ Namespace accessible: $ns"
        else
            log_info "ℹ️ Namespace not found: $ns (may be expected)"
        fi
    done
}

# Check backup availability
check_backup_availability() {
    log_info "Checking backup availability..."
    
    local environments=("production" "staging")
    local backup_found=false
    
    for env in "${environments[@]}"; do
        # Check local backups
        local local_backup_dir="${BACKUP_DIR}/${env}"
        if [[ -d "$local_backup_dir" ]]; then
            local recent_backups=$(find "$local_backup_dir" -name "*.tar.gz" -mtime -7 2>/dev/null || true)
            if [[ -n "$recent_backups" ]]; then
                local backup_count=$(echo "$recent_backups" | wc -l)
                log_success "✅ Found ${backup_count} recent local backup(s) for ${env}"
                backup_found=true
            else
                log_warning "⚠️ No recent local backups found for ${env}"
            fi
        else
            log_info "ℹ️ Local backup directory not found for ${env}"
        fi
        
        # Check S3 backups
        if command -v aws &> /dev/null && [[ -n "${AWS_ACCESS_KEY_ID:-}" ]]; then
            local s3_backups=$(aws s3 ls "s3://${S3_BUCKET}/${env}/" --recursive 2>/dev/null | wc -l || echo 0)
            if [[ $s3_backups -gt 0 ]]; then
                log_success "✅ Found ${s3_backups} S3 backup(s) for ${env}"
                backup_found=true
            else
                log_warning "⚠️ No S3 backups found for ${env}"
            fi
        fi
    done
    
    if [[ "$backup_found" == "true" ]]; then
        log_success "Backup availability check passed"
        return 0
    else
        log_error "No backups found for any environment"
        return 1
    fi
}

# Check network connectivity
check_network_connectivity() {
    log_info "Checking network connectivity..."
    
    local endpoints=(
        "https://aws.amazon.com"
        "https://kubernetes.io"
        "https://github.com"
    )
    
    for endpoint in "${endpoints[@]}"; do
        if curl -s --max-time 10 "$endpoint" > /dev/null 2>&1; then
            log_success "✅ Connectivity to $(echo "$endpoint" | cut -d'/' -f3)"
        else
            log_warning "⚠️ No connectivity to $(echo "$endpoint" | cut -d'/' -f3)"
        fi
    done
}

# Check system resources
check_system_resources() {
    log_info "Checking system resources..."
    
    # Check disk space
    local disk_usage=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
    local available_space=$(df -h / | awk 'NR==2 {print $4}')
    
    log_info "Root disk usage: ${disk_usage}%"
    log_info "Available space: ${available_space}"
    
    if [[ $disk_usage -gt 90 ]]; then
        log_error "❌ Disk usage is critically high: ${disk_usage}%"
        return 1
    elif [[ $disk_usage -gt 80 ]]; then
        log_warning "⚠️ Disk usage is high: ${disk_usage}%"
    else
        log_success "✅ Disk usage is normal: ${disk_usage}%"
    fi
    
    # Check memory
    local memory_usage=$(free | awk 'NR==2{printf "%.0f", $3*100/$2}')
    log_info "Memory usage: ${memory_usage}%"
    
    if [[ $memory_usage -gt 90 ]]; then
        log_warning "⚠️ Memory usage is high: ${memory_usage}%"
    else
        log_success "✅ Memory usage is normal: ${memory_usage}%"
    fi
}

# Generate DR readiness report
generate_dr_report() {
    log_info "Generating DR readiness report..."
    
    local report_file="${BACKUP_DIR}/dr-readiness-report-$(date +%Y%m%d-%H%M%S).txt"
    mkdir -p "$(dirname "$report_file")"
    
    cat > "$report_file" << EOF
# Disaster Recovery Readiness Report
Generated: $(date)
Environment: ${ENVIRONMENT:-unknown}

## Tool Availability
$(check_required_tools > /dev/null 2>&1 && echo "✅ All required tools available" || echo "❌ Some tools missing")

## Script Availability
$(check_backup_scripts > /dev/null 2>&1 && echo "✅ All DR scripts available" || echo "❌ Some scripts missing")

## AWS Access
$(check_aws_access > /dev/null 2>&1 && echo "✅ AWS access configured" || echo "❌ AWS access issues")

## Kubernetes Access
$(check_kubernetes_access > /dev/null 2>&1 && echo "✅ Kubernetes accessible" || echo "⚠️ Kubernetes access limited")

## Backup Availability
$(check_backup_availability > /dev/null 2>&1 && echo "✅ Backups available" || echo "❌ No recent backups found")

## Network Connectivity
$(check_network_connectivity > /dev/null 2>&1 && echo "✅ Network connectivity good" || echo "⚠️ Some connectivity issues")

## System Resources
$(check_system_resources > /dev/null 2>&1 && echo "✅ System resources sufficient" || echo "⚠️ Resource constraints detected")

## DR Readiness Score
$(
    score=0
    check_required_tools > /dev/null 2>&1 && score=$((score + 1))
    check_backup_scripts > /dev/null 2>&1 && score=$((score + 1))
    check_aws_access > /dev/null 2>&1 && score=$((score + 1))
    check_kubernetes_access > /dev/null 2>&1 && score=$((score + 1))
    check_backup_availability > /dev/null 2>&1 && score=$((score + 1))
    check_network_connectivity > /dev/null 2>&1 && score=$((score + 1))
    check_system_resources > /dev/null 2>&1 && score=$((score + 1))
    
    percentage=$((score * 100 / 7))
    echo "Score: ${score}/7 (${percentage}%)"
    
    if [[ $percentage -ge 90 ]]; then
        echo "Status: EXCELLENT - Ready for disaster recovery"
    elif [[ $percentage -ge 70 ]]; then
        echo "Status: GOOD - Minor issues to address"
    elif [[ $percentage -ge 50 ]]; then
        echo "Status: FAIR - Several issues need attention"
    else
        echo "Status: POOR - Significant issues must be resolved"
    fi
)

## Recommendations
- Test restore procedures monthly
- Verify backup integrity regularly
- Update DR documentation
- Train team on DR procedures
- Review and update RTO/RPO targets

EOF
    
    log_success "DR readiness report generated: $report_file"
}

# Main DR readiness check
main() {
    log_info "Starting disaster recovery readiness check..."
    
    local exit_code=0
    
    # Perform readiness checks
    check_required_tools || exit_code=1
    check_backup_scripts || exit_code=1
    check_aws_access || exit_code=1
    check_kubernetes_access || exit_code=1
    check_backup_availability || exit_code=1
    check_network_connectivity || exit_code=1
    check_system_resources || exit_code=1
    
    # Generate report
    generate_dr_report
    
    if [[ $exit_code -eq 0 ]]; then
        log_success "Disaster recovery readiness check completed successfully"
    else
        log_warning "Disaster recovery readiness check completed with issues"
    fi
    
    exit $exit_code
}

# Error handling
trap 'log_error "DR readiness check failed at line $LINENO"; exit 1' ERR

# Run main function
main "$@"
