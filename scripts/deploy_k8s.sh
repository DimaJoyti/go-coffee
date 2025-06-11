#!/bin/bash

echo "🚀 Deploying Go Coffee to Kubernetes"
echo "====================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    print_error "kubectl is not installed. Please install kubectl first."
    exit 1
fi

print_status "kubectl is available"

# Check if we can connect to Kubernetes cluster
if ! kubectl cluster-info &> /dev/null; then
    print_error "Cannot connect to Kubernetes cluster. Please check your kubeconfig."
    exit 1
fi

print_status "Connected to Kubernetes cluster"

# Build Docker images
print_info "Building Docker images..."

services=("ai-search" "auth-service" "kitchen-service" "communication-hub" "user-gateway" "redis-mcp-server")

for service in "${services[@]}"; do
    print_info "Building $service..."
    if docker build -t "go-coffee/$service:latest" -f "deployments/$service/Dockerfile" .; then
        print_status "$service image built successfully"
    else
        print_error "Failed to build $service image"
        exit 1
    fi
done

# Tag images for registry (if using external registry)
# Uncomment and modify if pushing to external registry
# REGISTRY="your-registry.com"
# for service in "${services[@]}"; do
#     docker tag "go-coffee/$service:latest" "$REGISTRY/go-coffee/$service:latest"
#     docker push "$REGISTRY/go-coffee/$service:latest"
# done

# Apply Kubernetes manifests
print_info "Applying Kubernetes manifests..."

# Create namespace
print_info "Creating namespace..."
kubectl apply -f k8s/namespace.yaml
print_status "Namespace created"

# Apply secrets and configmaps
print_info "Applying secrets and configmaps..."
kubectl apply -f k8s/secrets.yaml
kubectl apply -f k8s/configmap.yaml
print_status "Secrets and configmaps applied"

# Deploy infrastructure
print_info "Deploying infrastructure..."
kubectl apply -f k8s/redis.yaml
kubectl apply -f k8s/postgres.yaml
print_status "Infrastructure deployed"

# Wait for infrastructure to be ready
print_info "Waiting for infrastructure to be ready..."
kubectl wait --for=condition=ready pod -l app=redis -n go-coffee --timeout=300s
kubectl wait --for=condition=ready pod -l app=postgres -n go-coffee --timeout=300s
print_status "Infrastructure is ready"

# Deploy Go Coffee services
print_info "Deploying Go Coffee services..."
kubectl apply -f k8s/go-coffee-services.yaml
print_status "Go Coffee services deployed"

# Deploy monitoring
print_info "Deploying monitoring stack..."
kubectl apply -f k8s/monitoring.yaml
print_status "Monitoring stack deployed"

# Wait for services to be ready
print_info "Waiting for services to be ready..."
kubectl wait --for=condition=ready pod -l app=ai-search -n go-coffee --timeout=300s
kubectl wait --for=condition=ready pod -l app=auth-service -n go-coffee --timeout=300s
kubectl wait --for=condition=ready pod -l app=kitchen-service -n go-coffee --timeout=300s
kubectl wait --for=condition=ready pod -l app=communication-hub -n go-coffee --timeout=300s
kubectl wait --for=condition=ready pod -l app=user-gateway -n go-coffee --timeout=300s

print_status "All services are ready!"

# Get service information
echo ""
echo "🎯 **KUBERNETES DEPLOYMENT STATUS**"
echo "==================================="

echo ""
echo "📊 **Pods Status:**"
kubectl get pods -n go-coffee

echo ""
echo "🔧 **Services:**"
kubectl get services -n go-coffee

echo ""
echo "🌐 **Ingress/LoadBalancer:**"
kubectl get ingress -n go-coffee 2>/dev/null || echo "No ingress configured"

# Get external IP for user-gateway
EXTERNAL_IP=$(kubectl get service user-gateway-service -n go-coffee -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null)
if [ -z "$EXTERNAL_IP" ]; then
    EXTERNAL_IP=$(kubectl get service user-gateway-service -n go-coffee -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null)
fi

if [ -n "$EXTERNAL_IP" ]; then
    echo ""
    echo "🌍 **External Access:**"
    echo "  • User Gateway: http://$EXTERNAL_IP:8081"
    echo "  • API Health: curl http://$EXTERNAL_IP:8081/health"
else
    echo ""
    echo "🔧 **Port Forward Commands (for local access):**"
    echo "  • User Gateway: kubectl port-forward -n go-coffee service/user-gateway-service 8081:8081"
    echo "  • AI Search: kubectl port-forward -n go-coffee service/ai-search-service 8092:8092"
    echo "  • Auth Service: kubectl port-forward -n go-coffee service/auth-service-service 8080:8080"
    echo "  • Prometheus: kubectl port-forward -n go-coffee service/prometheus-service 9090:9090"
    echo "  • Grafana: kubectl port-forward -n go-coffee service/grafana-service 3000:3000"
fi

echo ""
echo "📋 **Useful Commands:**"
echo "  • View logs: kubectl logs -f deployment/[service-name] -n go-coffee"
echo "  • Scale service: kubectl scale deployment [service-name] --replicas=3 -n go-coffee"
echo "  • Delete deployment: kubectl delete -f k8s/"
echo "  • Get all resources: kubectl get all -n go-coffee"

echo ""
echo "🔍 **Monitoring:**"
echo "  • Prometheus: kubectl port-forward -n go-coffee service/prometheus-service 9090:9090"
echo "  • Grafana: kubectl port-forward -n go-coffee service/grafana-service 3000:3000"
echo "  • Jaeger: kubectl port-forward -n go-coffee service/jaeger-service 16686:16686"

echo ""
print_status "🎉 Go Coffee successfully deployed to Kubernetes! ☕🚀"

echo ""
print_info "To access services locally, use the port-forward commands above."
print_info "To clean up: kubectl delete namespace go-coffee"
