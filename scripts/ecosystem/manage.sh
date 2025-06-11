#!/bin/bash

# Go Coffee - Ecosystem Management
# Unified management interface for the entire Go Coffee platform ecosystem
# Version: 3.0.0
# Usage: ./manage.sh [OPTIONS] COMMAND
#   -e, --environment   Environment (development|staging|production)
#   -v, --verbose       Verbose output
#   -q, --quiet         Quiet mode
#   -h, --help          Show this help message

set -euo pipefail

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

# Source shared library
source "$PROJECT_ROOT/scripts/lib/common.sh" 2>/dev/null || {
    echo "❌ Cannot load shared library. Please run from project root."
    exit 1
}

print_header "🌟 Go Coffee Ecosystem Management"

# =============================================================================
# CONFIGURATION
# =============================================================================

ENVIRONMENT="${ENVIRONMENT:-development}"
VERBOSE_MODE=false
QUIET_MODE=false
COMMAND=""

# Ecosystem components
declare -A ECOSYSTEM_COMPONENTS=(
    ["core"]="Core microservices (27 services)"
    ["crypto"]="Crypto wallet services (8 services)"
    ["web-ui"]="Web UI stack (3 services)"
    ["monitoring"]="Observability stack (5 services)"
    ["infrastructure"]="Infrastructure services (Redis, PostgreSQL)"
)

# Management commands
declare -A MANAGEMENT_COMMANDS=(
    ["status"]="Show ecosystem status"
    ["start"]="Start ecosystem components"
    ["stop"]="Stop ecosystem components"
    ["restart"]="Restart ecosystem components"
    ["build"]="Build all components"
    ["test"]="Run comprehensive tests"
    ["deploy"]="Deploy to environment"
    ["monitor"]="Setup monitoring"
    ["security"]="Run security scans"
    ["docs"]="Generate documentation"
    ["backup"]="Create system backup"
    ["restore"]="Restore from backup"
    ["update"]="Update all components"
    ["clean"]="Clean up resources"
)

# =============================================================================
# COMMAND LINE PARSING
# =============================================================================

parse_ecosystem_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -v|--verbose)
                VERBOSE_MODE=true
                shift
                ;;
            -q|--quiet)
                QUIET_MODE=true
                shift
                ;;
            -h|--help)
                show_ecosystem_help
                exit 0
                ;;
            status|start|stop|restart|build|test|deploy|monitor|security|docs|backup|restore|update|clean)
                COMMAND="$1"
                shift
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    if [[ -z "$COMMAND" ]]; then
        print_error "No command specified"
        show_ecosystem_help
        exit 1
    fi
}

# Show ecosystem help
show_ecosystem_help() {
    cat << EOF
🌟 Go Coffee Ecosystem Management v3.0.0

Usage: $0 [OPTIONS] COMMAND

Commands:
    status      Show ecosystem status and health
    start       Start ecosystem components
    stop        Stop ecosystem components  
    restart     Restart ecosystem components
    build       Build all components
    test        Run comprehensive tests
    deploy      Deploy to environment
    monitor     Setup monitoring and observability
    security    Run security scans and analysis
    docs        Generate documentation
    backup      Create system backup
    restore     Restore from backup
    update      Update all components
    clean       Clean up resources

Options:
    -e, --environment   Environment (development|staging|production)
    -v, --verbose       Verbose output
    -q, --quiet         Quiet mode
    -h, --help          Show this help message

Examples:
    $0 status                           # Show ecosystem status
    $0 start -e development            # Start development environment
    $0 deploy -e production            # Deploy to production
    $0 test -v                         # Run tests with verbose output
    $0 monitor                         # Setup monitoring stack

Ecosystem Components:
EOF

    for component in "${!ECOSYSTEM_COMPONENTS[@]}"; do
        printf "    %-12s %s\n" "$component" "${ECOSYSTEM_COMPONENTS[$component]}"
    done
    
    echo ""
}

# =============================================================================
# ECOSYSTEM MANAGEMENT FUNCTIONS
# =============================================================================

