#!/bin/bash

# Generate comprehensive Kubernetes manifests for Go Coffee services
# This script creates production-ready K8s manifests with proper resource management,
# health checks, secrets handling, and environment-specific configurations

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
OUTPUT_DIR="$SCRIPT_DIR/manifests"

# Service definitions with their configurations
declare -A SERVICES=(
    # Core Business Services
    ["api-gateway"]="8080:2:4Gi:2Gi:gateway"
    ["auth-service"]="8091:2:2Gi:1Gi:auth"
    ["order-service"]="8094:3:3Gi:2Gi:order"
    ["payment-service"]="8093:2:2Gi:1Gi:payment"
    ["kitchen-service"]="8095:2:2Gi:1Gi:kitchen"
    ["user-gateway"]="8096:2:2Gi:1Gi:user"
    ["security-gateway"]="8097:2:3Gi:2Gi:security"
    ["communication-hub"]="8098:2:2Gi:1Gi:communication"
    
    # AI & ML Services
    ["ai-service"]="8100:1:8Gi:4Gi:ai"
    ["ai-search"]="8099:2:4Gi:2Gi:search"
    ["ai-arbitrage-service"]="8101:1:4Gi:2Gi:arbitrage"
    ["ai-order-service"]="8102:2:4Gi:2Gi:ai-order"
    ["llm-orchestrator"]="8106:1:16Gi:8Gi:llm"
    ["llm-orchestrator-simple"]="8107:1:8Gi:4Gi:llm-simple"
    ["mcp-ai-integration"]="8109:2:4Gi:2Gi:mcp-ai"
    
    # Infrastructure Services
    ["market-data-service"]="8103:2:3Gi:2Gi:market"
    ["defi-service"]="8104:2:3Gi:2Gi:defi"
    ["bright-data-hub-service"]="8105:2:4Gi:2Gi:bright-data"
    ["redis-mcp-server"]="8108:2:2Gi:1Gi:redis-mcp"
    ["web-ui-backend"]="3000:2:2Gi:1Gi:web-ui"
)

# GPU-enabled services
GPU_SERVICES=("ai-service" "llm-orchestrator" "llm-orchestrator-simple")

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

# Create namespace manifest
create_namespace() {
    local env="$1"
    local output_file="$OUTPUT_DIR/$env/00-namespace.yaml"
    
    mkdir -p "$(dirname "$output_file")"
    
    cat > "$output_file" <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: go-coffee-$env
  labels:
    name: go-coffee-$env
    environment: $env
    app.kubernetes.io/name: go-coffee
    app.kubernetes.io/component: namespace
    app.kubernetes.io/part-of: go-coffee-platform
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
  annotations:
    description: "Go Coffee $env environment namespace"
---
apiVersion: v1
kind: ResourceQuota
metadata:
  name: go-coffee-$env-quota
  namespace: go-coffee-$env
spec:
  hard:
    requests.cpu: "20"
    requests.memory: 40Gi
    limits.cpu: "40"
    limits.memory: 80Gi
    persistentvolumeclaims: "10"
    services: "30"
    secrets: "20"
    configmaps: "20"
---
apiVersion: v1
kind: LimitRange
metadata:
  name: go-coffee-$env-limits
  namespace: go-coffee-$env
spec:
  limits:
  - default:
      cpu: "1"
      memory: "2Gi"
    defaultRequest:
      cpu: "100m"
      memory: "256Mi"
    type: Container
  - default:
      storage: "10Gi"
    type: PersistentVolumeClaim
EOF
    
    log "Created namespace manifest for $env environment"
}

# Create secrets manifest
create_secrets() {
    local env="$1"
    local output_file="$OUTPUT_DIR/$env/01-secrets.yaml"
    
    cat > "$output_file" <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: go-coffee-database
  namespace: go-coffee-$env
  labels:
    app.kubernetes.io/name: go-coffee
    app.kubernetes.io/component: database
    environment: $env
type: Opaque
stringData:
  database-url: "postgres://go_coffee_user:CHANGE_ME@postgres:5432/go_coffee?sslmode=disable"
  postgres-user: "go_coffee_user"
  postgres-password: "CHANGE_ME"
  postgres-database: "go_coffee"
---
apiVersion: v1
kind: Secret
metadata:
  name: go-coffee-redis
  namespace: go-coffee-$env
  labels:
    app.kubernetes.io/name: go-coffee
    app.kubernetes.io/component: redis
    environment: $env
type: Opaque
stringData:
  redis-url: "redis://redis:6379"
  redis-password: ""
---
apiVersion: v1
kind: Secret
metadata:
  name: go-coffee-jwt
  namespace: go-coffee-$env
  labels:
    app.kubernetes.io/name: go-coffee
    app.kubernetes.io/component: auth
    environment: $env
type: Opaque
stringData:
  jwt-secret: "CHANGE_ME_TO_SECURE_JWT_SECRET"
  jwt-issuer: "go-coffee-$env"
---
apiVersion: v1
kind: Secret
metadata:
  name: go-coffee-api-keys
  namespace: go-coffee-$env
  labels:
    app.kubernetes.io/name: go-coffee
    app.kubernetes.io/component: api
    environment: $env
type: Opaque
stringData:
  openai-api-key: "CHANGE_ME"
  bright-data-api-key: "CHANGE_ME"
  market-data-api-key: "CHANGE_ME"
EOF
    
    log "Created secrets manifest for $env environment"
}

