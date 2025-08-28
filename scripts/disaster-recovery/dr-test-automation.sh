#!/bin/bash

# Go Coffee Platform - Disaster Recovery Testing Automation
# Comprehensive DR testing suite with automated validation

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
ENVIRONMENT="${ENVIRONMENT:-staging}"
NAMESPACE="${NAMESPACE:-go-coffee}"
DR_NAMESPACE="${DR_NAMESPACE:-disaster-recovery}"
TEST_NAMESPACE="${TEST_NAMESPACE:-dr-test}"

# Test configuration
TEST_TYPE="${TEST_TYPE:-backup_restore}"
DRY_RUN="${DRY_RUN:-false}"
CLEANUP_AFTER_TEST="${CLEANUP_AFTER_TEST:-true}"
NOTIFICATION_ENABLED="${NOTIFICATION_ENABLED:-true}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
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

log_test() {
    echo -e "${PURPLE}[TEST]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# Test results tracking
TESTS_PASSED=0
TESTS_FAILED=0
FAILED_TESTS=()

# Record test result
record_test_result() {
    local test_name="$1"
    local result="$2"
    
    if [[ "$result" == "PASS" ]]; then
        ((TESTS_PASSED++))
        log_success "✅ $test_name: PASSED"
    else
        ((TESTS_FAILED++))
        FAILED_TESTS+=("$test_name")
        log_error "❌ $test_name: FAILED"
    fi
}

# Send notification
send_notification() {
    local status="$1"
    local message="$2"
    
    if [[ "$NOTIFICATION_ENABLED" != "true" ]]; then
        return 0
    fi
    
    # Slack notification
    if [[ -n "${SLACK_WEBHOOK_URL:-}" ]]; then
        local color="good"
        local emoji="✅"
        
        if [[ "$status" != "success" ]]; then
            color="danger"
            emoji="❌"
        fi
        
        curl -X POST -H 'Content-type: application/json' \
            --data "{\"attachments\":[{\"color\":\"${color}\",\"title\":\"${emoji} Go Coffee DR Test ${status}\",\"text\":\"${message}\",\"fields\":[{\"title\":\"Environment\",\"value\":\"${ENVIRONMENT}\",\"short\":true},{\"title\":\"Test Type\",\"value\":\"${TEST_TYPE}\",\"short\":true},{\"title\":\"Timestamp\",\"value\":\"$(date)\",\"short\":true}]}]}" \
            "${SLACK_WEBHOOK_URL}" || true
    fi
}

# Setup test environment
setup_test_environment() {
    log_info "Setting up DR test environment..."
    
    # Create test namespace
    if ! kubectl get namespace "$TEST_NAMESPACE" &>/dev/null; then
        kubectl create namespace "$TEST_NAMESPACE"
        kubectl label namespace "$TEST_NAMESPACE" \
            purpose=dr-testing \
            environment="$ENVIRONMENT" \
            managed-by=automation
    fi
    
    log_success "Test environment ready"
}

# Cleanup test environment
cleanup_test_environment() {
    if [[ "$CLEANUP_AFTER_TEST" == "true" ]]; then
        log_info "Cleaning up test environment..."
        kubectl delete namespace "$TEST_NAMESPACE" --ignore-not-found=true
        log_success "Test environment cleaned up"
    fi
}

# Test 1: Backup Integrity
test_backup_integrity() {
    log_test "Testing backup integrity..."
    
    # Check if Velero is available
    if ! command -v velero &> /dev/null; then
        record_test_result "Backup Integrity" "FAIL"
        log_error "Velero CLI not found"
        return 1
    fi
    
    # Get latest backup
    local latest_backup
    latest_backup=$(velero backup get -o name 2>/dev/null | head -1 | cut -d'/' -f2)
    
    if [[ -z "$latest_backup" ]]; then
        record_test_result "Backup Integrity" "FAIL"
        log_error "No backups found"
        return 1
    fi
    
    log_info "Testing backup: $latest_backup"
    
    # Check backup status
    local backup_status
    backup_status=$(velero backup describe "$latest_backup" -o json | jq -r '.status.phase')
    
    if [[ "$backup_status" == "Completed" ]]; then
        record_test_result "Backup Integrity" "PASS"
        return 0
    else
        record_test_result "Backup Integrity" "FAIL"
        log_error "Backup status: $backup_status"
        return 1
    fi
}

# Test 2: Backup Restore
test_backup_restore() {
    log_test "Testing backup restore functionality..."
    
    # Get latest backup
    local latest_backup
    latest_backup=$(velero backup get -o name 2>/dev/null | head -1 | cut -d'/' -f2)
    
    if [[ -z "$latest_backup" ]]; then
        record_test_result "Backup Restore" "FAIL"
        log_error "No backups available for restore test"
        return 1
    fi
    
    local restore_name="dr-test-restore-$(date +%s)"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "DRY RUN: Would restore backup $latest_backup as $restore_name"
        record_test_result "Backup Restore" "PASS"
        return 0
    fi
    
    # Create restore with namespace mapping
    velero restore create "$restore_name" \
        --from-backup "$latest_backup" \
        --namespace-mappings "$NAMESPACE:$TEST_NAMESPACE" \
        --wait
    
    # Check restore status
    local restore_status
    restore_status=$(velero restore describe "$restore_name" -o json | jq -r '.status.phase')
    
    if [[ "$restore_status" == "Completed" ]]; then
        # Verify restored resources
        local restored_pods
        restored_pods=$(kubectl get pods -n "$TEST_NAMESPACE" --no-headers | wc -l)
        
        if [[ "$restored_pods" -gt 0 ]]; then
            record_test_result "Backup Restore" "PASS"
            log_info "Restored $restored_pods pods to test namespace"
        else
            record_test_result "Backup Restore" "FAIL"
            log_error "No pods found in restored namespace"
        fi
    else
        record_test_result "Backup Restore" "FAIL"
        log_error "Restore failed with status: $restore_status"
    fi
}

# Test 3: Database Connectivity
test_database_connectivity() {
    log_test "Testing database connectivity and replication..."
    
    # Find database pod
    local db_pod
    db_pod=$(kubectl get pods -n "$NAMESPACE" -l app=postgres -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    
    if [[ -z "$db_pod" ]]; then
        record_test_result "Database Connectivity" "FAIL"
        log_error "Database pod not found"
        return 1
    fi
    
    # Test database connection
    if kubectl exec "$db_pod" -n "$NAMESPACE" -- pg_isready -U postgres &>/dev/null; then
        log_info "Database connection successful"
        
        # Test replication status
        local replication_count
        replication_count=$(kubectl exec "$db_pod" -n "$NAMESPACE" -- \
            psql -U postgres -t -c "SELECT COUNT(*) FROM pg_stat_replication;" 2>/dev/null | xargs)
        
        if [[ "$replication_count" -gt 0 ]]; then
            record_test_result "Database Connectivity" "PASS"
            log_info "Database replication active with $replication_count replicas"
        else
            record_test_result "Database Connectivity" "FAIL"
            log_warning "No database replication detected"
        fi
    else
        record_test_result "Database Connectivity" "FAIL"
        log_error "Database connection failed"
    fi
}

# Test 4: Service Health Checks
test_service_health() {
    log_test "Testing service health endpoints..."
    
    local services=("coffee-service" "payment-service" "user-service" "order-service")
    local healthy_services=0
    
    for service in "${services[@]}"; do
        # Check if service exists
        if kubectl get service "$service" -n "$NAMESPACE" &>/dev/null; then
            # Port forward and test health endpoint
            local port
            port=$(kubectl get service "$service" -n "$NAMESPACE" -o jsonpath='{.spec.ports[0].port}')
            
            # Test health endpoint (assuming /health endpoint exists)
            if kubectl exec -n "$NAMESPACE" deployment/"$service" -- \
                curl -f "http://localhost:$port/health" &>/dev/null; then
                log_info "$service health check passed"
                ((healthy_services++))
            else
                log_warning "$service health check failed"
            fi
        else
            log_warning "$service not found"
        fi
    done
    
    if [[ "$healthy_services" -ge 3 ]]; then
        record_test_result "Service Health" "PASS"
        log_info "$healthy_services out of ${#services[@]} services are healthy"
    else
        record_test_result "Service Health" "FAIL"
        log_error "Only $healthy_services out of ${#services[@]} services are healthy"
    fi
}

# Test 5: Failover Controller
test_failover_controller() {
    log_test "Testing failover controller functionality..."
    
    # Check if failover controller is running
    local controller_pods
    controller_pods=$(kubectl get pods -n "$DR_NAMESPACE" -l app=failover-controller --field-selector=status.phase=Running --no-headers | wc -l)
    
    if [[ "$controller_pods" -gt 0 ]]; then
        log_info "Failover controller is running ($controller_pods pods)"
        
        # Check controller logs for recent activity
        local recent_logs
        recent_logs=$(kubectl logs -n "$DR_NAMESPACE" -l app=failover-controller --since=1h --tail=10 2>/dev/null | wc -l)
        
        if [[ "$recent_logs" -gt 0 ]]; then
            record_test_result "Failover Controller" "PASS"
            log_info "Failover controller is active with recent logs"
        else
            record_test_result "Failover Controller" "FAIL"
            log_warning "Failover controller has no recent activity"
        fi
    else
        record_test_result "Failover Controller" "FAIL"
        log_error "Failover controller is not running"
    fi
}

# Test 6: Cross-Region Connectivity
test_cross_region_connectivity() {
    log_test "Testing cross-region connectivity..."
    
    # This is a simplified test - in reality, you'd test actual cross-region endpoints
    local primary_region="${PRIMARY_REGION:-us-east-1}"
    local secondary_region="${SECONDARY_REGION:-us-west-2}"
    
    log_info "Testing connectivity between $primary_region and $secondary_region"
    
    # Test DNS resolution for DR endpoints
    if nslookup "api-dr.go-coffee.com" &>/dev/null; then
        log_info "DR endpoint DNS resolution successful"
        
        # Test HTTP connectivity (if available)
        if curl -f --connect-timeout 10 "https://api-dr.go-coffee.com/health" &>/dev/null; then
            record_test_result "Cross-Region Connectivity" "PASS"
            log_info "DR endpoint is accessible"
        else
            record_test_result "Cross-Region Connectivity" "FAIL"
            log_warning "DR endpoint is not accessible"
        fi
    else
        record_test_result "Cross-Region Connectivity" "FAIL"
        log_error "DR endpoint DNS resolution failed"
    fi
}

# Test 7: Monitoring and Alerting
test_monitoring_alerting() {
    log_test "Testing monitoring and alerting systems..."
    
    # Check if Prometheus is accessible
    if kubectl get pods -n monitoring -l app.kubernetes.io/name=prometheus &>/dev/null; then
        log_info "Prometheus is running"
        
        # Check if DR-specific alerts are configured
        local dr_alerts
        dr_alerts=$(kubectl get prometheusrules -n "$DR_NAMESPACE" --no-headers 2>/dev/null | wc -l)
        
        if [[ "$dr_alerts" -gt 0 ]]; then
            record_test_result "Monitoring Alerting" "PASS"
            log_info "DR monitoring rules are configured ($dr_alerts rules)"
        else
            record_test_result "Monitoring Alerting" "FAIL"
            log_warning "No DR monitoring rules found"
        fi
    else
        record_test_result "Monitoring Alerting" "FAIL"
        log_error "Prometheus is not running"
    fi
}

# Generate test report
generate_test_report() {
    local total_tests=$((TESTS_PASSED + TESTS_FAILED))
    local success_rate=0
    
    if [[ "$total_tests" -gt 0 ]]; then
        success_rate=$(( (TESTS_PASSED * 100) / total_tests ))
    fi
    
    log_info "Generating DR test report..."
    
    cat > "/tmp/dr-test-report-$(date +%Y%m%d-%H%M%S).md" << EOF
# Go Coffee Disaster Recovery Test Report

**Test Date**: $(date)
**Environment**: $ENVIRONMENT
**Test Type**: $TEST_TYPE
**Dry Run**: $DRY_RUN

## Summary

- **Total Tests**: $total_tests
- **Passed**: $TESTS_PASSED
- **Failed**: $TESTS_FAILED
- **Success Rate**: $success_rate%

## Test Results

### Passed Tests
$(for i in $(seq 1 $TESTS_PASSED); do echo "- ✅ Test $i"; done)

### Failed Tests
$(for test in "${FAILED_TESTS[@]}"; do echo "- ❌ $test"; done)

## Recommendations

$(if [[ $TESTS_FAILED -gt 0 ]]; then
    echo "### Action Items"
    echo "- Investigate and fix failed tests"
    echo "- Review DR procedures and documentation"
    echo "- Consider additional monitoring or automation"
else
    echo "### Status"
    echo "All DR tests passed successfully. System is ready for disaster scenarios."
fi)

## Next Steps

- Schedule next DR test for $(date -d "+1 month" +%Y-%m-%d)
- Review and update DR runbooks if needed
- Conduct team training on any identified gaps

---
*Report generated by DR Test Automation*
EOF
    
    log_success "Test report generated: /tmp/dr-test-report-$(date +%Y%m%d-%H%M%S).md"
}

# Main test execution
run_dr_tests() {
    local start_time=$(date +%s)
    
    log_info "Starting Go Coffee DR tests..."
    log_info "Environment: $ENVIRONMENT"
    log_info "Test Type: $TEST_TYPE"
    log_info "Dry Run: $DRY_RUN"
    
    # Setup
    setup_test_environment
    
    # Run tests based on type
    case "$TEST_TYPE" in
        "backup_restore")
            test_backup_integrity
            test_backup_restore
            ;;
        "failover")
            test_failover_controller
            test_cross_region_connectivity
            ;;
        "full_dr")
            test_backup_integrity
            test_backup_restore
            test_database_connectivity
            test_service_health
            test_failover_controller
            test_cross_region_connectivity
            test_monitoring_alerting
            ;;
        "health_check")
            test_database_connectivity
            test_service_health
            test_monitoring_alerting
            ;;
        *)
            log_error "Unknown test type: $TEST_TYPE"
            exit 1
            ;;
    esac
    
    # Cleanup
    cleanup_test_environment
    
    # Generate report
    generate_test_report
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    local total_tests=$((TESTS_PASSED + TESTS_FAILED))
    
    # Final summary
    if [[ $TESTS_FAILED -eq 0 ]]; then
        local message="All $total_tests DR tests passed successfully in ${duration}s"
        log_success "$message"
        send_notification "success" "$message"
    else
        local message="$TESTS_FAILED out of $total_tests DR tests failed in ${duration}s"
        log_error "$message"
        send_notification "failed" "$message"
        exit 1
    fi
}

