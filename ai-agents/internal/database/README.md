# Database Layer & Repository Pattern

This package provides a comprehensive database layer implementation using the Repository pattern with PostgreSQL, featuring connection management, migrations, observability, and clean architecture principles.

## Overview

The database layer implements:

1. **Connection Management**: Thread-safe database connections with pooling
2. **Repository Pattern**: Clean separation of data access logic
3. **Migration System**: Version-controlled database schema management
4. **Observability**: Full tracing, metrics, and logging integration
5. **Transaction Support**: ACID transaction management
6. **Health Monitoring**: Database health checks and statistics

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Database        │    │ Connection      │    │ Migration       │
│ Manager         │───▶│ Manager         │    │ Manager         │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Repository      │    │ Connection Pool │    │ Schema          │
│ Implementations │    │ & Health Checks │    │ Migrations      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Domain Entities │    │ PostgreSQL      │    │ Version Control │
│ & Business Logic│    │ Database        │    │ & Rollbacks     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Components

### 1. Database Manager (`manager.go`)

Central coordinator for all database operations:

```go
// Initialize database manager
manager := database.NewManager(dbConfig, logger, metrics, tracing)
if err := manager.Initialize(ctx); err != nil {
    log.Fatal("Failed to initialize database:", err)
}

// Get repositories
beverageRepo := manager.GetBeverageRepository()

// Start transaction
tx, err := manager.BeginTransaction(ctx)
if err != nil {
    return err
}
defer tx.Rollback(ctx)

// Use transactional repository
txBeverageRepo := tx.GetBeverageRepository()
if err := txBeverageRepo.Create(ctx, beverage); err != nil {
    return err
}

// Commit transaction
return tx.Commit(ctx)
```

### 2. Connection Manager (`connection.go`)

Manages PostgreSQL connections with observability:

```go
// Connection with observability
type DB struct {
    *sql.DB
    config  config.DatabaseConfig
    logger  *observability.StructuredLogger
    metrics *observability.MetricsCollector
    tracing *observability.TracingHelper
}

// All database operations are traced and logged
result, err := db.ExecContext(ctx, query, args...)
rows, err := db.QueryContext(ctx, query, args...)
row := db.QueryRowContext(ctx, query, args...)
```

#### Connection Pool Configuration
```go
// Connection pool settings
db.SetMaxOpenConns(config.MaxOpenConns)     // Maximum open connections
db.SetMaxIdleConns(config.MaxIdleConns)     // Maximum idle connections
db.SetConnMaxLifetime(config.ConnMaxLifetime) // Connection lifetime
db.SetConnMaxIdleTime(config.ConnMaxIdleTime) // Idle connection timeout
```

### 3. Repository Pattern (`repository.go`, `beverage_repository.go`)

Clean data access layer with business-specific operations:

```go
// Base repository interface
type Repository interface {
    Create(ctx context.Context, entity interface{}) error
    GetByID(ctx context.Context, id interface{}, dest interface{}) error
    Update(ctx context.Context, entity interface{}) error
    Delete(ctx context.Context, id interface{}) error
    FindAll(ctx context.Context, dest interface{}) error
    FindWhere(ctx context.Context, condition string, args []interface{}, dest interface{}) error
    Count(ctx context.Context, condition string, args []interface{}) (int64, error)
    WithTx(tx *Tx) Repository
}

// Beverage-specific repository
type BeverageRepository interface {
    Repository
    
    // Business-specific operations
    FindByTheme(ctx context.Context, theme string) ([]*entities.Beverage, error)
    FindByCreator(ctx context.Context, createdBy string) ([]*entities.Beverage, error)
    FindByIngredient(ctx context.Context, ingredient string) ([]*entities.Beverage, error)
    FindRecent(ctx context.Context, limit int) ([]*entities.Beverage, error)
    FindPopular(ctx context.Context, limit int) ([]*entities.Beverage, error)
    UpdateRating(ctx context.Context, id uuid.UUID, rating float64) error
    IncrementViewCount(ctx context.Context, id uuid.UUID) error
    Search(ctx context.Context, query string, filters BeverageFilters) ([]*entities.Beverage, error)
    GetStatistics(ctx context.Context) (*BeverageStatistics, error)
}
```

### 4. Migration System (`migrations.go`)

Version-controlled database schema management:

```go
// Migration structure
type Migration struct {
    Version     int
    Name        string
    UpSQL       string
    DownSQL     string
    Description string
}

// Run migrations
migrationManager := NewMigrationManager(db, logger, tracing)
if err := migrationManager.Migrate(ctx); err != nil {
    return fmt.Errorf("migration failed: %w", err)
}

// Rollback last migration
if err := migrationManager.Rollback(ctx); err != nil {
    return fmt.Errorf("rollback failed: %w", err)
}
```

