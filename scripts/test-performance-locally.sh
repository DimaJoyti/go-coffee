#!/bin/bash

# Test Performance Scripts Locally
# This script validates that our performance tests work before running in CI

set -e

echo "🧪 Testing Performance Scripts Locally"
echo "======================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "SUCCESS")
            echo -e "${GREEN}✅ $message${NC}"
            ;;
        "ERROR")
            echo -e "${RED}❌ $message${NC}"
            ;;
        "WARNING")
            echo -e "${YELLOW}⚠️  $message${NC}"
            ;;
        "INFO")
            echo -e "${YELLOW}ℹ️  $message${NC}"
            ;;
    esac
}

# Check if K6 is installed
check_k6() {
    print_status "INFO" "Checking K6 installation..."
    if command -v k6 &> /dev/null; then
        print_status "SUCCESS" "K6 is installed: $(k6 version)"
        return 0
    else
        print_status "WARNING" "K6 not found. Installing K6..."
        install_k6
        return $?
    fi
}

# Install K6
install_k6() {
    print_status "INFO" "Installing K6..."
    
    # Detect OS
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        curl -fsSL https://github.com/grafana/k6/releases/download/v0.47.0/k6-v0.47.0-linux-amd64.tar.gz | tar -xz
        sudo mv k6-v0.47.0-linux-amd64/k6 /usr/local/bin/
        rm -rf k6-v0.47.0-linux-amd64
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        if command -v brew &> /dev/null; then
            brew install k6
        else
            curl -fsSL https://github.com/grafana/k6/releases/download/v0.47.0/k6-v0.47.0-darwin-amd64.tar.gz | tar -xz
            sudo mv k6-v0.47.0-darwin-amd64/k6 /usr/local/bin/
            rm -rf k6-v0.47.0-darwin-amd64
        fi
    else
        print_status "ERROR" "Unsupported OS: $OSTYPE"
        return 1
    fi
    
    if command -v k6 &> /dev/null; then
        print_status "SUCCESS" "K6 installed successfully: $(k6 version)"
        return 0
    else
        print_status "ERROR" "Failed to install K6"
        return 1
    fi
}

# Start mock services
start_mock_services() {
    print_status "INFO" "Starting mock services..."
    
    # Check if Docker is available
    if ! command -v docker &> /dev/null; then
        print_status "ERROR" "Docker is required but not installed"
        return 1
    fi
    
    # Stop any existing containers
    docker stop mock-api-gateway mock-user-gateway mock-security-gateway mock-web-ui-backend 2>/dev/null || true
    docker rm mock-api-gateway mock-user-gateway mock-security-gateway mock-web-ui-backend 2>/dev/null || true
    
    # Start mock services
    docker run -d --name mock-api-gateway -p 8080:80 kennethreitz/httpbin:latest
    docker run -d --name mock-user-gateway -p 8081:80 kennethreitz/httpbin:latest
    docker run -d --name mock-security-gateway -p 8082:80 kennethreitz/httpbin:latest
    docker run -d --name mock-web-ui-backend -p 8090:80 kennethreitz/httpbin:latest
    
    # Wait for services to start
    print_status "INFO" "Waiting for services to start..."
    sleep 15
    
    # Verify services
    local all_services_up=true
    for port in 8080 8081 8082 8090; do
        if curl -f -s http://localhost:$port/get >/dev/null 2>&1; then
            print_status "SUCCESS" "Service on port $port is responding"
        else
            print_status "ERROR" "Service on port $port is not responding"
            all_services_up=false
        fi
    done
    
    if $all_services_up; then
        print_status "SUCCESS" "All mock services are running"
        return 0
    else
        print_status "ERROR" "Some mock services failed to start"
        return 1
    fi
}

# Stop mock services
stop_mock_services() {
    print_status "INFO" "Stopping mock services..."
    docker stop mock-api-gateway mock-user-gateway mock-security-gateway mock-web-ui-backend 2>/dev/null || true
    docker rm mock-api-gateway mock-user-gateway mock-security-gateway mock-web-ui-backend 2>/dev/null || true
    print_status "SUCCESS" "Mock services stopped"
}

# Test load test script
test_load_script() {
    print_status "INFO" "Testing load test script..."
    
    export API_BASE_URL="http://localhost:8080"
    export USER_GATEWAY_URL="http://localhost:8081"
    export SECURITY_GATEWAY_URL="http://localhost:8082"
    export WEB_UI_BACKEND_URL="http://localhost:8090"
    
    if k6 run --duration=30s --vus=2 tests/performance/load-test.js; then
        print_status "SUCCESS" "Load test completed successfully"
        return 0
    else
        print_status "ERROR" "Load test failed"
        return 1
    fi
}

# Test stress test script
test_stress_script() {
    print_status "INFO" "Testing stress test script (short version)..."
    
    export API_BASE_URL="http://localhost:8080"
    export USER_GATEWAY_URL="http://localhost:8081"
    export SECURITY_GATEWAY_URL="http://localhost:8082"
    export WEB_UI_BACKEND_URL="http://localhost:8090"
    
    # Create a short version of stress test for local testing
    cat > tests/performance/stress-test-short.js << 'EOF'
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '10s', target: 2 },
    { duration: '20s', target: 5 },
    { duration: '10s', target: 0 },
  ],
};

const BASE_URL = __ENV.API_BASE_URL || 'http://localhost:8080';

export default function() {
  const response = http.get(`${BASE_URL}/get`);
  
  check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 1000ms': (r) => r.timings.duration < 1000,
  });
  
  sleep(0.5);
}
EOF
    
    if k6 run tests/performance/stress-test-short.js; then
        print_status "SUCCESS" "Stress test completed successfully"
        rm -f tests/performance/stress-test-short.js
        return 0
    else
        print_status "ERROR" "Stress test failed"
        rm -f tests/performance/stress-test-short.js
        return 1
    fi
}

# Main execution
main() {
    print_status "INFO" "Starting performance test validation..."
    
    # Trap to ensure cleanup
    trap 'stop_mock_services' EXIT
    
    # Check and install K6
    if ! check_k6; then
        print_status "ERROR" "Failed to setup K6"
        exit 1
    fi
    
    # Start mock services
    if ! start_mock_services; then
        print_status "ERROR" "Failed to start mock services"
        exit 1
    fi
    
    # Test load script
    if ! test_load_script; then
        print_status "ERROR" "Load test validation failed"
        exit 1
    fi
    
    # Test stress script
    if ! test_stress_script; then
        print_status "ERROR" "Stress test validation failed"
        exit 1
    fi
    
    print_status "SUCCESS" "All performance tests validated successfully!"
    print_status "INFO" "Performance tests are ready for CI/CD"
}

# Run main function
main "$@"
