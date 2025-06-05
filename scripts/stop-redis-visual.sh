#!/bin/bash

# Redis 8 Visual Interface Stop Script
# This script stops the complete Redis 8 visual interface stack

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🛑 Stopping Redis 8 Visual Interface Stack${NC}"
echo "=================================================="

# Function to check if containers are running
check_containers() {
    local running_containers=$(docker-compose -f docker-compose.redis8.yml ps -q)
    if [ -z "$running_containers" ]; then
        echo -e "${YELLOW}ℹ️  No containers are currently running${NC}"
        return 1
    fi
    return 0
}

# Check if Docker Compose file exists
if [ ! -f "docker-compose.redis8.yml" ]; then
    echo -e "${RED}❌ docker-compose.redis8.yml not found${NC}"
    echo -e "${YELLOW}   Please run this script from the project root directory${NC}"
    exit 1
fi

# Check if containers are running
if ! check_containers; then
    echo -e "${GREEN}✅ All services are already stopped${NC}"
    exit 0
fi

# Show current status
echo -e "${BLUE}📊 Current container status:${NC}"
docker-compose -f docker-compose.redis8.yml ps

echo ""
read -p "Are you sure you want to stop all Redis Visual Interface services? (y/N): " -n 1 -r
echo

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}⏸️  Operation cancelled${NC}"
    exit 0
fi

# Stop services
echo -e "${YELLOW}🛑 Stopping services...${NC}"
docker-compose -f docker-compose.redis8.yml down

# Check if we should remove volumes
echo ""
read -p "Would you like to remove Redis data volumes? (This will delete all Redis data) (y/N): " -n 1 -r
echo

if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}🗑️  Removing volumes...${NC}"
    docker-compose -f docker-compose.redis8.yml down -v
    echo -e "${GREEN}✅ Volumes removed${NC}"
fi

# Check if we should remove images
echo ""
read -p "Would you like to remove Docker images? (This will free up disk space) (y/N): " -n 1 -r
echo

if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}🗑️  Removing images...${NC}"
    docker-compose -f docker-compose.redis8.yml down --rmi all
    echo -e "${GREEN}✅ Images removed${NC}"
fi

# Final status check
echo ""
echo -e "${BLUE}📊 Final status:${NC}"
if check_containers; then
    echo -e "${YELLOW}⚠️  Some containers are still running${NC}"
    docker-compose -f docker-compose.redis8.yml ps
else
    echo -e "${GREEN}✅ All Redis Visual Interface services have been stopped${NC}"
fi

echo ""
echo -e "${GREEN}🎉 Redis 8 Visual Interface stack stopped successfully!${NC}"
echo "=================================================="
echo -e "${BLUE}🔄 To start again:${NC}"
echo -e "  • ${GREEN}Quick start:${NC}     ./scripts/start-redis-visual.sh"
echo -e "  • ${GREEN}Manual start:${NC}    docker-compose -f docker-compose.redis8.yml up -d"
echo ""
echo -e "${BLUE}🧹 Cleanup commands:${NC}"
echo -e "  • ${YELLOW}Remove all containers:${NC} docker-compose -f docker-compose.redis8.yml down"
echo -e "  • ${YELLOW}Remove with volumes:${NC}   docker-compose -f docker-compose.redis8.yml down -v"
echo -e "  • ${YELLOW}Remove with images:${NC}    docker-compose -f docker-compose.redis8.yml down --rmi all"
echo ""
echo -e "${GREEN}👋 Goodbye!${NC}"
