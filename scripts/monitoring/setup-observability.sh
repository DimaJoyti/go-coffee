#!/bin/bash

# Go Coffee - Advanced Observability Setup
# Sets up comprehensive monitoring with Prometheus, Grafana, Jaeger, and custom dashboards
# Version: 3.0.0
# Usage: ./setup-observability.sh [OPTIONS]
#   -e, --environment   Environment (development|staging|production)
#   -m, --mode         Setup mode (docker|kubernetes|local)
#   -c, --custom       Include custom Go Coffee dashboards
#   -a, --alerts       Setup alerting rules
#   -s, --storage      Setup persistent storage
#   -h, --help         Show this help message

set -euo pipefail

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

# Source shared library
source "$PROJECT_ROOT/scripts/lib/common.sh" 2>/dev/null || {
    echo "❌ Cannot load shared library. Please run from project root."
    exit 1
}

print_header "📊 Go Coffee Advanced Observability Setup"

# =============================================================================
# CONFIGURATION
# =============================================================================

ENVIRONMENT="${ENVIRONMENT:-development}"
SETUP_MODE="docker"
INCLUDE_CUSTOM=false
SETUP_ALERTS=false
SETUP_STORAGE=false

# Monitoring stack configuration
PROMETHEUS_PORT=9090
GRAFANA_PORT=3000
JAEGER_PORT=16686
ALERTMANAGER_PORT=9093
LOKI_PORT=3100

# Service discovery configuration
declare -A MONITORING_SERVICES=(
    ["prometheus"]="$PROMETHEUS_PORT"
    ["grafana"]="$GRAFANA_PORT"
    ["jaeger"]="$JAEGER_PORT"
    ["alertmanager"]="$ALERTMANAGER_PORT"
    ["loki"]="$LOKI_PORT"
)

# =============================================================================
# COMMAND LINE PARSING
# =============================================================================

parse_observability_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -m|--mode)
                SETUP_MODE="$2"
                shift 2
                ;;
            -c|--custom)
                INCLUDE_CUSTOM=true
                shift
                ;;
            -a|--alerts)
                SETUP_ALERTS=true
                shift
                ;;
            -s|--storage)
                SETUP_STORAGE=true
                shift
                ;;
            -h|--help)
                show_usage "setup-observability.sh" \
                    "Advanced observability setup for Go Coffee platform" \
                    "  ./setup-observability.sh [OPTIONS]
  
  Options:
    -e, --environment   Environment (development|staging|production)
    -m, --mode         Setup mode (docker|kubernetes|local)
    -c, --custom       Include custom Go Coffee dashboards
    -a, --alerts       Setup alerting rules
    -s, --storage      Setup persistent storage
    -h, --help         Show this help message
  
  Examples:
    ./setup-observability.sh                    # Basic Docker setup
    ./setup-observability.sh -c -a             # With custom dashboards and alerts
    ./setup-observability.sh -m kubernetes -s  # Kubernetes with storage
    ./setup-observability.sh -e production -a  # Production with alerting"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
}

# =============================================================================
# OBSERVABILITY SETUP FUNCTIONS
# =============================================================================

# Create monitoring directories
create_monitoring_structure() {
    print_header "📁 Creating Monitoring Directory Structure"
    
    local dirs=(
        "monitoring/prometheus/config"
        "monitoring/prometheus/rules"
        "monitoring/prometheus/data"
        "monitoring/grafana/dashboards"
        "monitoring/grafana/provisioning/dashboards"
        "monitoring/grafana/provisioning/datasources"
        "monitoring/grafana/data"
        "monitoring/jaeger/data"
        "monitoring/loki/config"
        "monitoring/loki/data"
        "monitoring/alertmanager/config"
        "monitoring/alertmanager/data"
    )
    
    for dir in "${dirs[@]}"; do
        mkdir -p "$dir"
        print_status "Created: $dir"
    done
    
    print_success "Monitoring directory structure created"
}

