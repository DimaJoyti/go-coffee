#!/bin/bash

# Go Coffee Platform - Backup Monitoring Script
# This script monitors backup health and sends alerts for issues

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
BACKUP_DIR="${BACKUP_DIR:-/var/backups/go-coffee}"
S3_BUCKET="${S3_BUCKET:-go-coffee-backups}"
S3_REGION="${S3_REGION:-us-east-1}"
ALERT_THRESHOLD_HOURS="${ALERT_THRESHOLD_HOURS:-25}"  # Alert if backup is older than 25 hours
STORAGE_WARNING_THRESHOLD="${STORAGE_WARNING_THRESHOLD:-80}"  # Warning at 80% storage
STORAGE_CRITICAL_THRESHOLD="${STORAGE_CRITICAL_THRESHOLD:-95}"  # Critical at 95% storage

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

# Check local backup status
check_local_backups() {
    log_info "Checking local backup status..."
    
    local issues=0
    local environments=("production" "staging" "development")
    
    for env in "${environments[@]}"; do
        local backup_path="${BACKUP_DIR}/${env}"
        
        if [[ ! -d "$backup_path" ]]; then
            log_warning "Backup directory not found for environment: $env"
            ((issues++))
            continue
        fi
        
        # Find latest backup
        local latest_backup=$(find "$backup_path" -name "*.tar.gz" -type f -printf '%T@ %p\n' | sort -n | tail -1 | cut -d' ' -f2-)
        
        if [[ -z "$latest_backup" ]]; then
            log_error "No backups found for environment: $env"
            ((issues++))
            continue
        fi
        
        # Check backup age
        local backup_age_hours=$(( ($(date +%s) - $(stat -c %Y "$latest_backup")) / 3600 ))
        
        if [[ $backup_age_hours -gt $ALERT_THRESHOLD_HOURS ]]; then
            log_error "Backup for $env is too old: ${backup_age_hours} hours (threshold: ${ALERT_THRESHOLD_HOURS}h)"
            ((issues++))
        else
            log_success "Backup for $env is current: ${backup_age_hours} hours old"
        fi
        
        # Check backup integrity
        if [[ -f "${latest_backup}.sha256" ]]; then
            cd "$(dirname "$latest_backup")"
            if sha256sum -c "$(basename "$latest_backup").sha256" &>/dev/null; then
                log_success "Backup integrity verified for $env"
            else
                log_error "Backup integrity check failed for $env"
                ((issues++))
            fi
        else
            log_warning "No checksum file found for $env backup"
        fi
    done
    
    return $issues
}

# Check S3 backup status
check_s3_backups() {
    log_info "Checking S3 backup status..."
    
    if [[ -z "${AWS_ACCESS_KEY_ID:-}" ]] || [[ -z "${AWS_SECRET_ACCESS_KEY:-}" ]]; then
        log_warning "AWS credentials not found, skipping S3 backup check"
        return 0
    fi
    
    local issues=0
    local environments=("production" "staging")
    
    for env in "${environments[@]}"; do
        # Get latest backup from S3
        local latest_s3_backup=$(aws s3 ls "s3://${S3_BUCKET}/${env}/" --recursive | sort | tail -1)
        
        if [[ -z "$latest_s3_backup" ]]; then
            log_error "No S3 backups found for environment: $env"
            ((issues++))
            continue
        fi
        
        # Extract backup date from S3 listing
        local backup_date=$(echo "$latest_s3_backup" | awk '{print $1 " " $2}')
        local backup_timestamp=$(date -d "$backup_date" +%s)
        local current_timestamp=$(date +%s)
        local backup_age_hours=$(( (current_timestamp - backup_timestamp) / 3600 ))
        
        if [[ $backup_age_hours -gt $ALERT_THRESHOLD_HOURS ]]; then
            log_error "S3 backup for $env is too old: ${backup_age_hours} hours"
            ((issues++))
        else
            log_success "S3 backup for $env is current: ${backup_age_hours} hours old"
        fi
    done
    
    return $issues
}

