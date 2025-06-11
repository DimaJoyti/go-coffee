# 🚀 LLM Orchestrator Simple: Lightweight Control System for LLM Workloads

## Overview

The LLM Orchestrator Simple is a lightweight, standalone version of our cutting-edge control system for managing Large Language Model (LLM) workloads. This simplified version provides core orchestration capabilities without requiring Kubernetes, making it perfect for development, testing, and smaller deployments.

## ✨ Key Features

### 🎯 **Core Capabilities**
- **Workload Management**: Create, list, retrieve, and delete LLM workloads
- **Intelligent Scheduling**: Basic scheduling with resource-aware placement
- **Real-time Monitoring**: Live metrics collection and performance tracking
- **RESTful API**: Complete HTTP API for programmatic control
- **Configuration-Driven**: YAML-based configuration for easy customization

### 📊 **Monitoring & Observability**
- **Health Checks**: Built-in health monitoring endpoints
- **Performance Metrics**: CPU, memory, GPU utilization tracking
- **Workload Status**: Real-time status updates and phase tracking
- **Request Analytics**: Request rate, latency, and error rate monitoring

### 🔧 **Developer-Friendly**
- **Lightweight**: No external dependencies beyond Go runtime
- **Fast Startup**: Quick deployment and testing
- **Simple Configuration**: Easy YAML configuration
- **Comprehensive Logging**: Structured logging with configurable levels

## 🚀 Quick Start

### Prerequisites
- Go 1.24+ installed
- Basic understanding of REST APIs

### Installation & Build

1. **Clone the repository**:
```bash
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee
```

2. **Build the orchestrator**:
```bash
go build -o bin/llm-orchestrator-simple.exe ./cmd/llm-orchestrator-simple
```

3. **Run the orchestrator**:
```bash
./bin/llm-orchestrator-simple.exe --config=config/llm-orchestrator-simple.yaml --port=8080 --log-level=info
```

### Verify Installation

Test the health endpoint:
```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0"
}
```

## 📋 API Reference

### Health & Status Endpoints

#### GET /health
Returns the health status of the orchestrator.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0"
}
```

#### GET /status
Returns detailed orchestrator status.

**Response:**
```json
{
  "orchestrator": "running",
  "workloads": 3,
  "uptime": "2h30m15s",
  "version": "1.0.0",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### GET /metrics
Returns performance metrics.

**Response:**
```json
{
  "totalWorkloads": 5,
  "runningWorkloads": 3,
  "pendingWorkloads": 1,
  "failedWorkloads": 1,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Workload Management Endpoints

#### GET /workloads
List all workloads.

**Response:**
```json
[
  {
    "id": "workload-1642248600",
    "name": "llama2-7b",
    "modelName": "llama2",
    "modelType": "text-generation",
    "resources": {
      "cpu": "2000m",
      "memory": "8Gi",
      "gpu": 1
    },
    "status": {
      "phase": "running",
      "replicas": 1,
      "readyReplicas": 1,
      "lastUpdated": "2024-01-15T10:25:00Z"
    },
    "metrics": {
      "requestsPerSecond": 125.5,
      "averageLatency": "75ms",
      "errorRate": 0.01,
      "cpuUsage": 0.65,
      "memoryUsage": 0.72,
      "gpuUsage": 0.85,
      "lastUpdated": "2024-01-15T10:30:00Z"
    },
    "createdAt": "2024-01-15T10:00:00Z",
    "updatedAt": "2024-01-15T10:30:00Z"
  }
]
```

#### POST /workloads
Create a new workload.

**Request Body:**
```json
{
  "name": "my-llm-workload",
  "modelName": "llama2",
  "modelType": "text-generation",
  "resources": {
    "cpu": "2000m",
    "memory": "8Gi",
    "gpu": 1
  },
  "labels": {
    "environment": "development",
    "team": "ai-research"
  }
}
```

**Response:** (201 Created)
```json
{
  "id": "workload-1642248600",
  "name": "my-llm-workload",
  "modelName": "llama2",
  "modelType": "text-generation",
  "resources": {
    "cpu": "2000m",
    "memory": "8Gi",
    "gpu": 1
  },
  "status": {
    "phase": "pending",
    "replicas": 1,
    "readyReplicas": 0,
    "lastUpdated": "2024-01-15T10:30:00Z"
  },
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-15T10:30:00Z",
  "labels": {
    "environment": "development",
    "team": "ai-research"
  }
}
```

#### GET /workloads/{id}
Retrieve a specific workload.

**Response:**
```json
{
  "id": "workload-1642248600",
  "name": "my-llm-workload",
  "modelName": "llama2",
  "modelType": "text-generation",
  "status": {
    "phase": "running",
    "replicas": 1,
    "readyReplicas": 1,
    "lastUpdated": "2024-01-15T10:30:00Z"
  }
}
```

#### DELETE /workloads/{id}
Delete a workload.

**Response:** (204 No Content)

### Scheduling Endpoints

#### POST /schedule
Schedule a workload to a node.

**Request Body:**
```json
{
  "workloadId": "workload-1642248600"
}
```

**Response:**
```json
{
  "workloadId": "workload-1642248600",
  "scheduledNode": "node-1",
  "schedulingTime": "2024-01-15T10:30:00Z",
  "status": "scheduled"
}
```

## ⚙️ Configuration

### Basic Configuration

The orchestrator uses YAML configuration files. Here's a basic example:

```yaml
# Server Configuration
port: 8080
logLevel: "info"

# Resource Management
maxWorkloads: 100
defaultCPU: "1000m"
defaultMemory: "2Gi"
defaultGPU: 0

# Metrics and Monitoring
metricsInterval: "30s"

# Performance Tuning
performance:
  maxConcurrentRequests: 1000
  requestTimeout: "30s"
  healthCheckInterval: "10s"
```

### Advanced Configuration

For more advanced setups, you can configure:

- **Resource Profiles**: Predefined resource configurations
- **Model Configurations**: Model-specific settings
- **Monitoring**: Custom metrics and alerting
- **Security**: Authentication and authorization (future)

See `config/llm-orchestrator-simple.yaml` for a complete example.

## 🧪 Testing

### Automated Testing

Run the comprehensive test suite:

**PowerShell (Windows):**
```powershell
powershell -ExecutionPolicy Bypass -File scripts/test-llm-orchestrator.ps1
```

**Bash (Linux/macOS):**
```bash
chmod +x scripts/test-llm-orchestrator.sh
./scripts/test-llm-orchestrator.sh
```

### Manual Testing

1. **Start the orchestrator**:
```bash
./bin/llm-orchestrator-simple.exe --config=config/llm-orchestrator-simple.yaml
```

2. **Create a workload**:
```bash
curl -X POST http://localhost:8080/workloads \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-workload",
    "modelName": "llama2",
    "modelType": "text-generation",
    "resources": {
      "cpu": "1000m",
      "memory": "4Gi",
      "gpu": 0
    }
  }'