# Show ecosystem status
ecosystem_status() {
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_header "📊 Go Coffee Ecosystem Status"
    fi
    
    local total_components=0
    local healthy_components=0
    
    # Check core services
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_info "Checking core microservices..."
    fi
    
    if "$PROJECT_ROOT/scripts/health-check.sh" --quiet >/dev/null 2>&1; then
        ((healthy_components++))
        if [[ "$QUIET_MODE" != "true" ]]; then
            print_status "Core services: HEALTHY"
        fi
    else
        if [[ "$QUIET_MODE" != "true" ]]; then
            print_error "Core services: UNHEALTHY"
        fi
    fi
    ((total_components++))
    
    # Check crypto wallet services
    if [[ -f "$PROJECT_ROOT/crypto-wallet/run.sh" ]]; then
        if [[ "$QUIET_MODE" != "true" ]]; then
            print_info "Checking crypto wallet services..."
        fi
        
        if curl -s --max-time 5 "http://localhost:8080/health" >/dev/null 2>&1; then
            ((healthy_components++))
            if [[ "$QUIET_MODE" != "true" ]]; then
                print_status "Crypto services: HEALTHY"
            fi
        else
            if [[ "$QUIET_MODE" != "true" ]]; then
                print_error "Crypto services: UNHEALTHY"
            fi
        fi
        ((total_components++))
    fi
    
    # Check web UI services
    if [[ -f "$PROJECT_ROOT/web-ui/start-all.sh" ]]; then
        if [[ "$QUIET_MODE" != "true" ]]; then
            print_info "Checking web UI services..."
        fi
        
        if curl -s --max-time 5 "http://localhost:3000" >/dev/null 2>&1; then
            ((healthy_components++))
            if [[ "$QUIET_MODE" != "true" ]]; then
                print_status "Web UI: HEALTHY"
            fi
        else
            if [[ "$QUIET_MODE" != "true" ]]; then
                print_error "Web UI: UNHEALTHY"
            fi
        fi
        ((total_components++))
    fi
    
    # Check monitoring stack
    if curl -s --max-time 5 "http://localhost:9090" >/dev/null 2>&1; then
        ((healthy_components++))
        if [[ "$QUIET_MODE" != "true" ]]; then
            print_status "Monitoring: HEALTHY"
        fi
    else
        if [[ "$QUIET_MODE" != "true" ]]; then
            print_error "Monitoring: UNHEALTHY"
        fi
    fi
    ((total_components++))
    
    # Summary
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_header "📈 Ecosystem Health Summary"
        echo -e "${BOLD}Total Components:${NC} $total_components"
        echo -e "${GREEN}Healthy:${NC} $healthy_components"
        echo -e "${RED}Unhealthy:${NC} $((total_components - healthy_components))"
        
        local health_percentage=$((healthy_components * 100 / total_components))
        echo -e "${BLUE}Health Score:${NC} ${health_percentage}%"
        
        if [[ $healthy_components -eq $total_components ]]; then
            print_success "🎉 Ecosystem is fully operational!"
        elif [[ $healthy_components -gt $((total_components / 2)) ]]; then
            print_warning "⚠️  Ecosystem is partially operational"
        else
            print_error "❌ Ecosystem has critical issues"
        fi
    fi
}

# Start ecosystem
ecosystem_start() {
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_header "🚀 Starting Go Coffee Ecosystem"
    fi
    
    # Start infrastructure first
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_progress "Starting infrastructure services..."
    fi
    
    # Start core services
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_progress "Starting core microservices..."
    fi
    
    local start_args=()
    if [[ "$ENVIRONMENT" == "development" ]]; then
        start_args+=("--dev-mode")
    elif [[ "$ENVIRONMENT" == "production" ]]; then
        start_args+=("--production")
    fi
    
    "$PROJECT_ROOT/scripts/start-all-services.sh" "${start_args[@]}" &
    local core_pid=$!
    
    # Start crypto wallet services
    if [[ -f "$PROJECT_ROOT/crypto-wallet/run.sh" ]]; then
        if [[ "$QUIET_MODE" != "true" ]]; then
            print_progress "Starting crypto wallet services..."
        fi
        
        cd "$PROJECT_ROOT/crypto-wallet"
        ./run.sh "${start_args[@]}" &
        local crypto_pid=$!
        cd "$PROJECT_ROOT"
    fi
    
    # Start web UI services
    if [[ -f "$PROJECT_ROOT/web-ui/start-all.sh" ]]; then
        if [[ "$QUIET_MODE" != "true" ]]; then
            print_progress "Starting web UI services..."
        fi
        
        cd "$PROJECT_ROOT/web-ui"
        ./start-all.sh "${start_args[@]}" &
        local webui_pid=$!
        cd "$PROJECT_ROOT"
    fi
    
    # Wait for services to start
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_progress "Waiting for services to initialize..."
    fi
    sleep 30
    
    # Verify startup
    ecosystem_status
    
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_success "Ecosystem startup completed"
    fi
}

# Stop ecosystem
ecosystem_stop() {
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_header "🛑 Stopping Go Coffee Ecosystem"
    fi
    
    # Stop all services gracefully
    pkill -f "go-coffee" 2>/dev/null || true
    pkill -f "start-all-services" 2>/dev/null || true
    pkill -f "web-ui" 2>/dev/null || true
    pkill -f "crypto-wallet" 2>/dev/null || true
    
    # Stop Docker containers if running
    if command_exists docker-compose; then
        docker-compose down 2>/dev/null || true
    fi
    
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_success "Ecosystem stopped"
    fi
}