# Create ConfigMap
create_configmap() {
    local env="$1"
    local output_file="$OUTPUT_DIR/$env/02-configmap.yaml"
    
    cat > "$output_file" <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: go-coffee-config
  namespace: go-coffee-$env
  labels:
    app.kubernetes.io/name: go-coffee
    app.kubernetes.io/component: config
    environment: $env
data:
  environment: "$env"
  log-level: "info"
  metrics-enabled: "true"
  tracing-enabled: "true"
  
  # Database configuration
  database-max-connections: "25"
  database-max-idle: "5"
  database-connection-timeout: "30s"
  
  # Redis configuration
  redis-max-connections: "100"
  redis-connection-timeout: "5s"
  
  # HTTP configuration
  http-timeout: "30s"
  http-max-idle-connections: "100"
  
  # AI configuration
  ai-model-timeout: "60s"
  ai-max-tokens: "4096"
  
  # Monitoring configuration
  prometheus-metrics-path: "/metrics"
  health-check-path: "/health"
  readiness-check-path: "/ready"
EOF
    
    log "Created ConfigMap for $env environment"
}

# Generate service deployment manifest
generate_service_deployment() {
    local service="$1"
    local env="$2"
    local config="$3"
    
    IFS=':' read -r port replicas memory_limit memory_request service_type <<< "$config"
    
    local output_file="$OUTPUT_DIR/$env/10-$service-deployment.yaml"
    local is_gpu_service=false
    
    # Check if this is a GPU service
    for gpu_service in "${GPU_SERVICES[@]}"; do
        if [[ "$service" == "$gpu_service" ]]; then
            is_gpu_service=true
            break
        fi
    done
    
    cat > "$output_file" <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $service
  namespace: go-coffee-$env
  labels:
    app: $service
    app.kubernetes.io/name: $service
    app.kubernetes.io/component: $service_type
    app.kubernetes.io/part-of: go-coffee-platform
    app.kubernetes.io/version: "1.0.0"
    environment: $env
spec:
  replicas: $replicas
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 1
  selector:
    matchLabels:
      app: $service
      environment: $env
  template:
    metadata:
      labels:
        app: $service
        app.kubernetes.io/name: $service
        app.kubernetes.io/component: $service_type
        environment: $env
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "$port"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: go-coffee-$service
      securityContext:
        runAsNonRoot: true
        runAsUser: 1001
        runAsGroup: 1001
        fsGroup: 1001
        seccompProfile:
          type: RuntimeDefault
      containers:
      - name: $service
        image: ghcr.io/dimajoyti/go-coffee/$service:latest
        imagePullPolicy: Always
        ports:
        - containerPort: $port
          name: http
          protocol: TCP
        env:
        - name: PORT
          value: "$port"
        - name: ENVIRONMENT
          valueFrom:
            configMapKeyRef:
              name: go-coffee-config
              key: environment
        - name: LOG_LEVEL
          valueFrom:
            configMapKeyRef:
              name: go-coffee-config
              key: log-level
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: go-coffee-database
              key: database-url
        - name: REDIS_URL
          valueFrom:
            secretKeyRef:
              name: go-coffee-redis
              key: redis-url
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: go-coffee-jwt
              key: jwt-secret
        - name: METRICS_ENABLED
          valueFrom:
            configMapKeyRef:
              name: go-coffee-config
              key: metrics-enabled
        - name: TRACING_ENABLED
          valueFrom:
            configMapKeyRef:
              name: go-coffee-config
              key: tracing-enabled
        resources:
          requests:
            cpu: "100m"
            memory: $memory_request
          limits:
            cpu: "1000m"
            memory: $memory_limit
EOF

    # Add GPU resources for GPU services
    if [[ "$is_gpu_service" == true ]]; then
        cat >> "$output_file" <<EOF
            nvidia.com/gpu: 1
EOF
    fi

    cat >> "$output_file" <<EOF
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          runAsNonRoot: true
          runAsUser: 1001
          runAsGroup: 1001
          capabilities:
            drop:
            - ALL
        livenessProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
          successThreshold: 1
        readinessProbe:
          httpGet:
            path: /ready
            port: http
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
          successThreshold: 1
        startupProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 30
          successThreshold: 1
        volumeMounts:
        - name: tmp
          mountPath: /tmp
        - name: cache
          mountPath: /app/cache
      volumes:
      - name: tmp
        emptyDir: {}
      - name: cache
        emptyDir: {}
      nodeSelector:
EOF

    # Add node selector for GPU services
    if [[ "$is_gpu_service" == true ]]; then
        cat >> "$output_file" <<EOF
        accelerator: nvidia-tesla-k80
EOF
    else
        cat >> "$output_file" <<EOF
        kubernetes.io/arch: amd64
EOF
    fi

    cat >> "$output_file" <<EOF
      tolerations:
      - key: "node.kubernetes.io/not-ready"
        operator: "Exists"
        effect: "NoExecute"
        tolerationSeconds: 300
      - key: "node.kubernetes.io/unreachable"
        operator: "Exists"
        effect: "NoExecute"
        tolerationSeconds: 300
EOF

    # Add GPU tolerations for GPU services
    if [[ "$is_gpu_service" == true ]]; then
        cat >> "$output_file" <<EOF
      - key: "nvidia.com/gpu"
        operator: "Exists"
        effect: "NoSchedule"
EOF
    fi

    cat >> "$output_file" <<EOF
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - $service
              topologyKey: kubernetes.io/hostname
EOF
    
    log "Generated deployment manifest for $service"
}

