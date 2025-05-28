package rag

import (
	"time"
)

// Document represents a document in the RAG system
type Document struct {
	ID              string                 `json:"id"`
	Content         string                 `json:"content"`
	Title           string                 `json:"title"`
	Source          string                 `json:"source"`          // reddit_post, reddit_comment, etc.
	SourceID        string                 `json:"source_id"`       // original post/comment ID
	Metadata        map[string]interface{} `json:"metadata"`
	EmbeddingVector []float64              `json:"embedding_vector"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	IndexedAt       time.Time              `json:"indexed_at"`
	Version         int                    `json:"version"`
	Tags            []string               `json:"tags"`
	Category        string                 `json:"category"`
	Language        string                 `json:"language"`
	WordCount       int                    `json:"word_count"`
	Quality         float64                `json:"quality"`         // quality score 0-1
}

// DocumentChunk represents a chunk of a document for better retrieval
type DocumentChunk struct {
	ID              string                 `json:"id"`
	DocumentID      string                 `json:"document_id"`
	Content         string                 `json:"content"`
	ChunkIndex      int                    `json:"chunk_index"`
	StartOffset     int                    `json:"start_offset"`
	EndOffset       int                    `json:"end_offset"`
	EmbeddingVector []float64              `json:"embedding_vector"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
}

// Query represents a RAG query
type Query struct {
	ID              string                 `json:"id"`
	Text            string                 `json:"text"`
	UserID          string                 `json:"user_id"`
	Context         string                 `json:"context"`
	Filters         map[string]interface{} `json:"filters"`
	EmbeddingVector []float64              `json:"embedding_vector"`
	CreatedAt       time.Time              `json:"created_at"`
	Language        string                 `json:"language"`
	Intent          string                 `json:"intent"`
	Entities        []Entity               `json:"entities"`
}

// Entity represents a named entity in a query
type Entity struct {
	Text  string  `json:"text"`
	Type  string  `json:"type"`  // person, organization, location, etc.
	Score float64 `json:"score"`
}

// RetrievalResult represents a document retrieval result
type RetrievalResult struct {
	Document       *Document              `json:"document"`
	Chunk          *DocumentChunk         `json:"chunk,omitempty"`
	Score          float64                `json:"score"`
	Relevance      float64                `json:"relevance"`
	Explanation    string                 `json:"explanation"`
	MatchedTerms   []string               `json:"matched_terms"`
	Metadata       map[string]interface{} `json:"metadata"`
	RetrievedAt    time.Time              `json:"retrieved_at"`
}

// RAGResponse represents a complete RAG response
type RAGResponse struct {
	ID              string            `json:"id"`
	QueryID         string            `json:"query_id"`
	GeneratedText   string            `json:"generated_text"`
	Sources         []RetrievalResult `json:"sources"`
	Confidence      float64           `json:"confidence"`
	ModelUsed       string            `json:"model_used"`
	ProcessingTime  time.Duration     `json:"processing_time"`
	TokensUsed      int               `json:"tokens_used"`
	CreatedAt       time.Time         `json:"created_at"`
	Metadata        map[string]string `json:"metadata"`
	Citations       []Citation        `json:"citations"`
	FollowUpQueries []string          `json:"follow_up_queries"`
}

// Citation represents a citation in the generated response
type Citation struct {
	DocumentID   string `json:"document_id"`
	ChunkID      string `json:"chunk_id,omitempty"`
	StartIndex   int    `json:"start_index"`
	EndIndex     int    `json:"end_index"`
	Text         string `json:"text"`
	Source       string `json:"source"`
	URL          string `json:"url,omitempty"`
	Confidence   float64 `json:"confidence"`
}

// EmbeddingRequest represents a request to generate embeddings
type EmbeddingRequest struct {
	Texts    []string          `json:"texts"`
	Model    string            `json:"model"`
	Metadata map[string]string `json:"metadata"`
}

// EmbeddingResponse represents an embedding generation response
type EmbeddingResponse struct {
	Embeddings  [][]float64       `json:"embeddings"`
	Model       string            `json:"model"`
	Dimensions  int               `json:"dimensions"`
	TokensUsed  int               `json:"tokens_used"`
	ProcessedAt time.Time         `json:"processed_at"`
	Metadata    map[string]string `json:"metadata"`
}

// VectorSearchRequest represents a vector search request
type VectorSearchRequest struct {
	Vector      []float64              `json:"vector"`
	TopK        int                    `json:"top_k"`
	Filters     map[string]interface{} `json:"filters"`
	Threshold   float64                `json:"threshold"`
	IncludeMetadata bool               `json:"include_metadata"`
	Namespace   string                 `json:"namespace,omitempty"`
}

// VectorSearchResponse represents a vector search response
type VectorSearchResponse struct {
	Results     []VectorSearchResult `json:"results"`
	QueryTime   time.Duration        `json:"query_time"`
	TotalFound  int                  `json:"total_found"`
	SearchedAt  time.Time            `json:"searched_at"`
}

// VectorSearchResult represents a single vector search result
type VectorSearchResult struct {
	ID       string                 `json:"id"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata"`
	Content  string                 `json:"content,omitempty"`
}

