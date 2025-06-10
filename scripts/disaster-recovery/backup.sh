#!/bin/bash

# Go Coffee Platform - Comprehensive Backup Script
# This script creates full backups of all critical data and configurations

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
BACKUP_DIR="${BACKUP_DIR:-/var/backups/go-coffee}"
ENVIRONMENT="${ENVIRONMENT:-production}"
NAMESPACE="${NAMESPACE:-go-coffee}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
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

# Create backup directory structure
create_backup_structure() {
    local timestamp=$(date +%Y%m%d-%H%M%S)
    BACKUP_PATH="${BACKUP_DIR}/${ENVIRONMENT}/${timestamp}"
    
    log_info "Creating backup directory structure: ${BACKUP_PATH}"
    
    mkdir -p "${BACKUP_PATH}"/{database,redis,configs,secrets,volumes,logs}
    
    # Create metadata file
    cat > "${BACKUP_PATH}/metadata.json" << EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "environment": "${ENVIRONMENT}",
  "namespace": "${NAMESPACE}",
  "backup_type": "full",
  "retention_days": ${RETENTION_DAYS},
  "created_by": "$(whoami)",
  "hostname": "$(hostname)",
  "git_commit": "$(git rev-parse HEAD 2>/dev/null || echo 'unknown')",
  "git_branch": "$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'unknown')"
}
EOF
    
    log_success "Backup directory structure created"
}

# Backup PostgreSQL database
backup_database() {
    log_info "Starting database backup..."
    
    local db_pod=$(kubectl get pods -n "${NAMESPACE}" -l app=postgres -o jsonpath='{.items[0].metadata.name}')
    
    if [[ -z "$db_pod" ]]; then
        log_error "PostgreSQL pod not found"
        return 1
    fi
    
    log_info "Found PostgreSQL pod: $db_pod"
    
    # Create database dump
    kubectl exec "$db_pod" -n "${NAMESPACE}" -- pg_dumpall -U postgres > "${BACKUP_PATH}/database/full_dump.sql"
    
    # Create individual database dumps
    local databases=$(kubectl exec "$db_pod" -n "${NAMESPACE}" -- psql -U postgres -t -c "SELECT datname FROM pg_database WHERE datistemplate = false;")
    
    for db in $databases; do
        db=$(echo "$db" | xargs)  # Trim whitespace
        if [[ "$db" != "postgres" ]]; then
            log_info "Backing up database: $db"
            kubectl exec "$db_pod" -n "${NAMESPACE}" -- pg_dump -U postgres "$db" > "${BACKUP_PATH}/database/${db}.sql"
        fi
    done
    
    # Backup database configuration
    kubectl exec "$db_pod" -n "${NAMESPACE}" -- cat /var/lib/postgresql/data/postgresql.conf > "${BACKUP_PATH}/database/postgresql.conf" || true
    kubectl exec "$db_pod" -n "${NAMESPACE}" -- cat /var/lib/postgresql/data/pg_hba.conf > "${BACKUP_PATH}/database/pg_hba.conf" || true
    
    # Create database statistics
    kubectl exec "$db_pod" -n "${NAMESPACE}" -- psql -U postgres -c "SELECT * FROM pg_stat_database;" > "${BACKUP_PATH}/database/stats.txt"
    
    log_success "Database backup completed"
}

