#!/bin/bash

# Go Coffee Database & Cache Optimization Deployment Script
# This script helps deploy the advanced database and caching optimizations

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="go-coffee"
ENVIRONMENT="${ENVIRONMENT:-development}"
DRY_RUN="${DRY_RUN:-false}"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if kubectl is installed
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed. Please install kubectl first."
        exit 1
    fi
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go first."
        exit 1
    fi
    
    # Check if the namespace exists
    if ! kubectl get namespace $NAMESPACE &> /dev/null; then
        print_warning "Namespace $NAMESPACE does not exist. Creating it..."
        kubectl create namespace $NAMESPACE
    fi
    
    print_success "Prerequisites check completed"
}

# Function to build optimized components
build_optimizations() {
    print_status "Building optimization components..."
    
    # Build the optimization service
    print_status "Building optimization service..."
    go build -o bin/optimization-service ./pkg/optimization/
    
    # Build example integration
    print_status "Building example integration..."
    go build -o bin/example-integration ./examples/optimization_integration.go
    
    print_success "Build completed"
}

# Function to run tests
run_tests() {
    print_status "Running optimization tests..."
    
    # Test database optimization
    print_status "Testing database optimization..."
    go test -v ./pkg/database/... -timeout=30s
    
    # Test cache optimization
    print_status "Testing cache optimization..."
    go test -v ./pkg/cache/... -timeout=30s
    
    # Test memory optimization
    print_status "Testing memory optimization..."
    go test -v ./pkg/performance/... -timeout=30s
    
    print_success "All tests passed"
}

# Function to deploy database optimizations
deploy_database_optimizations() {
    print_status "Deploying database optimizations..."
    
    # Create database configuration
    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: database-optimization-config
  namespace: $NAMESPACE
data:
  config.yaml: |
    database:
      optimization:
        enabled: true
        connection_pool:
          max_connections: 50
          min_connections: 10
          max_connection_lifetime: "5m"
          max_connection_idle_time: "2m"
          health_check_period: "30s"
        query:
          default_timeout: "30s"
          slow_query_threshold: "1s"
        read_replicas:
          enabled: true
          failover: true
        monitoring:
          enabled: true
          metrics_interval: "30s"
EOF
    
    print_success "Database optimization configuration deployed"
}

# Function to deploy cache optimizations
deploy_cache_optimizations() {
    print_status "Deploying cache optimizations..."
    
    # Create cache configuration
    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: cache-optimization-config
  namespace: $NAMESPACE
data:
  config.yaml: |
    cache:
      optimization:
        enabled: true
        connection:
          pool_size: 50
          min_idle_connections: 10
          read_timeout: "3s"
          write_timeout: "3s"
        compression:
          enabled: true
          min_size: 1024
          algorithm: "gzip"
        warming:
          enabled: true
          interval: "5m"
          strategies:
            - name: "menu"
              enabled: true
              ttl: "1h"
            - name: "popular_items"
              enabled: true
              ttl: "30m"
        monitoring:
          enabled: true
          metrics_interval: "30s"
EOF
    
    print_success "Cache optimization configuration deployed"
}

# Function to deploy Redis cluster (if enabled)
deploy_redis_cluster() {
    if [ "$ENVIRONMENT" = "production" ]; then
        print_status "Deploying Redis cluster for production..."
        
        cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis-cluster
  namespace: $NAMESPACE
spec:
  serviceName: redis-cluster
  replicas: 3
  selector:
    matchLabels:
      app: redis-cluster
  template:
    metadata:
      labels:
        app: redis-cluster
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        ports:
        - containerPort: 6379
          name: redis
        - containerPort: 16379
          name: cluster
        command:
        - redis-server
        - /etc/redis/redis.conf
        - --cluster-enabled
        - "yes"
        - --cluster-config-file
        - /data/nodes.conf
        - --cluster-node-timeout
        - "5000"
        - --appendonly
        - "yes"
        volumeMounts:
        - name: redis-data
          mountPath: /data
        - name: redis-config
          mountPath: /etc/redis
        resources:
          requests:
            cpu: 100m
            memory: 256Mi
          limits:
            cpu: 500m
            memory: 1Gi
      volumes:
      - name: redis-config
        configMap:
          name: redis-config
  volumeClaimTemplates:
  - metadata:
      name: redis-data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 10Gi
---
apiVersion: v1
kind: Service
metadata:
  name: redis-cluster
  namespace: $NAMESPACE
spec:
  clusterIP: None
  selector:
    app: redis-cluster
  ports:
  - name: redis
    port: 6379
    targetPort: 6379
  - name: cluster
    port: 16379
    targetPort: 16379
EOF
        
        print_success "Redis cluster deployed"
    else
        print_status "Skipping Redis cluster deployment for $ENVIRONMENT environment"
    fi
}

# Function to deploy monitoring
deploy_monitoring() {
    print_status "Deploying optimization monitoring..."
    
    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: optimization-metrics
  namespace: $NAMESPACE
  labels:
    app: optimization-service
spec:
  selector:
    app: optimization-service
  ports:
  - name: metrics
    port: 9090
    targetPort: 9090
  - name: health
    port: 8080
    targetPort: 8080
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: optimization-metrics
  namespace: $NAMESPACE
spec:
  selector:
    matchLabels:
      app: optimization-service
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
  - port: health
    interval: 60s
    path: /health
EOF
    
    print_success "Optimization monitoring deployed"
}

