#!/bin/bash

# Redis 8 Visual Interface Startup Script
# This script starts the complete Redis 8 visual interface stack

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REDIS_PORT=6379
MCP_PORT=8080
WEB_PORT=3000
REDIS_INSIGHT_PORT=8001

echo -e "${BLUE}🚀 Starting Redis 8 Visual Interface Stack${NC}"
echo "=================================================="

# Function to check if port is available
check_port() {
    local port=$1
    local service=$2
    
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Port $port is already in use (required for $service)${NC}"
        echo -e "${YELLOW}   Please stop the service using this port or change the configuration${NC}"
        return 1
    fi
    return 0
}

# Function to wait for service to be ready
wait_for_service() {
    local url=$1
    local service=$2
    local max_attempts=30
    local attempt=1
    
    echo -e "${YELLOW}⏳ Waiting for $service to be ready...${NC}"
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "$url" >/dev/null 2>&1; then
            echo -e "${GREEN}✅ $service is ready!${NC}"
            return 0
        fi
        
        echo -n "."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo -e "${RED}❌ $service failed to start within expected time${NC}"
    return 1
}

# Check prerequisites
echo -e "${BLUE}🔍 Checking prerequisites...${NC}"

# Check Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed${NC}"
    exit 1
fi

# Check Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}❌ Docker Compose is not installed${NC}"
    exit 1
fi

# Check if Docker is running
if ! docker info &> /dev/null; then
    echo -e "${RED}❌ Docker is not running${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Prerequisites check passed${NC}"

# Check ports
echo -e "${BLUE}🔍 Checking port availability...${NC}"

check_port $REDIS_PORT "Redis" || exit 1
check_port $MCP_PORT "Redis MCP Server" || exit 1
check_port $WEB_PORT "Web UI" || exit 1
check_port $REDIS_INSIGHT_PORT "RedisInsight" || exit 1

echo -e "${GREEN}✅ All ports are available${NC}"

# Start the stack
echo -e "${BLUE}🐳 Starting Docker containers...${NC}"

# Pull latest images
echo -e "${YELLOW}📥 Pulling latest images...${NC}"
docker-compose -f docker-compose.redis8.yml pull

# Start services
echo -e "${YELLOW}🚀 Starting services...${NC}"
docker-compose -f docker-compose.redis8.yml up -d

# Wait for services to be ready
echo -e "${BLUE}⏳ Waiting for services to start...${NC}"

# Wait for Redis
wait_for_service "http://localhost:$REDIS_PORT" "Redis" || {
    echo -e "${RED}❌ Redis failed to start${NC}"
    docker-compose -f docker-compose.redis8.yml logs redis8
    exit 1
}

# Wait for MCP Server
wait_for_service "http://localhost:$MCP_PORT/api/v1/redis-mcp/health" "Redis MCP Server" || {
    echo -e "${RED}❌ Redis MCP Server failed to start${NC}"
    docker-compose -f docker-compose.redis8.yml logs redis-mcp-server
    exit 1
}

# Wait for Web UI
wait_for_service "http://localhost:$WEB_PORT" "Web UI" || {
    echo -e "${RED}❌ Web UI failed to start${NC}"
    docker-compose -f docker-compose.redis8.yml logs web-ui
    exit 1
}

# Success message
echo ""
echo -e "${GREEN}🎉 Redis 8 Visual Interface is now running!${NC}"
echo "=================================================="
echo -e "${BLUE}📊 Access Points:${NC}"
echo -e "  • ${GREEN}Web UI:${NC}          http://localhost:$WEB_PORT"
echo -e "  • ${GREEN}Redis MCP API:${NC}   http://localhost:$MCP_PORT"
echo -e "  • ${GREEN}RedisInsight:${NC}    http://localhost:$REDIS_INSIGHT_PORT"
echo -e "  • ${GREEN}Redis Direct:${NC}    localhost:$REDIS_PORT"
echo ""
echo -e "${BLUE}🛠️ Management Commands:${NC}"
echo -e "  • ${YELLOW}View logs:${NC}       docker-compose -f docker-compose.redis8.yml logs -f"
echo -e "  • ${YELLOW}Stop services:${NC}   docker-compose -f docker-compose.redis8.yml down"
echo -e "  • ${YELLOW}Restart:${NC}         docker-compose -f docker-compose.redis8.yml restart"
echo -e "  • ${YELLOW}Status:${NC}          docker-compose -f docker-compose.redis8.yml ps"
echo ""
echo -e "${BLUE}📖 Documentation:${NC}"
echo -e "  • ${GREEN}Redis Visual Interface:${NC} docs/REDIS_VISUAL_INTERFACE.md"
echo -e "  • ${GREEN}API Documentation:${NC}     http://localhost:$MCP_PORT/api/docs"
echo ""
echo -e "${GREEN}🚀 Happy Redis exploring!${NC}"

# Optional: Open browser
if command -v xdg-open &> /dev/null; then
    read -p "Would you like to open the Web UI in your browser? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        xdg-open "http://localhost:$WEB_PORT"
    fi
elif command -v open &> /dev/null; then
    read -p "Would you like to open the Web UI in your browser? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        open "http://localhost:$WEB_PORT"
    fi
fi