# Check storage usage
check_storage_usage() {
    log_info "Checking storage usage..."
    
    local issues=0
    
    # Check local storage
    if [[ -d "$BACKUP_DIR" ]]; then
        local usage=$(df "$BACKUP_DIR" | awk 'NR==2 {print $5}' | sed 's/%//')
        
        if [[ $usage -ge $STORAGE_CRITICAL_THRESHOLD ]]; then
            log_error "Local backup storage critically high: ${usage}% (threshold: ${STORAGE_CRITICAL_THRESHOLD}%)"
            ((issues++))
        elif [[ $usage -ge $STORAGE_WARNING_THRESHOLD ]]; then
            log_warning "Local backup storage high: ${usage}% (threshold: ${STORAGE_WARNING_THRESHOLD}%)"
        else
            log_success "Local backup storage usage: ${usage}%"
        fi
    fi
    
    # Check S3 storage
    if [[ -n "${AWS_ACCESS_KEY_ID:-}" ]] && [[ -n "${AWS_SECRET_ACCESS_KEY:-}" ]]; then
        local s3_size=$(aws s3api list-objects-v2 --bucket "$S3_BUCKET" --query 'sum(Contents[].Size)' --output text 2>/dev/null || echo "0")
        local s3_size_gb=$((s3_size / 1024 / 1024 / 1024))
        
        log_info "S3 backup storage usage: ${s3_size_gb} GB"
        
        # Check if approaching S3 limits (example: warn at 1TB)
        if [[ $s3_size_gb -gt 1000 ]]; then
            log_warning "S3 backup storage is large: ${s3_size_gb} GB"
        fi
    fi
    
    return $issues
}

# Check backup consistency across environments
check_backup_consistency() {
    log_info "Checking backup consistency..."
    
    local issues=0
    
    # Compare backup schedules
    local prod_backup_count=$(find "${BACKUP_DIR}/production" -name "*.tar.gz" -mtime -7 2>/dev/null | wc -l)
    local staging_backup_count=$(find "${BACKUP_DIR}/staging" -name "*.tar.gz" -mtime -7 2>/dev/null | wc -l)
    
    if [[ $prod_backup_count -lt 7 ]]; then
        log_warning "Production has fewer than expected daily backups: $prod_backup_count"
    fi
    
    if [[ $staging_backup_count -lt 7 ]]; then
        log_warning "Staging has fewer than expected daily backups: $staging_backup_count"
    fi
    
    # Check backup sizes for anomalies
    if [[ -d "${BACKUP_DIR}/production" ]]; then
        local avg_size=$(find "${BACKUP_DIR}/production" -name "*.tar.gz" -mtime -7 -exec stat -c%s {} \; | awk '{sum+=$1; count++} END {if(count>0) print sum/count; else print 0}')
        local latest_size=$(find "${BACKUP_DIR}/production" -name "*.tar.gz" -type f -printf '%T@ %s\n' | sort -n | tail -1 | cut -d' ' -f2)
        
        if [[ -n "$latest_size" ]] && [[ -n "$avg_size" ]] && [[ $avg_size -gt 0 ]]; then
            local size_ratio=$((latest_size * 100 / avg_size))
            
            if [[ $size_ratio -lt 50 ]]; then
                log_warning "Latest backup is significantly smaller than average (${size_ratio}% of average)"
                ((issues++))
            elif [[ $size_ratio -gt 200 ]]; then
                log_warning "Latest backup is significantly larger than average (${size_ratio}% of average)"
            fi
        fi
    fi
    
    return $issues
}

# Check backup automation
check_backup_automation() {
    log_info "Checking backup automation..."
    
    local issues=0
    
    # Check if backup cron job exists
    if crontab -l 2>/dev/null | grep -q "backup.sh"; then
        log_success "Backup cron job is configured"
    else
        log_warning "Backup cron job not found in current user's crontab"
    fi
    
    # Check if backup script is executable
    local backup_script="${PROJECT_ROOT}/scripts/disaster-recovery/backup.sh"
    if [[ -x "$backup_script" ]]; then
        log_success "Backup script is executable"
    else
        log_error "Backup script is not executable: $backup_script"
        ((issues++))
    fi
    
    # Check backup script dependencies
    local dependencies=("kubectl" "aws" "tar" "gzip" "sha256sum")
    for dep in "${dependencies[@]}"; do
        if command -v "$dep" &>/dev/null; then
            log_success "Dependency available: $dep"
        else
            log_error "Missing dependency: $dep"
            ((issues++))
        fi
    done
    
    return $issues
}

