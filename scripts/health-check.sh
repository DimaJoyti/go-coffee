#!/bin/bash

# Go Coffee - Comprehensive Health Check Script
# Performs health checks across all microservices and infrastructure components
# Version: 2.0.0
# Usage: ./health-check.sh [OPTIONS]
#   -e, --environment   Environment to check (development|staging|production)
#   -c, --comprehensive Run comprehensive health checks including performance
#   -m, --monitoring    Enable continuous monitoring mode
#   -r, --report        Generate detailed health report
#   -h, --help          Show this help message

set -euo pipefail

# Get script directory for relative imports
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source shared library
source "$SCRIPT_DIR/lib/common.sh" 2>/dev/null || {
    echo "❌ Cannot load shared library. Please run from project root."
    exit 1
}

print_header "🏥 Go Coffee Health Check System"

# =============================================================================
# CONFIGURATION
# =============================================================================

ENVIRONMENT="${ENVIRONMENT:-development}"
COMPREHENSIVE=false
MONITORING_MODE=false
GENERATE_REPORT=false
HEALTH_TIMEOUT=10
CHECK_INTERVAL=30

# Health check counters
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNING_CHECKS=0

# Service health endpoints (service:port:health_path)
declare -A SERVICE_HEALTH_ENDPOINTS=(
    ["auth-service"]="8091:/health"
    ["payment-service"]="8093:/health"
    ["order-service"]="8094:/health"
    ["kitchen-service"]="8095:/health"
    ["user-gateway"]="8096:/health"
    ["security-gateway"]="8097:/health"
    ["communication-hub"]="8098:/health"
    ["ai-search"]="8099:/health"
    ["ai-service"]="8100:/health"
    ["ai-arbitrage-service"]="8101:/health"
    ["ai-order-service"]="8102:/health"
    ["market-data-service"]="8103:/health"
    ["defi-service"]="8104:/health"
    ["bright-data-hub-service"]="8105:/health"
    ["llm-orchestrator"]="8106:/health"
    ["llm-orchestrator-simple"]="8107:/health"
    ["redis-mcp-server"]="8108:/health"
    ["mcp-ai-integration"]="8109:/health"
    ["task-cli"]="8110:/health"
    ["api-gateway"]="8080:/health"
)

# Infrastructure components
INFRASTRUCTURE_COMPONENTS=(
    "redis:6379"
    "postgres:5432"
)

# =============================================================================
# COMMAND LINE PARSING
# =============================================================================

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -c|--comprehensive)
                COMPREHENSIVE=true
                shift
                ;;
            -m|--monitoring)
                MONITORING_MODE=true
                shift
                ;;
            -r|--report)
                GENERATE_REPORT=true
                shift
                ;;
            -h|--help)
                show_usage "health-check.sh" \
                    "Comprehensive health check system for all Go Coffee microservices" \
                    "  ./health-check.sh [OPTIONS]

  Options:
    -e, --environment   Environment to check (development|staging|production)
    -c, --comprehensive Run comprehensive health checks including performance
    -m, --monitoring    Enable continuous monitoring mode
    -r, --report        Generate detailed health report
    -h, --help          Show this help message

  Examples:
    ./health-check.sh                    # Basic health check
    ./health-check.sh --comprehensive    # Detailed health check
    ./health-check.sh --monitoring       # Continuous monitoring
    ./health-check.sh -e production -r   # Production check with report"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                print_info "Use --help for usage information"
                exit 1
                ;;
        esac
    done
}

# =============================================================================
# HEALTH CHECK FUNCTIONS
# =============================================================================

# Increment check counter
increment_check() {
    ((TOTAL_CHECKS++))
}

# Log health check results
log_health_pass() {
    increment_check
    ((PASSED_CHECKS++))
    print_status "$1"
}

log_health_warning() {
    increment_check
    ((WARNING_CHECKS++))
    print_warning "$1"
}

log_health_fail() {
    increment_check
    ((FAILED_CHECKS++))
    print_error "$1"
}