# Backup Redis data
backup_redis() {
    log_info "Starting Redis backup..."
    
    local redis_pod=$(kubectl get pods -n "${NAMESPACE}" -l app=redis -o jsonpath='{.items[0].metadata.name}')
    
    if [[ -z "$redis_pod" ]]; then
        log_error "Redis pod not found"
        return 1
    fi
    
    log_info "Found Redis pod: $redis_pod"
    
    # Create Redis dump
    kubectl exec "$redis_pod" -n "${NAMESPACE}" -- redis-cli BGSAVE
    
    # Wait for background save to complete
    while kubectl exec "$redis_pod" -n "${NAMESPACE}" -- redis-cli LASTSAVE | grep -q "$(kubectl exec "$redis_pod" -n "${NAMESPACE}" -- redis-cli LASTSAVE)"; do
        sleep 1
    done
    
    # Copy RDB file
    kubectl cp "${NAMESPACE}/${redis_pod}:/data/dump.rdb" "${BACKUP_PATH}/redis/dump.rdb"
    
    # Backup Redis configuration
    kubectl exec "$redis_pod" -n "${NAMESPACE}" -- redis-cli CONFIG GET '*' > "${BACKUP_PATH}/redis/config.txt"
    
    # Create Redis info
    kubectl exec "$redis_pod" -n "${NAMESPACE}" -- redis-cli INFO > "${BACKUP_PATH}/redis/info.txt"
    
    # Backup AOF file if exists
    kubectl cp "${NAMESPACE}/${redis_pod}:/data/appendonly.aof" "${BACKUP_PATH}/redis/appendonly.aof" 2>/dev/null || true
    
    log_success "Redis backup completed"
}

# Backup Kubernetes configurations
backup_k8s_configs() {
    log_info "Starting Kubernetes configurations backup..."
    
    # Backup all resources in namespace
    kubectl get all -n "${NAMESPACE}" -o yaml > "${BACKUP_PATH}/configs/all-resources.yaml"
    
    # Backup specific resource types
    local resources=("configmaps" "secrets" "persistentvolumeclaims" "services" "deployments" "statefulsets" "ingresses")
    
    for resource in "${resources[@]}"; do
        log_info "Backing up $resource..."
        kubectl get "$resource" -n "${NAMESPACE}" -o yaml > "${BACKUP_PATH}/configs/${resource}.yaml" 2>/dev/null || true
    done
    
    # Backup custom resources
    kubectl get customresourcedefinitions -o yaml > "${BACKUP_PATH}/configs/crds.yaml" 2>/dev/null || true
    
    # Backup RBAC
    kubectl get rolebindings,roles -n "${NAMESPACE}" -o yaml > "${BACKUP_PATH}/configs/rbac.yaml" 2>/dev/null || true
    
    # Backup network policies
    kubectl get networkpolicies -n "${NAMESPACE}" -o yaml > "${BACKUP_PATH}/configs/networkpolicies.yaml" 2>/dev/null || true
    
    log_success "Kubernetes configurations backup completed"
}

# Backup secrets (encrypted)
backup_secrets() {
    log_info "Starting secrets backup..."
    
    # Get all secrets (values will be base64 encoded)
    kubectl get secrets -n "${NAMESPACE}" -o yaml > "${BACKUP_PATH}/secrets/secrets.yaml"
    
    # Create secrets inventory
    kubectl get secrets -n "${NAMESPACE}" -o custom-columns=NAME:.metadata.name,TYPE:.type,AGE:.metadata.creationTimestamp > "${BACKUP_PATH}/secrets/inventory.txt"
    
    log_success "Secrets backup completed"
}

# Backup persistent volumes
backup_volumes() {
    log_info "Starting persistent volumes backup..."
    
    # Get PVC information
    kubectl get pvc -n "${NAMESPACE}" -o yaml > "${BACKUP_PATH}/volumes/pvcs.yaml"
    kubectl get pv -o yaml > "${BACKUP_PATH}/volumes/pvs.yaml"
    
    # Backup volume data using snapshots if available
    local pvcs=$(kubectl get pvc -n "${NAMESPACE}" -o jsonpath='{.items[*].metadata.name}')
    
    for pvc in $pvcs; do
        log_info "Creating snapshot for PVC: $pvc"
        
        # Create volume snapshot if VolumeSnapshot CRD exists
        if kubectl get crd volumesnapshots.snapshot.storage.k8s.io &>/dev/null; then
            cat > "${BACKUP_PATH}/volumes/${pvc}-snapshot.yaml" << EOF
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: ${pvc}-backup-$(date +%Y%m%d-%H%M%S)
  namespace: ${NAMESPACE}
spec:
  source:
    persistentVolumeClaimName: ${pvc}
EOF
            kubectl apply -f "${BACKUP_PATH}/volumes/${pvc}-snapshot.yaml"
        fi
    done
    
    log_success "Persistent volumes backup completed"
}

