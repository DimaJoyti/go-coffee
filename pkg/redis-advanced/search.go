package redisadvanced

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// SearchClient provides advanced search capabilities using RediSearch
type SearchClient struct {
	redis  *redis.Client
	logger *zap.Logger
	config *SearchConfig
}

// SearchConfig contains configuration for search operations
type SearchConfig struct {
	IndexPrefix    string
	VectorDim      int
	DistanceMetric string
	MaxResults     int
	Timeout        time.Duration
}

// VectorSearchRequest represents a vector similarity search request
type VectorSearchRequest struct {
	Vector     []float32          `json:"vector"`
	TopK       int                `json:"top_k"`
	Filter     map[string]string  `json:"filter,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// SearchResult represents a search result
type SearchResult struct {
	ID       string                 `json:"id"`
	Score    float64                `json:"score"`
	Fields   map[string]interface{} `json:"fields"`
	Vector   []float32              `json:"vector,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SearchResponse contains search results and metadata
type SearchResponse struct {
	Results    []SearchResult         `json:"results"`
	Total      int                    `json:"total"`
	Duration   time.Duration          `json:"duration"`
	Query      string                 `json:"query"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// NewSearchClient creates a new search client
func NewSearchClient(redisClient *redis.Client, logger *zap.Logger, config *SearchConfig) *SearchClient {
	if config == nil {
		config = &SearchConfig{
			IndexPrefix:    "coffee_search",
			VectorDim:      128,
			DistanceMetric: "COSINE",
			MaxResults:     100,
			Timeout:        5 * time.Second,
		}
	}

	return &SearchClient{
		redis:  redisClient,
		logger: logger,
		config: config,
	}
}

// CreateVectorIndex creates a vector search index
func (sc *SearchClient) CreateVectorIndex(ctx context.Context, indexName string, schema map[string]interface{}) error {
	sc.logger.Info("Creating vector index", zap.String("index", indexName))

	// Build FT.CREATE command for vector index
	args := []interface{}{
		indexName,
		"ON", "HASH",
		"PREFIX", "1", sc.config.IndexPrefix + ":",
		"SCHEMA",
	}

	// Add vector field
	args = append(args,
		"vector", "VECTOR", "HNSW", "6",
		"TYPE", "FLOAT32",
		"DIM", sc.config.VectorDim,
		"DISTANCE_METRIC", sc.config.DistanceMetric,
	)

	// Add other fields from schema
	for fieldName, fieldType := range schema {
		args = append(args, fieldName, fieldType)
	}

	cmd := redis.NewCmd(ctx, append([]interface{}{"FT.CREATE"}, args...)...)
	if err := sc.redis.Process(ctx, cmd); err != nil {
		return fmt.Errorf("failed to create vector index: %w", err)
	}

	sc.logger.Info("Vector index created successfully", zap.String("index", indexName))
	return nil
}

// VectorSearch performs vector similarity search
func (sc *SearchClient) VectorSearch(ctx context.Context, indexName string, req *VectorSearchRequest) (*SearchResponse, error) {
	startTime := time.Now()
	
	sc.logger.Info("Performing vector search",
		zap.String("index", indexName),
		zap.Int("vector_dim", len(req.Vector)),
		zap.Int("top_k", req.TopK),
	)

	// Build vector query
	vectorStr := sc.vectorToString(req.Vector)
	query := fmt.Sprintf("*=>[KNN %d @vector $BLOB]", req.TopK)

	// Build FT.SEARCH command
	args := []interface{}{
		indexName,
		query,
		"PARAMS", "2", "BLOB", vectorStr,
		"SORTBY", "__vector_score",
		"LIMIT", "0", req.TopK,
		"RETURN", "3", "__vector_score", "id", "content",
	}

	cmd := redis.NewCmd(ctx, append([]interface{}{"FT.SEARCH"}, args...)...)
	if err := sc.redis.Process(ctx, cmd); err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// Parse results
	results, err := sc.parseSearchResults(cmd.Val())
	if err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}

	duration := time.Since(startTime)
	
	response := &SearchResponse{
		Results:  results,
		Total:    len(results),
		Duration: duration,
		Query:    query,
		Metadata: map[string]interface{}{
			"index_name": indexName,
			"vector_dim": len(req.Vector),
			"top_k":      req.TopK,
		},
	}

	sc.logger.Info("Vector search completed",
		zap.Int("results_count", len(results)),
		zap.Duration("duration", duration),
	)

	return response, nil
}

// SemanticSearch performs semantic text search
func (sc *SearchClient) SemanticSearch(ctx context.Context, indexName, query string, limit int) (*SearchResponse, error) {
	startTime := time.Now()
	
	sc.logger.Info("Performing semantic search",
		zap.String("index", indexName),
		zap.String("query", query),
		zap.Int("limit", limit),
	)

	// Build FT.SEARCH command for text search
	args := []interface{}{
		indexName,
		query,
		"LIMIT", "0", limit,
		"RETURN", "3", "id", "content", "score",
	}

	cmd := redis.NewCmd(ctx, append([]interface{}{"FT.SEARCH"}, args...)...)
	if err := sc.redis.Process(ctx, cmd); err != nil {
		return nil, fmt.Errorf("semantic search failed: %w", err)
	}

	// Parse results
	results, err := sc.parseSearchResults(cmd.Val())
	if err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}

	duration := time.Since(startTime)
	
	response := &SearchResponse{
		Results:  results,
		Total:    len(results),
		Duration: duration,
		Query:    query,
		Metadata: map[string]interface{}{
			"index_name": indexName,
			"search_type": "semantic",
		},
	}

	sc.logger.Info("Semantic search completed",
		zap.Int("results_count", len(results)),
		zap.Duration("duration", duration),
	)

	return response, nil
}

// HybridSearch combines vector and text search
func (sc *SearchClient) HybridSearch(ctx context.Context, indexName string, textQuery string, vector []float32, weights map[string]float64) (*SearchResponse, error) {
	startTime := time.Now()
	
	sc.logger.Info("Performing hybrid search",
		zap.String("index", indexName),
		zap.String("text_query", textQuery),
		zap.Int("vector_dim", len(vector)),
	)

	// Perform both searches
	textResults, err := sc.SemanticSearch(ctx, indexName, textQuery, sc.config.MaxResults/2)
	if err != nil {
		return nil, fmt.Errorf("text search failed: %w", err)
	}

	vectorResults, err := sc.VectorSearch(ctx, indexName, &VectorSearchRequest{
		Vector: vector,
		TopK:   sc.config.MaxResults / 2,
	})
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// Combine and rerank results
	combinedResults := sc.combineResults(textResults.Results, vectorResults.Results, weights)

	duration := time.Since(startTime)
	
	response := &SearchResponse{
		Results:  combinedResults,
		Total:    len(combinedResults),
		Duration: duration,
		Query:    fmt.Sprintf("hybrid: %s + vector", textQuery),
		Metadata: map[string]interface{}{
			"index_name":   indexName,
			"search_type":  "hybrid",
			"text_results": len(textResults.Results),
			"vector_results": len(vectorResults.Results),
		},
	}

	sc.logger.Info("Hybrid search completed",
		zap.Int("results_count", len(combinedResults)),
		zap.Duration("duration", duration),
	)

	return response, nil
}

// AddDocument adds a document to the search index
func (sc *SearchClient) AddDocument(ctx context.Context, docID string, fields map[string]interface{}, vector []float32) error {
	sc.logger.Info("Adding document to search index", zap.String("doc_id", docID))

	key := fmt.Sprintf("%s:%s", sc.config.IndexPrefix, docID)
	
	// Prepare fields for Redis hash
	hashFields := make(map[string]interface{})
	for k, v := range fields {
		hashFields[k] = v
	}
	
	// Add vector as binary data
	if len(vector) > 0 {
		hashFields["vector"] = sc.vectorToString(vector)
	}

	if err := sc.redis.HMSet(ctx, key, hashFields).Err(); err != nil {
		return fmt.Errorf("failed to add document: %w", err)
	}

	sc.logger.Info("Document added successfully", zap.String("doc_id", docID))
	return nil
}

// Helper methods

func (sc *SearchClient) vectorToString(vector []float32) string {
	strValues := make([]string, len(vector))
	for i, v := range vector {
		strValues[i] = strconv.FormatFloat(float64(v), 'f', -1, 32)
	}
	return strings.Join(strValues, ",")
}

func (sc *SearchClient) parseSearchResults(rawResults interface{}) ([]SearchResult, error) {
	// Parse Redis FT.SEARCH response format
	results := []SearchResult{}
	
	// Implementation depends on the specific format returned by Redis
	// This is a simplified version
	if resultSlice, ok := rawResults.([]interface{}); ok && len(resultSlice) > 0 {
		// First element is the count
		if len(resultSlice) > 1 {
			for i := 1; i < len(resultSlice); i += 2 {
				if i+1 < len(resultSlice) {
					docID := fmt.Sprintf("%v", resultSlice[i])
					fields := make(map[string]interface{})
					
					if fieldSlice, ok := resultSlice[i+1].([]interface{}); ok {
						for j := 0; j < len(fieldSlice); j += 2 {
							if j+1 < len(fieldSlice) {
								key := fmt.Sprintf("%v", fieldSlice[j])
								value := fieldSlice[j+1]
								fields[key] = value
							}
						}
					}
					
					result := SearchResult{
						ID:     docID,
						Fields: fields,
						Score:  0.0, // Extract from fields if available
					}
					results = append(results, result)
				}
			}
		}
	}
	
	return results, nil
}

func (sc *SearchClient) combineResults(textResults, vectorResults []SearchResult, weights map[string]float64) []SearchResult {
	// Simple combination strategy - merge and deduplicate
	resultMap := make(map[string]SearchResult)
	
	textWeight := weights["text"]
	vectorWeight := weights["vector"]
	if textWeight == 0 {
		textWeight = 0.5
	}
	if vectorWeight == 0 {
		vectorWeight = 0.5
	}
	
	// Add text results
	for _, result := range textResults {
		result.Score *= textWeight
		resultMap[result.ID] = result
	}
	
	// Add vector results (merge if exists)
	for _, result := range vectorResults {
		if existing, exists := resultMap[result.ID]; exists {
			existing.Score += result.Score * vectorWeight
			resultMap[result.ID] = existing
		} else {
			result.Score *= vectorWeight
			resultMap[result.ID] = result
		}
	}
	
	// Convert back to slice
	combined := make([]SearchResult, 0, len(resultMap))
	for _, result := range resultMap {
		combined = append(combined, result)
	}
	
	return combined
}
