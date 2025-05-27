#!/bin/bash

# Web3 Coffee Telegram Bot Startup Script
# This script sets up and starts the Telegram bot with all dependencies

set -e

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

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check environment variables
check_env_vars() {
    print_status "Checking environment variables..."
    
    if [ -z "$TELEGRAM_BOT_TOKEN" ]; then
        print_error "TELEGRAM_BOT_TOKEN is not set!"
        print_status "Please set your Telegram bot token:"
        print_status "export TELEGRAM_BOT_TOKEN='your_bot_token_here'"
        exit 1
    fi
    
    if [ -z "$GEMINI_API_KEY" ]; then
        print_warning "GEMINI_API_KEY is not set. Gemini AI will be disabled."
        print_status "To enable Gemini AI, set:"
        print_status "export GEMINI_API_KEY='your_gemini_api_key'"
    fi
    
    print_success "Environment variables checked"
}

# Function to check dependencies
check_dependencies() {
    print_status "Checking dependencies..."
    
    if ! command_exists docker; then
        print_error "Docker is not installed!"
        print_status "Please install Docker: https://docs.docker.com/get-docker/"
        exit 1
    fi
    
    if ! command_exists docker-compose; then
        print_error "Docker Compose is not installed!"
        print_status "Please install Docker Compose: https://docs.docker.com/compose/install/"
        exit 1
    fi
    
    if ! command_exists go; then
        print_warning "Go is not installed. Using Docker for build."
    fi
    
    print_success "Dependencies checked"
}

# Function to setup Ollama
setup_ollama() {
    print_status "Setting up Ollama..."
    
    # Wait for Ollama to be ready
    print_status "Waiting for Ollama to start..."
    timeout=60
    while [ $timeout -gt 0 ]; do
        if curl -s http://localhost:11434/api/tags >/dev/null 2>&1; then
            break
        fi
        sleep 2
        timeout=$((timeout - 2))
    done
    
    if [ $timeout -le 0 ]; then
        print_warning "Ollama is not responding. AI features may be limited."
        return
    fi
    
    # Pull required models
    print_status "Pulling Ollama models..."
    docker exec web3-coffee-ollama ollama pull llama3.1 || print_warning "Failed to pull llama3.1 model"
    
    print_success "Ollama setup completed"
}

# Function to create .env file
create_env_file() {
    print_status "Creating .env file..."
    
    cat > .env << EOF
# Telegram Bot Configuration
TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}
TELEGRAM_WEBHOOK_URL=${TELEGRAM_WEBHOOK_URL:-}

# AI Configuration
GEMINI_API_KEY=${GEMINI_API_KEY:-}

# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_NAME=web3_coffee
DB_USER=web3_user
DB_PASSWORD=web3_password

# Redis Configuration
REDIS_HOST=redis
REDIS_PORT=6379

# Ollama Configuration
OLLAMA_HOST=ollama
OLLAMA_PORT=11434

# Application Configuration
APP_ENV=development
LOG_LEVEL=debug
EOF
    
    print_success ".env file created"
}

# Function to start services
start_services() {
    print_status "Starting Web3 Coffee Telegram Bot..."
    
    # Navigate to deployment directory
    cd deployments/telegram-bot
    
    # Create .env file
    create_env_file
    
    # Start services with docker-compose
    print_status "Starting infrastructure services..."
    docker-compose up -d redis postgres ollama prometheus grafana
    
    # Wait for services to be ready
    print_status "Waiting for services to be ready..."
    sleep 10
    
    # Setup Ollama
    setup_ollama
    
    # Start the Telegram bot
    print_status "Starting Telegram bot..."
    docker-compose up -d telegram-bot
    
    print_success "All services started successfully!"
}

# Function to show status
show_status() {
    print_status "Service Status:"
    docker-compose ps
    
    echo ""
    print_status "Access URLs:"
    echo "  🤖 Telegram Bot: Running (check your Telegram bot)"
    echo "  📊 Grafana: http://localhost:3000 (admin/admin)"
    echo "  📈 Prometheus: http://localhost:9090"
    echo "  🔍 Redis: localhost:6379"
    echo "  🗄️  PostgreSQL: localhost:5432"
    echo "  🧠 Ollama: http://localhost:11434"
}

# Function to show logs
show_logs() {
    print_status "Showing Telegram bot logs..."
    docker-compose logs -f telegram-bot
}

# Function to stop services
stop_services() {
    print_status "Stopping Web3 Coffee Telegram Bot..."
    cd deployments/telegram-bot
    docker-compose down
    print_success "Services stopped"
}

# Function to clean up
cleanup() {
    print_status "Cleaning up Web3 Coffee Telegram Bot..."
    cd deployments/telegram-bot
    docker-compose down -v
    docker system prune -f
    print_success "Cleanup completed"
}

# Function to show help
show_help() {
    echo "Web3 Coffee Telegram Bot Management Script"
    echo ""
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  start     Start the Telegram bot and all dependencies"
    echo "  stop      Stop all services"
    echo "  restart   Restart all services"
    echo "  status    Show service status"
    echo "  logs      Show Telegram bot logs"
    echo "  cleanup   Stop services and remove volumes"
    echo "  help      Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  TELEGRAM_BOT_TOKEN  Required: Your Telegram bot token"
    echo "  GEMINI_API_KEY      Optional: Google Gemini API key"
    echo "  TELEGRAM_WEBHOOK_URL Optional: Webhook URL for production"
    echo ""
    echo "Examples:"
    echo "  export TELEGRAM_BOT_TOKEN='1234567890:ABC...'"
    echo "  export GEMINI_API_KEY='AIzaSy...'"
    echo "  $0 start"
}

# Main script logic
case "${1:-start}" in
    start)
        check_dependencies
        check_env_vars
        start_services
        show_status
        print_success "Telegram bot is running! Check your Telegram app."
        print_status "Use '$0 logs' to see bot logs"
        print_status "Use '$0 stop' to stop the bot"
        ;;
    stop)
        stop_services
        ;;
    restart)
        stop_services
        sleep 2
        check_dependencies
        check_env_vars
        start_services
        show_status
        ;;
    status)
        cd deployments/telegram-bot
        show_status
        ;;
    logs)
        cd deployments/telegram-bot
        show_logs
        ;;
    cleanup)
        cleanup
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        print_error "Unknown command: $1"
        show_help
        exit 1
        ;;
esac