# Build ecosystem
ecosystem_build() {
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_header "🔨 Building Go Coffee Ecosystem"
    fi
    
    # Build core services
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_progress "Building core services..."
    fi
    
    "$PROJECT_ROOT/build_all.sh" --core-only
    
    # Build crypto wallet
    if [[ -f "$PROJECT_ROOT/crypto-wallet/run.sh" ]]; then
        if [[ "$QUIET_MODE" != "true" ]]; then
            print_progress "Building crypto wallet services..."
        fi
        
        cd "$PROJECT_ROOT/crypto-wallet"
        ./run.sh --build-only
        cd "$PROJECT_ROOT"
    fi
    
    # Build web UI
    if [[ -f "$PROJECT_ROOT/web-ui/start-all.sh" ]]; then
        if [[ "$QUIET_MODE" != "true" ]]; then
            print_progress "Building web UI services..."
        fi
        
        cd "$PROJECT_ROOT/web-ui"
        ./start-all.sh --build-only
        cd "$PROJECT_ROOT"
    fi
    
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_success "Ecosystem build completed"
    fi
}

# Test ecosystem
ecosystem_test() {
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_header "🧪 Testing Go Coffee Ecosystem"
    fi
    
    local test_args=()
    if [[ "$VERBOSE_MODE" == "true" ]]; then
        test_args+=("--verbose")
    fi
    
    # Run comprehensive tests
    "$PROJECT_ROOT/scripts/test-all-services.sh" "${test_args[@]}"
    
    # Run security scans
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_progress "Running security scans..."
    fi
    
    "$PROJECT_ROOT/scripts/security/security-scan.sh" --scope full --quiet
    
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_success "Ecosystem testing completed"
    fi
}

# Deploy ecosystem
ecosystem_deploy() {
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_header "🚀 Deploying Go Coffee Ecosystem"
    fi
    
    # Deploy using enhanced deployment script
    "$PROJECT_ROOT/scripts/deploy.sh" --env "$ENVIRONMENT" --backup k8s
    
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_success "Ecosystem deployment completed"
    fi
}

# Setup monitoring
ecosystem_monitor() {
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_header "📊 Setting Up Ecosystem Monitoring"
    fi
    
    # Setup observability stack
    "$PROJECT_ROOT/scripts/monitoring/setup-observability.sh" --custom --alerts
    
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_success "Monitoring setup completed"
    fi
}

# Run security scans
ecosystem_security() {
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_header "🔒 Running Ecosystem Security Scans"
    fi
    
    # Run comprehensive security scan
    "$PROJECT_ROOT/scripts/security/security-scan.sh" --scope full --level comprehensive
    
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_success "Security scanning completed"
    fi
}

# Generate documentation
ecosystem_docs() {
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_header "📚 Generating Ecosystem Documentation"
    fi
    
    # Generate comprehensive documentation
    "$PROJECT_ROOT/scripts/docs/generate-docs.sh" --type all --format all
    
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_success "Documentation generation completed"
    fi
}

# Clean ecosystem
ecosystem_clean() {
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_header "🧹 Cleaning Go Coffee Ecosystem"
    fi
    
    # Stop all services
    ecosystem_stop
    
    # Clean build artifacts
    rm -rf bin/ build/ dist/ 2>/dev/null || true
    
    # Clean Docker resources
    if command_exists docker; then
        docker system prune -f 2>/dev/null || true
    fi
    
    # Clean temporary files
    find . -name "*.log" -delete 2>/dev/null || true
    find . -name "*.tmp" -delete 2>/dev/null || true
    
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_success "Ecosystem cleanup completed"
    fi
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    local start_time=$(date +%s)
    
    # Parse arguments
    parse_ecosystem_args "$@"
    
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_info "Ecosystem Management Configuration:"
        print_info "  Command: $COMMAND"
        print_info "  Environment: $ENVIRONMENT"
        print_info "  Verbose: $VERBOSE_MODE"
        print_info "  Quiet: $QUIET_MODE"
    fi
    
    # Execute command
    case "$COMMAND" in
        "status")
            ecosystem_status
            ;;
        "start")
            ecosystem_start
            ;;
        "stop")
            ecosystem_stop
            ;;
        "restart")
            ecosystem_stop
            sleep 5
            ecosystem_start
            ;;
        "build")
            ecosystem_build
            ;;
        "test")
            ecosystem_test
            ;;
        "deploy")
            ecosystem_deploy
            ;;
        "monitor")
            ecosystem_monitor
            ;;
        "security")
            ecosystem_security
            ;;
        "docs")
            ecosystem_docs
            ;;
        "clean")
            ecosystem_clean
            ;;
        *)
            print_error "Unknown command: $COMMAND"
            exit 1
            ;;
    esac
    
    # Calculate execution time
    local end_time=$(date +%s)
    local total_time=$((end_time - start_time))
    
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_success "🎉 Ecosystem command '$COMMAND' completed in ${total_time}s"
    fi
}

# Run main function with all arguments
main "$@"