# Check if service is running and healthy
check_service_health() {
    local service_name=$1
    local endpoint_info=${SERVICE_HEALTH_ENDPOINTS[$service_name]}
    local port=$(echo "$endpoint_info" | cut -d':' -f1)
    local health_path=$(echo "$endpoint_info" | cut -d':' -f2)
    local health_url="http://localhost:$port$health_path"

    print_progress "Checking $service_name health..."

    # Check if port is listening
    if ! lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        log_health_fail "$service_name: Port $port not listening"
        return 1
    fi

    # Check health endpoint
    local response_code
    response_code=$(curl -s -o /dev/null -w "%{http_code}" --max-time $HEALTH_TIMEOUT "$health_url" 2>/dev/null || echo "000")

    case $response_code in
        200)
            log_health_pass "$service_name: Healthy (HTTP $response_code)"
            return 0
            ;;
        000)
            log_health_fail "$service_name: Health endpoint unreachable"
            return 1
            ;;
        *)
            log_health_warning "$service_name: Health endpoint returned HTTP $response_code"
            return 1
            ;;
    esac
}

# Check all microservices
check_all_services() {
    print_header "🔍 Checking Microservices Health"

    local healthy_services=0
    local total_services=0

    for service_name in "${!SERVICE_HEALTH_ENDPOINTS[@]}"; do
        ((total_services++))
        if check_service_health "$service_name"; then
            ((healthy_services++))
        fi
    done

    print_info "Service Health Summary: $healthy_services/$total_services services healthy"

    if [[ $healthy_services -eq $total_services ]]; then
        log_health_pass "All microservices are healthy"
    elif [[ $healthy_services -gt $((total_services / 2)) ]]; then
        log_health_warning "Some microservices are unhealthy ($((total_services - healthy_services)) failed)"
    else
        log_health_fail "Critical: Majority of microservices are unhealthy"
    fi
}

# Check infrastructure components
check_infrastructure() {
    print_header "🏗️ Checking Infrastructure Components"

    for component in "${INFRASTRUCTURE_COMPONENTS[@]}"; do
        local service=$(echo "$component" | cut -d':' -f1)
        local port=$(echo "$component" | cut -d':' -f2)

        print_progress "Checking $service..."

        case $service in
            "redis")
                if command_exists redis-cli; then
                    if redis-cli -p $port ping 2>/dev/null | grep -q "PONG"; then
                        log_health_pass "Redis: Responding to ping"
                    else
                        log_health_fail "Redis: Not responding to ping"
                    fi
                else
                    if nc -z localhost $port 2>/dev/null; then
                        log_health_pass "Redis: Port $port is accessible"
                    else
                        log_health_fail "Redis: Port $port is not accessible"
                    fi
                fi
                ;;
            "postgres")
                if command_exists psql; then
                    if PGPASSWORD=postgres psql -h localhost -p $port -U postgres -d postgres -c "SELECT 1;" >/dev/null 2>&1; then
                        log_health_pass "PostgreSQL: Database connection successful"
                    else
                        log_health_fail "PostgreSQL: Database connection failed"
                    fi
                else
                    if nc -z localhost $port 2>/dev/null; then
                        log_health_pass "PostgreSQL: Port $port is accessible"
                    else
                        log_health_fail "PostgreSQL: Port $port is not accessible"
                    fi
                fi
                ;;
            *)
                if nc -z localhost $port 2>/dev/null; then
                    log_health_pass "$service: Port $port is accessible"
                else
                    log_health_fail "$service: Port $port is not accessible"
                fi
                ;;
        esac
    done
}

# Check system resources
check_system_resources() {
    if [[ "$COMPREHENSIVE" != "true" ]]; then
        return 0
    fi

    print_header "💻 Checking System Resources"

    # Check disk space
    local disk_usage=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
    if [[ $disk_usage -lt 80 ]]; then
        log_health_pass "Disk usage: ${disk_usage}% (healthy)"
    elif [[ $disk_usage -lt 90 ]]; then
        log_health_warning "Disk usage: ${disk_usage}% (warning)"
    else
        log_health_fail "Disk usage: ${disk_usage}% (critical)"
    fi

    # Check memory usage
    local memory_usage=$(free | grep Mem | awk '{printf "%.0f", $3/$2 * 100.0}')
    if [[ $memory_usage -lt 80 ]]; then
        log_health_pass "Memory usage: ${memory_usage}% (healthy)"
    elif [[ $memory_usage -lt 90 ]]; then
        log_health_warning "Memory usage: ${memory_usage}% (warning)"
    else
        log_health_fail "Memory usage: ${memory_usage}% (critical)"
    fi

    # Check CPU load
    local cpu_load=$(uptime | awk -F'load average:' '{print $2}' | awk '{print $1}' | sed 's/,//')
    local cpu_cores=$(nproc)
    local load_percentage=$(echo "scale=0; $cpu_load * 100 / $cpu_cores" | bc -l)

    if [[ $load_percentage -lt 80 ]]; then
        log_health_pass "CPU load: ${load_percentage}% (healthy)"
    elif [[ $load_percentage -lt 100 ]]; then
        log_health_warning "CPU load: ${load_percentage}% (warning)"
    else
        log_health_fail "CPU load: ${load_percentage}% (critical)"
    fi
}

