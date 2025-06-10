#!/bin/bash

# Go Coffee Platform - Disaster Recovery Restoration Script
# This script restores the platform from backup files

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
BACKUP_DIR="${BACKUP_DIR:-/var/backups/go-coffee}"
ENVIRONMENT="${ENVIRONMENT:-production}"
NAMESPACE="${NAMESPACE:-go-coffee}"
S3_BUCKET="${S3_BUCKET:-go-coffee-backups}"
S3_REGION="${S3_REGION:-us-east-1}"

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

# Show help
show_help() {
    cat << EOF
Go Coffee Platform - Disaster Recovery Restoration Script

Usage: $0 [OPTIONS] BACKUP_PATH

Options:
    -e, --env           Environment (development|staging|production) [default: production]
    -n, --namespace     Kubernetes namespace [default: go-coffee]
    -s, --s3            Download backup from S3 instead of local path
    -f, --force         Force restoration without confirmation
    -d, --dry-run       Show what would be restored without actually doing it
    -h, --help          Show this help message

Examples:
    $0 /var/backups/go-coffee/production/20231201-120000.tar.gz
    $0 -s s3://go-coffee-backups/production/2023/12/01/backup.tar.gz
    $0 -e staging -f /path/to/backup.tar.gz
    $0 -d /path/to/backup.tar.gz  # Dry run

EOF
}

# Parse command line arguments
parse_args() {
    BACKUP_PATH=""
    FORCE=false
    DRY_RUN=false
    FROM_S3=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--env)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -n|--namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            -s|--s3)
                FROM_S3=true
                shift
                ;;
            -f|--force)
                FORCE=true
                shift
                ;;
            -d|--dry-run)
                DRY_RUN=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                if [[ -z "$BACKUP_PATH" ]]; then
                    BACKUP_PATH="$1"
                else
                    log_error "Unknown option: $1"
                    show_help
                    exit 1
                fi
                shift
                ;;
        esac
    done

    if [[ -z "$BACKUP_PATH" ]]; then
        log_error "Backup path is required"
        show_help
        exit 1
    fi
}

# Download backup from S3
download_from_s3() {
    log_info "Downloading backup from S3: $BACKUP_PATH"
    
    local local_backup="/tmp/$(basename "$BACKUP_PATH")"
    
    aws s3 cp "$BACKUP_PATH" "$local_backup" --region "$S3_REGION"
    aws s3 cp "$BACKUP_PATH.sha256" "$local_backup.sha256" --region "$S3_REGION"
    aws s3 cp "$BACKUP_PATH.md5" "$local_backup.md5" --region "$S3_REGION"
    
    BACKUP_PATH="$local_backup"
    log_success "Backup downloaded to: $BACKUP_PATH"
}

# Verify backup integrity
verify_backup() {
    log_info "Verifying backup integrity..."
    
    if [[ ! -f "$BACKUP_PATH" ]]; then
        log_error "Backup file not found: $BACKUP_PATH"
        exit 1
    fi
    
    # Check checksums if available
    if [[ -f "$BACKUP_PATH.sha256" ]]; then
        cd "$(dirname "$BACKUP_PATH")"
        if ! sha256sum -c "$(basename "$BACKUP_PATH").sha256"; then
            log_error "SHA256 checksum verification failed"
            exit 1
        fi
    fi
    
    if [[ -f "$BACKUP_PATH.md5" ]]; then
        cd "$(dirname "$BACKUP_PATH")"
        if ! md5sum -c "$(basename "$BACKUP_PATH").md5"; then
            log_error "MD5 checksum verification failed"
            exit 1
        fi
    fi
    
    # Test extraction
    if ! tar -tzf "$BACKUP_PATH" > /dev/null; then
        log_error "Backup file is corrupted or not a valid tar.gz file"
        exit 1
    fi
    
    log_success "Backup integrity verified"
}

