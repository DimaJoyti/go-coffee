package redisadvanced

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// JSONClient provides advanced JSON operations using RedisJSON
type JSONClient struct {
	redis  *redis.Client
	logger *zap.Logger
	config *JSONConfig
}

// JSONConfig contains configuration for JSON operations
type JSONConfig struct {
	DefaultTTL time.Duration
	MaxDepth   int
	MaxSize    int64
}

// JSONDocument represents a JSON document with metadata
type JSONDocument struct {
	ID       string                 `json:"id"`
	Data     interface{}            `json:"data"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Version  int64                  `json:"version"`
	Created  time.Time              `json:"created"`
	Updated  time.Time              `json:"updated"`
}

// JSONQuery represents a JSON path query
type JSONQuery struct {
	Path      string      `json:"path"`
	Filter    interface{} `json:"filter,omitempty"`
	Transform interface{} `json:"transform,omitempty"`
}

// JSONUpdateOperation represents an update operation
type JSONUpdateOperation struct {
	Path      string      `json:"path"`
	Operation string      `json:"operation"` // SET, DELETE, APPEND, INCREMENT
	Value     interface{} `json:"value,omitempty"`
}

// NewJSONClient creates a new JSON client
func NewJSONClient(redisClient *redis.Client, logger *zap.Logger, config *JSONConfig) *JSONClient {
	if config == nil {
		config = &JSONConfig{
			DefaultTTL: 24 * time.Hour,
			MaxDepth:   10,
			MaxSize:    1024 * 1024, // 1MB
		}
	}

	return &JSONClient{
		redis:  redisClient,
		logger: logger,
		config: config,
	}
}

// Set stores a JSON document
func (jc *JSONClient) Set(ctx context.Context, key string, document interface{}, ttl time.Duration) error {
	jc.logger.Info("Setting JSON document", zap.String("key", key))

	// Serialize to JSON
	jsonData, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Check size limit
	if int64(len(jsonData)) > jc.config.MaxSize {
		return fmt.Errorf("document size exceeds limit: %d > %d", len(jsonData), jc.config.MaxSize)
	}

	// Use JSON.SET command
	cmd := redis.NewCmd(ctx, "JSON.SET", key, "$", string(jsonData))
	if err := jc.redis.Process(ctx, cmd); err != nil {
		return fmt.Errorf("failed to set JSON document: %w", err)
	}

	// Set TTL if specified
	if ttl > 0 {
		if err := jc.redis.Expire(ctx, key, ttl).Err(); err != nil {
			jc.logger.Warn("Failed to set TTL", zap.String("key", key), zap.Error(err))
		}
	}

	jc.logger.Info("JSON document set successfully", zap.String("key", key))
	return nil
}

// Get retrieves a JSON document
func (jc *JSONClient) Get(ctx context.Context, key string, path string) (interface{}, error) {
	jc.logger.Info("Getting JSON document", zap.String("key", key), zap.String("path", path))

	if path == "" {
		path = "$"
	}

	cmd := redis.NewCmd(ctx, "JSON.GET", key, path)
	if err := jc.redis.Process(ctx, cmd); err != nil {
		return nil, fmt.Errorf("failed to get JSON document: %w", err)
	}

	result := cmd.Val()
	if result == nil {
		return nil, fmt.Errorf("document not found")
	}

	// Parse JSON result
	var data interface{}
	if err := json.Unmarshal([]byte(result.(string)), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON result: %w", err)
	}

	jc.logger.Info("JSON document retrieved successfully", zap.String("key", key))
	return data, nil
}

// Update performs partial updates on a JSON document
func (jc *JSONClient) Update(ctx context.Context, key string, operations []JSONUpdateOperation) error {
	jc.logger.Info("Updating JSON document", 
		zap.String("key", key), 
		zap.Int("operations_count", len(operations)),
	)

	// Execute operations in a pipeline for atomicity
	pipe := jc.redis.Pipeline()

	for _, op := range operations {
		switch op.Operation {
		case "SET":
			valueJSON, err := json.Marshal(op.Value)
			if err != nil {
				return fmt.Errorf("failed to marshal value for SET operation: %w", err)
			}
			pipe.Process(ctx, redis.NewCmd(ctx, "JSON.SET", key, op.Path, string(valueJSON)))

		case "DELETE":
			pipe.Process(ctx, redis.NewCmd(ctx, "JSON.DEL", key, op.Path))

		case "APPEND":
			if op.Value != nil {
				valueJSON, err := json.Marshal(op.Value)
				if err != nil {
					return fmt.Errorf("failed to marshal value for APPEND operation: %w", err)
				}
				pipe.Process(ctx, redis.NewCmd(ctx, "JSON.ARRAPPEND", key, op.Path, string(valueJSON)))
			}

		case "INCREMENT":
			if num, ok := op.Value.(float64); ok {
				pipe.Process(ctx, redis.NewCmd(ctx, "JSON.NUMINCRBY", key, op.Path, num))
			} else {
				return fmt.Errorf("INCREMENT operation requires numeric value")
			}

		default:
			return fmt.Errorf("unsupported operation: %s", op.Operation)
		}
	}

	// Execute pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute update operations: %w", err)
	}

	jc.logger.Info("JSON document updated successfully", zap.String("key", key))
	return nil
}

// Query performs complex queries on JSON documents
func (jc *JSONClient) Query(ctx context.Context, key string, queries []JSONQuery) (map[string]interface{}, error) {
	jc.logger.Info("Querying JSON document", 
		zap.String("key", key), 
		zap.Int("queries_count", len(queries)),
	)

	results := make(map[string]interface{})

	for i, query := range queries {
		cmd := redis.NewCmd(ctx, "JSON.GET", key, query.Path)
		if err := jc.redis.Process(ctx, cmd); err != nil {
			jc.logger.Warn("Query failed", zap.String("path", query.Path), zap.Error(err))
			continue
		}

		result := cmd.Val()
		if result != nil {
			var data interface{}
			if err := json.Unmarshal([]byte(result.(string)), &data); err == nil {
				results[fmt.Sprintf("query_%d", i)] = data
			}
		}
	}

	jc.logger.Info("JSON queries completed", 
		zap.String("key", key), 
		zap.Int("results_count", len(results)),
	)

	return results, nil
}

// Search searches across multiple JSON documents
func (jc *JSONClient) Search(ctx context.Context, pattern string, queries []JSONQuery) (map[string]interface{}, error) {
	jc.logger.Info("Searching JSON documents", zap.String("pattern", pattern))

	// Get all keys matching pattern
	keys, err := jc.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %w", err)
	}

	results := make(map[string]interface{})

	// Query each key
	for _, key := range keys {
		keyResults, err := jc.Query(ctx, key, queries)
		if err != nil {
			jc.logger.Warn("Failed to query key", zap.String("key", key), zap.Error(err))
			continue
		}
		
		if len(keyResults) > 0 {
			results[key] = keyResults
		}
	}

	jc.logger.Info("JSON search completed", 
		zap.String("pattern", pattern), 
		zap.Int("keys_found", len(keys)),
		zap.Int("results_count", len(results)),
	)

	return results, nil
}

// GetType returns the type of a JSON value at a specific path
func (jc *JSONClient) GetType(ctx context.Context, key string, path string) (string, error) {
	if path == "" {
		path = "$"
	}

	cmd := redis.NewCmd(ctx, "JSON.TYPE", key, path)
	if err := jc.redis.Process(ctx, cmd); err != nil {
		return "", fmt.Errorf("failed to get JSON type: %w", err)
	}

	result := cmd.Val()
	if result == nil {
		return "", fmt.Errorf("path not found")
	}

	return result.(string), nil
}

// GetSize returns the size of a JSON value at a specific path
func (jc *JSONClient) GetSize(ctx context.Context, key string, path string) (int64, error) {
	if path == "" {
		path = "$"
	}

	// Get the JSON data first
	data, err := jc.Get(ctx, key, path)
	if err != nil {
		return 0, err
	}

	// Calculate size based on type
	jsonData, err := json.Marshal(data)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal for size calculation: %w", err)
	}

	return int64(len(jsonData)), nil
}

// CreateDocument creates a new JSON document with metadata
func (jc *JSONClient) CreateDocument(ctx context.Context, id string, data interface{}, metadata map[string]interface{}) error {
	now := time.Now()
	
	document := JSONDocument{
		ID:       id,
		Data:     data,
		Metadata: metadata,
		Version:  1,
		Created:  now,
		Updated:  now,
	}

	return jc.Set(ctx, fmt.Sprintf("doc:%s", id), document, jc.config.DefaultTTL)
}

// UpdateDocument updates an existing JSON document
func (jc *JSONClient) UpdateDocument(ctx context.Context, id string, updates map[string]interface{}) error {
	key := fmt.Sprintf("doc:%s", id)
	
	// Check if document exists
	_, err := jc.Get(ctx, key, "$")
	if err != nil {
		return fmt.Errorf("document not found: %w", err)
	}

	// Prepare update operations
	operations := []JSONUpdateOperation{
		{
			Path:      "$.updated",
			Operation: "SET",
			Value:     time.Now(),
		},
		{
			Path:      "$.version",
			Operation: "INCREMENT",
			Value:     1.0,
		},
	}

	// Add data updates
	for path, value := range updates {
		operations = append(operations, JSONUpdateOperation{
			Path:      fmt.Sprintf("$.data.%s", path),
			Operation: "SET",
			Value:     value,
		})
	}

	return jc.Update(ctx, key, operations)
}

// ListDocuments lists all documents with optional filtering
func (jc *JSONClient) ListDocuments(ctx context.Context, filter map[string]interface{}) ([]JSONDocument, error) {
	jc.logger.Info("Listing JSON documents")

	// Get all document keys
	keys, err := jc.redis.Keys(ctx, "doc:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get document keys: %w", err)
	}

	documents := make([]JSONDocument, 0, len(keys))

	for _, key := range keys {
		data, err := jc.Get(ctx, key, "$")
		if err != nil {
			jc.logger.Warn("Failed to get document", zap.String("key", key), zap.Error(err))
			continue
		}

		// Parse as JSONDocument
		jsonData, err := json.Marshal(data)
		if err != nil {
			continue
		}

		var doc JSONDocument
		if err := json.Unmarshal(jsonData, &doc); err != nil {
			continue
		}

		// Apply filter if specified
		if jc.matchesFilter(doc, filter) {
			documents = append(documents, doc)
		}
	}

	jc.logger.Info("Documents listed", zap.Int("count", len(documents)))
	return documents, nil
}

// Helper method to check if document matches filter
func (jc *JSONClient) matchesFilter(doc JSONDocument, filter map[string]interface{}) bool {
	if len(filter) == 0 {
		return true
	}

	// Simple filter implementation
	for key, expectedValue := range filter {
		if key == "id" && doc.ID != expectedValue {
			return false
		}
		// Add more filter logic as needed
	}

	return true
}
