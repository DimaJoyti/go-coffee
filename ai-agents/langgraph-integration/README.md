# 🤖 Go Coffee LangGraph Integration

A sophisticated **Go-based implementation** of LangGraph-inspired multi-agent orchestration for the Go Coffee platform. This system provides graph-based workflow execution, state management, and intelligent agent coordination using pure Go.

## 🎯 Overview

This implementation brings the power of LangGraph's graph-based agent orchestration to Go, providing:

- **Graph-Based Workflows** - Define complex multi-agent workflows as directed graphs
- **State Management** - Persistent state across agent executions with proper serialization
- **Conditional Routing** - Dynamic workflow paths based on agent outputs and conditions
- **Agent Orchestration** - Seamless coordination between specialized AI agents
- **Error Handling** - Robust retry policies and error recovery mechanisms
- **Observability** - Comprehensive metrics, logging, and monitoring

## 🏗️ Architecture

### Core Components

```
┌─────────────────────────────────────────────────────────────┐
│                 LangGraph Orchestrator                     │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   Graph     │  │   Agent     │  │   State     │        │
│  │  Engine     │  │ Registry    │  │ Manager     │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
├─────────────────────────────────────────────────────────────┤
│                    Agent Network                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │ Beverage    │  │ Inventory   │  │Social Media │        │
│  │ Inventor    │  │ Manager     │  │   Agent     │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

### Key Features

- **🔄 Graph Execution** - Execute workflows as directed graphs with conditional routing
- **🧠 State Persistence** - Maintain conversation and workflow state across executions
- **🎯 Agent Coordination** - Intelligent routing between specialized agents
- **⚡ Concurrent Execution** - Support for multiple simultaneous workflows
- **🔧 Tool Integration** - Seamless integration with external APIs and services
- **📊 Monitoring** - Real-time metrics and execution tracking

## 🚀 Quick Start

### 1. Installation

```bash
# Clone the repository
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee/ai-agents/langgraph-integration

# Install dependencies
go mod download

# Build the application
go build -o langgraph-server cmd/main.go
```

### 2. Run the Server

```bash
# Start the LangGraph server
./langgraph-server

# Or run directly
go run cmd/main.go
```

The server will start on port 8080 by default.

### 3. Execute a Workflow

```bash
# Create a coffee recipe using the beverage inventor agent
curl -X POST http://localhost:8080/api/v1/workflows/execute \
  -H "Content-Type: application/json" \
  -d '{
    "graph_id": "coffee_creation",
    "input_data": {
      "season": "winter",
      "flavor_profile": "spicy",
      "temperature": "hot",
      "occasion": "morning",
      "generate_variations": true
    },
    "priority": "medium"
  }'
```

### 4. Check Execution Status

```bash
# Get execution status
curl http://localhost:8080/api/v1/executions/{execution_id}

# List all active executions
curl http://localhost:8080/api/v1/executions

