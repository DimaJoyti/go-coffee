#!/bin/bash

echo "🚀 Starting Go Coffee Production Environment"
echo "============================================="

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

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker first."
    exit 1
fi

print_status "Docker is running"

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    print_error "docker-compose is not installed. Please install docker-compose first."
    exit 1
fi

print_status "docker-compose is available"

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    print_info "Creating .env file with default values..."
    cat > .env << EOF
# Go Coffee Environment Variables

# AI Services
GEMINI_API_KEY=your-gemini-api-key-here
OLLAMA_BASE_URL=http://ollama:11434

# Security
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Database
POSTGRES_DB=go_coffee
POSTGRES_USER=go_coffee_user
POSTGRES_PASSWORD=go_coffee_password

# Redis
REDIS_URL=redis://redis:6379

# Logging
LOG_LEVEL=info

# Environment
ENVIRONMENT=production
EOF
    print_warning "Please update .env file with your actual API keys and secrets!"
fi

# Build all services first
print_info "Building all Go Coffee services..."

echo "Building binaries..."
if ! ./build_all.sh; then
    print_error "Failed to build services. Please check the build output."
    exit 1
fi

print_status "All services built successfully"

# Stop any existing containers
print_info "Stopping existing containers..."
docker-compose down --remove-orphans

# Pull latest images
print_info "Pulling latest Docker images..."
docker-compose pull

# Start infrastructure services first
print_info "Starting infrastructure services..."
docker-compose up -d redis postgres prometheus grafana jaeger

# Wait for infrastructure to be ready
print_info "Waiting for infrastructure services to be ready..."
sleep 30

# Check if Redis is ready
print_info "Checking Redis connection..."
if ! docker-compose exec -T redis redis-cli ping > /dev/null 2>&1; then
    print_error "Redis is not ready. Please check the logs."
    docker-compose logs redis
    exit 1
fi
print_status "Redis is ready"

# Check if PostgreSQL is ready
print_info "Checking PostgreSQL connection..."
if ! docker-compose exec -T postgres pg_isready -U go_coffee_user -d go_coffee > /dev/null 2>&1; then
    print_error "PostgreSQL is not ready. Please check the logs."
    docker-compose logs postgres
    exit 1
fi
print_status "PostgreSQL is ready"

# Start Go Coffee services
print_info "Starting Go Coffee microservices..."
docker-compose up -d ai-search auth-service kitchen-service communication-hub redis-mcp-server

# Wait for services to start
print_info "Waiting for services to start..."
sleep 20

# Start user gateway (depends on other services)
print_info "Starting User Gateway..."
docker-compose up -d user-gateway

# Start load balancer
print_info "Starting Nginx load balancer..."
docker-compose up -d nginx

# Optional: Start Ollama for local AI
read -p "Do you want to start Ollama for local AI? (y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    print_info "Starting Ollama..."
    docker-compose up -d ollama
    print_info "Ollama will take a few minutes to download models..."
fi

# Wait for all services to be ready
print_info "Waiting for all services to be ready..."
sleep 30

# Health checks
print_info "Performing health checks..."

services=(
    "http://localhost:8092/api/v1/ai-search/health:AI Search"
    "http://localhost:8080/health:Auth Service"
    "http://localhost:8081/health:User Gateway"
    "http://localhost:8093/health:Redis MCP Server"
    "http://localhost/health:Load Balancer"
)

all_healthy=true

for service in "${services[@]}"; do
    IFS=':' read -r url name <<< "$service"
    if curl -f -s "$url" > /dev/null 2>&1; then
        print_status "$name is healthy"
    else
        print_error "$name is not responding"
        all_healthy=false
    fi
done

if [ "$all_healthy" = true ]; then
    print_status "All services are healthy!"
else
    print_warning "Some services are not responding. Check the logs for details."
fi

# Display service information
echo ""
echo "🎯 **GO COFFEE SERVICES STATUS**"
echo "================================"

echo ""
echo "📊 **Monitoring & Observability:**"
echo "  • Prometheus: http://localhost:9090"
echo "  • Grafana: http://localhost:3000 (admin/admin)"
echo "  • Jaeger: http://localhost:16686"

echo ""
echo "🔧 **Go Coffee Services:**"
echo "  • Load Balancer: http://localhost"
echo "  • User Gateway: http://localhost:8081"
echo "  • AI Search: http://localhost:8092"
echo "  • Auth Service: http://localhost:8080"
echo "  • Redis MCP Server: http://localhost:8093"

echo ""
echo "🗄️ **Infrastructure:**"
echo "  • Redis: localhost:6379"
echo "  • PostgreSQL: localhost:5432"
echo "  • Ollama (if started): http://localhost:11434"

echo ""
echo "📋 **API Endpoints:**"
echo "  • Health Check: curl http://localhost/health"
echo "  • User Registration: curl -X POST http://localhost/api/v1/auth/register"
echo "  • AI Search: curl -X POST http://localhost/api/v1/ai-search/semantic"
echo "  • Create Order: curl -X POST http://localhost/api/v1/orders"

echo ""
echo "🔍 **Useful Commands:**"
echo "  • View logs: docker-compose logs [service-name]"
echo "  • Stop all: docker-compose down"
echo "  • Restart service: docker-compose restart [service-name]"
echo "  • Scale service: docker-compose up -d --scale [service-name]=3"

echo ""
if [ "$all_healthy" = true ]; then
    print_status "🎉 Go Coffee Production Environment is READY! ☕🚀"
    echo ""
    print_info "You can now:"
    print_info "1. Access the API at http://localhost"
    print_info "2. Monitor services at http://localhost:9090 (Prometheus)"
    print_info "3. View dashboards at http://localhost:3000 (Grafana)"
    print_info "4. Trace requests at http://localhost:16686 (Jaeger)"
else
    print_warning "Some services need attention. Check logs with: docker-compose logs"
fi

echo ""
print_info "To stop all services: docker-compose down"
print_info "To view real-time logs: docker-compose logs -f"
