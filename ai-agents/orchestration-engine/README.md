# Go Coffee AI Agents Orchestration Engine

A sophisticated workflow orchestration engine that coordinates multiple AI agents to execute complex business processes. Built with Clean Architecture principles and designed for enterprise-scale automation.

## 🚀 Overview

The Orchestration Engine is the central nervous system of the Go Coffee AI Agents ecosystem. It manages workflow definitions, coordinates agent execution, handles error recovery, and provides comprehensive monitoring and analytics.

### Key Features

- **Visual Workflow Designer**: Define complex workflows with drag-and-drop interface
- **Multi-Agent Coordination**: Seamlessly orchestrate multiple specialized AI agents
- **Advanced Error Handling**: Sophisticated retry policies, fallback strategies, and error recovery
- **Real-time Monitoring**: Live workflow execution tracking with detailed metrics
- **Conditional Logic**: Complex branching and decision-making capabilities
- **Event-Driven Architecture**: Reactive workflows triggered by events, schedules, or APIs
- **Scalable Execution**: Horizontal scaling with configurable concurrency limits
- **Audit Trail**: Complete execution history and compliance tracking

## 🏗️ Architecture

### Core Components

```
┌─────────────────────────────────────────────────────────────┐
│                 Orchestration Engine                        │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  Workflow   │  │   Agent     │  │   Event     │        │
│  │   Engine    │  │  Registry   │  │ Publisher   │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  Workflow   │  │ Execution   │  │   Step      │        │
│  │ Repository  │  │ Repository  │  │  Executor   │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
├─────────────────────────────────────────────────────────────┤
│                    Agent Network                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │Social Media │  │  Feedback   │  │ Inventory   │        │
│  │Content Agent│  │Analyst Agent│  │   Agent     │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

### Workflow Definition Structure

```yaml
workflow:
  name: "AI-Powered Content Creation"
  type: "hybrid"
  triggers:
    - type: "schedule"
      schedule: "0 9 * * 1-5"
    - type: "event"
      event_type: "campaign.content_requested"
  
  steps:
    - id: "validate_input"
      type: "validation"
      parameters:
        required_fields: ["brand_id", "content_topic"]
    
    - id: "generate_content"
      type: "agent"
      agent_type: "social-media-content"
      action: "create_content"
      timeout: "5m"
      retry_policy:
        max_attempts: 3
        backoff_factor: 2.0
    
    - id: "quality_gate"
      type: "condition"
      conditions:
        - expression: "quality_score >= 75"
          operator: "greater_or_equal"
    
    - id: "optimize_content"
      type: "agent"
      agent_type: "social-media-content"
      action: "enhance_content"
      conditions:
        - expression: "quality_score < 75"
  
  connections:
    - from: "validate_input"
      to: "generate_content"
    - from: "generate_content"
      to: "quality_gate"
    - from: "quality_gate"
      to: "optimize_content"
      condition: "quality_score < 75"
```

## 🎯 Workflow Types

### 1. Sequential Workflows
Linear execution of steps in a predefined order.

**Use Cases:**
- Content creation pipelines
- Data processing workflows
- Approval processes

### 2. Parallel Workflows
Concurrent execution of independent steps.

**Use Cases:**
- Multi-platform content publishing
- Batch data processing
- Independent validation checks

### 3. Conditional Workflows
Dynamic routing based on runtime conditions.

**Use Cases:**
- Quality gates
- Approval routing
- Error handling

### 4. Event-Driven Workflows
Reactive workflows triggered by external events.

**Use Cases:**
- Real-time feedback processing
- Campaign automation
- Alert handling

### 5. Hybrid Workflows
Combination of multiple workflow types.

**Use Cases:**
- Complex business processes
- Multi-stage campaigns
- Comprehensive automation

## 🔧 Agent Integration

### Supported Agents

#### Social Media Content Agent
- **Actions**: `create_content`, `schedule_content`, `publish_content`, `analyze_content`
- **Capabilities**: AI content generation, platform optimization, scheduling
- **Input**: Content requirements, brand guidelines, target platforms
- **Output**: Generated content, scheduling confirmation, analytics

#### Feedback Analyst Agent
- **Actions**: `analyze_feedback`, `generate_response`, `categorize_feedback`
- **Capabilities**: Sentiment analysis, response generation, trend identification
- **Input**: Customer feedback, review data, social mentions
- **Output**: Analysis results, response suggestions, insights

#### Inventory Agent
- **Actions**: `check_stock`, `update_inventory`, `forecast_demand`
- **Capabilities**: Stock management, demand forecasting, reorder automation
- **Input**: Product IDs, quantities, sales data
- **Output**: Stock levels, reorder recommendations, forecasts

#### Notifier Agent
- **Actions**: `send_notification`, `create_alert`, `update_status`
- **Capabilities**: Multi-channel notifications, alert management
- **Input**: Message content, recipients, channels
- **Output**: Delivery confirmation, status updates

### Agent Registration

```go
// Register agents with the orchestration engine
agentRegistry := services.NewDefaultAgentRegistry(logger)

socialMediaAgent := services.NewSocialMediaContentAgent(
    "http://localhost:8081", 
    httpClient, 
    logger,
)
agentRegistry.RegisterAgent("social-media-content", socialMediaAgent)