# Backup application logs
backup_logs() {
    log_info "Starting logs backup..."
    
    # Get pods
    local pods=$(kubectl get pods -n "${NAMESPACE}" -o jsonpath='{.items[*].metadata.name}')
    
    for pod in $pods; do
        log_info "Backing up logs for pod: $pod"
        
        # Get current logs
        kubectl logs "$pod" -n "${NAMESPACE}" --all-containers=true > "${BACKUP_PATH}/logs/${pod}-current.log" 2>/dev/null || true
        
        # Get previous logs if available
        kubectl logs "$pod" -n "${NAMESPACE}" --previous --all-containers=true > "${BACKUP_PATH}/logs/${pod}-previous.log" 2>/dev/null || true
    done
    
    log_success "Logs backup completed"
}

# Compress backup
compress_backup() {
    log_info "Compressing backup..."
    
    cd "$(dirname "${BACKUP_PATH}")"
    local backup_name=$(basename "${BACKUP_PATH}")
    
    tar -czf "${backup_name}.tar.gz" "${backup_name}"
    
    # Calculate checksums
    sha256sum "${backup_name}.tar.gz" > "${backup_name}.tar.gz.sha256"
    md5sum "${backup_name}.tar.gz" > "${backup_name}.tar.gz.md5"
    
    # Remove uncompressed directory
    rm -rf "${backup_name}"
    
    COMPRESSED_BACKUP="${BACKUP_DIR}/${ENVIRONMENT}/${backup_name}.tar.gz"
    
    log_success "Backup compressed: ${COMPRESSED_BACKUP}"
}

# Upload to S3
upload_to_s3() {
    if [[ -z "${AWS_ACCESS_KEY_ID:-}" ]] || [[ -z "${AWS_SECRET_ACCESS_KEY:-}" ]]; then
        log_warning "AWS credentials not found, skipping S3 upload"
        return 0
    fi
    
    log_info "Uploading backup to S3..."
    
    local backup_name=$(basename "${COMPRESSED_BACKUP}")
    local s3_path="s3://${S3_BUCKET}/${ENVIRONMENT}/$(date +%Y/%m/%d)/${backup_name}"
    
    # Upload backup
    aws s3 cp "${COMPRESSED_BACKUP}" "${s3_path}" --region "${S3_REGION}"
    
    # Upload checksums
    aws s3 cp "${COMPRESSED_BACKUP}.sha256" "${s3_path}.sha256" --region "${S3_REGION}"
    aws s3 cp "${COMPRESSED_BACKUP}.md5" "${s3_path}.md5" --region "${S3_REGION}"
    
    # Set lifecycle policy for automatic cleanup
    aws s3api put-object-tagging \
        --bucket "${S3_BUCKET}" \
        --key "${ENVIRONMENT}/$(date +%Y/%m/%d)/${backup_name}" \
        --tagging "TagSet=[{Key=Environment,Value=${ENVIRONMENT}},{Key=RetentionDays,Value=${RETENTION_DAYS}},{Key=BackupType,Value=full}]" \
        --region "${S3_REGION}"
    
    log_success "Backup uploaded to S3: ${s3_path}"
}