# Generate Prometheus configuration
generate_prometheus_config() {
    print_header "⚙️ Generating Prometheus Configuration"
    
    cat > monitoring/prometheus/config/prometheus.yml <<EOF
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    environment: '$ENVIRONMENT'
    cluster: 'go-coffee'

rule_files:
  - "/etc/prometheus/rules/*.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

scrape_configs:
  # Prometheus itself
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  # Go Coffee Core Services
  - job_name: 'go-coffee-core'
    static_configs:
      - targets:
        - 'auth-service:8091'
        - 'payment-service:8093'
        - 'order-service:8094'
        - 'kitchen-service:8095'
        - 'user-gateway:8096'
        - 'security-gateway:8097'
        - 'communication-hub:8098'
        - 'api-gateway:8080'
    metrics_path: '/metrics'
    scrape_interval: 10s

  # Go Coffee AI Services
  - job_name: 'go-coffee-ai'
    static_configs:
      - targets:
        - 'ai-search:8099'
        - 'ai-service:8100'
        - 'ai-arbitrage-service:8101'
        - 'ai-order-service:8102'
        - 'llm-orchestrator:8106'
        - 'llm-orchestrator-simple:8107'
        - 'mcp-ai-integration:8109'
    metrics_path: '/metrics'
    scrape_interval: 15s

  # Go Coffee Infrastructure
  - job_name: 'go-coffee-infrastructure'
    static_configs:
      - targets:
        - 'redis:6379'
        - 'postgres:5432'
        - 'redis-mcp-server:8108'
    metrics_path: '/metrics'
    scrape_interval: 30s

  # Crypto Wallet Services
  - job_name: 'crypto-wallet'
    static_configs:
      - targets:
        - 'wallet-service:8081'
        - 'transaction-service:8082'
        - 'smart-contract-service:8083'
        - 'security-service:8084'
        - 'defi-service:8085'
        - 'fintech-api:8086'
        - 'telegram-bot:8087'
    metrics_path: '/metrics'
    scrape_interval: 10s

  # Web UI Services
  - job_name: 'web-ui'
    static_configs:
      - targets:
        - 'mcp-server:3001'
        - 'web-ui-backend:8090'
        - 'web-ui-frontend:3000'
    metrics_path: '/metrics'
    scrape_interval: 15s

  # Node Exporter (System Metrics)
  - job_name: 'node-exporter'
    static_configs:
      - targets: ['node-exporter:9100']

  # cAdvisor (Container Metrics)
  - job_name: 'cadvisor'
    static_configs:
      - targets: ['cadvisor:8080']
    metrics_path: '/metrics'
EOF

    print_status "Prometheus configuration generated"
}

# Generate alerting rules
generate_alerting_rules() {
    if [[ "$SETUP_ALERTS" != "true" ]]; then
        return 0
    fi
    
    print_header "🚨 Generating Alerting Rules"
    
    cat > monitoring/prometheus/rules/go-coffee-alerts.yml <<EOF
groups:
  - name: go-coffee-services
    rules:
      - alert: ServiceDown
        expr: up == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Service {{ \$labels.instance }} is down"
          description: "{{ \$labels.job }} service on {{ \$labels.instance }} has been down for more than 1 minute."

      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate on {{ \$labels.instance }}"
          description: "Error rate is {{ \$value }} errors per second on {{ \$labels.instance }}"

      - alert: HighLatency
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 0.5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High latency on {{ \$labels.instance }}"
          description: "95th percentile latency is {{ \$value }}s on {{ \$labels.instance }}"

      - alert: HighMemoryUsage
        expr: (node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / node_memory_MemTotal_bytes > 0.9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage"
          description: "Memory usage is above 90% on {{ \$labels.instance }}"

      - alert: HighCPUUsage
        expr: 100 - (avg by(instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High CPU usage"
          description: "CPU usage is above 80% on {{ \$labels.instance }}"

  - name: go-coffee-business
    rules:
      - alert: OrderProcessingDelay
        expr: increase(order_processing_duration_seconds[5m]) > 300
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Order processing delays detected"
          description: "Order processing is taking longer than expected"

      - alert: PaymentFailures
        expr: rate(payment_failures_total[5m]) > 0.05
        for: 3m
        labels:
          severity: critical
        annotations:
          summary: "High payment failure rate"
          description: "Payment failure rate is {{ \$value }} per second"

      - alert: CryptoWalletIssues
        expr: rate(crypto_wallet_errors_total[5m]) > 0.01
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Crypto wallet errors detected"
          description: "Crypto wallet error rate is {{ \$value }} per second"
EOF

    print_status "Alerting rules generated"
}

# Generate Grafana datasources
generate_grafana_datasources() {
    print_header "📈 Generating Grafana Datasources"

    cat > monitoring/grafana/provisioning/datasources/datasources.yml <<EOF
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true

  - name: Jaeger
    type: jaeger
    access: proxy
    url: http://jaeger:16686
    editable: true

  - name: Loki
    type: loki
    access: proxy
    url: http://loki:3100
    editable: true
EOF

    print_status "Grafana datasources configured"
}

# Generate custom Go Coffee dashboards
generate_custom_dashboards() {
    if [[ "$INCLUDE_CUSTOM" != "true" ]]; then
        return 0
    fi

    print_header "📊 Generating Custom Go Coffee Dashboards"

    # Dashboard provisioning config
    cat > monitoring/grafana/provisioning/dashboards/dashboards.yml <<EOF
apiVersion: 1

providers:
  - name: 'go-coffee-dashboards'
    orgId: 1
    folder: 'Go Coffee'
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /var/lib/grafana/dashboards
EOF

    # Go Coffee Overview Dashboard
    cat > monitoring/grafana/dashboards/go-coffee-overview.json <<'EOF'
{
  "dashboard": {
    "id": null,
    "title": "Go Coffee - Platform Overview",
    "tags": ["go-coffee", "overview"],
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "Service Health Status",
        "type": "stat",
        "targets": [
          {
            "expr": "up{job=~\"go-coffee.*\"}",
            "legendFormat": "{{instance}}"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "color": {
              "mode": "thresholds"
            },
            "thresholds": {
              "steps": [
                {"color": "red", "value": 0},
                {"color": "green", "value": 1}
              ]
            }
          }
        },
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0}
      },
      {
        "id": 2,
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total[5m])) by (job)",
            "legendFormat": "{{job}}"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0}
      },
      {
        "id": 3,
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{status=~\"5..\"}[5m])) by (job)",
            "legendFormat": "{{job}} errors"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 8}
      },
      {
        "id": 4,
        "title": "Response Time (95th percentile)",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, job))",
            "legendFormat": "{{job}} p95"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 8}
      }
    ],
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "refresh": "5s"
  }
}
EOF

    print_status "Custom dashboards generated"
}

