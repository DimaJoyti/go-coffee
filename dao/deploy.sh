#!/bin/bash

# Developer DAO Platform Deployment Script
set -e

echo "🚀 Developer DAO Platform Deployment"
echo "===================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
    print_success "Docker is running"
}

# Check if Docker Compose is available
check_docker_compose() {
    if command -v docker-compose &> /dev/null; then
        DOCKER_COMPOSE_CMD="docker-compose"
    elif docker compose version &> /dev/null; then
        DOCKER_COMPOSE_CMD="docker compose"
    else
        print_error "Docker Compose is not available. Please install Docker Compose and try again."
        exit 1
    fi
    print_success "Docker Compose is available ($DOCKER_COMPOSE_CMD)"
}

# Create .env file if it doesn't exist
setup_env() {
    if [ ! -f .env ]; then
        print_warning ".env file not found. Creating from .env.example..."
        cp .env.example .env
        print_warning "Please edit .env file with your API keys before running the platform"
        print_warning "Required: OPENAI_API_KEY, GITHUB_TOKEN, WALLET_CONNECT_PROJECT_ID"
    else
        print_success ".env file exists"
    fi
}

# Build Go services
build_services() {
    print_status "Building Go services..."
    
    # Build bounty service
    print_status "Building bounty service..."
    go build -o bin/bounty-service ./cmd/bounty-service
    
    # Build marketplace service
    print_status "Building marketplace service..."
    go build -o bin/marketplace-service ./cmd/marketplace-service
    
    # Build metrics service
    print_status "Building metrics service..."
    go build -o bin/metrics-service ./cmd/metrics-service
    
    # Build DAO governance service
    print_status "Building DAO governance service..."
    go build -o bin/dao-governance-service ./cmd/dao-governance-service
    
    print_success "All Go services built successfully"
}

# Build Docker images
build_docker_images() {
    print_status "Building Docker images..."
    
    # Build bounty service image
    print_status "Building bounty service Docker image..."
    docker build -f cmd/bounty-service/Dockerfile -t developer-dao/bounty-service:latest .
    
    # Build marketplace service image
    print_status "Building marketplace service Docker image..."
    docker build -f cmd/marketplace-service/Dockerfile -t developer-dao/marketplace-service:latest .
    
    # Build metrics service image
    print_status "Building metrics service Docker image..."
    docker build -f cmd/metrics-service/Dockerfile -t developer-dao/metrics-service:latest .
    
    # Build DAO governance service image
    print_status "Building DAO governance service Docker image..."
    docker build -f cmd/dao-governance-service/Dockerfile -t developer-dao/dao-governance-service:latest .
    
    print_success "All Docker images built successfully"
}

# Deploy infrastructure services
deploy_infrastructure() {
    print_status "Deploying infrastructure services..."
    
    # Start PostgreSQL and Redis
    $DOCKER_COMPOSE_CMD up -d postgres redis
    
    # Wait for services to be ready
    print_status "Waiting for database to be ready..."
    sleep 10
    
    # Run database migrations
    print_status "Running database migrations..."
    make migrate-up || print_warning "Migration failed - database might already be initialized"
    
    print_success "Infrastructure services deployed"
}

# Deploy backend services
deploy_backend() {
    print_status "Deploying backend services..."
    
    # Start all backend services
    $DOCKER_COMPOSE_CMD up -d bounty-service marketplace-service metrics-service dao-governance-service
    
    # Wait for services to start
    print_status "Waiting for backend services to start..."
    sleep 15
    
    print_success "Backend services deployed"
}

# Deploy AI service
deploy_ai_service() {
    print_status "Deploying AI service..."
    
    # Check if AI service directory exists
    if [ -d "ai-service" ]; then
        cd ai-service
        $DOCKER_COMPOSE_CMD up -d
        cd ..
        print_success "AI service deployed"
    else
        print_warning "AI service directory not found. Skipping AI service deployment."
    fi
}