# Check API Gateway integration
check_api_gateway() {
    print_header "🌐 Checking API Gateway Integration"

    if [[ -z "${SERVICE_HEALTH_ENDPOINTS[api-gateway]:-}" ]]; then
        log_health_warning "API Gateway not configured for health checks"
        return 1
    fi

    local gateway_port=$(echo "${SERVICE_HEALTH_ENDPOINTS[api-gateway]}" | cut -d':' -f1)
    local gateway_url="http://localhost:$gateway_port"

    # Check gateway health
    if check_service_health "api-gateway"; then
        # Test service discovery
        print_progress "Testing service discovery..."
        local status_response
        status_response=$(curl -s --max-time $HEALTH_TIMEOUT "$gateway_url/api/v1/status" 2>/dev/null || echo "")

        if [[ -n "$status_response" ]]; then
            log_health_pass "API Gateway: Service discovery working"
        else
            log_health_warning "API Gateway: Service discovery endpoint not responding"
        fi

        # Test routing
        print_progress "Testing routing..."
        local routes_response
        routes_response=$(curl -s --max-time $HEALTH_TIMEOUT "$gateway_url/api/v1/routes" 2>/dev/null || echo "")

        if [[ -n "$routes_response" ]]; then
            log_health_pass "API Gateway: Routing configuration accessible"
        else
            log_health_warning "API Gateway: Routing configuration not accessible"
        fi
    else
        log_health_fail "API Gateway: Not healthy, skipping integration tests"
    fi
}

# Check service dependencies
check_service_dependencies() {
    if [[ "$COMPREHENSIVE" != "true" ]]; then
        return 0
    fi

    print_header "🔗 Checking Service Dependencies"

    # Check if auth service is accessible from other services
    if check_service_health "auth-service"; then
        print_progress "Testing auth service integration..."

        # Test auth endpoint
        local auth_response
        auth_response=$(curl -s --max-time $HEALTH_TIMEOUT "http://localhost:8091/api/v1/auth/status" 2>/dev/null || echo "")

        if [[ -n "$auth_response" ]]; then
            log_health_pass "Auth Service: API endpoints accessible"
        else
            log_health_warning "Auth Service: API endpoints not responding"
        fi
    fi

    # Check database connectivity from services
    print_progress "Testing database connectivity..."
    local db_connected_services=0
    local total_db_services=0

    for service in "auth-service" "order-service" "payment-service"; do
        if [[ -n "${SERVICE_HEALTH_ENDPOINTS[$service]:-}" ]]; then
            ((total_db_services++))
            local port=$(echo "${SERVICE_HEALTH_ENDPOINTS[$service]}" | cut -d':' -f1)
            local db_check_response
            db_check_response=$(curl -s --max-time $HEALTH_TIMEOUT "http://localhost:$port/health/db" 2>/dev/null || echo "")

            if [[ -n "$db_check_response" ]]; then
                ((db_connected_services++))
            fi
        fi
    done

    if [[ $db_connected_services -eq $total_db_services ]]; then
        log_health_pass "Database connectivity: All services connected"
    elif [[ $db_connected_services -gt 0 ]]; then
        log_health_warning "Database connectivity: $db_connected_services/$total_db_services services connected"
    else
        log_health_fail "Database connectivity: No services connected"
    fi
}

