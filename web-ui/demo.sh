#!/bin/bash

# Go Coffee Epic UI - Demo Script
# Demonstration script to showcase all Epic UI capabilities

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# Emojis
ROCKET="🚀"
COFFEE="☕"
CHART="📊"
ROBOT="🤖"
SEARCH="🔍"
MONEY="💰"
CHECK="✅"
CROSS="❌"
INFO="ℹ️"
WARNING="⚠️"

print_header() {
    echo -e "${PURPLE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${WHITE}${ROCKET} Go Coffee Epic UI - Demonstration${NC}"
    echo -e "${PURPLE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}Revolutionary Web3 interface for coffee ecosystem${NC}"
    echo -e "${PURPLE}═══════════════════════════════════════════════════════════════${NC}"
    echo ""
}

print_section() {
    echo -e "${BLUE}▶ $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

print_step() {
    echo -e "${GREEN}${CHECK} $1${NC}"
}

print_info() {
    echo -e "${CYAN}${INFO} $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}${WARNING} $1${NC}"
}

print_error() {
    echo -e "${RED}${CROSS} $1${NC}"
}

check_dependencies() {
    print_section "Checking Dependencies"

    # Check Docker
    if command -v docker &> /dev/null; then
        print_step "Docker installed"
    else
        print_error "Docker not found. Please install Docker"
        exit 1
    fi

    # Check Docker Compose
    if command -v docker-compose &> /dev/null; then
        print_step "Docker Compose installed"
    else
        print_error "Docker Compose not found"
        exit 1
    fi

    # Check Node.js
    if command -v node &> /dev/null; then
        NODE_VERSION=$(node --version)
        print_step "Node.js installed: $NODE_VERSION"
    else
        print_warning "Node.js not found (needed for development)"
    fi

    # Check Go
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version)
        print_step "Go installed: $GO_VERSION"
    else
        print_warning "Go not found (needed for development)"
    fi

    echo ""
}

show_features() {
    print_section "Epic UI Features"

    echo -e "${COFFEE} ${WHITE}Coffee Orders:${NC}"
    echo -e "   • Real-time order management"
    echo -e "   • Inventory and warehouse"
    echo -e "   • Location map"
    echo ""

    echo -e "${MONEY} ${WHITE}DeFi Portfolio:${NC}"
    echo -e "   • Cryptocurrency balances"
    echo -e "   • Automatic trading strategies"
    echo -e "   • P&L analytics"
    echo ""

    echo -e "${ROBOT} ${WHITE}AI Agents:${NC}"
    echo -e "   • 9 specialized agents"
    echo -e "   • Real-time monitoring"
    echo -e "   • Process automation"
    echo ""

    echo -e "${SEARCH} ${WHITE}Bright Data Analytics:${NC}"
    echo -e "   • Competitor web scraping"
    echo -e "   • Market data"
    echo -e "   • Industry news"
    echo ""

    echo -e "${CHART} ${WHITE}Analytics:${NC}"
    echo -e "   • Interactive charts"
    echo -e "   • Reports and metrics"
    echo -e "   • Data export"
    echo ""
}

demo_bright_data() {
    print_section "Bright Data MCP Demonstration"

    print_info "Starting web scraping demonstration..."

    if [ -f "test-bright-data-mcp.go" ]; then
        go run test-bright-data-mcp.go
    else
        print_warning "File test-bright-data-mcp.go not found"
        echo -e "${CYAN}Bright Data MCP Simulation:${NC}"
        echo "🔍 Scraping Starbucks menu..."
        echo "✅ Found: Grande Latte $5.45 (+$0.15)"
        echo "🔍 Scraping Coffee Futures..."
        echo "✅ Arabica: $1.85/lb (+2.3%)"
        echo "🔍 Searching coffee news..."
        echo "✅ Found 15 articles about sustainable practices"
    fi

    echo ""
}

start_services() {
    print_section "Starting Services"

    print_info "Creating .env file..."
    if [ ! -f ".env" ]; then
        cp .env.example .env
        print_step ".env file created"
    else
        print_step ".env file already exists"
    fi

    print_info "Starting Docker containers..."
    docker-compose -f docker-compose.ui.yml up -d --build

    print_step "Services started"

    print_info "Waiting for services to be ready..."
    sleep 15

    # Check if services are running
    if docker-compose -f docker-compose.ui.yml ps | grep -q "Up"; then
        print_step "All services are running"
    else
        print_error "Some services failed to start"
        docker-compose -f docker-compose.ui.yml logs
        exit 1
    fi

    echo ""
}

