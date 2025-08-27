#!/bin/bash

# ☕ Go Coffee - Performance Optimization Script
# Implements comprehensive performance optimizations for all services

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Functions
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
    exit 1
}

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $1${NC}"
}

success() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] SUCCESS: $1${NC}"
}

# Performance optimization functions
optimize_go_services() {
    log "🚀 Optimizing Go services for performance..."
    
    # Set Go performance environment variables
    export GOGC=100
    export GOMAXPROCS=$(nproc)
    export GOMEMLIMIT=1GiB
    
    info "Setting Go runtime optimizations:"
    info "  - GOGC=100 (garbage collection target)"
    info "  - GOMAXPROCS=$(nproc) (CPU cores)"
    info "  - GOMEMLIMIT=1GiB (memory limit)"
    
    # Rebuild services with optimizations
    log "Rebuilding services with performance optimizations..."
    
    cd "$PROJECT_ROOT"
    
    # Build with optimizations
    local build_flags="-ldflags '-s -w' -trimpath"
    
    if [ -f "bin/api-gateway" ]; then
        info "Rebuilding API Gateway with optimizations..."
        go build -ldflags '-s -w' -trimpath -o bin/api-gateway-optimized ./cmd/api-gateway/main.go
    fi

    if [ -f "bin/redis-mcp-server" ]; then
        info "Rebuilding Redis MCP Server with optimizations..."
        go build -ldflags '-s -w' -trimpath -o bin/redis-mcp-server-optimized ./cmd/redis-mcp-server/main.go
    fi

    if [ -f "bin/ai-search" ]; then
        info "Rebuilding AI Search with optimizations..."
        go build -ldflags '-s -w' -trimpath -o bin/ai-search-optimized ./cmd/ai-search/main.go
    fi
    
    success "Go services optimized successfully!"
}

optimize_system_settings() {
    log "⚙️ Optimizing system settings for performance..."
    
    # Check if we can modify system settings
    if [ "$EUID" -eq 0 ]; then
        info "Running as root, applying system optimizations..."
        
        # Increase file descriptor limits
        echo "* soft nofile 65536" >> /etc/security/limits.conf
        echo "* hard nofile 65536" >> /etc/security/limits.conf
        
        # Optimize network settings
        echo "net.core.somaxconn = 65536" >> /etc/sysctl.conf
        echo "net.core.netdev_max_backlog = 5000" >> /etc/sysctl.conf
        echo "net.ipv4.tcp_max_syn_backlog = 65536" >> /etc/sysctl.conf
        
        sysctl -p
        
        success "System settings optimized!"
    else
        warn "Not running as root, skipping system optimizations"
        info "For full optimization, run: sudo $0"
    fi
}

optimize_redis_settings() {
    log "🔧 Optimizing Redis settings for performance..."
    
    # Check if Redis is running
    if pgrep -x "redis-server" > /dev/null; then
        info "Redis is running, applying optimizations..."
        
        # Connect to Redis and apply optimizations
        redis-cli CONFIG SET maxmemory-policy allkeys-lru 2>/dev/null || warn "Could not set Redis maxmemory-policy"
        redis-cli CONFIG SET tcp-keepalive 60 2>/dev/null || warn "Could not set Redis tcp-keepalive"
        redis-cli CONFIG SET timeout 0 2>/dev/null || warn "Could not set Redis timeout"
        
        success "Redis optimizations applied!"
    else
        warn "Redis is not running, skipping Redis optimizations"
    fi
}

create_performance_configs() {
    log "📝 Creating performance configuration files..."
    
    # Create optimized environment file
    cat > "$PROJECT_ROOT/.env.performance" << EOF
# Go Coffee Performance Optimizations

# Go Runtime Settings
GOGC=100
GOMAXPROCS=$(nproc)
GOMEMLIMIT=1GiB

# Service Settings
API_GATEWAY_WORKERS=4
REDIS_MCP_WORKERS=2
AI_SEARCH_WORKERS=2

# Connection Pool Settings
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=300s

# Cache Settings
CACHE_TTL=300s
CACHE_MAX_SIZE=100MB

# Rate Limiting
RATE_LIMIT_REQUESTS=1000
RATE_LIMIT_WINDOW=60s

# Monitoring
METRICS_ENABLED=true
TRACING_ENABLED=true
PROFILING_ENABLED=false

# Performance Tuning
HTTP_READ_TIMEOUT=30s
HTTP_WRITE_TIMEOUT=30s
HTTP_IDLE_TIMEOUT=120s
HTTP_MAX_HEADER_BYTES=1048576

# AI Search Optimizations
AI_SEARCH_BATCH_SIZE=100
AI_SEARCH_CACHE_SIZE=1000
AI_SEARCH_VECTOR_CACHE=true

# Redis Optimizations
REDIS_POOL_SIZE=10
REDIS_MIN_IDLE_CONNS=5
REDIS_DIAL_TIMEOUT=5s
REDIS_READ_TIMEOUT=3s
REDIS_WRITE_TIMEOUT=3s
EOF

    success "Performance configuration created at .env.performance"
}