# Extract backup
extract_backup() {
    log_info "Extracting backup..."
    
    RESTORE_DIR="/tmp/go-coffee-restore-$$"
    mkdir -p "$RESTORE_DIR"
    
    tar -xzf "$BACKUP_PATH" -C "$RESTORE_DIR"
    
    # Find the extracted directory
    BACKUP_CONTENT_DIR=$(find "$RESTORE_DIR" -maxdepth 1 -type d -name "*-*" | head -1)
    
    if [[ -z "$BACKUP_CONTENT_DIR" ]]; then
        log_error "Could not find backup content directory"
        exit 1
    fi
    
    log_success "Backup extracted to: $BACKUP_CONTENT_DIR"
}

# Show backup metadata
show_backup_info() {
    log_info "Backup Information:"
    
    if [[ -f "$BACKUP_CONTENT_DIR/metadata.json" ]]; then
        cat "$BACKUP_CONTENT_DIR/metadata.json" | jq .
    else
        log_warning "Backup metadata not found"
    fi
}

# Confirm restoration
confirm_restoration() {
    if [[ "$FORCE" == "true" ]]; then
        return 0
    fi
    
    echo
    log_warning "This will restore the Go Coffee platform from backup."
    log_warning "Environment: $ENVIRONMENT"
    log_warning "Namespace: $NAMESPACE"
    log_warning "This operation will:"
    echo "  - Stop all running services"
    echo "  - Restore database from backup"
    echo "  - Restore Redis data"
    echo "  - Restore Kubernetes configurations"
    echo "  - Restart all services"
    echo
    
    read -p "Are you sure you want to continue? (yes/no): " -r
    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        log_info "Restoration cancelled by user"
        exit 0
    fi
}

# Scale down services
scale_down_services() {
    log_info "Scaling down services..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would scale down all deployments in namespace $NAMESPACE"
        return 0
    fi
    
    # Scale down all deployments
    kubectl scale deployment --all --replicas=0 -n "$NAMESPACE" || true
    
    # Wait for pods to terminate
    log_info "Waiting for pods to terminate..."
    kubectl wait --for=delete pod --all -n "$NAMESPACE" --timeout=300s || true
    
    log_success "Services scaled down"
}

