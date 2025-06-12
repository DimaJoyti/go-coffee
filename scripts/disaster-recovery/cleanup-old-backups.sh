#!/bin/bash

# Go Coffee Platform - Backup Cleanup Script
# This script removes old backups based on retention policies

set -euo pipefail

# Configuration
BACKUP_DIR="${BACKUP_DIR:-/var/backups/go-coffee}"
RETENTION_DAYS="${RETENTION_DAYS:-90}"
S3_BUCKET="${S3_BUCKET:-go-coffee-backups}"
S3_REGION="${S3_REGION:-us-east-1}"
DRY_RUN="${DRY_RUN:-false}"

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

# Cleanup local backups
cleanup_local_backups() {
    log_info "Starting local backup cleanup..."
    log_info "Retention policy: ${RETENTION_DAYS} days"
    
    if [[ ! -d "$BACKUP_DIR" ]]; then
        log_warning "Backup directory not found: $BACKUP_DIR"
        return 0
    fi
    
    # Find old backup files
    local old_files=$(find "$BACKUP_DIR" -name "*.tar.gz" -mtime +${RETENTION_DAYS} 2>/dev/null || true)
    local old_checksums=$(find "$BACKUP_DIR" -name "*.sha256" -o -name "*.md5" -mtime +${RETENTION_DAYS} 2>/dev/null || true)
    
    if [[ -z "$old_files" && -z "$old_checksums" ]]; then
        log_info "No old backup files found for cleanup"
        return 0
    fi
    
    # Count files to be deleted
    local file_count=0
    if [[ -n "$old_files" ]]; then
        file_count=$(echo "$old_files" | wc -l)
    fi
    
    local checksum_count=0
    if [[ -n "$old_checksums" ]]; then
        checksum_count=$(echo "$old_checksums" | wc -l)
    fi
    
    log_info "Found ${file_count} old backup files and ${checksum_count} checksum files for cleanup"
    
    # Calculate space to be freed
    local space_to_free=0
    if [[ -n "$old_files" ]]; then
        while IFS= read -r file; do
            if [[ -f "$file" ]]; then
                local file_size=$(stat -c%s "$file" 2>/dev/null || echo 0)
                space_to_free=$((space_to_free + file_size))
            fi
        done <<< "$old_files"
    fi
    
    log_info "Space to be freed: $(numfmt --to=iec $space_to_free)"
    
    # Perform cleanup
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would delete the following files:"
        if [[ -n "$old_files" ]]; then
            echo "$old_files"
        fi
        if [[ -n "$old_checksums" ]]; then
            echo "$old_checksums"
        fi
    else
        # Delete old backup files
        if [[ -n "$old_files" ]]; then
            while IFS= read -r file; do
                if [[ -f "$file" ]]; then
                    log_info "Deleting old backup: $(basename "$file")"
                    rm -f "$file"
                fi
            done <<< "$old_files"
        fi
        
        # Delete old checksum files
        if [[ -n "$old_checksums" ]]; then
            while IFS= read -r file; do
                if [[ -f "$file" ]]; then
                    log_info "Deleting old checksum: $(basename "$file")"
                    rm -f "$file"
                fi
            done <<< "$old_checksums"
        fi
        
        log_success "Local backup cleanup completed"
    fi
}

# Cleanup S3 backups
cleanup_s3_backups() {
    if [[ -z "${AWS_ACCESS_KEY_ID:-}" ]] || [[ -z "${AWS_SECRET_ACCESS_KEY:-}" ]]; then
        log_warning "AWS credentials not found, skipping S3 cleanup"
        return 0
    fi
    
    if ! command -v aws &> /dev/null; then
        log_warning "AWS CLI not found, skipping S3 cleanup"
        return 0
    fi
    
    log_info "Starting S3 backup cleanup..."
    
    # Check if bucket exists
    if ! aws s3api head-bucket --bucket "$S3_BUCKET" --region "$S3_REGION" &>/dev/null; then
        log_warning "S3 bucket not found or not accessible: $S3_BUCKET"
        return 0
    fi
    
    # Calculate cutoff date
    local cutoff_date
    if command -v gdate &> /dev/null; then
        # macOS
        cutoff_date=$(gdate -d "${RETENTION_DAYS} days ago" +%Y-%m-%d)
    else
        # Linux
        cutoff_date=$(date -d "${RETENTION_DAYS} days ago" +%Y-%m-%d)
    fi
    
    log_info "S3 cleanup cutoff date: $cutoff_date"
    
    # List old objects
    local old_objects=$(aws s3api list-objects-v2 \
        --bucket "$S3_BUCKET" \
        --query "Contents[?LastModified<='${cutoff_date}T23:59:59.000Z'].Key" \
        --output text \
        --region "$S3_REGION" 2>/dev/null || echo "")
    
    if [[ -z "$old_objects" || "$old_objects" == "None" ]]; then
        log_info "No old S3 objects found for cleanup"
        return 0
    fi
    
    # Count objects and calculate size
    local object_count=$(echo "$old_objects" | wc -w)
    log_info "Found ${object_count} old S3 objects for cleanup"
    
    # Perform S3 cleanup
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would delete the following S3 objects:"
        echo "$old_objects" | tr '\t' '\n'
    else
        echo "$old_objects" | tr '\t' '\n' | while read -r key; do
            if [[ -n "$key" && "$key" != "None" ]]; then
                log_info "Deleting S3 object: $key"
                aws s3 rm "s3://${S3_BUCKET}/${key}" --region "$S3_REGION" || log_warning "Failed to delete: $key"
            fi
        done
        
        log_success "S3 backup cleanup completed"
    fi
}

