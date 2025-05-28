# Content Analysis System Documentation

## Overview

The Content Analysis System is a comprehensive solution for collecting, analyzing, and processing Reddit content using advanced AI techniques including LLM-based classification, sentiment analysis, topic modeling, and Retrieval-Augmented Generation (RAG).

## Architecture

### Core Components

1. **Content Analysis Service** - Main HTTP service providing APIs for content analysis
2. **Content Analysis Agent** - Kafka-based agent for processing Reddit content
3. **Reddit Integration** - Client for collecting content from Reddit API
4. **AI Analysis Engine** - LLM-powered content classification and analysis
5. **RAG System** - Retrieval-Augmented Generation for enhanced content understanding
6. **Vector Database** - Storage for embeddings and semantic search
7. **Message Queue** - Kafka for asynchronous processing
8. **Caching Layer** - Redis for performance optimization

### Technology Stack

- **Backend**: Go 1.21+
- **AI/ML**: Gemini Pro, OpenAI, Ollama
- **Vector DB**: Qdrant
- **Message Queue**: Apache Kafka
- **Cache**: Redis
- **Database**: PostgreSQL
- **Monitoring**: Prometheus + Grafana
- **Containerization**: Docker + Docker Compose

## Features

### Content Collection
- **Reddit API Integration**: Automated collection from multiple subreddits
- **Real-time Processing**: Continuous monitoring and analysis
- **Rate Limiting**: Compliant with Reddit API limits
- **Content Filtering**: Configurable filters for quality and relevance

### AI-Powered Analysis
- **Content Classification**: Automatic categorization into predefined categories
- **Sentiment Analysis**: Emotion and sentiment detection
- **Topic Modeling**: Extraction of key themes and topics
- **Trend Analysis**: Identification of emerging trends and viral content

### RAG System
- **Semantic Search**: Vector-based content retrieval
- **Context-Aware Generation**: Enhanced responses using retrieved context
- **Iterative Refinement**: Continuous improvement through feedback
- **Multi-modal Support**: Text, images, and metadata processing

### Scalability & Performance
- **Horizontal Scaling**: Microservices architecture
- **Asynchronous Processing**: Kafka-based event streaming
- **Caching Strategy**: Multi-level caching for optimal performance
- **Load Balancing**: Nginx reverse proxy with SSL termination

## Quick Start

### Prerequisites

```bash
# Required software
- Docker 20.10+
- Docker Compose 2.0+
- Go 1.21+ (for development)
```

### Environment Setup

1. **Clone the repository**:
```bash
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee
```

2. **Set up environment variables**:
```bash
cp .env.example .env
# Edit .env with your API keys and configuration
```

3. **Required API Keys**:
```bash
# Reddit API
REDDIT_CLIENT_ID=your_reddit_client_id
REDDIT_CLIENT_SECRET=your_reddit_client_secret
REDDIT_USERNAME=your_reddit_username
REDDIT_PASSWORD=your_reddit_password

# AI Services
GEMINI_API_KEY=your_gemini_api_key
OPENAI_API_KEY=your_openai_api_key

# Optional integrations
SLACK_WEBHOOK_URL=your_slack_webhook
VECTOR_DB_URL=your_vector_db_url
```

### Running the System

1. **Start all services**:
```bash
docker-compose -f docker-compose.content-analysis.yml up -d
```

2. **Verify services are running**:
```bash
# Check service status
curl http://localhost:8085/health

# View logs
docker-compose -f docker-compose.content-analysis.yml logs -f content-analysis-service
```

3. **Access web interfaces**:
- **Content Analysis API**: http://localhost:8085
- **Kafka UI**: http://localhost:8086
- **Grafana Dashboard**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090

## API Documentation

### Content Analysis Endpoints

#### Health Check
```http
GET /health
```

#### Service Status
```http
GET /status
```

#### Reddit Content Analysis
```http
POST /api/v1/reddit/analyze/post
Content-Type: application/json

{
  "id": "post_id",
  "title": "Post title",
  "content": "Post content",
  "subreddit": "MachineLearning",
  "author": "username",
  "score": 150
}
```

#### Content Classification
```http
POST /api/v1/analysis/classify
Content-Type: application/json

{
  "text": "Content to classify"
}
```

#### Sentiment Analysis
```http
POST /api/v1/analysis/sentiment
Content-Type: application/json

{
  "text": "Content to analyze"
}
```

#### RAG Query
```http
POST /api/v1/rag/query
Content-Type: application/json

{
  "query": "What are the latest trends in AI?"
}
```

