# Redis 8 Visual Interface & Query Builder

## 🎯 Overview

The Redis 8 Visual Interface provides a comprehensive, AI-powered solution for interacting with Redis data through natural language queries and visual exploration. Built with modern web technologies and optimized for Redis 8's advanced search capabilities.

## ✨ Features

### 🔍 **Data Explorer**
- **Interactive Key Browser**: Browse Redis keys with real-time filtering and search
- **Multi-Data Structure Support**: Visualize strings, hashes, lists, sets, sorted sets, and streams
- **Advanced Search**: Pattern-based key discovery with wildcard support
- **Memory Usage Analysis**: Track memory consumption per key
- **TTL Management**: Monitor and manage key expiration times

### 🛠️ **Visual Query Builder**
- **Drag-and-Drop Interface**: Build Redis commands visually without syntax knowledge
- **Operation Templates**: Pre-built templates for common Redis operations
- **Query Validation**: Real-time validation with helpful error messages
- **Command Preview**: See generated Redis commands before execution
- **Query History**: Save and reuse frequently used queries

### 📊 **Real-Time Monitoring**
- **Live Metrics Dashboard**: Monitor Redis performance in real-time
- **Command Tracking**: Track command execution and performance
- **Memory Analytics**: Visualize memory usage patterns
- **Connection Monitoring**: Monitor client connections and activity

### 🤖 **AI-Powered Features**
- **Natural Language Queries**: Convert plain English to Redis commands
- **Smart Suggestions**: AI-powered query completion and optimization
- **Performance Insights**: Automated performance recommendations
- **Anomaly Detection**: Identify unusual patterns in data access

## 🏗️ Architecture

### Backend Components

#### **Redis MCP Server** (`pkg/redis-mcp/`)
- **Visual Query Builder** (`visual_query_builder.go`): Core query building logic
- **Data Handlers** (`visual_handlers.go`): REST API endpoints for data operations
- **WebSocket Support** (`websocket_handlers.go`): Real-time data streaming
- **Performance Monitoring**: Built-in metrics and observability

#### **API Endpoints**

```
POST /api/v1/redis-mcp/visual/explore          # Data exploration
GET  /api/v1/redis-mcp/visual/keys             # Key listing
GET  /api/v1/redis-mcp/visual/key/:key         # Key details
GET  /api/v1/redis-mcp/visual/search           # Data search

POST /api/v1/redis-mcp/visual/query/build      # Query building
POST /api/v1/redis-mcp/visual/query/validate   # Query validation
GET  /api/v1/redis-mcp/visual/query/templates  # Query templates
GET  /api/v1/redis-mcp/visual/query/suggestions # Query suggestions

POST /api/v1/redis-mcp/visual/visualize        # Data visualization
GET  /api/v1/redis-mcp/visual/metrics          # Redis metrics
GET  /api/v1/redis-mcp/visual/performance      # Performance metrics

GET  /api/v1/redis-mcp/visual/stream           # WebSocket connection
POST /api/v1/redis-mcp/visual/stream/subscribe # Stream subscription
```

### Frontend Components

#### **React Components** (`web-ui/frontend/src/components/redis/`)
- **RedisDashboard** (`redis-dashboard.tsx`): Main dashboard interface
- **RedisExplorer** (`redis-explorer.tsx`): Data exploration component
- **QueryBuilder** (`query-builder.tsx`): Visual query building interface

#### **Custom Hooks** (`web-ui/frontend/src/hooks/`)
- **useRedisData**: Data fetching and management
- **useRedisQuery**: Query building and execution
- **useRedisVisualization**: Metrics and visualization data

## 🚀 Getting Started

### Prerequisites

- **Docker & Docker Compose**: For containerized deployment
- **Go 1.21+**: For local development
- **Node.js 18+**: For frontend development
- **Redis 8**: With required modules (RediSearch, RedisJSON, etc.)

### Quick Start with Docker

1. **Start Redis 8 with modules**:
```bash
docker-compose -f docker-compose.redis8.yml up -d
```

2. **Access the interfaces**:
- **Web UI**: http://localhost:3000
- **Redis MCP API**: http://localhost:8080
- **RedisInsight**: http://localhost:8001

### Local Development