#### Available Migrations

| Version | Name | Description |
|---------|------|-------------|
| 1 | `create_schema_migrations_table` | Migration tracking table |
| 2 | `create_beverages_table` | Main beverages table with indexes |
| 3 | `create_beverage_ratings_table` | Individual beverage ratings |
| 4 | `create_beverage_tags_table` | Tagging system |
| 5 | `create_beverage_favorites_table` | User favorites |
| 6 | `add_beverage_analytics_columns` | Analytics columns |
| 7 | `create_beverage_preparation_logs_table` | Preparation tracking |
| 8 | `create_beverage_search_functions` | Full-text search |

## Database Schema

### Beverages Table

```sql
CREATE TABLE beverages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    theme VARCHAR(100) NOT NULL,
    ingredients JSONB NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_by VARCHAR(255) NOT NULL,
    rating DECIMAL(3,2) DEFAULT 0.0 CHECK (rating >= 0.0 AND rating <= 5.0),
    view_count INTEGER DEFAULT 0 CHECK (view_count >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Analytics columns (added in migration 6)
    last_viewed_at TIMESTAMP WITH TIME ZONE,
    favorite_count INTEGER DEFAULT 0,
    rating_count INTEGER DEFAULT 0,
    average_rating DECIMAL(3,2) DEFAULT 0.0,
    
    -- Search vector (added in migration 8)
    search_vector tsvector
);
```

### Indexes for Performance

```sql
-- Basic indexes
CREATE INDEX idx_beverages_theme ON beverages(theme);
CREATE INDEX idx_beverages_created_by ON beverages(created_by);
CREATE INDEX idx_beverages_rating ON beverages(rating DESC);
CREATE INDEX idx_beverages_created_at ON beverages(created_at DESC);

-- JSON indexes
CREATE INDEX idx_beverages_ingredients_gin ON beverages USING GIN (ingredients);
CREATE INDEX idx_beverages_metadata_gin ON beverages USING GIN (metadata);

-- Composite indexes
CREATE INDEX idx_beverages_popularity ON beverages(rating DESC, view_count DESC, created_at DESC);

-- Full-text search
CREATE INDEX idx_beverages_search_vector ON beverages USING GIN (search_vector);
```

## Usage Examples

### Basic Repository Operations

```go
// Create a new beverage
beverage := &entities.Beverage{
    ID:          uuid.New(),
    Name:        "Cosmic Coffee",
    Description: "A stellar blend for space explorers",
    Theme:       "Mars Base",
    Ingredients: []entities.Ingredient{
        {Name: "Coffee Beans", Quantity: "200g", Type: "Base"},
        {Name: "Martian Dust", Quantity: "1tsp", Type: "Flavor"},
    },
    CreatedBy:   "beverage-inventor-agent",
    CreatedAt:   time.Now(),
    UpdatedAt:   time.Now(),
}

// Save to database
if err := beverageRepo.Create(ctx, beverage); err != nil {
    return fmt.Errorf("failed to create beverage: %w", err)
}

// Retrieve by ID
var retrievedBeverage entities.Beverage
if err := beverageRepo.GetByID(ctx, beverage.ID, &retrievedBeverage); err != nil {
    if err == database.ErrNotFound {
        return fmt.Errorf("beverage not found")
    }
    return fmt.Errorf("failed to get beverage: %w", err)
}

// Update beverage
retrievedBeverage.Rating = 4.5
if err := beverageRepo.Update(ctx, &retrievedBeverage); err != nil {
    return fmt.Errorf("failed to update beverage: %w", err)
}
```

### Business-Specific Queries

```go
// Find beverages by theme
marsBeverages, err := beverageRepo.FindByTheme(ctx, "Mars Base")
if err != nil {
    return fmt.Errorf("failed to find Mars beverages: %w", err)
}

// Find recent beverages
recentBeverages, err := beverageRepo.FindRecent(ctx, 10)
if err != nil {
    return fmt.Errorf("failed to find recent beverages: %w", err)
}

// Find popular beverages
popularBeverages, err := beverageRepo.FindPopular(ctx, 5)
if err != nil {
    return fmt.Errorf("failed to find popular beverages: %w", err)
}

// Search by ingredient
coffeeBeverages, err := beverageRepo.FindByIngredient(ctx, "coffee")
if err != nil {
    return fmt.Errorf("failed to find coffee beverages: %w", err)
}

// Update rating
if err := beverageRepo.UpdateRating(ctx, beverageID, 4.8); err != nil {
    return fmt.Errorf("failed to update rating: %w", err)
}

// Increment view count
if err := beverageRepo.IncrementViewCount(ctx, beverageID); err != nil {
    return fmt.Errorf("failed to increment view count: %w", err)
}
```

