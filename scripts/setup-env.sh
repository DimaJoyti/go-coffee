#!/bin/bash

# =============================================================================
# Go Coffee Environment Setup Script
# =============================================================================
# This script helps set up environment files for Go Coffee project
# =============================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Emojis
ROCKET="🚀"
CHECK="✅"
WARNING="⚠️"
ERROR="❌"
INFO="ℹ️"
GEAR="⚙️"
LOCK="🔒"
FIRE="🔥"

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Change to project root
cd "$PROJECT_ROOT"

echo -e "${CYAN}${ROCKET} Go Coffee Environment Setup${NC}"
echo -e "${CYAN}=================================${NC}"
echo ""

# Function to print colored output
print_status() {
    echo -e "${GREEN}${CHECK} $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}${WARNING} $1${NC}"
}

print_error() {
    echo -e "${RED}${ERROR} $1${NC}"
}

print_info() {
    echo -e "${BLUE}${INFO} $1${NC}"
}

print_gear() {
    echo -e "${PURPLE}${GEAR} $1${NC}"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to generate secure random string
generate_secret() {
    local length=${1:-32}
    if command_exists openssl; then
        openssl rand -hex $((length/2))
    elif command_exists head && [ -f /dev/urandom ]; then
        head -c $((length/2)) /dev/urandom | xxd -p | tr -d '\n'
    else
        # Fallback to date-based random
        date +%s | sha256sum | head -c $length
    fi
}

# Function to prompt for input with default
prompt_with_default() {
    local prompt="$1"
    local default="$2"
    local result
    
    read -p "$prompt [$default]: " result
    echo "${result:-$default}"
}

# Function to prompt for yes/no
prompt_yes_no() {
    local prompt="$1"
    local default="$2"
    local result
    
    while true; do
        read -p "$prompt (y/n) [$default]: " result
        result="${result:-$default}"
        case $result in
            [Yy]* ) echo "yes"; break;;
            [Nn]* ) echo "no"; break;;
            * ) echo "Please answer yes or no.";;
        esac
    done
}

# Function to setup basic environment
setup_basic_env() {
    print_gear "Setting up basic environment files..."
    
    # Create .env from .env.example if it doesn't exist
    if [ ! -f .env ]; then
        if [ -f .env.example ]; then
            cp .env.example .env
            print_status "Created .env from .env.example"
        else
            print_error ".env.example not found!"
            exit 1
        fi
    else
        print_info ".env already exists"
    fi
    
    # Create .env.local if it doesn't exist
    if [ ! -f .env.local ]; then
        cat > .env.local << EOF
# =============================================================================
# LOCAL ENVIRONMENT OVERRIDES
# =============================================================================
# Add your local-specific environment variables here
# This file should not be committed to version control
# =============================================================================

# Local development overrides
DEBUG=true
LOG_LEVEL=debug

# Local database (if different from default)
# DATABASE_HOST=localhost
# DATABASE_PASSWORD=your-local-password

# Local Redis (if different from default)
# REDIS_HOST=localhost
# REDIS_PASSWORD=your-local-redis-password

# Local API keys for development
# OPENAI_API_KEY=your-local-openai-key
# GEMINI_API_KEY=your-local-gemini-key

EOF
        print_status "Created .env.local"
    else
        print_info ".env.local already exists"
    fi
}

# Function to generate secure secrets
generate_secrets() {
    print_gear "Generating secure secrets..."
    
    local jwt_secret=$(generate_secret 64)
    local api_key_secret=$(generate_secret 48)
    local webhook_secret=$(generate_secret 48)
    local encryption_key=$(generate_secret 32)
    
    echo ""
    print_info "Generated secure secrets (copy these to your .env file):"
    echo ""
    echo -e "${CYAN}JWT_SECRET=${jwt_secret}${NC}"
    echo -e "${CYAN}API_KEY_SECRET=${api_key_secret}${NC}"
    echo -e "${CYAN}WEBHOOK_SECRET=${webhook_secret}${NC}"
    echo -e "${CYAN}ENCRYPTION_KEY=${encryption_key}${NC}"
    echo ""
    
    # Ask if user wants to automatically update .env file
    if [ "$(prompt_yes_no "Do you want to automatically update .env file with these secrets?" "yes")" = "yes" ]; then
        # Update .env file with generated secrets
        sed -i.bak \
            -e "s/JWT_SECRET=.*/JWT_SECRET=${jwt_secret}/" \
            -e "s/API_KEY_SECRET=.*/API_KEY_SECRET=${api_key_secret}/" \
            -e "s/WEBHOOK_SECRET=.*/WEBHOOK_SECRET=${webhook_secret}/" \
            -e "s/ENCRYPTION_KEY=.*/ENCRYPTION_KEY=${encryption_key}/" \
            .env
        
        print_status "Updated .env file with secure secrets"
        print_info "Backup saved as .env.bak"
    fi
}

# Function to configure database
configure_database() {
    print_gear "Configuring database settings..."
    
    local db_host=$(prompt_with_default "Database host" "localhost")
    local db_port=$(prompt_with_default "Database port" "5432")
    local db_name=$(prompt_with_default "Database name" "go_coffee")
    local db_user=$(prompt_with_default "Database user" "postgres")
    local db_password=$(prompt_with_default "Database password" "postgres")
    
    # Update .env file
    sed -i.bak \
        -e "s/DATABASE_HOST=.*/DATABASE_HOST=${db_host}/" \
        -e "s/DATABASE_PORT=.*/DATABASE_PORT=${db_port}/" \
        -e "s/DATABASE_NAME=.*/DATABASE_NAME=${db_name}/" \
        -e "s/DATABASE_USER=.*/DATABASE_USER=${db_user}/" \
        -e "s/DATABASE_PASSWORD=.*/DATABASE_PASSWORD=${db_password}/" \
        .env
    
    print_status "Database configuration updated"
}