show_endpoints() {
    print_section "Available Endpoints"

    echo -e "${WHITE}🌐 Frontend:${NC}"
    echo -e "   ${CYAN}http://localhost:3000${NC} - Main interface"
    echo ""

    echo -e "${WHITE}🔗 Backend API:${NC}"
    echo -e "   ${CYAN}http://localhost:8090${NC} - API server"
    echo -e "   ${CYAN}http://localhost:8090/health${NC} - Health check"
    echo ""

    echo -e "${WHITE}🔌 WebSocket:${NC}"
    echo -e "   ${CYAN}ws://localhost:8090/ws/realtime${NC} - Real-time data"
    echo ""

    echo -e "${WHITE}📊 API endpoints:${NC}"
    echo -e "   ${CYAN}/api/v1/dashboard/metrics${NC} - Dashboard metrics"
    echo -e "   ${CYAN}/api/v1/coffee/orders${NC} - Coffee orders"
    echo -e "   ${CYAN}/api/v1/defi/portfolio${NC} - DeFi portfolio"
    echo -e "   ${CYAN}/api/v1/agents/status${NC} - AI agents status"
    echo -e "   ${CYAN}/api/v1/scraping/data${NC} - Bright Data analytics"
    echo ""
}

test_api() {
    print_section "Testing API"

    print_info "Checking health endpoint..."
    if curl -s http://localhost:8090/health > /dev/null; then
        print_step "Backend API is working"
    else
        print_error "Backend API not responding"
    fi

    print_info "Checking dashboard metrics..."
    if curl -s http://localhost:8090/api/v1/dashboard/metrics > /dev/null; then
        print_step "Dashboard API is working"
    else
        print_error "Dashboard API not responding"
    fi

    print_info "Checking frontend..."
    if curl -s http://localhost:3000 > /dev/null; then
        print_step "Frontend is working"
    else
        print_error "Frontend not responding"
    fi

    echo ""
}

show_usage_examples() {
    print_section "Приклади використання"
    
    echo -e "${WHITE}📱 Мобільний досвід:${NC}"
    echo -e "   • Responsive дизайн для всіх пристроїв"
    echo -e "   • PWA підтримка"
    echo -e "   • Офлайн режим"
    echo ""
    
    echo -e "${WHITE}⚡ Real-time оновлення:${NC}"
    echo -e "   • WebSocket підключення"
    echo -e "   • Миттєві нотифікації"
    echo -e "   • Live графіки"
    echo ""
    
    echo -e "${WHITE}🎨 Кастомізація:${NC}"
    echo -e "   • Темна/світла теми"
    echo -e "   • Drag & drop дашборд"
    echo -e "   • Персональні налаштування"
    echo ""
}

show_next_steps() {
    print_section "Наступні кроки"
    
    echo -e "${WHITE}🔧 Розробка:${NC}"
    echo -e "   ${CYAN}make dev${NC} - Запуск в режимі розробки"
    echo -e "   ${CYAN}make test${NC} - Запуск тестів"
    echo -e "   ${CYAN}make build${NC} - Збірка проекту"
    echo ""
    
    echo -e "${WHITE}🐳 Docker:${NC}"
    echo -e "   ${CYAN}make start${NC} - Запуск всіх сервісів"
    echo -e "   ${CYAN}make stop${NC} - Зупинка сервісів"
    echo -e "   ${CYAN}make logs${NC} - Перегляд логів"
    echo ""
    
    echo -e "${WHITE}🔍 Моніторинг:${NC}"
    echo -e "   ${CYAN}make health${NC} - Перевірка здоров'я сервісів"
    echo -e "   ${CYAN}make status${NC} - Статус контейнерів"
    echo ""
    
    echo -e "${WHITE}🧹 Очищення:${NC}"
    echo -e "   ${CYAN}make clean${NC} - Очищення Docker ресурсів"
    echo ""
}

main() {
    clear
    print_header
    
    # Check if help is requested
    if [[ "$1" == "--help" || "$1" == "-h" ]]; then
        echo -e "${WHITE}Usage:${NC}"
        echo -e "  $0 [options]"
        echo ""
        echo -e "${WHITE}Options:${NC}"
        echo -e "  --help, -h     Show this help"
        echo -e "  --quick, -q    Quick start without demonstration"
        echo -e "  --stop         Stop services"
        echo ""
        exit 0
    fi

    # Stop services if requested
    if [[ "$1" == "--stop" ]]; then
        print_section "Stopping Services"
        docker-compose -f docker-compose.ui.yml down
        print_step "Services stopped"
        exit 0
    fi
    
    # Quick start mode
    if [[ "$1" == "--quick" || "$1" == "-q" ]]; then
        check_dependencies
        start_services
        show_endpoints
        exit 0
    fi
    
    # Full demo
    check_dependencies
    show_features
    demo_bright_data
    start_services
    show_endpoints
    test_api
    show_usage_examples
    show_next_steps
    
    echo -e "${GREEN}${CHECK} Demonstration completed!${NC}"
    echo -e "${CYAN}${INFO} Open http://localhost:3000 to view Epic UI${NC}"
    echo ""
    echo -e "${YELLOW}To stop services run: $0 --stop${NC}"
}

# Run main function with all arguments
main "$@"
