# Enhanced Beverage Inventor Agent v2.0

An enterprise-grade AI-powered agent that creates innovative beverage recipes using advanced machine learning, comprehensive analysis, and intelligent optimization. The agent has been significantly enhanced with advanced domain services for nutritional analysis, cost calculation, ingredient compatibility, and recipe optimization.

## Architecture

This agent implements Clean Architecture with the following layers:

### Domain Layer (`internal/domain/`)
- **Entities**: Core business objects (Beverage, Ingredient, etc.)
- **Services**: Business logic services (BeverageGeneratorService)
- **Repositories**: Interface definitions for data access

### Application Layer (`internal/application/`)
- **Use Cases**: Application-specific business rules (BeverageInventorUseCase)

### Infrastructure Layer (`internal/infrastructure/`)
- **Config**: Configuration management
- **Kafka**: Message broker integration
- **Logger**: Structured logging implementation

### Interface Layer (`internal/interfaces/`)
- **Handlers**: Kafka and HTTP request handlers

## 🚀 Enhanced Features

### Core Capabilities
- **🧠 AI-Powered Recipe Generation**: Uses advanced AI models to create unique beverage combinations
- **🧪 Comprehensive Nutritional Analysis**: Detailed health scoring, dietary compatibility, and allergen detection
- **💰 Advanced Cost Calculation**: Real-time pricing, supplier integration, and profitability analysis
- **🔬 Ingredient Compatibility Analysis**: AI-powered flavor harmony and chemical compatibility checking
- **⚡ Multi-Objective Recipe Optimization**: Genetic algorithms for taste, cost, and nutrition optimization
- **📊 Market Intelligence**: Customer preference prediction and market fit analysis
- **🎯 Personalized Recommendations**: Dietary profile-based suggestions and enhancements

### Advanced Domain Services
- **Nutritional Analyzer**: Health scoring (0-100), glycemic index calculation, inflammation analysis
- **Cost Calculator**: Supplier pricing, bulk discounts, seasonal variations, profitability analysis
- **Compatibility Analyzer**: Flavor synergies, conflict detection, substitution suggestions
- **Recipe Optimizer**: Pareto optimization, genetic algorithms, market alignment

### Integration & Architecture
- **Event-Driven Architecture**: Uses Kafka for inter-agent communication
- **Enhanced Task Management**: Rich barista tasks with comprehensive analysis results
- **Smart Notifications**: Intelligent Slack notifications with visual indicators
- **Clean Architecture**: Modular, testable, and maintainable design
- **Mock Databases**: Complete mock implementations for development and testing

## Configuration

The agent is configured via YAML files and environment variables. See `configs/config.yaml` for the complete configuration structure.

### Key Environment Variables

```bash
# AI Providers
GEMINI_API_KEY=your-gemini-api-key
OPENAI_API_KEY=your-openai-api-key

# External Services
CLICKUP_API_KEY=your-clickup-api-key
SLACK_BOT_TOKEN=your-slack-bot-token

# Database
DATABASE_URL=postgres://user:pass@localhost/go_coffee

# Kafka
KAFKA_BROKERS=localhost:9092
```

## Running the Agent

### Prerequisites

1. Go 1.22+
2. Kafka cluster running
3. PostgreSQL database (optional)
4. AI provider API keys

### Development

```bash
# Navigate to the agent directory
cd ai-agents/beverage-inventor-agent

# Install dependencies
go mod tidy

# Run the agent
go run cmd/main.go
```

### Production

```bash
# Build the binary
go build -o bin/beverage-inventor-agent cmd/main.go

# Run with configuration
./bin/beverage-inventor-agent
```

## API

### Kafka Topics

#### Consumed Topics
- `recipe.requests`: Recipe creation requests
- `ingredient.discovered`: New ingredient discovery events

#### Published Topics
- `beverage.created`: New beverage creation events
- `beverage.updated`: Beverage update events
- `task.created`: Task creation events

### Event Schemas

#### Recipe Request Event
```json
{
  "request_id": "uuid",
  "ingredients": ["espresso", "dragon fruit"],
  "theme": "Mars Base",
  "requested_by": "user-id",
  "requested_at": "2024-01-01T12:00:00Z",
  "use_ai": true,
  "constraints": {
    "max_cost": 5.00,
    "max_calories": 200,
    "allergen_free": ["dairy"]
  }
}
```

#### Beverage Created Event
```json
{
  "beverage_id": "uuid",
  "name": "Martian Dragon Fruit Elixir",
  "description": "A cosmic concoction featuring dragon fruit...",
  "theme": "Mars Base",
  "ingredients": [
    {
      "name": "dragon fruit",
      "quantity": 50.0,
      "unit": "g",
      "source": "Local Supplier",
      "cost": 2.50
    }
  ],
  "created_by": "user-id",
  "created_at": "2024-01-01T12:00:00Z",
  "estimated_cost": 3.75,
  "metadata": {
    "preparation_time": 5,
    "difficulty": "Easy",
    "tags": ["AI-Generated", "Space-Themed"]
  }
}
```

## Testing

```bash
# Run unit tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/domain/services/...
```

## Monitoring

The agent provides structured logging and metrics for monitoring:

- **Logs**: JSON-formatted logs with correlation IDs
- **Metrics**: Custom metrics for beverage creation, AI usage, and errors
- **Health Checks**: HTTP endpoint for health monitoring

## Development

### Adding New AI Providers

1. Implement the `AIProvider` interface in `internal/domain/repositories/`
2. Add the provider implementation in `internal/infrastructure/ai/`
3. Update the configuration to include the new provider

### Adding New External Integrations

1. Define the interface in `internal/domain/repositories/`
2. Implement the integration in `internal/infrastructure/external/`
3. Update the dependency injection in `cmd/main.go`

## Architecture Decisions

### Why Clean Architecture?

- **Testability**: Easy to unit test business logic
- **Maintainability**: Clear separation of concerns
- **Flexibility**: Easy to swap implementations
- **Independence**: Business logic independent of frameworks

### Why Event-Driven?

- **Scalability**: Loose coupling between agents
- **Reliability**: Message persistence and replay
- **Flexibility**: Easy to add new consumers
- **Observability**: Clear audit trail of events

### Why Multiple AI Providers?

- **Reliability**: Fallback options if one provider fails
- **Cost Optimization**: Use different providers for different use cases
- **Feature Diversity**: Leverage unique capabilities of each provider
- **Vendor Independence**: Avoid lock-in to a single provider