# Cleanup empty directories
cleanup_empty_directories() {
    log_info "Cleaning up empty directories..."
    
    if [[ ! -d "$BACKUP_DIR" ]]; then
        return 0
    fi
    
    # Find and remove empty directories
    local empty_dirs=$(find "$BACKUP_DIR" -type d -empty 2>/dev/null || true)
    
    if [[ -z "$empty_dirs" ]]; then
        log_info "No empty directories found"
        return 0
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would remove the following empty directories:"
        echo "$empty_dirs"
    else
        while IFS= read -r dir; do
            if [[ -d "$dir" && "$dir" != "$BACKUP_DIR" ]]; then
                log_info "Removing empty directory: $dir"
                rmdir "$dir" 2>/dev/null || true
            fi
        done <<< "$empty_dirs"
        
        log_success "Empty directory cleanup completed"
    fi
}

# Generate cleanup report
generate_cleanup_report() {
    log_info "Generating cleanup report..."
    
    local report_file="${BACKUP_DIR}/cleanup-report-$(date +%Y%m%d-%H%M%S).txt"
    mkdir -p "$(dirname "$report_file")"
    
    cat > "$report_file" << EOF
# Backup Cleanup Report
Generated: $(date)
Retention Policy: ${RETENTION_DAYS} days
Dry Run: ${DRY_RUN}

## Local Backup Status
Backup Directory: ${BACKUP_DIR}
$(if [[ -d "$BACKUP_DIR" ]]; then
    echo "Current backup files:"
    find "$BACKUP_DIR" -name "*.tar.gz" -type f -exec ls -lh {} \; 2>/dev/null | head -10
    echo ""
    echo "Total backup files: $(find "$BACKUP_DIR" -name "*.tar.gz" -type f | wc -l)"
    echo "Total backup size: $(find "$BACKUP_DIR" -name "*.tar.gz" -type f -exec stat -c%s {} \; 2>/dev/null | awk '{sum+=$1} END {print sum}' | numfmt --to=iec)"
else
    echo "Backup directory not found"
fi)

## S3 Backup Status
$(if command -v aws &> /dev/null && [[ -n "${AWS_ACCESS_KEY_ID:-}" ]]; then
    echo "S3 Bucket: ${S3_BUCKET}"
    aws s3 ls "s3://${S3_BUCKET}/" --recursive --human-readable --summarize 2>/dev/null | tail -2 || echo "Unable to access S3 bucket"
else
    echo "AWS CLI not available or credentials not configured"
fi)

## Cleanup Summary
- Local cleanup: $(if [[ "$DRY_RUN" == "true" ]]; then echo "DRY RUN"; else echo "EXECUTED"; fi)
- S3 cleanup: $(if [[ "$DRY_RUN" == "true" ]]; then echo "DRY RUN"; else echo "EXECUTED"; fi)
- Empty directory cleanup: $(if [[ "$DRY_RUN" == "true" ]]; then echo "DRY RUN"; else echo "EXECUTED"; fi)

## Recommendations
- Monitor backup storage usage regularly
- Adjust retention policies based on compliance requirements
- Verify backup integrity before cleanup
- Test restore procedures periodically

EOF
    
    log_success "Cleanup report generated: $report_file"
}

# Main cleanup function
main() {
    log_info "Starting backup cleanup process..."
    log_info "Retention policy: ${RETENTION_DAYS} days"
    log_info "Dry run mode: ${DRY_RUN}"
    
    # Perform cleanup operations
    cleanup_local_backups
    cleanup_s3_backups
    cleanup_empty_directories
    generate_cleanup_report
    
    log_success "Backup cleanup process completed"
}

# Error handling
trap 'log_error "Backup cleanup failed at line $LINENO"; exit 1' ERR

# Run main function
main "$@"