# Generate backup health report
generate_health_report() {
    local total_issues=$1
    
    log_info "Generating backup health report..."
    
    local report_file="/tmp/backup-health-report-$(date +%Y%m%d-%H%M%S).json"
    
    cat > "$report_file" << EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "total_issues": $total_issues,
  "status": "$(if [[ $total_issues -eq 0 ]]; then echo "healthy"; elif [[ $total_issues -lt 5 ]]; then echo "warning"; else echo "critical"; fi)",
  "checks": {
    "local_backups": "$(if check_local_backups &>/dev/null; then echo "pass"; else echo "fail"; fi)",
    "s3_backups": "$(if check_s3_backups &>/dev/null; then echo "pass"; else echo "fail"; fi)",
    "storage_usage": "$(if check_storage_usage &>/dev/null; then echo "pass"; else echo "fail"; fi)",
    "backup_consistency": "$(if check_backup_consistency &>/dev/null; then echo "pass"; else echo "fail"; fi)",
    "backup_automation": "$(if check_backup_automation &>/dev/null; then echo "pass"; else echo "fail"; fi)"
  },
  "recommendations": [
    $(if [[ $total_issues -gt 0 ]]; then echo '"Review backup logs for detailed error information",'; fi)
    $(if [[ $total_issues -gt 5 ]]; then echo '"Consider immediate backup system maintenance",'; fi)
    "Monitor backup storage usage regularly",
    "Test backup restoration procedures monthly"
  ]
}
EOF
    
    echo "$report_file"
}

# Send alerts
send_alerts() {
    local total_issues=$1
    local report_file=$2
    
    if [[ $total_issues -eq 0 ]]; then
        return 0
    fi
    
    local severity="warning"
    if [[ $total_issues -ge 5 ]]; then
        severity="critical"
    fi
    
    local message="Backup monitoring detected $total_issues issue(s). Severity: $severity"
    
    # Slack notification
    if [[ -n "${SLACK_WEBHOOK_URL:-}" ]]; then
        local color="warning"
        if [[ "$severity" == "critical" ]]; then
            color="danger"
        fi
        
        curl -X POST -H 'Content-type: application/json' \
            --data "{\"attachments\":[{\"color\":\"${color}\",\"title\":\"Backup Health Alert\",\"text\":\"${message}\",\"fields\":[{\"title\":\"Issues Found\",\"value\":\"${total_issues}\",\"short\":true},{\"title\":\"Severity\",\"value\":\"${severity}\",\"short\":true}]}]}" \
            "${SLACK_WEBHOOK_URL}"
    fi
    
    # Email notification
    if command -v mail &>/dev/null && [[ -n "${NOTIFICATION_EMAIL:-}" ]]; then
        echo "$message" | mail -s "Go Coffee Backup Health Alert - $severity" -a "$report_file" "${NOTIFICATION_EMAIL}"
    fi
    
    # PagerDuty integration (if configured)
    if [[ -n "${PAGERDUTY_INTEGRATION_KEY:-}" ]] && [[ "$severity" == "critical" ]]; then
        curl -X POST \
            -H "Content-Type: application/json" \
            -d "{
                \"routing_key\": \"${PAGERDUTY_INTEGRATION_KEY}\",
                \"event_action\": \"trigger\",
                \"payload\": {
                    \"summary\": \"Critical backup issues detected\",
                    \"source\": \"backup-monitor\",
                    \"severity\": \"critical\",
                    \"custom_details\": {
                        \"issues_count\": $total_issues,
                        \"report_file\": \"$report_file\"
                    }
                }
            }" \
            https://events.pagerduty.com/v2/enqueue
    fi
}

# Main monitoring function
main() {
    log_info "Starting backup health monitoring..."
    
    local total_issues=0
    
    # Run all checks
    check_local_backups || total_issues=$((total_issues + $?))
    check_s3_backups || total_issues=$((total_issues + $?))
    check_storage_usage || total_issues=$((total_issues + $?))
    check_backup_consistency || total_issues=$((total_issues + $?))
    check_backup_automation || total_issues=$((total_issues + $?))
    
    # Generate report
    local report_file=$(generate_health_report $total_issues)
    
    # Send alerts if needed
    send_alerts $total_issues "$report_file"
    
    # Summary
    if [[ $total_issues -eq 0 ]]; then
        log_success "Backup health monitoring completed - no issues found"
    else
        log_warning "Backup health monitoring completed - $total_issues issue(s) found"
        log_info "Detailed report: $report_file"
    fi
    
    return $total_issues
}

# Run main function
main "$@"