### Response Format

All API responses follow this structure:
```json
{
  "status": "success|error",
  "data": {},
  "message": "Optional message",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## Configuration

### Content Analysis Agent Configuration

The agent is configured via `ai-agents/content-analysis-agent/config.yaml`:

```yaml
# Key configuration sections
reddit:
  enabled: true
  subreddits:
    - "MachineLearning"
    - "artificial"
    - "Coffee"
  poll_interval: "5m"

ai:
  provider: "gemini"
  model: "gemini-pro"
  temperature: 0.1

processing:
  enable_sentiment: true
  enable_topic_modeling: true
  enable_trend_analysis: true
  min_confidence: 0.6
```

### Subreddit Configuration

Add or modify subreddits to monitor:
```yaml
reddit:
  subreddits:
    # AI/Tech subreddits
    - "MachineLearning"
    - "artificial"
    - "technology"
    - "programming"
    
    # Coffee subreddits
    - "Coffee"
    - "espresso"
    - "roasting"
    
    # Custom subreddits
    - "your_custom_subreddit"
```

## Monitoring & Observability

### Metrics

The system exposes metrics for:
- **Content Processing**: Posts/comments processed per minute
- **AI Analysis**: Classification accuracy and confidence scores
- **System Performance**: Response times, error rates, throughput
- **Resource Usage**: CPU, memory, disk usage

### Dashboards

Pre-configured Grafana dashboards include:
- **Content Analysis Overview**: High-level system metrics
- **Reddit Collection**: Subreddit-specific collection stats
- **AI Performance**: Model accuracy and processing times
- **Infrastructure**: System resource utilization

### Alerts

Configurable alerts for:
- High error rates (>5% in 5 minutes)
- Processing delays (>30 seconds average)
- Low confidence scores (<0.5 average)
- System resource exhaustion

## Development

### Local Development Setup

1. **Install dependencies**:
```bash
go mod download
```

2. **Run tests**:
```bash
go test ./...
```

3. **Run locally**:
```bash
# Start dependencies
docker-compose -f docker-compose.content-analysis.yml up -d redis kafka postgres

# Run service
go run web3-wallet-backend/cmd/content-analysis-service/main.go

# Run agent
go run ai-agents/content-analysis-agent/main.go
```

### Adding New Analysis Features

1. **Extend the analyzer**:
```go
// Add new analysis method
func (a *Analyzer) AnalyzeCustomFeature(ctx context.Context, content string) (*CustomResult, error) {
    // Implementation
}
```

2. **Update configuration**:
```yaml
processing:
  enable_custom_feature: true
  custom_feature_config:
    threshold: 0.7
```

3. **Add API endpoint**:
```go
func (s *Service) analyzeCustomFeature(c *gin.Context) {
    // Handler implementation
}
```

## Troubleshooting

### Common Issues

1. **Reddit API Rate Limiting**:
   - Check rate limit configuration
   - Verify API credentials
   - Monitor request frequency

2. **AI Service Errors**:
   - Verify API keys are valid
   - Check model availability
   - Monitor token usage

3. **Kafka Connection Issues**:
   - Ensure Kafka is running
   - Check topic creation
   - Verify consumer group configuration

4. **Vector Database Issues**:
   - Check Qdrant connectivity
   - Verify index configuration
   - Monitor embedding generation

### Logs

View service logs:
```bash
# All services
docker-compose -f docker-compose.content-analysis.yml logs

# Specific service
docker-compose -f docker-compose.content-analysis.yml logs content-analysis-service

# Follow logs
docker-compose -f docker-compose.content-analysis.yml logs -f content-analysis-agent
```

## Security

### API Security
- Rate limiting on all endpoints
- Input validation and sanitization
- CORS configuration
- SSL/TLS encryption

### Data Protection
- Encrypted data at rest
- Secure API key management
- Audit logging
- Access control

### Reddit API Compliance
- Respect rate limits
- Follow Reddit API terms
- User privacy protection
- Content attribution

## Performance Optimization

### Caching Strategy
- Redis for frequently accessed data
- In-memory caching for hot paths
- TTL-based cache invalidation
- Cache warming strategies

### Database Optimization
- Indexed queries
- Connection pooling
- Query optimization
- Partitioning for large datasets

### AI Model Optimization
- Batch processing for efficiency
- Model caching
- Prompt optimization
- Token usage monitoring

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support and questions:
- Create an issue on GitHub
- Check the documentation
- Review existing issues and discussions