# Deploy frontend applications
deploy_frontend() {
    print_status "Deploying frontend applications..."
    
    # Check if web directory exists
    if [ -d "web" ]; then
        cd web
        
        # Install dependencies for shared components
        if [ -d "shared" ]; then
            print_status "Installing shared component dependencies..."
            cd shared
            npm install
            npm run build
            cd ..
        fi
        
        # Build and start frontend services
        $DOCKER_COMPOSE_CMD up -d dao-portal governance-ui
        cd ..
        print_success "Frontend applications deployed"
    else
        print_warning "Web directory not found. Skipping frontend deployment."
    fi
}

# Health check
health_check() {
    print_status "Performing health checks..."
    
    # Wait a bit for services to fully start
    sleep 10
    
    # Check backend services
    services=("bounty-service:8080" "marketplace-service:8081" "metrics-service:8082" "dao-governance-service:8084")
    
    for service in "${services[@]}"; do
        name=$(echo $service | cut -d: -f1)
        port=$(echo $service | cut -d: -f2)
        
        if curl -f -s http://localhost:$port/health > /dev/null; then
            print_success "$name is healthy"
        else
            print_warning "$name health check failed"
        fi
    done
    
    # Check frontend applications
    if curl -f -s http://localhost:3000 > /dev/null; then
        print_success "Developer Portal is accessible"
    else
        print_warning "Developer Portal health check failed"
    fi
    
    if curl -f -s http://localhost:3001 > /dev/null; then
        print_success "Governance UI is accessible"
    else
        print_warning "Governance UI health check failed"
    fi
}

# Show access URLs
show_urls() {
    echo ""
    echo "🎉 Developer DAO Platform Deployed Successfully!"
    echo "=============================================="
    echo ""
    echo "📱 Frontend Applications:"
    echo "   Developer Portal:    http://localhost:3000"
    echo "   Governance UI:       http://localhost:3001"
    echo ""
    echo "🔧 Backend Services:"
    echo "   Bounty Service:      http://localhost:8080"
    echo "   Marketplace Service: http://localhost:8081"
    echo "   Metrics Service:     http://localhost:8082"
    echo "   DAO Governance:      http://localhost:8084"
    echo ""
    echo "🤖 AI Service:"
    echo "   AI Service:          http://localhost:8083"
    echo ""
    echo "📊 Monitoring:"
    echo "   Prometheus:          http://localhost:9090"
    echo "   Grafana:            http://localhost:3003"
    echo ""
    echo "💾 Infrastructure:"
    echo "   PostgreSQL:          localhost:5432"
    echo "   Redis:              localhost:6379"
    echo "   Qdrant:             http://localhost:6333"
    echo ""
    echo "📚 API Documentation:"
    echo "   Swagger UI:         http://localhost:8080/swagger/index.html"
    echo ""
}

# Main deployment function
main() {
    echo "Starting deployment process..."
    
    # Preliminary checks
    check_docker
    check_docker_compose
    setup_env
    
    # Build phase
    print_status "Phase 1: Building services..."
    build_services
    build_docker_images
    
    # Deployment phase
    print_status "Phase 2: Deploying infrastructure..."
    deploy_infrastructure
    
    print_status "Phase 3: Deploying backend services..."
    deploy_backend
    
    print_status "Phase 4: Deploying AI service..."
    deploy_ai_service
    
    print_status "Phase 5: Deploying frontend applications..."
    deploy_frontend
    
    # Verification phase
    print_status "Phase 6: Health checks..."
    health_check
    
    # Success
    show_urls
}

# Handle script arguments
case "${1:-}" in
    "infrastructure")
        check_docker
        check_docker_compose
        setup_env
        deploy_infrastructure
        ;;
    "backend")
        build_services
        build_docker_images
        deploy_backend
        ;;
    "frontend")
        deploy_frontend
        ;;
    "ai")
        deploy_ai_service
        ;;
    "health")
        health_check
        ;;
    "clean")
        print_status "Cleaning up Docker containers and images..."
        $DOCKER_COMPOSE_CMD down -v
        docker system prune -f
        print_success "Cleanup completed"
        ;;
    *)
        main
        ;;
esac