# Generate health report
generate_health_report() {
    if [[ "$GENERATE_REPORT" != "true" ]]; then
        return 0
    fi

    print_header "📊 Generating Health Report"

    local report_file="health-report-$(date +%Y%m%d-%H%M%S).json"
    local success_rate=0

    if [[ $TOTAL_CHECKS -gt 0 ]]; then
        success_rate=$(echo "scale=2; $PASSED_CHECKS * 100 / $TOTAL_CHECKS" | bc -l)
    fi

    local overall_status="healthy"
    if [[ $FAILED_CHECKS -gt 0 ]]; then
        overall_status="unhealthy"
    elif [[ $WARNING_CHECKS -gt 3 ]]; then
        overall_status="degraded"
    elif [[ $WARNING_CHECKS -gt 0 ]]; then
        overall_status="warning"
    fi

    cat > "$report_file" <<EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "environment": "$ENVIRONMENT",
  "overall_status": "$overall_status",
  "summary": {
    "total_checks": $TOTAL_CHECKS,
    "passed": $PASSED_CHECKS,
    "failed": $FAILED_CHECKS,
    "warnings": $WARNING_CHECKS,
    "success_rate": $success_rate
  },
  "services": {
EOF

    # Add service status
    local first=true
    for service_name in "${!SERVICE_HEALTH_ENDPOINTS[@]}"; do
        if [[ "$first" == "true" ]]; then
            first=false
        else
            echo "," >> "$report_file"
        fi

        local endpoint_info=${SERVICE_HEALTH_ENDPOINTS[$service_name]}
        local port=$(echo "$endpoint_info" | cut -d':' -f1)
        local health_path=$(echo "$endpoint_info" | cut -d':' -f2)
        local health_url="http://localhost:$port$health_path"

        local status="unknown"
        local response_time="N/A"

        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            local start_time=$(date +%s%N)
            local response_code
            response_code=$(curl -s -o /dev/null -w "%{http_code}" --max-time $HEALTH_TIMEOUT "$health_url" 2>/dev/null || echo "000")
            local end_time=$(date +%s%N)
            response_time=$(echo "scale=3; ($end_time - $start_time) / 1000000" | bc -l)

            case $response_code in
                200) status="healthy" ;;
                000) status="unreachable" ;;
                *) status="unhealthy" ;;
            esac
        else
            status="not_running"
        fi

        echo "    \"$service_name\": {" >> "$report_file"
        echo "      \"status\": \"$status\"," >> "$report_file"
        echo "      \"port\": $port," >> "$report_file"
        echo "      \"response_time_ms\": \"$response_time\"" >> "$report_file"
        echo -n "    }" >> "$report_file"
    done

    echo "" >> "$report_file"
    echo "  }," >> "$report_file"
    echo "  \"recommendations\": [" >> "$report_file"

    if [[ $FAILED_CHECKS -gt 0 ]]; then
        echo "    \"Investigate failed services immediately\"," >> "$report_file"
    fi
    if [[ $WARNING_CHECKS -gt 3 ]]; then
        echo "    \"Review services with warnings\"," >> "$report_file"
    fi

    echo "    \"Monitor resource usage trends\"," >> "$report_file"
    echo "    \"Ensure backup procedures are working\"" >> "$report_file"
    echo "  ]" >> "$report_file"
    echo "}" >> "$report_file"

    print_status "Health report generated: $report_file"
}

