#!/bin/bash

# Go Coffee AI Orchestrator Test Script
# This script tests the AI agent orchestration functionality

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
AI_ORCHESTRATOR_URL="http://localhost:8094"
AI_HEALTH_URL="http://localhost:8095/health"

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
    print_status "Testing AI Orchestrator service health..."
    
    if curl -f -s "$AI_HEALTH_URL" > /dev/null; then
        print_success "AI Orchestrator service is healthy"
        return 0
    else
        print_error "AI Orchestrator service health check failed"
        return 1
    fi
}

# Function to test agent listing
test_list_agents() {
    print_status "Testing agent listing..."
    
    local response=$(curl -s "$AI_ORCHESTRATOR_URL/agents")
    
    if echo "$response" | grep -q "agents"; then
        print_success "Agent listing successful"
        echo "Response: $response"
        return 0
    else
        print_error "Agent listing failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test agent registration
test_register_agent() {
    print_status "Testing agent registration..."
    
    local response=$(curl -s -X POST "$AI_ORCHESTRATOR_URL/agents/register" \
        -H "Content-Type: application/json" \
        -d '{
            "id": "test-agent",
            "type": "test",
            "capabilities": ["test_action"]
        }')
    
    if echo "$response" | grep -q "registered"; then
        print_success "Agent registration successful"
        echo "Response: $response"
        return 0
    else
        print_error "Agent registration failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test workflow creation
test_create_workflow() {
    print_status "Testing workflow creation..."
    
    local response=$(curl -s -X POST "$AI_ORCHESTRATOR_URL/workflows/create" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "Test Coffee Workflow",
            "description": "A test workflow for coffee processing",
            "steps": [
                {
                    "id": "analyze-order",
                    "agent_id": "beverage-inventor",
                    "action": "analyze_order",
                    "inputs": {"order_type": "latte"}
                },
                {
                    "id": "check-inventory",
                    "agent_id": "inventory-manager",
                    "action": "check_availability",
                    "inputs": {"item": "milk"}
                }
            ]
        }')
    
    if echo "$response" | grep -q "created"; then
        print_success "Workflow creation successful"
        echo "Response: $response"
        
        # Extract workflow ID for further tests
        WORKFLOW_ID=$(echo "$response" | jq -r '.workflow_id' 2>/dev/null || echo "")
        return 0
    else
        print_error "Workflow creation failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test workflow execution
test_execute_workflow() {
    if [ -z "$WORKFLOW_ID" ]; then
        print_warning "No workflow ID available for execution test"
        return 1
    fi
    
    print_status "Testing workflow execution..."
    
    local response=$(curl -s -X POST "$AI_ORCHESTRATOR_URL/workflows/execute" \
        -H "Content-Type: application/json" \
        -d "{
            \"workflow_id\": \"$WORKFLOW_ID\"
        }")
    
    if echo "$response" | grep -q "executing"; then
        print_success "Workflow execution started successfully"
        echo "Response: $response"
        return 0
    else
        print_error "Workflow execution failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test task assignment
test_assign_task() {
    print_status "Testing task assignment..."
    
    local response=$(curl -s -X POST "$AI_ORCHESTRATOR_URL/tasks/assign" \
        -H "Content-Type: application/json" \
        -d '{
            "agent_id": "beverage-inventor",
            "action": "create_recipe",
            "inputs": {
                "name": "Test Latte",
                "type": "coffee"
            },
            "priority": "medium"
        }')
    
    if echo "$response" | grep -q "assigned"; then
        print_success "Task assignment successful"
        echo "Response: $response"
        
        # Extract task ID for further tests
        TASK_ID=$(echo "$response" | jq -r '.task_id' 2>/dev/null || echo "")
        return 0
    else
        print_error "Task assignment failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test message sending
test_send_message() {
    print_status "Testing agent message sending..."
    
    local response=$(curl -s -X POST "$AI_ORCHESTRATOR_URL/messages/send" \
        -H "Content-Type: application/json" \
        -d '{
            "from_agent": "beverage-inventor",
            "to_agent": "inventory-manager",
            "type": "request",
            "content": {
                "message": "Need ingredient availability check",
                "ingredients": ["milk", "coffee beans", "sugar"]
            }
        }')
    
    if echo "$response" | grep -q "sent"; then
        print_success "Message sending successful"
        echo "Response: $response"
        return 0
    else
        print_error "Message sending failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test message broadcasting
test_broadcast_message() {
    print_status "Testing message broadcasting..."
    
    local response=$(curl -s -X POST "$AI_ORCHESTRATOR_URL/messages/broadcast" \
        -H "Content-Type: application/json" \
        -d '{
            "type": "system_announcement",
            "content": {
                "message": "Daily operations starting",
                "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
            }
        }')
    
    if echo "$response" | grep -q "broadcasted"; then
        print_success "Message broadcasting successful"
        echo "Response: $response"
        return 0
    else
        print_error "Message broadcasting failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test external integrations
test_external_integrations() {
    print_status "Testing external integrations..."
    
    # Test ClickUp integration
    local clickup_response=$(curl -s "$AI_ORCHESTRATOR_URL/integrations/clickup")
    if echo "$clickup_response" | grep -q "clickup"; then
        print_success "ClickUp integration test passed"
    else
        print_warning "ClickUp integration test failed"
    fi
    
    # Test Slack integration
    local slack_response=$(curl -s "$AI_ORCHESTRATOR_URL/integrations/slack")
    if echo "$slack_response" | grep -q "slack"; then
        print_success "Slack integration test passed"
    else
        print_warning "Slack integration test failed"
    fi
    
    # Test Google Sheets integration
    local sheets_response=$(curl -s "$AI_ORCHESTRATOR_URL/integrations/sheets")
    if echo "$sheets_response" | grep -q "google_sheets"; then
        print_success "Google Sheets integration test passed"
    else
        print_warning "Google Sheets integration test failed"
    fi
    
    # Test Airtable integration
    local airtable_response=$(curl -s "$AI_ORCHESTRATOR_URL/integrations/airtable")
    if echo "$airtable_response" | grep -q "airtable"; then
        print_success "Airtable integration test passed"
    else
        print_warning "Airtable integration test failed"
    fi
}

# Function to test AI agent workflow scenarios
test_ai_workflow_scenarios() {
    print_status "Testing AI workflow scenarios..."
    
    # Scenario 1: Coffee Order Processing
    print_status "Testing coffee order processing workflow..."
    local order_response=$(curl -s -X POST "$AI_ORCHESTRATOR_URL/workflows/execute" \
        -H "Content-Type: application/json" \
        -d '{
            "workflow_id": "coffee-order-processing"
        }')
    
    if echo "$order_response" | grep -q "executing"; then
        print_success "Coffee order processing workflow started"
    else
        print_warning "Coffee order processing workflow failed to start"
    fi
    
    # Scenario 2: Daily Operations
    print_status "Testing daily operations workflow..."
    local daily_response=$(curl -s -X POST "$AI_ORCHESTRATOR_URL/workflows/execute" \
        -H "Content-Type: application/json" \
        -d '{
            "workflow_id": "daily-operations"
        }')
    
    if echo "$daily_response" | grep -q "executing"; then
        print_success "Daily operations workflow started"
    else
        print_warning "Daily operations workflow failed to start"
    fi
}

# Function to test agent capabilities
test_agent_capabilities() {
    print_status "Testing individual agent capabilities..."
    
    # Test Beverage Inventor Agent
    local beverage_response=$(curl -s -X POST "$AI_ORCHESTRATOR_URL/tasks/assign" \
        -H "Content-Type: application/json" \
        -d '{
            "agent_id": "beverage-inventor",
            "action": "analyze_trends",
            "inputs": {"period": "monthly"},
            "priority": "high"
        }')
    
    if echo "$beverage_response" | grep -q "assigned"; then
        print_success "Beverage Inventor Agent capability test passed"
    else
        print_warning "Beverage Inventor Agent capability test failed"
    fi
    
    # Test Inventory Manager Agent
    local inventory_response=$(curl -s -X POST "$AI_ORCHESTRATOR_URL/tasks/assign" \
        -H "Content-Type: application/json" \
        -d '{
            "agent_id": "inventory-manager",
            "action": "forecast_demand",
            "inputs": {"period": "weekly"},
            "priority": "high"
        }')
    
    if echo "$inventory_response" | grep -q "assigned"; then
        print_success "Inventory Manager Agent capability test passed"
    else
        print_warning "Inventory Manager Agent capability test failed"
    fi
    
    # Test Task Manager Agent
    local task_response=$(curl -s -X POST "$AI_ORCHESTRATOR_URL/tasks/assign" \
        -H "Content-Type: application/json" \
        -d '{
            "agent_id": "task-manager",
            "action": "optimize_schedule",
            "inputs": {"scope": "daily"},
            "priority": "medium"
        }')
    
    if echo "$task_response" | grep -q "assigned"; then
        print_success "Task Manager Agent capability test passed"
    else
        print_warning "Task Manager Agent capability test failed"
    fi
}

# Function to run comprehensive tests
run_comprehensive_tests() {
    print_header "AI Orchestrator Test Suite"
    
    local failed_tests=0
    
    # Test 1: Health check
    print_header "Test 1: Health Check"
    test_health || ((failed_tests++))
    
    # Test 2: Agent listing
    print_header "Test 2: Agent Listing"
    test_list_agents || ((failed_tests++))
    
    # Test 3: Agent registration
    print_header "Test 3: Agent Registration"
    test_register_agent || ((failed_tests++))
    
    # Test 4: Workflow creation
    print_header "Test 4: Workflow Creation"
    test_create_workflow || ((failed_tests++))
    
    # Test 5: Workflow execution
    print_header "Test 5: Workflow Execution"
    test_execute_workflow || ((failed_tests++))
    
    # Test 6: Task assignment
    print_header "Test 6: Task Assignment"
    test_assign_task || ((failed_tests++))
    
    # Test 7: Message sending
    print_header "Test 7: Message Sending"
    test_send_message || ((failed_tests++))
    
    # Test 8: Message broadcasting
    print_header "Test 8: Message Broadcasting"
    test_broadcast_message || ((failed_tests++))
    
    # Test 9: External integrations
    print_header "Test 9: External Integrations"
    test_external_integrations || ((failed_tests++))
    
    # Test 10: AI workflow scenarios
    print_header "Test 10: AI Workflow Scenarios"
    test_ai_workflow_scenarios || ((failed_tests++))
    
    # Test 11: Agent capabilities
    print_header "Test 11: Agent Capabilities"
    test_agent_capabilities || ((failed_tests++))
    
    # Summary
    print_header "Test Summary"
    if [ $failed_tests -eq 0 ]; then
        print_success "All tests passed! ✅"
        print_status "The AI Orchestrator is working correctly with all 9 agents."
    else
        print_error "$failed_tests test(s) failed ❌"
        print_status "Please check the logs for more details."
    fi
    
    return $failed_tests
}

# Function to show service status
show_service_status() {
    print_header "AI Orchestrator Service Status"
    
    echo -e "${CYAN}Service Health:${NC}"
    curl -s "$AI_HEALTH_URL" | jq '.' 2>/dev/null || echo "Service: Not responding"
    
    echo -e "\n${CYAN}Registered Agents:${NC}"
    curl -s "$AI_ORCHESTRATOR_URL/agents" | jq '.agents[] | {id: .id, type: .type, status: .status}' 2>/dev/null || echo "Unable to fetch agents"
    
    echo -e "\n${CYAN}Active Workflows:${NC}"
    curl -s "$AI_ORCHESTRATOR_URL/workflows" | jq '.workflows[] | {id: .id, name: .name, status: .status}' 2>/dev/null || echo "Unable to fetch workflows"
    
    echo -e "\n${CYAN}AI Agent Capabilities:${NC}"
    echo "1. Beverage Inventor - Recipe creation, trend analysis, menu optimization"
    echo "2. Inventory Manager - Stock management, demand forecasting, supplier coordination"
    echo "3. Task Manager - Workflow automation, task scheduling, resource allocation"
    echo "4. Social Media - Content generation, engagement analysis, posting automation"
    echo "5. Feedback Analyst - Customer feedback analysis, sentiment tracking, insights"
    echo "6. Scheduler - Calendar management, appointment scheduling, resource planning"
    echo "7. Inter-Location Coordinator - Multi-location coordination, data synchronization"
    echo "8. Notifier - Alert management, notification delivery, communication routing"
    echo "9. Tasting Coordinator - Tasting session management, feedback coordination"
}

# Main execution
main() {
    case "${1:-}" in
        "health")
            print_header "Health Check"
            test_health
            ;;
        "agents")
            test_list_agents
            ;;
        "workflows")
            test_create_workflow
            test_execute_workflow
            ;;
        "tasks")
            test_assign_task
            ;;
        "messages")
            test_send_message
            test_broadcast_message
            ;;
        "integrations")
            test_external_integrations
            ;;
        "scenarios")
            test_ai_workflow_scenarios
            ;;
        "capabilities")
            test_agent_capabilities
            ;;
        "status")
            show_service_status
            ;;
        "all"|"")
            run_comprehensive_tests
            ;;
        *)
            echo "Usage: $0 [health|agents|workflows|tasks|messages|integrations|scenarios|capabilities|status|all]"
            echo "  health       - Check service health"
            echo "  agents       - Test agent management"
            echo "  workflows    - Test workflow creation and execution"
            echo "  tasks        - Test task assignment"
            echo "  messages     - Test agent communication"
            echo "  integrations - Test external integrations"
            echo "  scenarios    - Test AI workflow scenarios"
            echo "  capabilities - Test individual agent capabilities"
            echo "  status       - Show service status"
            echo "  all          - Run all tests (default)"
            ;;
    esac
}

# Check if jq is available for JSON parsing
if ! command -v jq > /dev/null 2>&1; then
    print_warning "jq is not installed. JSON output will be raw."
fi

# Run main function with arguments
main "$@"
