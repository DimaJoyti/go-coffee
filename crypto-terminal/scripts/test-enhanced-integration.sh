#!/bin/bash

# Enhanced Bright Data Integration Test Script
# This script tests all the new features implemented in Phase 1

set -e

echo "🚀 Testing Enhanced Bright Data Integration - Phase 1"
echo "=================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Base URL for API testing
BASE_URL="http://localhost:8080"

# Function to test API endpoint
test_endpoint() {
    local endpoint=$1
    local description=$2
    
    echo -e "${BLUE}Testing:${NC} $description"
    echo -e "${YELLOW}Endpoint:${NC} $endpoint"
    
    response=$(curl -s -w "%{http_code}" -o /tmp/response.json "$BASE_URL$endpoint")
    http_code="${response: -3}"
    
    if [ "$http_code" -eq 200 ]; then
        echo -e "${GREEN}✅ SUCCESS${NC} - HTTP $http_code"
        # Show first few lines of response
        echo -e "${BLUE}Response preview:${NC}"
        head -n 5 /tmp/response.json | jq . 2>/dev/null || head -n 5 /tmp/response.json
        echo ""
    else
        echo -e "${RED}❌ FAILED${NC} - HTTP $http_code"
        cat /tmp/response.json
        echo ""
    fi
}

# Function to check if service is running
check_service() {
    echo -e "${BLUE}Checking if crypto-terminal service is running...${NC}"
    
    if curl -s "$BASE_URL/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Service is running${NC}"
    else
        echo -e "${RED}❌ Service is not running${NC}"
        echo -e "${YELLOW}Please start the service with: go run cmd/terminal/main.go${NC}"
        exit 1
    fi
    echo ""
}

# Function to test Bright Data MCP integration
test_bright_data_mcp() {
    echo -e "${BLUE}=== Testing Bright Data MCP Integration ===${NC}"
    
    # Test search functionality
    test_endpoint "/api/v2/intelligence/news/search?q=bitcoin" "News Search"
    
    # Test scraping functionality
    test_endpoint "/api/v2/intelligence/quality/metrics" "Data Quality Metrics"
    
    echo ""
}

# Function to test 3commas integration
test_3commas_integration() {
    echo -e "${BLUE}=== Testing 3commas Integration ===${NC}"
    
    # Test 3commas bots
    test_endpoint "/api/v2/3commas/bots" "3commas Trading Bots"
    
    # Test 3commas signals
    test_endpoint "/api/v2/3commas/signals" "3commas Trading Signals"
    
    # Test 3commas deals
    test_endpoint "/api/v2/3commas/deals" "3commas Trading Deals"
    
    echo ""
}

# Function to test trading signals
test_trading_signals() {
    echo -e "${BLUE}=== Testing Trading Signals API ===${NC}"
    
    # Test all trading signals
    test_endpoint "/api/v2/trading/signals" "All Trading Signals"
    
    # Test signals by symbol
    test_endpoint "/api/v2/trading/signals/BTCUSDT" "Bitcoin Trading Signals"
    
    # Test signal search with filters
    test_endpoint "/api/v2/trading/signals/search?source=3commas&type=buy&risk_level=medium" "Filtered Trading Signals"
    
    # Test trading bots
    test_endpoint "/api/v2/trading/bots" "Trading Bots"
    
    # Test top trading bots
    test_endpoint "/api/v2/trading/bots/top?limit=10" "Top Trading Bots"
    
    echo ""
}

# Function to test technical analysis
test_technical_analysis() {
    echo -e "${BLUE}=== Testing Technical Analysis API ===${NC}"
    
    # Test all technical analysis
    test_endpoint "/api/v2/trading/analysis" "All Technical Analysis"
    
    # Test technical analysis by symbol
    test_endpoint "/api/v2/trading/analysis/BTCUSDT" "Bitcoin Technical Analysis"
    
    # Test with multiple symbols
    test_endpoint "/api/v2/trading/analysis?symbols=BTCUSDT,ETHUSDT,ADAUSDT" "Multi-Symbol Technical Analysis"
    
    echo ""
}

# Function to test active deals
test_active_deals() {
    echo -e "${BLUE}=== Testing Active Deals API ===${NC}"
    
    # Test all active deals
    test_endpoint "/api/v2/trading/deals" "All Trading Deals"
    
    # Test active deals only
    test_endpoint "/api/v2/trading/deals/active" "Active Trading Deals Only"
    
    echo ""
}

# Function to test enhanced TradingView integration
test_tradingview_enhanced() {
    echo -e "${BLUE}=== Testing Enhanced TradingView Integration ===${NC}"
    
    # Test TradingView market data
    test_endpoint "/api/v2/tradingview/market-data" "TradingView Market Data"
    
    # Test TradingView coins
    test_endpoint "/api/v2/tradingview/coins" "TradingView Coins"
    
    # Test trending coins
    test_endpoint "/api/v2/tradingview/trending" "TradingView Trending Coins"
    
    # Test gainers and losers
    test_endpoint "/api/v2/tradingview/gainers" "TradingView Top Gainers"
    test_endpoint "/api/v2/tradingview/losers" "TradingView Top Losers"
    
    echo ""
}