### Transaction Management

```go
// Start transaction
tx, err := dbManager.BeginTransaction(ctx)
if err != nil {
    return fmt.Errorf("failed to begin transaction: %w", err)
}
defer tx.Rollback(ctx) // Rollback if not committed

// Use transactional repositories
txBeverageRepo := tx.GetBeverageRepository()

// Create beverage
if err := txBeverageRepo.Create(ctx, beverage); err != nil {
    return fmt.Errorf("failed to create beverage: %w", err)
}

// Update related data
if err := txBeverageRepo.UpdateRating(ctx, beverage.ID, 5.0); err != nil {
    return fmt.Errorf("failed to update rating: %w", err)
}

// Commit transaction
if err := tx.Commit(ctx); err != nil {
    return fmt.Errorf("failed to commit transaction: %w", err)
}
```

### Health Monitoring

```go
// Database health check
healthStatus, err := dbManager.HealthCheck(ctx)
if err != nil {
    log.Printf("Health check failed: %v", err)
    return
}

if !healthStatus.Healthy {
    log.Printf("Database unhealthy: %s", healthStatus.Error)
    return
}

log.Printf("Database healthy - response time: %v", healthStatus.ResponseTime)

// Get database statistics
stats, err := dbManager.GetStatistics(ctx)
if err != nil {
    log.Printf("Failed to get statistics: %v", err)
    return
}

log.Printf("Connection pool: %d open, %d in use, %d idle", 
    stats.ConnectionPool.OpenConnections,
    stats.ConnectionPool.InUse,
    stats.ConnectionPool.Idle)
```

## Configuration

### Database Configuration

```yaml
database:
  driver: postgres
  host: localhost
  port: 5432
  database: gocoffee_dev
  username: gocoffee
  password: password
  ssl_mode: disable
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m
  conn_max_idle_time: 5m
  migrations_path: ./migrations
  enable_logging: true
```

### Environment Variables

```bash
# Database connection
GOCOFFEE_DATABASE_HOST=localhost
GOCOFFEE_DATABASE_PORT=5432
GOCOFFEE_DATABASE_DATABASE=gocoffee_dev
GOCOFFEE_DATABASE_USERNAME=gocoffee
GOCOFFEE_DATABASE_PASSWORD=password

# Connection pool
GOCOFFEE_DATABASE_MAX_OPEN_CONNS=25
GOCOFFEE_DATABASE_MAX_IDLE_CONNS=5
GOCOFFEE_DATABASE_CONN_MAX_LIFETIME=5m

# Security
GOCOFFEE_DATABASE_SSL_MODE=require  # For production
```

## Observability Integration

### Tracing

All database operations are automatically traced:

```go
// Automatic span creation for database operations
ctx, span := tracing.StartDatabaseSpan(ctx, "INSERT", "beverages")
defer span.End()

// Query information added to spans
tracing.SetAttributes(span,
    observability.Attribute("db.statement", query),
    observability.Attribute("db.operation", "insert"))

// Error recording
if err != nil {
    tracing.RecordError(span, err, "Database operation failed")
    return err
}

tracing.RecordSuccess(span, "Database operation completed")
```

### Metrics

Database metrics are automatically collected:

- `database_operations_total`: Total database operations
- `database_query_duration_seconds`: Query execution time
- Connection pool metrics (open, idle, in-use connections)
- Table-level statistics (inserts, updates, deletes)

### Logging

Structured logging with trace correlation:

```json
{
  "timestamp": "2024-01-01T12:00:00Z",
  "level": "info",
  "message": "Beverage created successfully",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",
  "beverage_id": "123e4567-e89b-12d3-a456-426614174000",
  "name": "Cosmic Coffee",
  "theme": "Mars Base",
  "duration_ms": 15
}
```

## Best Practices

### 1. Repository Pattern
- Keep repositories focused on data access
- Implement business-specific query methods
- Use interfaces for testability
- Support transactions through `WithTx()`

### 2. Error Handling
- Use typed errors (`ErrNotFound`, `ErrConflict`)
- Wrap errors with context
- Log errors with relevant metadata
- Handle database-specific errors appropriately

### 3. Performance
- Use appropriate indexes for query patterns
- Implement connection pooling
- Monitor query performance
- Use transactions for related operations

### 4. Security
- Use parameterized queries to prevent SQL injection
- Enable SSL/TLS in production
- Implement proper access controls
- Audit sensitive operations

### 5. Migrations
- Keep migrations small and focused
- Test migrations on staging data
- Provide rollback scripts
- Document breaking changes

This database layer provides a robust, observable, and maintainable foundation for data persistence in the Go Coffee AI agent ecosystem, following clean architecture principles and modern best practices.
