#!/bin/bash

# Go Coffee Platform - Comprehensive Health Check Script
# This script performs comprehensive health checks across all platform components

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
ENVIRONMENT="${ENVIRONMENT:-production}"
NAMESPACE="${NAMESPACE:-go-coffee}"
TIMEOUT="${TIMEOUT:-300}"
COMPREHENSIVE="${COMPREHENSIVE:-false}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Counters
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNING_CHECKS=0

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
    ((PASSED_CHECKS++))
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
    ((WARNING_CHECKS++))
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
    ((FAILED_CHECKS++))
}

# Increment total checks counter
check() {
    ((TOTAL_CHECKS++))
}

# Show help
show_help() {
    cat << EOF
Go Coffee Platform - Health Check Script

Usage: $0 [OPTIONS]

Options:
    -e, --environment   Environment to check (development|staging|production) [default: production]
    -n, --namespace     Kubernetes namespace [default: go-coffee]
    -t, --timeout       Timeout in seconds [default: 300]
    -c, --comprehensive Run comprehensive health checks
    -h, --help          Show this help message

Examples:
    $0                                    # Basic health check
    $0 -e staging -c                     # Comprehensive check for staging
    $0 -n go-coffee-dev -t 600          # Custom namespace with longer timeout

EOF
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -n|--namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            -t|--timeout)
                TIMEOUT="$2"
                shift 2
                ;;
            -c|--comprehensive)
                COMPREHENSIVE=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# Check Kubernetes cluster connectivity
check_k8s_connectivity() {
    log_info "Checking Kubernetes cluster connectivity..."
    check
    
    if kubectl cluster-info &>/dev/null; then
        log_success "Kubernetes cluster is accessible"
    else
        log_error "Cannot connect to Kubernetes cluster"
        return 1
    fi
    
    # Check namespace exists
    check
    if kubectl get namespace "$NAMESPACE" &>/dev/null; then
        log_success "Namespace '$NAMESPACE' exists"
    else
        log_error "Namespace '$NAMESPACE' not found"
        return 1
    fi
}

# Check pod status
check_pod_status() {
    log_info "Checking pod status..."
    
    local pods=$(kubectl get pods -n "$NAMESPACE" -o json)
    local pod_count=$(echo "$pods" | jq '.items | length')
    
    if [[ $pod_count -eq 0 ]]; then
        check
        log_error "No pods found in namespace '$NAMESPACE'"
        return 1
    fi
    
    # Check each pod
    echo "$pods" | jq -r '.items[] | "\(.metadata.name) \(.status.phase)"' | while read -r pod_name pod_phase; do
        check
        case "$pod_phase" in
            "Running")
                log_success "Pod $pod_name is running"
                ;;
            "Pending")
                log_warning "Pod $pod_name is pending"
                ;;
            "Failed"|"CrashLoopBackOff")
                log_error "Pod $pod_name is in failed state: $pod_phase"
                ;;
            *)
                log_warning "Pod $pod_name is in unknown state: $pod_phase"
                ;;
        esac
    done
}

# Check service endpoints
check_service_endpoints() {
    log_info "Checking service endpoints..."
    
    local services=("auth-service" "postgres" "redis")
    
    for service in "${services[@]}"; do
        check
        if kubectl get service "$service" -n "$NAMESPACE" &>/dev/null; then
            local endpoints=$(kubectl get endpoints "$service" -n "$NAMESPACE" -o jsonpath='{.subsets[*].addresses[*].ip}' 2>/dev/null || echo "")
            
            if [[ -n "$endpoints" ]]; then
                log_success "Service $service has endpoints: $endpoints"
            else
                log_error "Service $service has no endpoints"
            fi
        else
            log_warning "Service $service not found"
        fi
    done
}