# Get orchestrator statistics
curl http://localhost:8080/api/v1/stats
```

## 📁 Project Structure

```
ai-agents/langgraph-integration/
├── cmd/
│   └── main.go                 # Main application entry point
├── pkg/
│   ├── graph/
│   │   ├── state.go           # State management and types
│   │   └── graph.go           # Graph execution engine
│   ├── agents/
│   │   ├── base.go            # Base agent interface and implementation
│   │   └── beverage_inventor.go # Beverage inventor agent
│   └── orchestrator/
│       └── orchestrator.go    # Main orchestration engine
├── examples/
│   └── workflows/             # Example workflow definitions
├── go.mod                     # Go module definition
└── README.md                  # This file
```

## 🤖 Available Agents

### Beverage Inventor Agent

Creates innovative coffee beverages and recipes based on:
- **Seasonal preferences** (spring, summer, fall, winter)
- **Flavor profiles** (sweet, spicy, fruity, nutty, floral)
- **Dietary restrictions** (vegan, gluten-free, etc.)
- **Temperature preferences** (hot, cold)
- **Occasion** (daily, special, morning, evening)

**Example Usage:**
```json
{
  "season": "winter",
  "flavor_profile": "spicy",
  "dietary_restrictions": ["vegan"],
  "caffeine_level": "high",
  "temperature": "hot",
  "occasion": "morning"
}
```

**Output:**
- Complete recipe with ingredients and instructions
- Nutritional information
- Quality analysis and feasibility assessment
- Recipe variations (if requested)

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `ENVIRONMENT` | Environment (development/production) | `development` |
| `LOG_LEVEL` | Logging level | `info` |
| `ENABLE_METRICS` | Enable metrics collection | `true` |

### Orchestrator Configuration

```go
config := &orchestrator.OrchestratorConfig{
    MaxConcurrentExecutions: 50,
    DefaultTimeout:          10 * time.Minute,
    RetryPolicy: orchestrator.RetryPolicy{
        MaxRetries:    3,
        InitialDelay:  1 * time.Second,
        BackoffFactor: 2.0,
        MaxDelay:      30 * time.Second,
    },
    MonitoringEnabled: true,
}
```

## 📊 API Reference

### Execute Workflow

**POST** `/api/v1/workflows/execute`

```json
{
  "graph_id": "coffee_creation",
  "input_data": {
    "season": "spring",
    "flavor_profile": "fruity"
  },
  "priority": "high",
  "config": {}
}
```

**Response:**
```json
{
  "workflow_id": "uuid",
  "execution_id": "uuid",
  "status": "completed",
  "result": {
    "beverage_inventor": {
      "recipe": {...},
      "analysis": {...}
    }
  },
  "duration_seconds": 2.5
}
```

### Get Execution Status

**GET** `/api/v1/executions/{id}`

**Response:**
```json
{
  "workflow_id": "uuid",
  "execution_id": "uuid",
  "graph_id": "coffee_creation",
  "status": "running",
  "current_node": "beverage_inventor",
  "executed_nodes": ["start"],
  "start_time": "2024-01-01T10:00:00Z",
  "last_update": "2024-01-01T10:01:30Z"
}
```

### List Active Executions

**GET** `/api/v1/executions`

**Response:**
```json
{
  "executions": [...],
  "count": 5
}
```

### Cancel Execution

**POST** `/api/v1/executions/{id}/cancel`

**Response:**
```json
{
  "status": "cancelled",
  "message": "Execution cancelled successfully"
}
```

### Get Statistics

**GET** `/api/v1/stats`

**Response:**
```json
{
  "registered_agents": 1,
  "registered_graphs": 1,
  "active_executions": 3,
  "total_executions": 150,
  "successful_executions": 142,
  "failed_executions": 8,
  "success_rate": 0.947,
  "last_execution": "2024-01-01T10:05:00Z"
}
```

## 🔍 Monitoring & Observability

### Health Check

```bash
curl http://localhost:8080/health
```

### Metrics

The system provides comprehensive metrics including:
- **Execution metrics** - Success rate, duration, throughput
- **Agent metrics** - Individual agent performance and health
- **System metrics** - Resource usage, concurrent executions
- **Error metrics** - Error rates, retry counts, failure analysis

### Logging

Structured logging with contextual information:
- Workflow execution tracking
- Agent performance monitoring
- Error reporting and debugging
- State transitions and routing decisions

## 🧪 Testing

```bash
# Run unit tests
go test ./pkg/...

# Run integration tests
go test -tags=integration ./...

# Run with coverage
go test -cover ./pkg/...

# Benchmark tests
go test -bench=. ./pkg/...
```

## 🚀 Deployment

### Docker

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o langgraph-server cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/langgraph-server .
CMD ["./langgraph-server"]
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: langgraph-orchestrator
spec:
  replicas: 3
  selector:
    matchLabels:
      app: langgraph-orchestrator
  template:
    metadata:
      labels:
        app: langgraph-orchestrator
    spec:
      containers:
      - name: orchestrator
        image: go-coffee/langgraph-orchestrator:latest
        ports:
        - containerPort: 8080
        env:
        - name: PORT
          value: "8080"
        - name: ENVIRONMENT
          value: "production"
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: [Go Coffee Docs](../docs/)
- **Issues**: [GitHub Issues](https://github.com/DimaJoyti/go-coffee/issues)
- **Discussions**: [GitHub Discussions](https://github.com/DimaJoyti/go-coffee/discussions)

---

**Built with ❤️ by the Go Coffee AI Team using Go 1.24+**