# Function to configure Redis
configure_redis() {
    print_gear "Configuring Redis settings..."
    
    local redis_host=$(prompt_with_default "Redis host" "localhost")
    local redis_port=$(prompt_with_default "Redis port" "6379")
    local redis_password=$(prompt_with_default "Redis password (leave empty for no auth)" "")
    
    # Update .env file
    sed -i.bak \
        -e "s/REDIS_HOST=.*/REDIS_HOST=${redis_host}/" \
        -e "s/REDIS_PORT=.*/REDIS_PORT=${redis_port}/" \
        -e "s/REDIS_PASSWORD=.*/REDIS_PASSWORD=${redis_password}/" \
        .env
    
    print_status "Redis configuration updated"
}

# Function to configure AI services
configure_ai() {
    print_gear "Configuring AI services..."
    
    echo ""
    print_info "AI service configuration (you can skip these and configure later)"
    echo ""
    
    local openai_key=$(prompt_with_default "OpenAI API Key (optional)" "")
    local gemini_key=$(prompt_with_default "Google Gemini API Key (optional)" "")
    local ollama_url=$(prompt_with_default "Ollama URL (for local AI)" "http://localhost:11434")
    
    if [ -n "$openai_key" ]; then
        sed -i.bak "s/OPENAI_API_KEY=.*/OPENAI_API_KEY=${openai_key}/" .env
        print_status "OpenAI API key configured"
    fi
    
    if [ -n "$gemini_key" ]; then
        sed -i.bak "s/GEMINI_API_KEY=.*/GEMINI_API_KEY=${gemini_key}/" .env
        print_status "Gemini API key configured"
    fi
    
    sed -i.bak "s|OLLAMA_URL=.*|OLLAMA_URL=${ollama_url}|" .env
    print_status "Ollama URL configured"
}

# Function to validate environment
validate_environment() {
    print_gear "Validating environment configuration..."
    
    if command_exists go; then
        if go run cmd/config-test/main.go validate; then
            print_status "Environment validation passed!"
        else
            print_warning "Environment validation failed. Please check your configuration."
        fi
    else
        print_warning "Go not found. Skipping validation."
    fi
}

# Function to show next steps
show_next_steps() {
    echo ""
    echo -e "${CYAN}${FIRE} Setup Complete!${NC}"
    echo -e "${CYAN}=================${NC}"
    echo ""
    print_info "Your environment has been configured. Here are the next steps:"
    echo ""
    echo "1. Review your .env file and make any necessary adjustments"
    echo "2. Start the required services (PostgreSQL, Redis, Kafka)"
    echo "3. Run the configuration test: make env-test"
    echo "4. Start the Go Coffee services: make run"
    echo ""
    print_info "Useful commands:"
    echo "  make env-test     - Test environment configuration"
    echo "  make env-validate - Validate environment settings"
    echo "  make env-show     - Show current configuration"
    echo "  make help         - Show all available commands"
    echo ""
    print_warning "Remember to:"
    echo "  - Never commit .env files with real secrets to version control"
    echo "  - Use different configurations for different environments"
    echo "  - Regularly rotate your secrets and API keys"
    echo ""
}

# Main setup flow
main() {
    echo "This script will help you set up environment files for Go Coffee."
    echo ""
    
    # Check if we're in the right directory
    if [ ! -f "go.mod" ] || ! grep -q "go-coffee" go.mod; then
        print_error "This doesn't appear to be the Go Coffee project directory."
        print_info "Please run this script from the project root."
        exit 1
    fi
    
    # Setup basic environment files
    setup_basic_env
    
    # Ask what to configure
    echo ""
    print_info "What would you like to configure?"
    echo ""
    
    if [ "$(prompt_yes_no "Generate secure secrets?" "yes")" = "yes" ]; then
        generate_secrets
    fi
    
    if [ "$(prompt_yes_no "Configure database settings?" "yes")" = "yes" ]; then
        configure_database
    fi
    
    if [ "$(prompt_yes_no "Configure Redis settings?" "yes")" = "yes" ]; then
        configure_redis
    fi
    
    if [ "$(prompt_yes_no "Configure AI services?" "no")" = "yes" ]; then
        configure_ai
    fi
    
    # Validate configuration
    if [ "$(prompt_yes_no "Validate environment configuration?" "yes")" = "yes" ]; then
        validate_environment
    fi
    
    # Show next steps
    show_next_steps
}

# Handle command line arguments
case "${1:-}" in
    "secrets")
        generate_secrets
        ;;
    "database")
        configure_database
        ;;
    "redis")
        configure_redis
        ;;
    "ai")
        configure_ai
        ;;
    "validate")
        validate_environment
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  secrets   - Generate secure secrets only"
        echo "  database  - Configure database settings only"
        echo "  redis     - Configure Redis settings only"
        echo "  ai        - Configure AI services only"
        echo "  validate  - Validate environment only"
        echo "  help      - Show this help"
        echo ""
        echo "Run without arguments for interactive setup."
        ;;
    "")
        main
        ;;
    *)
        print_error "Unknown command: $1"
        echo "Run '$0 help' for usage information."
        exit 1
        ;;
esac