1. **Start Redis 8**:
```bash
# Using Docker
docker run -d --name redis8 \
  -p 6379:6379 \
  redis/redis-stack:latest

# Or using local Redis with modules
redis-server --loadmodule /path/to/redisearch.so \
             --loadmodule /path/to/rejson.so \
             --loadmodule /path/to/redistimeseries.so
```

2. **Start the MCP Server**:
```bash
cd cmd/redis-mcp-server
go run main.go
```

3. **Start the Web UI**:
```bash
cd web-ui/frontend
npm install
npm run dev
```

## 🔧 Configuration

### Redis Configuration (`config/redis.conf`)

```conf
# Redis 8 Module Configuration
loadmodule /opt/redis-stack/lib/redisearch.so
loadmodule /opt/redis-stack/lib/rejson.so
loadmodule /opt/redis-stack/lib/redistimeseries.so
loadmodule /opt/redis-stack/lib/redisbloom.so
loadmodule /opt/redis-stack/lib/redisgraph.so

# AI Search Configuration
search.default_dialect 2
search.gc_scansize 100
search.workers 4

# Performance Settings
maxmemory 2gb
maxmemory-policy allkeys-lru
```

### Environment Variables

```bash
# Redis Connection
REDIS_URL=redis://localhost:6379
REDIS_PASSWORD=your_password
REDIS_DB=0

# Server Configuration
PORT=8080
LOG_LEVEL=info

# Frontend Configuration
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080
```

## 📖 Usage Examples

### Data Exploration

```javascript
// Explore keys with pattern
const keys = await exploreKeys({ 
  pattern: "user:*", 
  limit: 100 
});

// Get detailed key information
const details = await getKeyDetails("user:123");

// Search within data
const results = await searchData("john", "values");
```

### Query Building

```javascript
// Build a query visually
const query = await buildQuery({
  operation: "HSET",
  key: "user:123",
  field: "name",
  value: "John Doe",
  preview: true
});

// Execute the query
const result = await executeQuery({
  operation: "HGET",
  key: "user:123",
  field: "name"
});
```

### Real-Time Monitoring

```javascript
// Subscribe to key changes
const subscription = await subscribeToStream({
  keys: ["user:*"],
  events: ["keyspace", "commands"]
});

// Get performance metrics
const metrics = await getPerformanceMetrics();
```

## 🎨 UI Components

### Dashboard Overview
- **Metrics Cards**: Key performance indicators
- **Keyspace Overview**: Database statistics
- **Quick Actions**: Common operations shortcuts
- **Performance Graphs**: Real-time charts

### Data Explorer
- **Key Browser**: Hierarchical key navigation
- **Data Viewer**: Type-specific data visualization
- **Search Interface**: Advanced filtering options
- **Export Tools**: Data export capabilities

### Query Builder
- **Visual Editor**: Drag-and-drop query construction
- **Template Library**: Pre-built query templates
- **Validation Engine**: Real-time error checking
- **Execution Panel**: Query results and history

## 🔒 Security Features

- **Input Validation**: Comprehensive input sanitization
- **Rate Limiting**: Protection against abuse
- **Authentication**: JWT-based access control
- **CORS Configuration**: Secure cross-origin requests
- **Audit Logging**: Complete operation tracking

## 🚀 Performance Optimizations

- **Connection Pooling**: Efficient Redis connection management
- **Query Caching**: Intelligent result caching
- **Lazy Loading**: On-demand data fetching
- **WebSocket Streaming**: Real-time updates
- **Compression**: Optimized data transfer

## 🧪 Testing

```bash
# Run backend tests
go test ./pkg/redis-mcp/...

# Run frontend tests
cd web-ui/frontend
npm test

# Integration tests
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

## 📊 Monitoring & Observability

- **Prometheus Metrics**: Built-in metrics export
- **Structured Logging**: JSON-formatted logs
- **Health Checks**: Comprehensive health endpoints
- **Tracing**: Distributed tracing support
- **Alerting**: Performance threshold monitoring

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Implement your changes
4. Add comprehensive tests
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.

## 🆘 Support

- **Documentation**: [Full API Documentation](./API.md)
- **Issues**: [GitHub Issues](https://github.com/DimaJoyti/go-coffee/issues)
- **Discussions**: [GitHub Discussions](https://github.com/DimaJoyti/go-coffee/discussions)
- **Email**: support@go-coffee.dev