# Check application health endpoints
check_app_health() {
    log_info "Checking application health endpoints..."
    
    local services=("auth-service" "ai-search" "producer" "consumer" "streams")
    
    for service in "${services[@]}"; do
        check
        local service_ip=$(kubectl get service "$service" -n "$NAMESPACE" -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "")
        
        if [[ -n "$service_ip" ]]; then
            # Use a temporary pod to check health endpoint
            if kubectl run health-check-pod --rm -i --restart=Never --image=curlimages/curl --timeout="$TIMEOUT" -- \
                curl -f -s "http://$service_ip:8080/health" &>/dev/null; then
                log_success "Service $service health endpoint is responding"
            else
                log_error "Service $service health endpoint is not responding"
            fi
        else
            log_warning "Service $service not found or has no cluster IP"
        fi
    done
}

# Check database connectivity
check_database() {
    log_info "Checking database connectivity..."
    
    local db_pod=$(kubectl get pods -n "$NAMESPACE" -l app=postgres -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
    
    if [[ -z "$db_pod" ]]; then
        check
        log_error "PostgreSQL pod not found"
        return 1
    fi
    
    # Test database connection
    check
    if kubectl exec "$db_pod" -n "$NAMESPACE" -- pg_isready -U postgres &>/dev/null; then
        log_success "Database is accepting connections"
    else
        log_error "Database is not accepting connections"
        return 1
    fi
    
    # Check database size and connections
    if [[ "$COMPREHENSIVE" == "true" ]]; then
        check
        local db_size=$(kubectl exec "$db_pod" -n "$NAMESPACE" -- psql -U postgres -t -c "SELECT pg_size_pretty(pg_database_size('go_coffee'));" 2>/dev/null | xargs || echo "unknown")
        log_info "Database size: $db_size"
        
        check
        local active_connections=$(kubectl exec "$db_pod" -n "$NAMESPACE" -- psql -U postgres -t -c "SELECT count(*) FROM pg_stat_activity WHERE state = 'active';" 2>/dev/null | xargs || echo "unknown")
        log_info "Active database connections: $active_connections"
        
        if [[ "$active_connections" =~ ^[0-9]+$ ]] && [[ $active_connections -lt 100 ]]; then
            log_success "Database connection count is healthy: $active_connections"
        elif [[ "$active_connections" =~ ^[0-9]+$ ]]; then
            log_warning "High database connection count: $active_connections"
        fi
    fi
}

# Check Redis connectivity
check_redis() {
    log_info "Checking Redis connectivity..."
    
    local redis_pod=$(kubectl get pods -n "$NAMESPACE" -l app=redis -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
    
    if [[ -z "$redis_pod" ]]; then
        check
        log_error "Redis pod not found"
        return 1
    fi
    
    # Test Redis connection
    check
    if kubectl exec "$redis_pod" -n "$NAMESPACE" -- redis-cli ping | grep -q "PONG"; then
        log_success "Redis is responding to ping"
    else
        log_error "Redis is not responding to ping"
        return 1
    fi
    
    # Check Redis memory usage
    if [[ "$COMPREHENSIVE" == "true" ]]; then
        check
        local memory_usage=$(kubectl exec "$redis_pod" -n "$NAMESPACE" -- redis-cli info memory | grep "used_memory_human" | cut -d: -f2 | tr -d '\r' || echo "unknown")
        log_info "Redis memory usage: $memory_usage"
        
        check
        local connected_clients=$(kubectl exec "$redis_pod" -n "$NAMESPACE" -- redis-cli info clients | grep "connected_clients" | cut -d: -f2 | tr -d '\r' || echo "unknown")
        log_info "Redis connected clients: $connected_clients"
        
        if [[ "$connected_clients" =~ ^[0-9]+$ ]] && [[ $connected_clients -lt 1000 ]]; then
            log_success "Redis client count is healthy: $connected_clients"
        elif [[ "$connected_clients" =~ ^[0-9]+$ ]]; then
            log_warning "High Redis client count: $connected_clients"
        fi
    fi
}

# Check resource usage
check_resource_usage() {
    if [[ "$COMPREHENSIVE" != "true" ]]; then
        return 0
    fi
    
    log_info "Checking resource usage..."
    
    # Check node resource usage
    local nodes=$(kubectl get nodes -o json)
    echo "$nodes" | jq -r '.items[] | .metadata.name' | while read -r node; do
        check
        local cpu_usage=$(kubectl top node "$node" --no-headers 2>/dev/null | awk '{print $3}' | sed 's/%//' || echo "unknown")
        local memory_usage=$(kubectl top node "$node" --no-headers 2>/dev/null | awk '{print $5}' | sed 's/%//' || echo "unknown")
        
        if [[ "$cpu_usage" =~ ^[0-9]+$ ]]; then
            if [[ $cpu_usage -lt 80 ]]; then
                log_success "Node $node CPU usage: ${cpu_usage}%"
            else
                log_warning "Node $node high CPU usage: ${cpu_usage}%"
            fi
        fi
        
        if [[ "$memory_usage" =~ ^[0-9]+$ ]]; then
            if [[ $memory_usage -lt 80 ]]; then
                log_success "Node $node memory usage: ${memory_usage}%"
            else
                log_warning "Node $node high memory usage: ${memory_usage}%"
            fi
        fi
    done
    
    # Check pod resource usage
    kubectl top pods -n "$NAMESPACE" --no-headers 2>/dev/null | while read -r pod cpu memory; do
        check
        local cpu_value=$(echo "$cpu" | sed 's/m$//')
        local memory_value=$(echo "$memory" | sed 's/Mi$//')
        
        if [[ "$cpu_value" =~ ^[0-9]+$ ]] && [[ $cpu_value -gt 1000 ]]; then
            log_warning "Pod $pod high CPU usage: $cpu"
        else
            log_success "Pod $pod CPU usage: $cpu"
        fi
        
        if [[ "$memory_value" =~ ^[0-9]+$ ]] && [[ $memory_value -gt 1000 ]]; then
            log_warning "Pod $pod high memory usage: $memory"
        else
            log_success "Pod $pod memory usage: $memory"
        fi
    done
}

# Check persistent volumes
check_persistent_volumes() {
    if [[ "$COMPREHENSIVE" != "true" ]]; then
        return 0
    fi
    
    log_info "Checking persistent volumes..."
    
    local pvcs=$(kubectl get pvc -n "$NAMESPACE" -o json)
    local pvc_count=$(echo "$pvcs" | jq '.items | length')
    
    if [[ $pvc_count -eq 0 ]]; then
        check
        log_warning "No persistent volume claims found"
        return 0
    fi
    
    echo "$pvcs" | jq -r '.items[] | "\(.metadata.name) \(.status.phase)"' | while read -r pvc_name pvc_phase; do
        check
        case "$pvc_phase" in
            "Bound")
                log_success "PVC $pvc_name is bound"
                ;;
            "Pending")
                log_warning "PVC $pvc_name is pending"
                ;;
            *)
                log_error "PVC $pvc_name is in unexpected state: $pvc_phase"
                ;;
        esac
    done
}