# Generate Docker Compose for monitoring stack
generate_monitoring_compose() {
    print_header "🐳 Generating Monitoring Docker Compose"

    cat > monitoring/docker-compose.monitoring.yml <<EOF
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:latest
    container_name: go-coffee-prometheus
    ports:
      - "${PROMETHEUS_PORT}:9090"
    volumes:
      - ./prometheus/config:/etc/prometheus
      - ./prometheus/rules:/etc/prometheus/rules
      - ./prometheus/data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=200h'
      - '--web.enable-lifecycle'
      - '--web.enable-admin-api'
    networks:
      - go-coffee-monitoring

  grafana:
    image: grafana/grafana:latest
    container_name: go-coffee-grafana
    ports:
      - "${GRAFANA_PORT}:3000"
    volumes:
      - ./grafana/data:/var/lib/grafana
      - ./grafana/provisioning:/etc/grafana/provisioning
      - ./grafana/dashboards:/var/lib/grafana/dashboards
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
      - GF_INSTALL_PLUGINS=grafana-piechart-panel
    networks:
      - go-coffee-monitoring

  jaeger:
    image: jaegertracing/all-in-one:latest
    container_name: go-coffee-jaeger
    ports:
      - "${JAEGER_PORT}:16686"
      - "14268:14268"
      - "6831:6831/udp"
    environment:
      - COLLECTOR_OTLP_ENABLED=true
    volumes:
      - ./jaeger/data:/tmp
    networks:
      - go-coffee-monitoring

  loki:
    image: grafana/loki:latest
    container_name: go-coffee-loki
    ports:
      - "${LOKI_PORT}:3100"
    volumes:
      - ./loki/config:/etc/loki
      - ./loki/data:/tmp/loki
    command: -config.file=/etc/loki/local-config.yaml
    networks:
      - go-coffee-monitoring

  alertmanager:
    image: prom/alertmanager:latest
    container_name: go-coffee-alertmanager
    ports:
      - "${ALERTMANAGER_PORT}:9093"
    volumes:
      - ./alertmanager/config:/etc/alertmanager
      - ./alertmanager/data:/alertmanager
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
      - '--storage.path=/alertmanager'
      - '--web.external-url=http://localhost:9093'
    networks:
      - go-coffee-monitoring

  node-exporter:
    image: prom/node-exporter:latest
    container_name: go-coffee-node-exporter
    ports:
      - "9100:9100"
    volumes:
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - /:/rootfs:ro
    command:
      - '--path.procfs=/host/proc'
      - '--path.rootfs=/rootfs'
      - '--path.sysfs=/host/sys'
      - '--collector.filesystem.mount-points-exclude=^/(sys|proc|dev|host|etc)($$|/)'
    networks:
      - go-coffee-monitoring

  cadvisor:
    image: gcr.io/cadvisor/cadvisor:latest
    container_name: go-coffee-cadvisor
    ports:
      - "8080:8080"
    volumes:
      - /:/rootfs:ro
      - /var/run:/var/run:rw
      - /sys:/sys:ro
      - /var/lib/docker/:/var/lib/docker:ro
    networks:
      - go-coffee-monitoring

networks:
  go-coffee-monitoring:
    driver: bridge
    external: false

volumes:
  prometheus-data:
  grafana-data:
  jaeger-data:
  loki-data:
  alertmanager-data:
EOF

    print_status "Monitoring Docker Compose generated"
}