// IndexingJob represents a document indexing job
type IndexingJob struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // document, chunk, batch
	Status      string                 `json:"status"` // pending, processing, completed, failed
	DocumentIDs []string               `json:"document_ids"`
	Progress    float64                `json:"progress"`
	Config      map[string]interface{} `json:"config"`
	Results     IndexingResults        `json:"results"`
	Error       string                 `json:"error"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at"`
	Metadata    map[string]string      `json:"metadata"`
}

// IndexingResults represents indexing job results
type IndexingResults struct {
	DocumentsProcessed int               `json:"documents_processed"`
	ChunksCreated      int               `json:"chunks_created"`
	EmbeddingsGenerated int              `json:"embeddings_generated"`
	Errors             []IndexingError   `json:"errors"`
	ProcessingTime     time.Duration     `json:"processing_time"`
	Statistics         IndexingStats     `json:"statistics"`
}

// IndexingError represents an indexing error
type IndexingError struct {
	DocumentID string `json:"document_id"`
	Error      string `json:"error"`
	Timestamp  time.Time `json:"timestamp"`
}

// IndexingStats represents indexing statistics
type IndexingStats struct {
	AverageChunkSize    int     `json:"average_chunk_size"`
	AverageWordCount    int     `json:"average_word_count"`
	AverageQualityScore float64 `json:"average_quality_score"`
	LanguageDistribution map[string]int `json:"language_distribution"`
	CategoryDistribution map[string]int `json:"category_distribution"`
}

// RetrievalConfig represents retrieval configuration
type RetrievalConfig struct {
	TopK              int                    `json:"top_k"`
	ScoreThreshold    float64                `json:"score_threshold"`
	MaxDocuments      int                    `json:"max_documents"`
	ContextWindow     int                    `json:"context_window"`
	RerankerModel     string                 `json:"reranker_model"`
	UseSemanticSearch bool                   `json:"use_semantic_search"`
	UseKeywordSearch  bool                   `json:"use_keyword_search"`
	HybridWeight      float64                `json:"hybrid_weight"` // 0.0 = pure keyword, 1.0 = pure semantic
	Filters           map[string]interface{} `json:"filters"`
	BoostFactors      map[string]float64     `json:"boost_factors"`
}

// GenerationConfig represents text generation configuration
type GenerationConfig struct {
	Model            string            `json:"model"`
	Temperature      float64           `json:"temperature"`
	MaxTokens        int               `json:"max_tokens"`
	TopP             float64           `json:"top_p"`
	FrequencyPenalty float64           `json:"frequency_penalty"`
	PresencePenalty  float64           `json:"presence_penalty"`
	StopSequences    []string          `json:"stop_sequences"`
	SystemPrompt     string            `json:"system_prompt"`
	ContextTemplate  string            `json:"context_template"`
	CitationStyle    string            `json:"citation_style"`
	IncludeSources   bool              `json:"include_sources"`
	Metadata         map[string]string `json:"metadata"`
}

// FeedbackData represents user feedback on RAG responses
type FeedbackData struct {
	ID           string                 `json:"id"`
	QueryID      string                 `json:"query_id"`
	ResponseID   string                 `json:"response_id"`
	UserID       string                 `json:"user_id"`
	Rating       int                    `json:"rating"` // 1-5 scale
	Helpful      bool                   `json:"helpful"`
	Accurate     bool                   `json:"accurate"`
	Complete     bool                   `json:"complete"`
	Relevant     bool                   `json:"relevant"`
	Comments     string                 `json:"comments"`
	Improvements []string               `json:"improvements"`
	CreatedAt    time.Time              `json:"created_at"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// AnalyticsData represents RAG system analytics
type AnalyticsData struct {
	ID                string                 `json:"id"`
	Timeframe         string                 `json:"timeframe"` // hourly, daily, weekly
	QueryCount        int                    `json:"query_count"`
	AvgResponseTime   time.Duration          `json:"avg_response_time"`
	AvgConfidence     float64                `json:"avg_confidence"`
	AvgRating         float64                `json:"avg_rating"`
	TopQueries        []QueryAnalytics       `json:"top_queries"`
	TopSources        []SourceAnalytics      `json:"top_sources"`
	ErrorRate         float64                `json:"error_rate"`
	CacheHitRate      float64                `json:"cache_hit_rate"`
	ModelUsage        map[string]int         `json:"model_usage"`
	UserEngagement    UserEngagementMetrics  `json:"user_engagement"`
	GeneratedAt       time.Time              `json:"generated_at"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// QueryAnalytics represents analytics for specific queries
type QueryAnalytics struct {
	Query       string  `json:"query"`
	Count       int     `json:"count"`
	AvgRating   float64 `json:"avg_rating"`
	SuccessRate float64 `json:"success_rate"`
}

// SourceAnalytics represents analytics for document sources
type SourceAnalytics struct {
	Source      string  `json:"source"`
	DocumentID  string  `json:"document_id"`
	RetrievalCount int  `json:"retrieval_count"`
	AvgRelevance float64 `json:"avg_relevance"`
	AvgRating   float64 `json:"avg_rating"`
}

// UserEngagementMetrics represents user engagement metrics
type UserEngagementMetrics struct {
	ActiveUsers       int     `json:"active_users"`
	QueriesPerUser    float64 `json:"queries_per_user"`
	AvgSessionLength  time.Duration `json:"avg_session_length"`
	ReturnUserRate    float64 `json:"return_user_rate"`
	FeedbackRate      float64 `json:"feedback_rate"`
}