# Check ingress and networking
check_networking() {
    if [[ "$COMPREHENSIVE" != "true" ]]; then
        return 0
    fi
    
    log_info "Checking networking..."
    
    # Check ingress
    local ingresses=$(kubectl get ingress -n "$NAMESPACE" -o json 2>/dev/null || echo '{"items":[]}')
    local ingress_count=$(echo "$ingresses" | jq '.items | length')
    
    if [[ $ingress_count -gt 0 ]]; then
        echo "$ingresses" | jq -r '.items[] | .metadata.name' | while read -r ingress_name; do
            check
            log_success "Ingress $ingress_name is configured"
        done
    else
        check
        log_info "No ingress resources found"
    fi
    
    # Check network policies
    local netpols=$(kubectl get networkpolicy -n "$NAMESPACE" -o json 2>/dev/null || echo '{"items":[]}')
    local netpol_count=$(echo "$netpols" | jq '.items | length')
    
    check
    if [[ $netpol_count -gt 0 ]]; then
        log_success "Network policies are configured ($netpol_count policies)"
    else
        log_info "No network policies found"
    fi
}

# Check secrets and configmaps
check_configs() {
    log_info "Checking configurations..."
    
    # Check secrets
    local secrets=$(kubectl get secrets -n "$NAMESPACE" -o json)
    local secret_count=$(echo "$secrets" | jq '.items | length')
    
    check
    if [[ $secret_count -gt 0 ]]; then
        log_success "Secrets are configured ($secret_count secrets)"
    else
        log_warning "No secrets found"
    fi
    
    # Check configmaps
    local configmaps=$(kubectl get configmaps -n "$NAMESPACE" -o json)
    local configmap_count=$(echo "$configmaps" | jq '.items | length')
    
    check
    if [[ $configmap_count -gt 0 ]]; then
        log_success "ConfigMaps are configured ($configmap_count configmaps)"
    else
        log_warning "No ConfigMaps found"
    fi
}