# Function to create optimization deployment
deploy_optimization_service() {
    print_status "Deploying optimization service..."
    
    cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: optimization-service
  namespace: $NAMESPACE
spec:
  replicas: 2
  selector:
    matchLabels:
      app: optimization-service
  template:
    metadata:
      labels:
        app: optimization-service
    spec:
      containers:
      - name: optimization-service
        image: go-coffee/optimization-service:latest
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: metrics
        env:
        - name: ENVIRONMENT
          value: "$ENVIRONMENT"
        - name: LOG_LEVEL
          value: "info"
        - name: METRICS_ENABLED
          value: "true"
        resources:
          requests:
            cpu: 200m
            memory: 256Mi
          limits:
            cpu: 1000m
            memory: 1Gi
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        volumeMounts:
        - name: config
          mountPath: /etc/optimization
          readOnly: true
      volumes:
      - name: config
        configMap:
          name: optimization-config
EOF
    
    print_success "Optimization service deployed"
}

# Function to verify deployment
verify_deployment() {
    print_status "Verifying deployment..."
    
    # Wait for pods to be ready
    print_status "Waiting for pods to be ready..."
    kubectl wait --for=condition=ready pod -l app=optimization-service -n $NAMESPACE --timeout=300s
    
    # Check pod status
    print_status "Checking pod status..."
    kubectl get pods -n $NAMESPACE -l app=optimization-service
    
    # Check service endpoints
    print_status "Checking service endpoints..."
    kubectl get endpoints -n $NAMESPACE optimization-metrics
    
    # Test health endpoint
    print_status "Testing health endpoint..."
    POD_NAME=$(kubectl get pods -n $NAMESPACE -l app=optimization-service -o jsonpath='{.items[0].metadata.name}')
    kubectl port-forward -n $NAMESPACE $POD_NAME 8080:8080 &
    PORT_FORWARD_PID=$!
    
    sleep 5
    
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        print_success "Health check passed"
    else
        print_warning "Health check failed"
    fi
    
    kill $PORT_FORWARD_PID 2>/dev/null || true
    
    print_success "Deployment verification completed"
}

# Function to run performance benchmark
run_benchmark() {
    print_status "Running performance benchmark..."
    
    # Create a simple benchmark job
    cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: optimization-benchmark
  namespace: $NAMESPACE
spec:
  template:
    spec:
      containers:
      - name: benchmark
        image: go-coffee/benchmark:latest
        command:
        - /bin/sh
        - -c
        - |
          echo "Running optimization benchmark..."
          
          # Test database performance
          echo "Testing database performance..."
          time curl -s http://optimization-service:8080/metrics | grep database
          
          # Test cache performance
          echo "Testing cache performance..."
          time curl -s http://optimization-service:8080/metrics | grep cache
          
          echo "Benchmark completed"
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 200m
            memory: 256Mi
      restartPolicy: Never
  backoffLimit: 3
EOF
    
    print_status "Benchmark job created. Check logs with: kubectl logs -n $NAMESPACE job/optimization-benchmark"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS] COMMAND"
    echo ""
    echo "Commands:"
    echo "  deploy-all      Deploy all optimizations"
    echo "  deploy-db       Deploy database optimizations only"
    echo "  deploy-cache    Deploy cache optimizations only"
    echo "  test            Run optimization tests"
    echo "  benchmark       Run performance benchmark"
    echo "  verify          Verify deployment"
    echo "  cleanup         Remove all optimization deployments"
    echo ""
    echo "Options:"
    echo "  --environment   Set environment (development|staging|production)"
    echo "  --namespace     Set Kubernetes namespace (default: go-coffee)"
    echo "  --dry-run       Show what would be deployed without actually deploying"
    echo "  --help          Show this help message"
}

# Function to cleanup deployments
cleanup() {
    print_status "Cleaning up optimization deployments..."
    
    kubectl delete deployment optimization-service -n $NAMESPACE --ignore-not-found=true
    kubectl delete service optimization-metrics -n $NAMESPACE --ignore-not-found=true
    kubectl delete configmap database-optimization-config -n $NAMESPACE --ignore-not-found=true
    kubectl delete configmap cache-optimization-config -n $NAMESPACE --ignore-not-found=true
    kubectl delete job optimization-benchmark -n $NAMESPACE --ignore-not-found=true
    
    print_success "Cleanup completed"
}

# Main execution
main() {
    case "${1:-}" in
        "deploy-all")
            check_prerequisites
            build_optimizations
            deploy_database_optimizations
            deploy_cache_optimizations
            deploy_redis_cluster
            deploy_optimization_service
            deploy_monitoring
            verify_deployment
            ;;
        "deploy-db")
            check_prerequisites
            deploy_database_optimizations
            ;;
        "deploy-cache")
            check_prerequisites
            deploy_cache_optimizations
            deploy_redis_cluster
            ;;
        "test")
            run_tests
            ;;
        "benchmark")
            run_benchmark
            ;;
        "verify")
            verify_deployment
            ;;
        "cleanup")
            cleanup
            ;;
        "help"|"--help"|"-h")
            show_usage
            ;;
        *)
            print_error "Unknown command: ${1:-}"
            show_usage
            exit 1
            ;;
    esac
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        --namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN="true"
            shift
            ;;
        --help|-h)
            show_usage
            exit 0
            ;;
        *)
            break
            ;;
    esac
done

# Execute main function
main "$@"
