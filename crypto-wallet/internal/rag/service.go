package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/internal/ai"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/config"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/redis"
)

// Service provides RAG (Retrieval-Augmented Generation) capabilities
type Service struct {
	config         config.RAGConfig
	logger         *logger.Logger
	aiService      ai.Service
	cache          redis.Client
	vectorStore    VectorStore
	embeddings     EmbeddingsService
	retriever      Retriever
	generator      Generator
}

// VectorStore interface for vector database operations
type VectorStore interface {
	Index(ctx context.Context, documents []Document) error
	Search(ctx context.Context, req VectorSearchRequest) (*VectorSearchResponse, error)
	Delete(ctx context.Context, ids []string) error
	Update(ctx context.Context, documents []Document) error
	GetStats(ctx context.Context) (map[string]interface{}, error)
}

// EmbeddingsService interface for generating embeddings
type EmbeddingsService interface {
	GenerateEmbeddings(ctx context.Context, req EmbeddingRequest) (*EmbeddingResponse, error)
	GetDimensions() int
	GetModel() string
}

// Retriever interface for document retrieval
type Retriever interface {
	Retrieve(ctx context.Context, query Query, config RetrievalConfig) ([]RetrievalResult, error)
	RetrieveWithReranking(ctx context.Context, query Query, config RetrievalConfig) ([]RetrievalResult, error)
}

// Generator interface for text generation
type Generator interface {
	Generate(ctx context.Context, query Query, sources []RetrievalResult, config GenerationConfig) (*RAGResponse, error)
}

// NewService creates a new RAG service
func NewService(
	cfg config.RAGConfig,
	logger *logger.Logger,
	aiService ai.Service,
	cache redis.Client,
	vectorStore VectorStore,
	embeddings EmbeddingsService,
) *Service {
	service := &Service{
		config:      cfg,
		logger:      logger,
		aiService:   aiService,
		cache:       cache,
		vectorStore: vectorStore,
		embeddings:  embeddings,
	}

	// Initialize retriever and generator
	service.retriever = newRetriever(vectorStore, embeddings, logger)
	service.generator = newGenerator(aiService, logger)

	logger.Info("RAG service initialized successfully")
	return service
}