# Generate health report
generate_health_report() {
    log_info "Generating health report..."
    
    local overall_status="healthy"
    if [[ $FAILED_CHECKS -gt 0 ]]; then
        overall_status="unhealthy"
    elif [[ $WARNING_CHECKS -gt 5 ]]; then
        overall_status="degraded"
    elif [[ $WARNING_CHECKS -gt 0 ]]; then
        overall_status="warning"
    fi
    
    local report_file="/tmp/health-report-$(date +%Y%m%d-%H%M%S).json"
    
    cat > "$report_file" << EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "environment": "$ENVIRONMENT",
  "namespace": "$NAMESPACE",
  "overall_status": "$overall_status",
  "summary": {
    "total_checks": $TOTAL_CHECKS,
    "passed": $PASSED_CHECKS,
    "failed": $FAILED_CHECKS,
    "warnings": $WARNING_CHECKS
  },
  "success_rate": $(echo "scale=2; $PASSED_CHECKS * 100 / $TOTAL_CHECKS" | bc -l),
  "recommendations": [
    $(if [[ $FAILED_CHECKS -gt 0 ]]; then echo '"Investigate failed checks immediately",'; fi)
    $(if [[ $WARNING_CHECKS -gt 5 ]]; then echo '"Review warning conditions",'; fi)
    "Monitor resource usage trends",
    "Ensure backup procedures are working"
  ]
}
EOF
    
    echo "$report_file"
}

# Main health check function
main() {
    parse_args "$@"
    
    log_info "Starting health check for environment: $ENVIRONMENT"
    log_info "Namespace: $NAMESPACE"
    log_info "Comprehensive mode: $COMPREHENSIVE"
    
    # Run health checks
    check_k8s_connectivity
    check_pod_status
    check_service_endpoints
    check_app_health
    check_database
    check_redis
    check_resource_usage
    check_persistent_volumes
    check_networking
    check_configs
    
    # Generate report
    local report_file=$(generate_health_report)
    
    # Summary
    echo
    log_info "Health Check Summary:"
    log_info "Total Checks: $TOTAL_CHECKS"
    log_success "Passed: $PASSED_CHECKS"
    log_warning "Warnings: $WARNING_CHECKS"
    log_error "Failed: $FAILED_CHECKS"
    
    local success_rate=$(echo "scale=1; $PASSED_CHECKS * 100 / $TOTAL_CHECKS" | bc -l)
    log_info "Success Rate: ${success_rate}%"
    log_info "Detailed Report: $report_file"
    
    # Exit with appropriate code
    if [[ $FAILED_CHECKS -gt 0 ]]; then
        log_error "Health check failed - $FAILED_CHECKS critical issues found"
        exit 1
    elif [[ $WARNING_CHECKS -gt 5 ]]; then
        log_warning "Health check completed with warnings - $WARNING_CHECKS issues found"
        exit 2
    else
        log_success "Health check passed successfully"
        exit 0
    fi
}

# Run main function
main "$@"