# Restore database
restore_database() {
    log_info "Restoring database..."
    
    if [[ ! -d "$BACKUP_CONTENT_DIR/database" ]]; then
        log_warning "Database backup not found, skipping database restoration"
        return 0
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would restore database from: $BACKUP_CONTENT_DIR/database/"
        return 0
    fi
    
    # Start PostgreSQL if not running
    kubectl scale deployment postgres --replicas=1 -n "$NAMESPACE"
    kubectl wait --for=condition=ready pod -l app=postgres -n "$NAMESPACE" --timeout=300s
    
    local db_pod=$(kubectl get pods -n "$NAMESPACE" -l app=postgres -o jsonpath='{.items[0].metadata.name}')
    
    # Drop existing databases (except system databases)
    log_info "Dropping existing databases..."
    kubectl exec "$db_pod" -n "$NAMESPACE" -- psql -U postgres -c "DROP DATABASE IF EXISTS go_coffee;"
    
    # Restore from full dump
    if [[ -f "$BACKUP_CONTENT_DIR/database/full_dump.sql" ]]; then
        log_info "Restoring from full database dump..."
        kubectl exec -i "$db_pod" -n "$NAMESPACE" -- psql -U postgres < "$BACKUP_CONTENT_DIR/database/full_dump.sql"
    else
        # Restore individual databases
        for sql_file in "$BACKUP_CONTENT_DIR/database"/*.sql; do
            if [[ -f "$sql_file" ]] && [[ "$(basename "$sql_file")" != "full_dump.sql" ]]; then
                local db_name=$(basename "$sql_file" .sql)
                log_info "Restoring database: $db_name"
                kubectl exec -i "$db_pod" -n "$NAMESPACE" -- psql -U postgres < "$sql_file"
            fi
        done
    fi
    
    log_success "Database restoration completed"
}

# Restore Redis
restore_redis() {
    log_info "Restoring Redis..."
    
    if [[ ! -d "$BACKUP_CONTENT_DIR/redis" ]]; then
        log_warning "Redis backup not found, skipping Redis restoration"
        return 0
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would restore Redis from: $BACKUP_CONTENT_DIR/redis/"
        return 0
    fi
    
    # Start Redis if not running
    kubectl scale deployment redis --replicas=1 -n "$NAMESPACE"
    kubectl wait --for=condition=ready pod -l app=redis -n "$NAMESPACE" --timeout=300s
    
    local redis_pod=$(kubectl get pods -n "$NAMESPACE" -l app=redis -o jsonpath='{.items[0].metadata.name}')
    
    # Stop Redis to restore data
    kubectl exec "$redis_pod" -n "$NAMESPACE" -- redis-cli SHUTDOWN NOSAVE || true
    
    # Copy RDB file
    if [[ -f "$BACKUP_CONTENT_DIR/redis/dump.rdb" ]]; then
        log_info "Restoring Redis RDB file..."
        kubectl cp "$BACKUP_CONTENT_DIR/redis/dump.rdb" "$NAMESPACE/$redis_pod:/data/dump.rdb"
    fi
    
    # Copy AOF file if exists
    if [[ -f "$BACKUP_CONTENT_DIR/redis/appendonly.aof" ]]; then
        log_info "Restoring Redis AOF file..."
        kubectl cp "$BACKUP_CONTENT_DIR/redis/appendonly.aof" "$NAMESPACE/$redis_pod:/data/appendonly.aof"
    fi
    
    # Restart Redis
    kubectl delete pod "$redis_pod" -n "$NAMESPACE"
    kubectl wait --for=condition=ready pod -l app=redis -n "$NAMESPACE" --timeout=300s
    
    log_success "Redis restoration completed"
}

# Restore Kubernetes configurations
restore_k8s_configs() {
    log_info "Restoring Kubernetes configurations..."
    
    if [[ ! -d "$BACKUP_CONTENT_DIR/configs" ]]; then
        log_warning "Kubernetes configurations backup not found, skipping"
        return 0
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would restore Kubernetes configurations from: $BACKUP_CONTENT_DIR/configs/"
        return 0
    fi
    
    # Restore ConfigMaps
    if [[ -f "$BACKUP_CONTENT_DIR/configs/configmaps.yaml" ]]; then
        log_info "Restoring ConfigMaps..."
        kubectl apply -f "$BACKUP_CONTENT_DIR/configs/configmaps.yaml"
    fi
    
    # Restore Secrets
    if [[ -f "$BACKUP_CONTENT_DIR/secrets/secrets.yaml" ]]; then
        log_info "Restoring Secrets..."
        kubectl apply -f "$BACKUP_CONTENT_DIR/secrets/secrets.yaml"
    fi
    
    # Restore Services
    if [[ -f "$BACKUP_CONTENT_DIR/configs/services.yaml" ]]; then
        log_info "Restoring Services..."
        kubectl apply -f "$BACKUP_CONTENT_DIR/configs/services.yaml"
    fi
    
    # Restore PVCs
    if [[ -f "$BACKUP_CONTENT_DIR/volumes/pvcs.yaml" ]]; then
        log_info "Restoring PersistentVolumeClaims..."
        kubectl apply -f "$BACKUP_CONTENT_DIR/volumes/pvcs.yaml"
    fi
    
    log_success "Kubernetes configurations restoration completed"
}

# Restore deployments
restore_deployments() {
    log_info "Restoring deployments..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would restore deployments"
        return 0
    fi
    
    # Restore Deployments
    if [[ -f "$BACKUP_CONTENT_DIR/configs/deployments.yaml" ]]; then
        log_info "Restoring Deployments..."
        kubectl apply -f "$BACKUP_CONTENT_DIR/configs/deployments.yaml"
    fi
    
    # Wait for deployments to be ready
    log_info "Waiting for deployments to be ready..."
    kubectl wait --for=condition=available deployment --all -n "$NAMESPACE" --timeout=600s
    
    log_success "Deployments restoration completed"
}

# Verify restoration
verify_restoration() {
    log_info "Verifying restoration..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would verify restoration"
        return 0
    fi
    
    # Check pod status
    local unhealthy_pods=$(kubectl get pods -n "$NAMESPACE" --field-selector=status.phase!=Running --no-headers 2>/dev/null | wc -l)
    
    if [[ $unhealthy_pods -eq 0 ]]; then
        log_success "All pods are running"
    else
        log_warning "$unhealthy_pods pods are not running"
        kubectl get pods -n "$NAMESPACE" --field-selector=status.phase!=Running
    fi
    
    # Test service endpoints
    local services=("auth-service")
    for service in "${services[@]}"; do
        local service_ip=$(kubectl get service "$service" -n "$NAMESPACE" -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "")
        if [[ -n "$service_ip" ]]; then
            if kubectl run test-pod --rm -i --restart=Never --image=curlimages/curl -- curl -f "http://$service_ip:8080/health" &>/dev/null; then
                log_success "$service health check passed"
            else
                log_warning "$service health check failed"
            fi
        fi
    done
}

# Cleanup
cleanup() {
    log_info "Cleaning up temporary files..."
    
    if [[ -n "${RESTORE_DIR:-}" ]] && [[ -d "$RESTORE_DIR" ]]; then
        rm -rf "$RESTORE_DIR"
    fi
    
    if [[ "$FROM_S3" == "true" ]] && [[ -f "$BACKUP_PATH" ]]; then
        rm -f "$BACKUP_PATH" "$BACKUP_PATH.sha256" "$BACKUP_PATH.md5"
    fi
    
    log_success "Cleanup completed"
}

# Send notification
send_notification() {
    local status=$1
    local message=$2
    
    # Slack notification
    if [[ -n "${SLACK_WEBHOOK_URL:-}" ]]; then
        local color="good"
        if [[ "$status" != "success" ]]; then
            color="danger"
        fi
        
        curl -X POST -H 'Content-type: application/json' \
            --data "{\"attachments\":[{\"color\":\"${color}\",\"title\":\"Go Coffee Restoration ${status}\",\"text\":\"${message}\",\"fields\":[{\"title\":\"Environment\",\"value\":\"${ENVIRONMENT}\",\"short\":true},{\"title\":\"Timestamp\",\"value\":\"$(date)\",\"short\":true}]}]}" \
            "${SLACK_WEBHOOK_URL}"
    fi
}

# Main restoration function
main() {
    local start_time=$(date +%s)
    
    log_info "Starting Go Coffee restoration for environment: $ENVIRONMENT"
    
    # Check prerequisites
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed"
        exit 1
    fi
    
    if ! kubectl cluster-info &> /dev/null; then
        log_error "kubectl is not configured or cluster is not accessible"
        exit 1
    fi
    
    # Download from S3 if needed
    if [[ "$FROM_S3" == "true" ]]; then
        download_from_s3
    fi
    
    # Verify and extract backup
    verify_backup
    extract_backup
    show_backup_info
    
    # Confirm restoration
    confirm_restoration
    
    # Perform restoration
    scale_down_services
    restore_database
    restore_redis
    restore_k8s_configs
    restore_deployments
    
    # Verify and cleanup
    verify_restoration
    cleanup
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    local success_message="Restoration completed successfully in ${duration} seconds"
    log_success "$success_message"
    
    send_notification "success" "$success_message"
}

# Error handling
trap 'log_error "Restoration failed at line $LINENO"; cleanup; send_notification "failed" "Restoration failed at line $LINENO"; exit 1' ERR

# Parse arguments and run
parse_args "$@"
main