run_performance_tests() {
    log "🧪 Running performance tests..."
    
    # Test API Gateway performance
    info "Testing API Gateway performance..."
    local api_gateway_time=$(curl -o /dev/null -s -w '%{time_total}' http://localhost:8080/health)
    info "API Gateway response time: ${api_gateway_time}s"
    
    # Test Redis MCP Server performance
    info "Testing Redis MCP Server performance..."
    local redis_mcp_time=$(curl -o /dev/null -s -w '%{time_total}' http://localhost:8108/api/v1/redis-mcp/health)
    info "Redis MCP Server response time: ${redis_mcp_time}s"
    
    # Test AI Search performance
    info "Testing AI Search performance..."
    local ai_search_time=$(curl -o /dev/null -s -w '%{time_total}' -X POST http://localhost:8092/api/v1/ai-search/semantic -H "Content-Type: application/json" -d '{"query":"performance test","limit":1}')
    info "AI Search response time: ${ai_search_time}s"
    
    # Load test with concurrent requests
    info "Running concurrent load test..."
    for i in {1..10}; do
        curl -s http://localhost:8080/health > /dev/null &
    done
    wait
    
    success "Performance tests completed!"
}

generate_performance_report() {
    log "📊 Generating performance report..."
    
    local report_file="$PROJECT_ROOT/performance-report-$(date +%Y%m%d-%H%M%S).json"
    
    # Get current metrics from performance monitor
    curl -s http://localhost:9999/metrics > "$report_file" 2>/dev/null || warn "Could not fetch metrics from performance monitor"
    
    if [ -f "$report_file" ]; then
        success "Performance report saved to: $report_file"
        
        # Display summary
        info "Performance Summary:"
        if command -v jq >/dev/null 2>&1; then
            jq '.summary' "$report_file" 2>/dev/null || info "Report saved but could not parse JSON"
        else
            info "Install 'jq' to see formatted performance summary"
        fi
    fi
}

monitor_performance() {
    log "📈 Starting continuous performance monitoring..."
    
    info "Performance monitoring dashboard: http://localhost:9999/dashboard"
    info "Prometheus metrics: http://localhost:9999/metrics/prometheus"
    
    # Monitor for 60 seconds and show key metrics
    for i in {1..12}; do
        sleep 5
        local timestamp=$(date '+%H:%M:%S')
        local cpu_usage=$(curl -s http://localhost:9999/metrics 2>/dev/null | jq -r '.system_metrics.cpu_usage // "N/A"')
        local memory_usage=$(curl -s http://localhost:9999/metrics 2>/dev/null | jq -r '.system_metrics.memory_usage // "N/A"')
        local healthy_services=$(curl -s http://localhost:9999/metrics 2>/dev/null | jq -r '.summary.healthy_services // "N/A"')
        
        info "[$timestamp] CPU: ${cpu_usage}% | Memory: ${memory_usage}% | Healthy Services: $healthy_services/6"
    done
    
    success "Performance monitoring completed!"
}

# Main execution
main() {
    echo -e "${PURPLE}"
    echo "☕ ═══════════════════════════════════════════════════════════════════════════════ ☕"
    echo "   ██████╗  ██████╗       ██████╗ ██████╗ ███████╗███████╗███████╗███████╗"
    echo "  ██╔════╝ ██╔═══██╗     ██╔════╝██╔═══██╗██╔════╝██╔════╝██╔════╝██╔════╝"
    echo "  ██║  ███╗██║   ██║     ██║     ██║   ██║█████╗  █████╗  █████╗  █████╗"
    echo "  ██║   ██║██║   ██║     ██║     ██║   ██║██╔══╝  ██╔══╝  ██╔══╝  ██╔══╝"
    echo "  ╚██████╔╝╚██████╔╝     ╚██████╗╚██████╔╝██║     ██║     ███████╗███████╗"
    echo "   ╚═════╝  ╚═════╝       ╚═════╝ ╚═════╝ ╚═╝     ╚═╝     ╚══════╝╚══════╝"
    echo ""
    echo "                    🚀 Performance Optimization Suite 📊"
    echo "☕ ═══════════════════════════════════════════════════════════════════════════════ ☕"
    echo -e "${NC}"
    
    log "Starting Go Coffee Performance Optimization..."
    
    case "${1:-all}" in
        "optimize")
            optimize_go_services
            optimize_system_settings
            optimize_redis_settings
            create_performance_configs
            ;;
        "test")
            run_performance_tests
            ;;
        "report")
            generate_performance_report
            ;;
        "monitor")
            monitor_performance
            ;;
        "all")
            optimize_go_services
            optimize_system_settings
            optimize_redis_settings
            create_performance_configs
            run_performance_tests
            generate_performance_report
            monitor_performance
            ;;
        *)
            echo "Usage: $0 [optimize|test|report|monitor|all]"
            echo ""
            echo "Commands:"
            echo "  optimize  - Apply performance optimizations"
            echo "  test      - Run performance tests"
            echo "  report    - Generate performance report"
            echo "  monitor   - Start performance monitoring"
            echo "  all       - Run all optimization steps (default)"
            exit 1
            ;;
    esac
    
    success "🎉 Performance optimization completed successfully!"
    info "📊 Dashboard: http://localhost:9999/dashboard"
    info "📈 Metrics: http://localhost:9999/metrics"
}

# Run main function with all arguments
main "$@"