# Function to test social media integration
test_social_media() {
    echo -e "${BLUE}=== Testing Social Media Integration ===${NC}"
    
    # Test sentiment analysis
    test_endpoint "/api/v2/intelligence/sentiment" "Social Media Sentiment"
    
    # Test sentiment by symbol
    test_endpoint "/api/v2/intelligence/sentiment/BTC" "Bitcoin Social Sentiment"
    
    # Test trending topics
    test_endpoint "/api/v2/intelligence/sentiment/trending" "Trending Social Topics"
    
    echo ""
}

# Function to test portfolio analytics
test_portfolio_analytics() {
    echo -e "${BLUE}=== Testing Portfolio Analytics ===${NC}"
    
    # Test market heatmap
    test_endpoint "/api/v2/market/heatmap" "Market Heatmap"
    
    # Test sector performance
    test_endpoint "/api/v2/market/sectors" "Sector Performance"
    
    # Test correlation matrix
    test_endpoint "/api/v2/market/correlation" "Asset Correlation Matrix"
    
    echo ""
}

# Function to test data quality and monitoring
test_data_quality() {
    echo -e "${BLUE}=== Testing Data Quality & Monitoring ===${NC}"
    
    # Test quality metrics
    test_endpoint "/api/v2/intelligence/quality/metrics" "Data Quality Metrics"
    
    # Test service status
    test_endpoint "/api/v2/intelligence/quality/status" "Service Status"
    
    echo ""
}

# Function to run performance tests
test_performance() {
    echo -e "${BLUE}=== Running Performance Tests ===${NC}"
    
    echo -e "${YELLOW}Testing API response times...${NC}"
    
    # Test response time for trading signals
    start_time=$(date +%s%N)
    curl -s "$BASE_URL/api/v2/trading/signals?limit=50" > /dev/null
    end_time=$(date +%s%N)
    duration=$(( (end_time - start_time) / 1000000 ))
    
    echo -e "Trading Signals API: ${duration}ms"
    
    if [ $duration -lt 500 ]; then
        echo -e "${GREEN}✅ Performance: Excellent (<500ms)${NC}"
    elif [ $duration -lt 1000 ]; then
        echo -e "${YELLOW}⚠️  Performance: Good (<1000ms)${NC}"
    else
        echo -e "${RED}❌ Performance: Needs improvement (>1000ms)${NC}"
    fi
    
    echo ""
}

# Function to test frontend components
test_frontend() {
    echo -e "${BLUE}=== Testing Frontend Components ===${NC}"
    
    # Check if frontend is running
    if curl -s "http://localhost:3000" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Frontend is running on port 3000${NC}"
        
        # Test specific component endpoints
        echo -e "${YELLOW}Testing component data endpoints...${NC}"
        
        # These would be the endpoints that the React components call
        test_endpoint "/api/v2/trading/signals?limit=20" "TradingSignalsWidget Data"
        test_endpoint "/api/v2/3commas/bots?limit=10" "CommasIntegrationWidget Data"
        
    else
        echo -e "${YELLOW}⚠️  Frontend is not running${NC}"
        echo -e "${YELLOW}Start with: cd web && npm run dev${NC}"
    fi
    
    echo ""
}

# Function to generate test report
generate_report() {
    echo -e "${BLUE}=== Test Report Summary ===${NC}"
    
    total_tests=$(grep -c "test_endpoint" /tmp/test_log.txt 2>/dev/null || echo "0")
    passed_tests=$(grep -c "✅ SUCCESS" /tmp/test_log.txt 2>/dev/null || echo "0")
    failed_tests=$(grep -c "❌ FAILED" /tmp/test_log.txt 2>/dev/null || echo "0")
    
    echo -e "Total Tests: $total_tests"
    echo -e "Passed: ${GREEN}$passed_tests${NC}"
    echo -e "Failed: ${RED}$failed_tests${NC}"
    
    if [ $failed_tests -eq 0 ]; then
        echo -e "${GREEN}🎉 All tests passed!${NC}"
    else
        echo -e "${RED}⚠️  Some tests failed. Check the output above.${NC}"
    fi
    
    echo ""
    echo -e "${BLUE}=== Next Steps ===${NC}"
    echo -e "1. Review any failed tests and fix issues"
    echo -e "2. Test the frontend components in browser"
    echo -e "3. Monitor real-time data updates"
    echo -e "4. Proceed to Phase 2 implementation"
    echo ""
}

# Main execution
main() {
    # Redirect output to log file for report generation
    exec > >(tee /tmp/test_log.txt)
    
    echo -e "${GREEN}Starting Enhanced Bright Data Integration Tests${NC}"
    echo -e "${YELLOW}Timestamp: $(date)${NC}"
    echo ""
    
    # Run all tests
    check_service
    test_bright_data_mcp
    test_3commas_integration
    test_trading_signals
    test_technical_analysis
    test_active_deals
    test_tradingview_enhanced
    test_social_media
    test_portfolio_analytics
    test_data_quality
    test_performance
    test_frontend
    
    # Generate final report
    generate_report
}

# Run main function
main "$@"