# Continuous monitoring mode
run_monitoring() {
    print_header "🔄 Starting Continuous Health Monitoring"
    print_info "Monitoring interval: ${CHECK_INTERVAL}s"
    print_info "Press Ctrl+C to stop monitoring"

    local iteration=0
    while true; do
        ((iteration++))

        print_header "📊 Health Check Iteration #$iteration ($(date))"

        # Reset counters
        TOTAL_CHECKS=0
        PASSED_CHECKS=0
        FAILED_CHECKS=0
        WARNING_CHECKS=0

        # Run health checks
        check_all_services
        check_infrastructure
        check_api_gateway

        # Show summary
        local success_rate=0
        if [[ $TOTAL_CHECKS -gt 0 ]]; then
            success_rate=$(echo "scale=1; $PASSED_CHECKS * 100 / $TOTAL_CHECKS" | bc -l)
        fi

        print_info "Iteration #$iteration Summary: $PASSED_CHECKS/$TOTAL_CHECKS passed (${success_rate}%)"

        if [[ $FAILED_CHECKS -gt 0 ]]; then
            print_warning "$FAILED_CHECKS critical issues detected"
        fi

        if [[ $WARNING_CHECKS -gt 0 ]]; then
            print_info "$WARNING_CHECKS warnings detected"
        fi

        print_info "Next check in ${CHECK_INTERVAL}s..."
        sleep $CHECK_INTERVAL
    done
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    local start_time=$(date +%s)

    # Parse command line arguments
    parse_args "$@"

    # Check dependencies
    local deps=("curl" "bc")
    if [[ "$COMPREHENSIVE" == "true" ]]; then
        deps+=("free" "df" "uptime" "nproc")
    fi
    check_dependencies "${deps[@]}" || exit 1

    print_info "Environment: $ENVIRONMENT"
    print_info "Comprehensive mode: $COMPREHENSIVE"
    print_info "Monitoring mode: $MONITORING_MODE"
    print_info "Generate report: $GENERATE_REPORT"

    # Run monitoring mode if requested
    if [[ "$MONITORING_MODE" == "true" ]]; then
        run_monitoring
        return 0
    fi

    # Run health checks
    print_header "🏥 Starting Health Check Suite"

    check_all_services
    check_infrastructure
    check_api_gateway
    check_system_resources
    check_service_dependencies

    # Calculate execution time
    local end_time=$(date +%s)
    local execution_time=$((end_time - start_time))

    # Generate report if requested
    generate_health_report

    # Show final summary
    print_header "📊 Health Check Summary"

    local success_rate=0
    if [[ $TOTAL_CHECKS -gt 0 ]]; then
        success_rate=$(echo "scale=1; $PASSED_CHECKS * 100 / $TOTAL_CHECKS" | bc -l)
    fi

    echo -e "${BOLD}Environment:${NC} $ENVIRONMENT"
    echo -e "${BOLD}Total Checks:${NC} $TOTAL_CHECKS"
    echo -e "${GREEN}Passed:${NC} $PASSED_CHECKS"
    echo -e "${YELLOW}Warnings:${NC} $WARNING_CHECKS"
    echo -e "${RED}Failed:${NC} $FAILED_CHECKS"
    echo -e "${BLUE}Success Rate:${NC} ${success_rate}%"
    echo -e "${BLUE}Execution Time:${NC} ${execution_time}s"

    # Determine overall health status
    if [[ $FAILED_CHECKS -eq 0 && $WARNING_CHECKS -eq 0 ]]; then
        print_success "🎉 All systems are healthy!"

        print_header "🚀 System Status"
        print_status "✅ All microservices are running"
        print_status "✅ Infrastructure components are healthy"
        print_status "✅ API Gateway is functioning"
        print_status "✅ Service dependencies are working"

        if [[ "$COMPREHENSIVE" == "true" ]]; then
            print_status "✅ System resources are within normal limits"
        fi

        exit 0

    elif [[ $FAILED_CHECKS -eq 0 ]]; then
        print_warning "⚠️  System is healthy with $WARNING_CHECKS warnings"

        print_header "🔍 Recommendations"
        print_info "• Monitor services with warnings"
        print_info "• Review system logs for potential issues"
        print_info "• Consider running comprehensive checks"

        exit 1

    else
        print_error "❌ System has critical issues ($FAILED_CHECKS failures)"

        print_header "🚨 Critical Issues Detected"
        print_error "• $FAILED_CHECKS critical failures require immediate attention"
        if [[ $WARNING_CHECKS -gt 0 ]]; then
            print_warning "• $WARNING_CHECKS additional warnings detected"
        fi

        print_header "🔧 Immediate Actions Required"
        print_info "• Check service logs: tail -f logs/SERVICE.log"
        print_info "• Restart failed services: ./scripts/start-all-services.sh"
        print_info "• Verify infrastructure: docker ps, systemctl status"
        print_info "• Check resource availability: df -h, free -h"

        if [[ "$GENERATE_REPORT" == "true" ]]; then
            print_info "• Review detailed report for specific issues"
        fi

        exit 2
    fi
}

# Run main function with all arguments
main "$@"

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
