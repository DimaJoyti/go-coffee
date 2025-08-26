#!/bin/bash

# Go Coffee Web3 Payment Service Test Script
# This script tests the Web3 payment functionality

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
WEB3_PAYMENT_URL="http://localhost:8083"
WEB3_HEALTH_URL="http://localhost:8084/health"

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

print_header() {
    echo -e "${CYAN}================================${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}================================${NC}"
}

# Function to test service health
test_health() {
    print_status "Testing Web3 Payment service health..."
    
    if curl -f -s "$WEB3_HEALTH_URL" > /dev/null; then
        print_success "Web3 Payment service is healthy"
        return 0
    else
        print_error "Web3 Payment service health check failed"
        return 1
    fi
}

# Function to test payment creation
test_payment_creation() {
    print_status "Testing payment creation..."
    
    local response=$(curl -s -X POST "$WEB3_PAYMENT_URL/payment/create" \
        -H "Content-Type: application/json" \
        -d '{
            "order_id": "order_123",
            "customer_address": "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b1",
            "amount": "5.0",
            "currency": "USDC",
            "chain": "ethereum",
            "metadata": {
                "coffee_type": "Latte",
                "customer_name": "John Doe"
            }
        }')
    
    if echo "$response" | grep -q "payment"; then
        print_success "Payment created successfully"
        echo "Response: $response"
        
        # Extract payment ID for further tests
        PAYMENT_ID=$(echo "$response" | jq -r '.payment.id' 2>/dev/null || echo "")
        return 0
    else
        print_error "Payment creation failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test payment status
test_payment_status() {
    if [ -z "$PAYMENT_ID" ]; then
        print_warning "No payment ID available for status test"
        return 1
    fi
    
    print_status "Testing payment status retrieval..."
    
    local response=$(curl -s "$WEB3_PAYMENT_URL/payment/status/$PAYMENT_ID")
    
    if echo "$response" | grep -q "pending\|confirmed\|failed"; then
        print_success "Payment status retrieved successfully"
        echo "Status: $response"
        return 0
    else
        print_error "Payment status retrieval failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test wallet balance
test_wallet_balance() {
    print_status "Testing wallet balance retrieval..."
    
    local address="0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b1"
    local response=$(curl -s "$WEB3_PAYMENT_URL/wallet/balance/$address")
    
    if echo "$response" | grep -q "balances"; then
        print_success "Wallet balance retrieved successfully"
        echo "Balance: $response"
        return 0
    else
        print_error "Wallet balance retrieval failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test token price
test_token_price() {
    print_status "Testing token price retrieval..."
    
    local response=$(curl -s "$WEB3_PAYMENT_URL/token/price/ETH")
    
    if echo "$response" | grep -q "price_usd"; then
        print_success "Token price retrieved successfully"
        echo "Price: $response"
        return 0
    else
        print_error "Token price retrieval failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test token swap
test_token_swap() {
    print_status "Testing token swap..."
    
    local response=$(curl -s -X POST "$WEB3_PAYMENT_URL/token/swap" \
        -H "Content-Type: application/json" \
        -d '{
            "from_token": "ETH",
            "to_token": "USDC",
            "amount": "1.0",
            "chain": "ethereum",
            "slippage": 0.5
        }')
    
    if echo "$response" | grep -q "swap_id"; then
        print_success "Token swap initiated successfully"
        echo "Swap: $response"
        return 0
    else
        print_error "Token swap failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test yield opportunities
test_yield_opportunities() {
    print_status "Testing yield opportunities..."
    
    local response=$(curl -s "$WEB3_PAYMENT_URL/defi/yield")
    
    if echo "$response" | grep -q "opportunities"; then
        print_success "Yield opportunities retrieved successfully"
        echo "Opportunities: $response"
        return 0
    else
        print_error "Yield opportunities retrieval failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test staking
test_staking() {
    print_status "Testing token staking..."
    
    local response=$(curl -s -X POST "$WEB3_PAYMENT_URL/defi/stake" \
        -H "Content-Type: application/json" \
        -d '{
            "token": "COFFEE",
            "amount": "100.0",
            "protocol": "Coffee Staking",
            "duration_days": 30
        }')
    
    if echo "$response" | grep -q "stake_id"; then
        print_success "Token staking initiated successfully"
        echo "Stake: $response"
        return 0
    else
        print_error "Token staking failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test multiple chains
test_multiple_chains() {
    print_status "Testing payments on multiple chains..."
    
    local chains=("ethereum" "bsc" "polygon" "solana")
    local currencies=("ETH" "BNB" "MATIC" "SOL")
    local success_count=0
    
    for i in "${!chains[@]}"; do
        local chain="${chains[$i]}"
        local currency="${currencies[$i]}"
        
        local response=$(curl -s -X POST "$WEB3_PAYMENT_URL/payment/create" \
            -H "Content-Type: application/json" \
            -d "{
                \"order_id\": \"order_${chain}_$(date +%s)\",
                \"customer_address\": \"0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b1\",
                \"amount\": \"1.0\",
                \"currency\": \"$currency\",
                \"chain\": \"$chain\"
            }")
        
        if echo "$response" | grep -q "payment"; then
            ((success_count++))
            print_success "Payment created on $chain with $currency"
        else
            print_error "Failed to create payment on $chain"
        fi
        
        # Small delay between requests
        sleep 0.5
    done
    
    print_status "Successfully created payments on $success_count out of ${#chains[@]} chains"
}

# Function to test payment flow
test_payment_flow() {
    print_status "Testing complete payment flow..."
    
    # Create payment
    local create_response=$(curl -s -X POST "$WEB3_PAYMENT_URL/payment/create" \
        -H "Content-Type: application/json" \
        -d '{
            "order_id": "flow_test_order",
            "customer_address": "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b1",
            "amount": "10.0",
            "currency": "USDC",
            "chain": "ethereum"
        }')
    
    local payment_id=$(echo "$create_response" | jq -r '.payment.id' 2>/dev/null || echo "")
    
    if [ -z "$payment_id" ] || [ "$payment_id" = "null" ]; then
        print_error "Failed to create payment for flow test"
        return 1
    fi
    
    print_success "Payment created: $payment_id"
    
    # Check status
    sleep 1
    local status_response=$(curl -s "$WEB3_PAYMENT_URL/payment/status/$payment_id")
    print_status "Payment status: $status_response"
    
    # Simulate confirmation (in test mode)
    local confirm_response=$(curl -s -X POST "$WEB3_PAYMENT_URL/payment/confirm" \
        -H "Content-Type: application/json" \
        -d "{
            \"payment_id\": \"$payment_id\",
            \"transaction_hash\": \"0x$(openssl rand -hex 32)\"
        }")
    
    if echo "$confirm_response" | grep -q "confirmed"; then
        print_success "Payment confirmed successfully"
        return 0
    else
        print_error "Payment confirmation failed"
        echo "Response: $confirm_response"
        return 1
    fi
}

# Function to run comprehensive tests
run_comprehensive_tests() {
    print_header "Web3 Payment Service Test Suite"
    
    local failed_tests=0
    
    # Test 1: Health check
    print_header "Test 1: Health Check"
    test_health || ((failed_tests++))
    
    # Test 2: Payment creation
    print_header "Test 2: Payment Creation"
    test_payment_creation || ((failed_tests++))
    
    # Test 3: Payment status
    print_header "Test 3: Payment Status"
    test_payment_status || ((failed_tests++))
    
    # Test 4: Wallet balance
    print_header "Test 4: Wallet Balance"
    test_wallet_balance || ((failed_tests++))
    
    # Test 5: Token price
    print_header "Test 5: Token Price"
    test_token_price || ((failed_tests++))
    
    # Test 6: Token swap
    print_header "Test 6: Token Swap"
    test_token_swap || ((failed_tests++))
    
    # Test 7: Yield opportunities
    print_header "Test 7: Yield Opportunities"
    test_yield_opportunities || ((failed_tests++))
    
    # Test 8: Staking
    print_header "Test 8: Token Staking"
    test_staking || ((failed_tests++))
    
    # Test 9: Multiple chains
    print_header "Test 9: Multiple Chains"
    test_multiple_chains || ((failed_tests++))
    
    # Test 10: Payment flow
    print_header "Test 10: Complete Payment Flow"
    test_payment_flow || ((failed_tests++))
    
    # Summary
    print_header "Test Summary"
    if [ $failed_tests -eq 0 ]; then
        print_success "All tests passed! ✅"
        print_status "The Web3 Payment service is working correctly."
    else
        print_error "$failed_tests test(s) failed ❌"
        print_status "Please check the logs for more details."
    fi
    
    return $failed_tests
}

# Function to show service status
show_service_status() {
    print_header "Web3 Payment Service Status"
    
    echo -e "${CYAN}Service Health:${NC}"
    curl -s "$WEB3_HEALTH_URL" | jq '.' 2>/dev/null || echo "Service: Not responding"
    
    echo -e "\n${CYAN}Supported Chains:${NC}"
    echo "- Ethereum (ETH, USDC, USDT)"
    echo "- BSC (BNB, USDC, USDT)"
    echo "- Polygon (MATIC, USDC, USDT)"
    echo "- Solana (SOL, USDC, USDT)"
    echo "- Coffee Token (COFFEE)"
}

# Main execution
main() {
    case "${1:-}" in
        "health")
            print_header "Health Check"
            test_health
            ;;
        "payment")
            test_payment_creation
            ;;
        "flow")
            test_payment_flow
            ;;
        "chains")
            test_multiple_chains
            ;;
        "status")
            show_service_status
            ;;
        "all"|"")
            run_comprehensive_tests
            ;;
        *)
            echo "Usage: $0 [health|payment|flow|chains|status|all]"
            echo "  health   - Check service health"
            echo "  payment  - Test payment creation"
            echo "  flow     - Test complete payment flow"
            echo "  chains   - Test multiple blockchain support"
            echo "  status   - Show service status"
            echo "  all      - Run all tests (default)"
            ;;
    esac
}

# Check if jq is available for JSON parsing
if ! command -v jq > /dev/null 2>&1; then
    print_warning "jq is not installed. JSON output will be raw."
fi

# Run main function with arguments
main "$@"