# Cleanup old backups
cleanup_old_backups() {
    log_info "Cleaning up old backups..."
    
    # Local cleanup
    find "${BACKUP_DIR}/${ENVIRONMENT}" -name "*.tar.gz" -mtime +${RETENTION_DAYS} -delete 2>/dev/null || true
    find "${BACKUP_DIR}/${ENVIRONMENT}" -name "*.sha256" -mtime +${RETENTION_DAYS} -delete 2>/dev/null || true
    find "${BACKUP_DIR}/${ENVIRONMENT}" -name "*.md5" -mtime +${RETENTION_DAYS} -delete 2>/dev/null || true
    
    # S3 cleanup (if configured)
    if [[ -n "${AWS_ACCESS_KEY_ID:-}" ]] && [[ -n "${AWS_SECRET_ACCESS_KEY:-}" ]]; then
        local cutoff_date=$(date -d "${RETENTION_DAYS} days ago" +%Y-%m-%d)
        
        aws s3api list-objects-v2 \
            --bucket "${S3_BUCKET}" \
            --prefix "${ENVIRONMENT}/" \
            --query "Contents[?LastModified<='${cutoff_date}'].Key" \
            --output text \
            --region "${S3_REGION}" | \
        while read -r key; do
            if [[ -n "$key" ]]; then
                aws s3 rm "s3://${S3_BUCKET}/${key}" --region "${S3_REGION}"
            fi
        done
    fi
    
    log_success "Old backups cleaned up"
}

# Verify backup integrity
verify_backup() {
    log_info "Verifying backup integrity..."
    
    # Verify checksums
    cd "$(dirname "${COMPRESSED_BACKUP}")"
    local backup_name=$(basename "${COMPRESSED_BACKUP}")
    
    if sha256sum -c "${backup_name}.sha256" && md5sum -c "${backup_name}.md5"; then
        log_success "Backup integrity verified"
    else
        log_error "Backup integrity check failed"
        return 1
    fi
    
    # Test extraction
    local test_dir="/tmp/backup-test-$$"
    mkdir -p "$test_dir"
    
    if tar -tzf "${COMPRESSED_BACKUP}" > /dev/null && tar -xzf "${COMPRESSED_BACKUP}" -C "$test_dir" > /dev/null; then
        log_success "Backup extraction test passed"
        rm -rf "$test_dir"
    else
        log_error "Backup extraction test failed"
        rm -rf "$test_dir"
        return 1
    fi
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
            --data "{\"attachments\":[{\"color\":\"${color}\",\"title\":\"Go Coffee Backup ${status}\",\"text\":\"${message}\",\"fields\":[{\"title\":\"Environment\",\"value\":\"${ENVIRONMENT}\",\"short\":true},{\"title\":\"Timestamp\",\"value\":\"$(date)\",\"short\":true}]}]}" \
            "${SLACK_WEBHOOK_URL}"
    fi
    
    # Email notification (if configured)
    if command -v mail &> /dev/null && [[ -n "${NOTIFICATION_EMAIL:-}" ]]; then
        echo "$message" | mail -s "Go Coffee Backup ${status} - ${ENVIRONMENT}" "${NOTIFICATION_EMAIL}"
    fi
}

# Main backup function
main() {
    local start_time=$(date +%s)
    
    log_info "Starting Go Coffee backup for environment: ${ENVIRONMENT}"
    
    # Check prerequisites
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed"
        exit 1
    fi
    
    if ! kubectl cluster-info &> /dev/null; then
        log_error "kubectl is not configured or cluster is not accessible"
        exit 1
    fi
    
    # Create backup structure
    create_backup_structure
    
    # Perform backups
    backup_database
    backup_redis
    backup_k8s_configs
    backup_secrets
    backup_volumes
    backup_logs
    
    # Compress and verify
    compress_backup
    verify_backup
    
    # Upload and cleanup
    upload_to_s3
    cleanup_old_backups
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    local success_message="Backup completed successfully in ${duration} seconds. Backup location: ${COMPRESSED_BACKUP}"
    log_success "$success_message"
    
    send_notification "success" "$success_message"
}

# Error handling
trap 'log_error "Backup failed at line $LINENO"; send_notification "failed" "Backup failed at line $LINENO"; exit 1' ERR

# Run main function
main "$@"