# Show usage
show_usage() {
    cat << EOF
Go Coffee Platform - DR Test Automation

Usage: $0 [OPTIONS]

Options:
    -t, --type TYPE         Test type (backup_restore|failover|full_dr|health_check) [default: backup_restore]
    -e, --env ENV          Environment (dev|staging|prod) [default: staging]
    -n, --namespace NS     Kubernetes namespace [default: go-coffee]
    -d, --dry-run          Perform dry run without making changes
    --no-cleanup           Don't cleanup test resources after completion
    --no-notifications     Disable notifications
    -h, --help             Show this help message

Examples:
    $0                                    # Run basic backup/restore test
    $0 -t full_dr -e prod                # Run full DR test in production
    $0 -t failover -d                    # Dry run failover test
    $0 -t health_check --no-notifications # Health check without notifications

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--type)
            TEST_TYPE="$2"
            shift 2
            ;;
        -e|--env)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -n|--namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        -d|--dry-run)
            DRY_RUN="true"
            shift
            ;;
        --no-cleanup)
            CLEANUP_AFTER_TEST="false"
            shift
            ;;
        --no-notifications)
            NOTIFICATION_ENABLED="false"
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Validate test type
case "$TEST_TYPE" in
    "backup_restore"|"failover"|"full_dr"|"health_check")
        ;;
    *)
        log_error "Invalid test type: $TEST_TYPE"
        show_usage
        exit 1
        ;;
esac

# Run the tests
run_dr_tests