```

3. **Check workload status**:
```bash
curl http://localhost:8080/workloads
```

## 🐳 Docker Deployment

### Build Docker Image

```bash
docker build -f docker/Dockerfile.llm-orchestrator-simple -t llm-orchestrator-simple:latest .
```

### Run with Docker

```bash
docker run -p 8080:8080 llm-orchestrator-simple:latest
```

### Docker Compose

```yaml
version: '3.8'
services:
  llm-orchestrator:
    build:
      context: .
      dockerfile: docker/Dockerfile.llm-orchestrator-simple
    ports:
      - "8080:8080"
    environment:
      - LOG_LEVEL=info
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3
```

## 🔍 Monitoring & Observability

### Metrics Collection

The orchestrator automatically collects and updates metrics every 30 seconds:

- **Workload Metrics**: Count by status (running, pending, failed)
- **Performance Metrics**: Request rate, latency, error rate
- **Resource Metrics**: CPU, memory, GPU utilization
- **System Metrics**: Uptime, version, timestamp

### Health Monitoring

Built-in health checks monitor:
- API endpoint responsiveness
- Internal component status
- Resource availability
- Configuration validity

### Logging

Structured JSON logging with configurable levels:
- **DEBUG**: Detailed debugging information
- **INFO**: General operational messages
- **WARN**: Warning conditions
- **ERROR**: Error conditions

## 🚀 Production Deployment

### Performance Tuning

For production deployments:

1. **Increase resource limits**:
```yaml
performance:
  maxConcurrentRequests: 5000
  requestTimeout: "60s"
```

2. **Enable advanced monitoring**:
```yaml
monitoring:
  enabled: true
  metricsInterval: "15s"
```

3. **Configure logging**:
```yaml
logging:
  level: "warn"
  format: "json"
```

### Security Considerations

- Run as non-root user
- Use TLS for external communication
- Implement authentication (future feature)
- Regular security updates

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](../CONTRIBUTING.md) for details.

### Development Setup

1. Clone the repository
2. Install Go 1.24+
3. Run tests: `go test ./...`
4. Build: `go build -o bin/llm-orchestrator-simple.exe ./cmd/llm-orchestrator-simple`

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.

## 🆘 Support

- 📖 [Documentation](https://github.com/DimaJoyti/go-coffee/docs)
- 🐛 [Issue Tracker](https://github.com/DimaJoyti/go-coffee/issues)
- 💬 [Discussions](https://github.com/DimaJoyti/go-coffee/discussions)

---

**Built with ❤️ by the Go Coffee Team**