// Query performs a complete RAG query with retrieval and generation
func (s *Service) Query(ctx context.Context, queryText string, userID string, options map[string]interface{}) (*RAGResponse, error) {
	s.logger.Info(fmt.Sprintf("Processing RAG query for user %s", userID))

	// Create query object
	query := Query{
		ID:        generateQueryID(),
		Text:      queryText,
		UserID:    userID,
		CreatedAt: time.Now(),
		Language:  "en", // Default to English, could be detected
	}

	// Check cache first
	cacheKey := fmt.Sprintf("rag:query:%s", hashQuery(queryText))
	if s.config.Enabled {
		if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != "" {
			var response RAGResponse
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				s.logger.Debug(fmt.Sprintf("Cache hit for RAG query: %s", queryText))
				return &response, nil
			}
		}
	}

	startTime := time.Now()

	// Generate query embedding
	embeddingReq := EmbeddingRequest{
		Texts: []string{queryText},
		Model: s.config.Embeddings.Model,
	}

	embeddingResp, err := s.embeddings.GenerateEmbeddings(ctx, embeddingReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	if len(embeddingResp.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings generated for query")
	}

	query.EmbeddingVector = embeddingResp.Embeddings[0]

	// Configure retrieval
	retrievalConfig := RetrievalConfig{
		TopK:              s.config.Retrieval.TopK,
		ScoreThreshold:    s.config.Retrieval.ScoreThreshold,
		MaxDocuments:      s.config.Retrieval.MaxDocuments,
		ContextWindow:     s.config.Retrieval.ContextWindow,
		RerankerModel:     s.config.Retrieval.RerankerModel,
		UseSemanticSearch: true,
		UseKeywordSearch:  true,
		HybridWeight:      0.7, // Favor semantic search
	}

	// Apply user-provided options
	if options != nil {
		if topK, ok := options["top_k"].(int); ok {
			retrievalConfig.TopK = topK
		}
		if threshold, ok := options["threshold"].(float64); ok {
			retrievalConfig.ScoreThreshold = threshold
		}
		if filters, ok := options["filters"].(map[string]interface{}); ok {
			retrievalConfig.Filters = filters
		}
	}

	// Retrieve relevant documents
	sources, err := s.retriever.RetrieveWithReranking(ctx, query, retrievalConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve documents: %w", err)
	}

	if len(sources) == 0 {
		s.logger.Warn(fmt.Sprintf("No relevant documents found for query: %s", queryText))
		return &RAGResponse{
			ID:             generateResponseID(),
			QueryID:        query.ID,
			GeneratedText:  "I couldn't find any relevant information to answer your question.",
			Sources:        []RetrievalResult{},
			Confidence:     0.0,
			ModelUsed:      "none",
			ProcessingTime: time.Since(startTime),
			CreatedAt:      time.Now(),
		}, nil
	}

	// Configure generation
	generationConfig := GenerationConfig{
		Model:            s.config.Embeddings.Model, // Use same model for consistency
		Temperature:      s.config.Embeddings.Temperature,
		MaxTokens:        s.config.Embeddings.MaxTokens,
		TopP:             0.9,
		FrequencyPenalty: 0.1,
		PresencePenalty:  0.1,
		SystemPrompt:     s.buildSystemPrompt(),
		ContextTemplate:  s.buildContextTemplate(),
		CitationStyle:    "numbered",
		IncludeSources:   true,
	}

	// Generate response
	response, err := s.generator.Generate(ctx, query, sources, generationConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	response.ProcessingTime = time.Since(startTime)

	// Cache the response
	if s.config.Enabled {
		if data, err := json.Marshal(response); err == nil {
			cacheTTL := 1 * time.Hour // Cache for 1 hour
			if err := s.cache.Set(ctx, cacheKey, data, cacheTTL); err != nil {
				s.logger.Warn(fmt.Sprintf("Failed to cache RAG response: %v", err))
			}
		}
	}

	s.logger.Info(fmt.Sprintf("RAG query completed in %v (confidence: %.2f, sources: %d)", 
		response.ProcessingTime, response.Confidence, len(response.Sources)))

	return response, nil
}

// IndexDocuments indexes documents for retrieval
func (s *Service) IndexDocuments(ctx context.Context, documents []Document) (*IndexingJob, error) {
	s.logger.Info(fmt.Sprintf("Starting indexing job for %d documents", len(documents)))

	job := &IndexingJob{
		ID:        generateJobID(),
		Type:      "batch",
		Status:    "processing",
		CreatedAt: time.Now(),
		StartedAt: &[]time.Time{time.Now()}[0],
		Results: IndexingResults{
			Errors: []IndexingError{},
		},
	}

	startTime := time.Now()

	// Process documents in batches
	batchSize := s.config.Embeddings.BatchSize
	if batchSize <= 0 {
		batchSize = 10
	}

	var allTexts []string
	var documentMap = make(map[int]*Document)

	// Prepare texts for embedding generation
	for i, doc := range documents {
		if doc.Content != "" {
			allTexts = append(allTexts, doc.Content)
			documentMap[len(allTexts)-1] = &documents[i]
		}
	}

	// Generate embeddings for all documents
	embeddingReq := EmbeddingRequest{
		Texts: allTexts,
		Model: s.config.Embeddings.Model,
	}

	embeddingResp, err := s.embeddings.GenerateEmbeddings(ctx, embeddingReq)
	if err != nil {
		job.Status = "failed"
		job.Error = fmt.Sprintf("Failed to generate embeddings: %v", err)
		return job, err
	}

	// Assign embeddings to documents
	for i, embedding := range embeddingResp.Embeddings {
		if doc, exists := documentMap[i]; exists {
			doc.EmbeddingVector = embedding
			doc.IndexedAt = time.Now()
		}
	}

	// Index documents in vector store
	if err := s.vectorStore.Index(ctx, documents); err != nil {
		job.Status = "failed"
		job.Error = fmt.Sprintf("Failed to index documents: %v", err)
		return job, err
	}

	// Update job results
	job.Status = "completed"
	job.CompletedAt = &[]time.Time{time.Now()}[0]
	job.Results.DocumentsProcessed = len(documents)
	job.Results.EmbeddingsGenerated = len(embeddingResp.Embeddings)
	job.Results.ProcessingTime = time.Since(startTime)

	s.logger.Info(fmt.Sprintf("Indexing job completed: %d documents processed in %v", 
		len(documents), job.Results.ProcessingTime))

	return job, nil
}

// SearchDocuments performs a semantic search over indexed documents
func (s *Service) SearchDocuments(ctx context.Context, queryText string, filters map[string]interface{}, topK int) ([]RetrievalResult, error) {
	s.logger.Info(fmt.Sprintf("Searching documents for: %s", queryText))

	// Generate query embedding
	embeddingReq := EmbeddingRequest{
		Texts: []string{queryText},
		Model: s.config.Embeddings.Model,
	}

	embeddingResp, err := s.embeddings.GenerateEmbeddings(ctx, embeddingReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	if len(embeddingResp.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings generated for query")
	}

	// Perform vector search
	searchReq := VectorSearchRequest{
		Vector:          embeddingResp.Embeddings[0],
		TopK:            topK,
		Filters:         filters,
		Threshold:       s.config.Retrieval.ScoreThreshold,
		IncludeMetadata: true,
	}

	searchResp, err := s.vectorStore.Search(ctx, searchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to search vector store: %w", err)
	}

	// Convert to retrieval results
	results := make([]RetrievalResult, len(searchResp.Results))
	for i, result := range searchResp.Results {
		results[i] = RetrievalResult{
			Document: &Document{
				ID:      result.ID,
				Content: result.Content,
			},
			Score:       result.Score,
			Relevance:   result.Score, // Use score as relevance for now
			RetrievedAt: time.Now(),
		}
	}

	s.logger.Info(fmt.Sprintf("Found %d documents for query: %s", len(results), queryText))
	return results, nil
}

// buildSystemPrompt builds the system prompt for generation
func (s *Service) buildSystemPrompt() string {
	return `You are a helpful AI assistant that provides accurate and informative responses based on the given context. 

Guidelines:
1. Use only the information provided in the context to answer questions
2. If the context doesn't contain enough information, say so clearly
3. Cite your sources using numbered references [1], [2], etc.
4. Be concise but comprehensive in your responses
5. If asked about recent events or real-time information, clarify the limitations of your knowledge

Always strive to be helpful, accurate, and transparent about the sources of your information.`
}

// buildContextTemplate builds the context template for generation
func (s *Service) buildContextTemplate() string {
	return `Context information:
{{range $i, $source := .Sources}}
[{{add $i 1}}] {{$source.Document.Title}}
{{$source.Document.Content}}

{{end}}

Question: {{.Query.Text}}

Answer:`
}

// Helper functions
func generateQueryID() string {
	return fmt.Sprintf("query_%d", time.Now().UnixNano())
}

func generateResponseID() string {
	return fmt.Sprintf("response_%d", time.Now().UnixNano())
}

func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}

func hashQuery(query string) string {
	// Simple hash function for caching
	return fmt.Sprintf("%x", []byte(strings.ToLower(strings.TrimSpace(query))))
}

// newRetriever creates a new retriever instance (placeholder implementation)
func newRetriever(vectorStore VectorStore, embeddings EmbeddingsService, logger *logger.Logger) Retriever {
	// TODO: Implement proper retriever
	return &mockRetriever{
		vectorStore: vectorStore,
		embeddings:  embeddings,
		logger:      logger,
	}
}

// newGenerator creates a new generator instance (placeholder implementation)
func newGenerator(aiService interface{}, logger *logger.Logger) Generator {
	// TODO: Implement proper generator
	return &mockGenerator{
		aiService: aiService,
		logger:    logger,
	}
}

// mockRetriever is a placeholder implementation of Retriever
type mockRetriever struct {
	vectorStore VectorStore
	embeddings  EmbeddingsService
	logger      *logger.Logger
}

func (r *mockRetriever) Retrieve(ctx context.Context, query Query, config RetrievalConfig) ([]RetrievalResult, error) {
	// Simple mock implementation
	return []RetrievalResult{}, nil
}

func (r *mockRetriever) RetrieveWithReranking(ctx context.Context, query Query, config RetrievalConfig) ([]RetrievalResult, error) {
	// Simple mock implementation
	return []RetrievalResult{}, nil
}

// mockGenerator is a placeholder implementation of Generator
type mockGenerator struct {
	aiService interface{}
	logger    *logger.Logger
}

func (g *mockGenerator) Generate(ctx context.Context, query Query, sources []RetrievalResult, config GenerationConfig) (*RAGResponse, error) {
	// Simple mock implementation
	return &RAGResponse{
		ID:            generateResponseID(),
		QueryID:       query.ID,
		GeneratedText: "Mock response - RAG functionality not yet fully implemented",
		Sources:       sources,
		Confidence:    0.5,
		ModelUsed:     "mock",
		CreatedAt:     time.Now(),
	}, nil
}


