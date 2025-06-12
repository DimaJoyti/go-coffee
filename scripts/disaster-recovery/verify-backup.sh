#!/bin/bash

# Go Coffee Platform - Backup Verification Script
# This script verifies the integrity and completeness of backups

set -euo pipefail

# Configuration
BACKUP_DIR="${BACKUP_DIR:-/var/backups/go-coffee}"
ENVIRONMENT="${ENVIRONMENT:-production}"

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

# Verify backup integrity
verify_backup_integrity() {
    log_info "Starting backup integrity verification..."
    
    local backup_path="${BACKUP_DIR}/${ENVIRONMENT}"
    
    if [[ ! -d "$backup_path" ]]; then
        log_error "Backup directory not found: $backup_path"
        return 1
    fi
    
    # Find the latest backup
    local latest_backup=$(find "$backup_path" -name "*.tar.gz" -type f -printf '%T@ %p\n' | sort -n | tail -1 | cut -d' ' -f2-)
    
    if [[ -z "$latest_backup" ]]; then
        log_error "No backup files found in $backup_path"
        return 1
    fi
    
    log_info "Verifying backup: $(basename "$latest_backup")"
    
    # Check if backup file exists and is readable
    if [[ ! -r "$latest_backup" ]]; then
        log_error "Backup file is not readable: $latest_backup"
        return 1
    fi
    
    # Verify file size (should be > 0)
    local file_size=$(stat -c%s "$latest_backup")
    if [[ $file_size -eq 0 ]]; then
        log_error "Backup file is empty: $latest_backup"
        return 1
    fi
    
    log_info "Backup file size: $(numfmt --to=iec $file_size)"
    
    # Verify tar archive integrity
    if tar -tzf "$latest_backup" > /dev/null 2>&1; then
        log_success "Backup archive integrity verified"
    else
        log_error "Backup archive is corrupted"
        return 1
    fi
    
    # Verify checksums if available
    local backup_dir=$(dirname "$latest_backup")
    local backup_name=$(basename "$latest_backup")
    
    if [[ -f "${backup_dir}/${backup_name}.sha256" ]]; then
        log_info "Verifying SHA256 checksum..."
        cd "$backup_dir"
        if sha256sum -c "${backup_name}.sha256" > /dev/null 2>&1; then
            log_success "SHA256 checksum verified"
        else
            log_error "SHA256 checksum verification failed"
            return 1
        fi
    else
        log_warning "SHA256 checksum file not found"
    fi
    
    if [[ -f "${backup_dir}/${backup_name}.md5" ]]; then
        log_info "Verifying MD5 checksum..."
        cd "$backup_dir"
        if md5sum -c "${backup_name}.md5" > /dev/null 2>&1; then
            log_success "MD5 checksum verified"
        else
            log_error "MD5 checksum verification failed"
            return 1
        fi
    else
        log_warning "MD5 checksum file not found"
    fi
    
    log_success "Backup integrity verification completed successfully"
}

# Test backup extraction
test_backup_extraction() {
    log_info "Testing backup extraction..."
    
    local backup_path="${BACKUP_DIR}/${ENVIRONMENT}"
    local latest_backup=$(find "$backup_path" -name "*.tar.gz" -type f -printf '%T@ %p\n' | sort -n | tail -1 | cut -d' ' -f2-)
    
    if [[ -z "$latest_backup" ]]; then
        log_error "No backup files found for extraction test"
        return 1
    fi
    
    # Create temporary directory for extraction test
    local test_dir="/tmp/backup-extraction-test-$$"
    mkdir -p "$test_dir"
    
    # Extract backup to test directory
    if tar -xzf "$latest_backup" -C "$test_dir" > /dev/null 2>&1; then
        log_success "Backup extraction test passed"
        
        # List extracted contents
        log_info "Extracted backup contents:"
        find "$test_dir" -type f | head -10 | while read -r file; do
            echo "  - $(basename "$file")"
        done
        
        # Cleanup test directory
        rm -rf "$test_dir"
    else
        log_error "Backup extraction test failed"
        rm -rf "$test_dir"
        return 1
    fi
}

# Verify backup completeness
verify_backup_completeness() {
    log_info "Verifying backup completeness..."
    
    local backup_path="${BACKUP_DIR}/${ENVIRONMENT}"
    local latest_backup=$(find "$backup_path" -name "*.tar.gz" -type f -printf '%T@ %p\n' | sort -n | tail -1 | cut -d' ' -f2-)
    
    if [[ -z "$latest_backup" ]]; then
        log_error "No backup files found for completeness check"
        return 1
    fi
    
    # Check for required backup components
    local required_components=("metadata" "configs" "logs")
    local missing_components=()
    
    for component in "${required_components[@]}"; do
        if tar -tzf "$latest_backup" | grep -q "/$component/" 2>/dev/null; then
            log_success "Found backup component: $component"
        else
            log_warning "Missing backup component: $component"
            missing_components+=("$component")
        fi
    done
    
    # Check for optional components
    local optional_components=("database" "redis" "secrets" "volumes")
    
    for component in "${optional_components[@]}"; do
        if tar -tzf "$latest_backup" | grep -q "/$component/" 2>/dev/null; then
            log_success "Found optional component: $component"
        else
            log_info "Optional component not found: $component"
        fi
    done
    
    if [[ ${#missing_components[@]} -eq 0 ]]; then
        log_success "Backup completeness verification passed"
    else
        log_warning "Backup completeness verification completed with warnings"
        log_warning "Missing components: ${missing_components[*]}"
    fi
}

# Generate verification report
generate_verification_report() {
    log_info "Generating verification report..."
    
    local backup_path="${BACKUP_DIR}/${ENVIRONMENT}"
    local report_file="${backup_path}/verification-report-$(date +%Y%m%d-%H%M%S).txt"
    
    cat > "$report_file" << EOF
# Backup Verification Report
Generated: $(date)
Environment: ${ENVIRONMENT}
Backup Directory: ${backup_path}

## Backup Files
$(find "$backup_path" -name "*.tar.gz" -type f -exec ls -lh {} \; 2>/dev/null || echo "No backup files found")

## Verification Results
- Integrity Check: $(tar -tzf "$(find "$backup_path" -name "*.tar.gz" -type f | head -1)" > /dev/null 2>&1 && echo "PASSED" || echo "FAILED")
- Extraction Test: $(test_backup_extraction > /dev/null 2>&1 && echo "PASSED" || echo "FAILED")
- Completeness Check: COMPLETED

## Recommendations
- Regular verification should be performed
- Test restore procedures periodically
- Monitor backup storage usage
- Ensure backup retention policies are followed

EOF
    
    log_success "Verification report generated: $report_file"
}

# Main verification function
main() {
    log_info "Starting backup verification for environment: ${ENVIRONMENT}"
    
    # Run verification checks
    verify_backup_integrity
    test_backup_extraction
    verify_backup_completeness
    generate_verification_report
    
    log_success "Backup verification completed successfully"
}

# Error handling
trap 'log_error "Backup verification failed at line $LINENO"; exit 1' ERR

# Run main function
main "$@"