feedbackAgent := services.NewFeedbackAnalystAgent(
    "http://localhost:8082", 
    httpClient, 
    logger,
)
agentRegistry.RegisterAgent("feedback-analyst", feedbackAgent)
```

## 📊 Monitoring & Analytics

### Workflow Metrics
- **Execution Count**: Total, successful, failed executions
- **Performance**: Average execution time, throughput
- **Success Rate**: Percentage of successful completions
- **Error Rate**: Failure analysis and trends
- **Resource Usage**: CPU, memory, network utilization

### Agent Health Monitoring
- **Status**: Online, offline, busy, error states
- **Response Time**: Average and percentile metrics
- **Load**: Current processing load and capacity
- **Error Rate**: Agent-specific failure rates
- **Availability**: Uptime and reliability metrics

### Real-time Dashboard
- Live workflow execution status
- Agent health and performance
- System resource utilization
- Alert and notification center
- Historical trends and analytics

## 🔒 Security & Compliance

### Authentication & Authorization
- **JWT-based Authentication**: Secure API access
- **Role-Based Access Control**: Granular permissions
- **API Key Management**: Secure agent communication
- **Audit Logging**: Complete action tracking

### Data Protection
- **Encryption**: Data at rest and in transit
- **PII Handling**: Privacy-compliant data processing
- **Secure Storage**: Encrypted workflow definitions
- **Access Controls**: Restricted data access

### Compliance Features
- **Audit Trail**: Complete execution history
- **Data Retention**: Configurable retention policies
- **Compliance Reporting**: Automated compliance checks
- **Change Management**: Version control and approvals

## 🚀 Getting Started

### Prerequisites
- Go 1.22 or later
- Docker & Docker Compose (recommended)
- PostgreSQL 14+ (if running locally)
- Redis 6+ (if running locally)
- Apache Kafka 3.0+ (if running locally)

### Quick Start with Docker

1. **Clone the repository**
```bash
git clone https://github.com/your-org/go-coffee-ai-agents.git
cd go-coffee-ai-agents/ai-agents/orchestration-engine
```

2. **Start the complete stack**
```bash
docker-compose up -d
```

3. **Verify services are running**
```bash
# Check orchestration engine
curl http://localhost:8080/health

# Check WebSocket endpoint
curl http://localhost:8080/ws

# View logs
docker-compose logs -f orchestration-engine
```

### Local Development Setup

1. **Install dependencies**
```bash
go mod download
```

2. **Set up environment variables**
```bash
export PORT=8080
export DATABASE_URL=postgres://localhost/orchestration
export REDIS_URL=redis://localhost:6379
export KAFKA_BROKERS=localhost:9092
export LOG_LEVEL=info
export MAX_CONCURRENT_WORKFLOWS=100
```

3. **Start the orchestration engine**
```bash
go run cmd/server/main.go
```

### Quick Start Example

```bash
# Create a workflow
curl -X POST http://localhost:8080/api/v1/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Content Creation Workflow",
    "type": "sequential",
    "steps": [...]
  }'

# Execute a workflow
curl -X POST http://localhost:8080/api/v1/workflows/execute \
  -H "Content-Type: application/json" \
  -d '{
    "workflow_id": "uuid",
    "input": {
      "brand_id": "uuid",
      "content_topic": "New Product Launch"
    }
  }'

# Check execution status
curl http://localhost:8080/api/v1/executions/{execution_id}
```

## 📚 API Documentation

### Workflow Management

#### Create Workflow
```http
POST /api/v1/workflows
Content-Type: application/json

{
  "name": "Workflow Name",
  "description": "Workflow Description",
  "type": "sequential",
  "definition": {...}
}
```

#### Execute Workflow
```http
POST /api/v1/workflows/{id}/execute
Content-Type: application/json

{
  "input": {...},
  "trigger_type": "manual"
}
```

#### Get Execution Status
```http
GET /api/v1/executions/{id}
```

### Agent Management

#### List Agents
```http
GET /api/v1/agents
```

#### Get Agent Health
```http
GET /api/v1/agents/{type}/health
```

### Monitoring

#### Get Workflow Metrics
```http
GET /api/v1/workflows/{id}/metrics
```

#### Get System Status
```http
GET /api/v1/status
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://localhost/orchestration` |
| `REDIS_URL` | Redis connection string | `redis://localhost:6379` |
| `KAFKA_BROKERS` | Kafka broker addresses | `localhost:9092` |
| `LOG_LEVEL` | Logging level | `info` |
| `MAX_CONCURRENT_WORKFLOWS` | Maximum concurrent executions | `100` |

### Agent Endpoints

| Agent | Environment Variable | Default |
|-------|---------------------|---------|
| Social Media Content | `SOCIAL_MEDIA_AGENT_URL` | `http://localhost:8081` |
| Feedback Analyst | `FEEDBACK_ANALYST_AGENT_URL` | `http://localhost:8082` |
| Inventory | `INVENTORY_AGENT_URL` | `http://localhost:8083` |
| Notifier | `NOTIFIER_AGENT_URL` | `http://localhost:8084` |

## 🧪 Testing

### Unit Tests
```bash
go test ./internal/...
```

### Integration Tests
```bash
go test -tags=integration ./tests/integration/...
```

### Load Testing
```bash
# Using k6
k6 run tests/load/workflow_execution.js
```

## 📈 Performance

### Benchmarks
- **Workflow Creation**: < 10ms
- **Execution Start**: < 50ms
- **Step Execution**: < 100ms (excluding agent processing)
- **Throughput**: 1000+ workflows/minute
- **Concurrent Executions**: 100+ simultaneous workflows

### Optimization
- Connection pooling for database and Redis
- Async processing for non-blocking operations
- Efficient workflow state management
- Optimized agent communication

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines
- Follow Clean Architecture principles
- Write comprehensive tests
- Use conventional commits
- Update documentation
- Ensure observability

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/your-org/go-coffee-ai-agents/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-org/go-coffee-ai-agents/discussions)
- **Email**: support@gocoffee.com

---

**Built with ❤️ by the Go Coffee AI Team**