# Start monitoring stack
start_monitoring_stack() {
    print_header "🚀 Starting Monitoring Stack"

    case "$SETUP_MODE" in
        "docker")
            print_progress "Starting monitoring stack with Docker Compose..."
            cd monitoring
            docker-compose -f docker-compose.monitoring.yml up -d
            cd ..
            ;;
        "kubernetes")
            print_progress "Deploying monitoring stack to Kubernetes..."
            # TODO: Add Kubernetes deployment
            print_warning "Kubernetes deployment not yet implemented"
            ;;
        "local")
            print_progress "Starting monitoring services locally..."
            # TODO: Add local startup
            print_warning "Local deployment not yet implemented"
            ;;
    esac

    # Wait for services to start
    print_progress "Waiting for monitoring services to be ready..."
    sleep 30

    # Check service health
    check_monitoring_health
}

# Check monitoring stack health
check_monitoring_health() {
    print_header "🏥 Checking Monitoring Stack Health"

    local healthy_count=0
    local total_count=${#MONITORING_SERVICES[@]}

    for service_name in "${!MONITORING_SERVICES[@]}"; do
        local port=${MONITORING_SERVICES[$service_name]}

        if curl -s --max-time 10 "http://localhost:$port" >/dev/null 2>&1; then
            print_status "$service_name (port $port) - HEALTHY"
            ((healthy_count++))
        else
            print_error "$service_name (port $port) - UNHEALTHY"
        fi
    done

    print_info "Monitoring Health: $healthy_count/$total_count services healthy"

    if [[ $healthy_count -eq $total_count ]]; then
        print_success "🎉 All monitoring services are healthy!"
        show_monitoring_endpoints
    else
        print_warning "⚠️  Some monitoring services need attention"
    fi
}

# Show monitoring endpoints
show_monitoring_endpoints() {
    print_header "🌐 Monitoring Endpoints"
    print_info "Prometheus:    http://localhost:$PROMETHEUS_PORT"
    print_info "Grafana:       http://localhost:$GRAFANA_PORT (admin/admin)"
    print_info "Jaeger:        http://localhost:$JAEGER_PORT"
    print_info "AlertManager:  http://localhost:$ALERTMANAGER_PORT"
    print_info "Loki:          http://localhost:$LOKI_PORT"
    print_info "Node Exporter: http://localhost:9100"
    print_info "cAdvisor:      http://localhost:8080"

    if [[ "$INCLUDE_CUSTOM" == "true" ]]; then
        print_info ""
        print_info "📊 Custom Dashboards:"
        print_info "Go Coffee Overview: http://localhost:$GRAFANA_PORT/d/go-coffee-overview"
    fi
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    local start_time=$(date +%s)

    # Parse arguments
    parse_observability_args "$@"

    # Check dependencies
    local deps=("docker" "docker-compose" "curl")
    check_dependencies "${deps[@]}" || exit 1

    print_info "Observability Setup Configuration:"
    print_info "  Environment: $ENVIRONMENT"
    print_info "  Setup Mode: $SETUP_MODE"
    print_info "  Custom Dashboards: $INCLUDE_CUSTOM"
    print_info "  Alerting: $SETUP_ALERTS"
    print_info "  Persistent Storage: $SETUP_STORAGE"

    # Create monitoring structure
    create_monitoring_structure

    # Generate configurations
    generate_prometheus_config
    generate_alerting_rules
    generate_grafana_datasources
    generate_custom_dashboards
    generate_monitoring_compose

    # Start monitoring stack
    start_monitoring_stack

    # Calculate setup time
    local end_time=$(date +%s)
    local total_time=$((end_time - start_time))

    print_success "🎉 Observability setup completed in ${total_time}s"

    print_header "🎯 Next Steps"
    print_info "1. Access Grafana at http://localhost:$GRAFANA_PORT"
    print_info "2. Import additional dashboards if needed"
    print_info "3. Configure alert notifications"
    print_info "4. Set up log aggregation"
    print_info "5. Configure service discovery"
}

# Run main function with all arguments
main "$@"