# Generate service manifest
generate_service_manifest() {
    local service="$1"
    local env="$2"
    local config="$3"
    
    IFS=':' read -r port replicas memory_limit memory_request service_type <<< "$config"
    
    local output_file="$OUTPUT_DIR/$env/20-$service-service.yaml"
    
    cat > "$output_file" <<EOF
apiVersion: v1
kind: Service
metadata:
  name: $service
  namespace: go-coffee-$env
  labels:
    app: $service
    app.kubernetes.io/name: $service
    app.kubernetes.io/component: $service_type
    environment: $env
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "$port"
    prometheus.io/path: "/metrics"
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: $port
    protocol: TCP
    name: http
  selector:
    app: $service
    environment: $env
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: go-coffee-$service
  namespace: go-coffee-$env
  labels:
    app: $service
    app.kubernetes.io/name: $service
    app.kubernetes.io/component: $service_type
    environment: $env
automountServiceAccountToken: false
EOF
    
    log "Generated service manifest for $service"
}

# Main execution
main() {
    local environments=("staging" "production")
    
    log "Starting Kubernetes manifest generation for Go Coffee services..."
    
    # Create output directory
    mkdir -p "$OUTPUT_DIR"
    
    # Generate manifests for each environment
    for env in "${environments[@]}"; do
        info "Generating manifests for $env environment..."
        
        # Create base manifests
        create_namespace "$env"
        create_secrets "$env"
        create_configmap "$env"
        
        # Generate service manifests
        for service in "${!SERVICES[@]}"; do
            generate_service_deployment "$service" "$env" "${SERVICES[$service]}"
            generate_service_manifest "$service" "$env" "${SERVICES[$service]}"
        done
        
        log "Completed manifest generation for $env environment"
    done
    
    log "Kubernetes manifest generation completed successfully!"
    log "Generated manifests for ${#SERVICES[@]} services across ${#environments[@]} environments"
    
    echo ""
    echo -e "${BLUE}Generated manifests in:${NC}"
    echo "  $OUTPUT_DIR/staging/"
    echo "  $OUTPUT_DIR/production/"
    echo ""
    echo -e "${BLUE}Next steps:${NC}"
    echo "1. Review generated manifests"
    echo "2. Update secrets with actual values"
    echo "3. Apply to cluster: kubectl apply -f $OUTPUT_DIR/staging/"
    echo "4. Verify deployments: kubectl get pods -n go-coffee-staging"
}

# Handle command line arguments
case "${1:-generate}" in
    "generate")
        main
        ;;
    "clean")
        log "Cleaning generated manifests..."
        rm -rf "$OUTPUT_DIR"
        log "Cleanup completed"
        ;;
    *)
        echo "Usage: $0 [generate|clean]"
        exit 1
        ;;
esac
