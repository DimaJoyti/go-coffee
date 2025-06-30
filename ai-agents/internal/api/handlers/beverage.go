package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/internal/ai"
	"go-coffee-ai-agents/internal/common"
	"go-coffee-ai-agents/internal/httputils"
	"go-coffee-ai-agents/internal/observability"
)

// BeverageHandler handles beverage-related HTTP requests
type BeverageHandler struct {
	logger  *observability.StructuredLogger
	metrics *observability.MetricsCollector
	tracing *observability.TracingHelper
}

// NewBeverageHandler creates a new beverage handler
func NewBeverageHandler(
	logger *observability.StructuredLogger,
	metrics *observability.MetricsCollector,
	tracing *observability.TracingHelper,
) *BeverageHandler {
	return &BeverageHandler{
		logger:  logger,
		metrics: metrics,
		tracing: tracing,
	}
}

// BeverageRequest represents a beverage creation/update request
type BeverageRequest struct {
	Name        string                 `json:"name" validate:"required,min=1,max=100"`
	Description string                 `json:"description" validate:"max=500"`
	Theme       string                 `json:"theme" validate:"required,min=1,max=50"`
	Ingredients []IngredientRequest    `json:"ingredients" validate:"required,min=1"`
	Instructions []string              `json:"instructions" validate:"required,min=1"`
	PrepTime    int                    `json:"prep_time_minutes" validate:"min=1,max=480"`
	Servings    int                    `json:"servings" validate:"min=1,max=20"`
	Difficulty  string                 `json:"difficulty" validate:"required,oneof=easy medium hard"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// IngredientRequest represents an ingredient in a beverage request
type IngredientRequest struct {
	Name       string  `json:"name" validate:"required,min=1,max=100"`
	Quantity   string  `json:"quantity" validate:"required,min=1,max=50"`
	Unit       string  `json:"unit" validate:"max=20"`
	Type       string  `json:"type" validate:"required,oneof=base flavor garnish sweetener"`
	Optional   bool    `json:"optional"`
	Substitute string  `json:"substitute" validate:"max=100"`
}

// BeverageGenerateRequest represents a beverage generation request
type BeverageGenerateRequest struct {
	Theme        string   `json:"theme" validate:"required,min=1,max=100"`
	Ingredients  []string `json:"ingredients"`
	Dietary      []string `json:"dietary_restrictions"`
	Complexity   string   `json:"complexity" validate:"oneof=simple medium complex"`
	Style        string   `json:"style" validate:"oneof=traditional modern fusion"`
	Temperature  string   `json:"temperature" validate:"oneof=hot cold room"`
	Count        int      `json:"count" validate:"min=1,max=10"`
}

// BeverageResponse represents a beverage response
type BeverageResponse struct {
	ID           uuid.UUID              `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Theme        string                 `json:"theme"`
	Ingredients  []IngredientResponse   `json:"ingredients"`
	Instructions []string               `json:"instructions"`
	PrepTime     int                    `json:"prep_time_minutes"`
	Servings     int                    `json:"servings"`
	Difficulty   string                 `json:"difficulty"`
	Tags         []string               `json:"tags"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	CreatedBy    string                 `json:"created_by"`
	URL          string                 `json:"url"`
}

// IngredientResponse represents an ingredient in a beverage response
type IngredientResponse struct {
	Name       string `json:"name"`
	Quantity   string `json:"quantity"`
	Unit       string `json:"unit"`
	Type       string `json:"type"`
	Optional   bool   `json:"optional"`
	Substitute string `json:"substitute,omitempty"`
}

// List handles GET /api/v1/beverages
func (h *BeverageHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "GET", "/api/v1/beverages", r.UserAgent())
	defer span.End()

	h.logger.InfoContext(ctx, "Listing beverages")

	// Parse pagination parameters
	pagination := httputils.ParsePaginationParams(r)
	sort := httputils.ParseSortParams(r, "created_at")
	filter := httputils.ParseFilterParams(r)

	h.logger.DebugContext(ctx, "Parsed request parameters",
		"page", pagination.Page,
		"per_page", pagination.PerPage,
		"sort_field", sort.Field,
		"sort_order", sort.Order,
		"search", filter.Search)

	// TODO: Implement actual beverage listing from database
	// For now, return mock data
	beverages := []BeverageResponse{
		{
			ID:          uuid.New(),
			Name:        "Mars Colony Coffee",
			Description: "A robust coffee blend designed for the harsh Martian environment",
			Theme:       "Mars Base",
			Ingredients: []IngredientResponse{
				{Name: "Coffee Beans", Quantity: "200", Unit: "g", Type: "base"},
				{Name: "Protein Powder", Quantity: "30", Unit: "g", Type: "flavor"},
			},
			Instructions: []string{"Grind coffee beans", "Mix with protein powder", "Brew with hot water"},
			PrepTime:     15,
			Servings:     2,
			Difficulty:   "medium",
			Tags:         []string{"coffee", "protein", "mars"},
			CreatedAt:    time.Now().Add(-24 * time.Hour),
			UpdatedAt:    time.Now().Add(-24 * time.Hour),
			CreatedBy:    "beverage-inventor-agent",
			URL:          httputils.BuildResourceURL(r, "beverages", uuid.New()),
		},
	}

	total := len(beverages)

	h.tracing.RecordSuccess(span, "Beverages listed successfully")
	h.logger.InfoContext(ctx, "Beverages listed successfully",
		"count", len(beverages),
		"total", total)

	httputils.WritePaginatedResponse(w, http.StatusOK, beverages, pagination, total)
}

// Create handles POST /api/v1/beverages
func (h *BeverageHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "POST", "/api/v1/beverages", r.UserAgent())
	defer span.End()

	h.logger.InfoContext(ctx, "Creating beverage")

	// Decode request body
	var req BeverageRequest
	if err := httputils.DecodeJSONBody(r, &req); err != nil {
		h.tracing.RecordError(span, err, "Failed to decode request body")
		httputils.WriteErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// TODO: Validate request using validator
	// TODO: Create beverage in database

	// Create response
	beverageID := uuid.New()
	response := BeverageResponse{
		ID:           beverageID,
		Name:         req.Name,
		Description:  req.Description,
		Theme:        req.Theme,
		Ingredients:  convertIngredients(req.Ingredients),
		Instructions: req.Instructions,
		PrepTime:     req.PrepTime,
		Servings:     req.Servings,
		Difficulty:   req.Difficulty,
		Tags:         req.Tags,
		Metadata:     req.Metadata,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		CreatedBy:    httputils.GetUserID(r),
		URL:          httputils.BuildResourceURL(r, "beverages", beverageID),
	}

	h.tracing.RecordSuccess(span, "Beverage created successfully")
	h.logger.InfoContext(ctx, "Beverage created successfully",
		"beverage_id", beverageID,
		"name", req.Name,
		"theme", req.Theme)

	httputils.WriteJSONResponse(w, http.StatusCreated, response)
}

// Get handles GET /api/v1/beverages/{id}
func (h *BeverageHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "GET", "/api/v1/beverages/{id}", r.UserAgent())
	defer span.End()

	// Parse beverage ID
	beverageID, err := httputils.ParseUUIDParam(r, "id")
	if err != nil {
		h.tracing.RecordError(span, err, "Invalid beverage ID")
		httputils.WriteErrorResponse(w, http.StatusBadRequest, "invalid_id", "Invalid beverage ID")
		return
	}

	h.logger.InfoContext(ctx, "Getting beverage", "beverage_id", beverageID)

	// TODO: Get beverage from database
	// For now, return mock data
	response := BeverageResponse{
		ID:          beverageID,
		Name:        "Mars Colony Coffee",
		Description: "A robust coffee blend designed for the harsh Martian environment",
		Theme:       "Mars Base",
		Ingredients: []IngredientResponse{
			{Name: "Coffee Beans", Quantity: "200", Unit: "g", Type: "base"},
			{Name: "Protein Powder", Quantity: "30", Unit: "g", Type: "flavor"},
		},
		Instructions: []string{"Grind coffee beans", "Mix with protein powder", "Brew with hot water"},
		PrepTime:     15,
		Servings:     2,
		Difficulty:   "medium",
		Tags:         []string{"coffee", "protein", "mars"},
		CreatedAt:    time.Now().Add(-24 * time.Hour),
		UpdatedAt:    time.Now().Add(-24 * time.Hour),
		CreatedBy:    "beverage-inventor-agent",
		URL:          httputils.BuildResourceURL(r, "beverages", beverageID),
	}

	h.tracing.RecordSuccess(span, "Beverage retrieved successfully")
	h.logger.InfoContext(ctx, "Beverage retrieved successfully", "beverage_id", beverageID)

	httputils.WriteJSONResponse(w, http.StatusOK, response)
}

// Update handles PUT /api/v1/beverages/{id}
func (h *BeverageHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "PUT", "/api/v1/beverages/{id}", r.UserAgent())
	defer span.End()

	// Parse beverage ID
	beverageID, err := httputils.ParseUUIDParam(r, "id")
	if err != nil {
		h.tracing.RecordError(span, err, "Invalid beverage ID")
		httputils.WriteErrorResponse(w, http.StatusBadRequest, "invalid_id", "Invalid beverage ID")
		return
	}

	h.logger.InfoContext(ctx, "Updating beverage", "beverage_id", beverageID)

	// Decode request body
	var req BeverageRequest
	if err := httputils.DecodeJSONBody(r, &req); err != nil {
		h.tracing.RecordError(span, err, "Failed to decode request body")
		httputils.WriteErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// TODO: Update beverage in database

	// Create response
	response := BeverageResponse{
		ID:           beverageID,
		Name:         req.Name,
		Description:  req.Description,
		Theme:        req.Theme,
		Ingredients:  convertIngredients(req.Ingredients),
		Instructions: req.Instructions,
		PrepTime:     req.PrepTime,
		Servings:     req.Servings,
		Difficulty:   req.Difficulty,
		Tags:         req.Tags,
		Metadata:     req.Metadata,
		CreatedAt:    time.Now().Add(-24 * time.Hour), // TODO: Get from database
		UpdatedAt:    time.Now(),
		CreatedBy:    httputils.GetUserID(r),
		URL:          httputils.BuildResourceURL(r, "beverages", beverageID),
	}

	h.tracing.RecordSuccess(span, "Beverage updated successfully")
	h.logger.InfoContext(ctx, "Beverage updated successfully",
		"beverage_id", beverageID,
		"name", req.Name)

	httputils.WriteJSONResponse(w, http.StatusOK, response)
}

// Delete handles DELETE /api/v1/beverages/{id}
func (h *BeverageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "DELETE", "/api/v1/beverages/{id}", r.UserAgent())
	defer span.End()

	// Parse beverage ID
	beverageID, err := httputils.ParseUUIDParam(r, "id")
	if err != nil {
		h.tracing.RecordError(span, err, "Invalid beverage ID")
		httputils.WriteErrorResponse(w, http.StatusBadRequest, "invalid_id", "Invalid beverage ID")
		return
	}

	h.logger.InfoContext(ctx, "Deleting beverage", "beverage_id", beverageID)

	// TODO: Delete beverage from database

	h.tracing.RecordSuccess(span, "Beverage deleted successfully")
	h.logger.InfoContext(ctx, "Beverage deleted successfully", "beverage_id", beverageID)

	httputils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Beverage deleted successfully",
		"id":      beverageID,
	})
}

// Generate handles POST /api/v1/beverages/generate
func (h *BeverageHandler) Generate(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "POST", "/api/v1/beverages/generate", r.UserAgent())
	defer span.End()

	h.logger.InfoContext(ctx, "Generating beverages with AI")

	// Decode request body
	var req BeverageGenerateRequest
	if err := httputils.DecodeJSONBody(r, &req); err != nil {
		h.tracing.RecordError(span, err, "Failed to decode request body")
		httputils.WriteErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Set defaults
	if req.Count == 0 {
		req.Count = 1
	}
	if req.Complexity == "" {
		req.Complexity = "medium"
	}

	h.logger.InfoContext(ctx, "AI beverage generation request",
		"theme", req.Theme,
		"count", req.Count,
		"complexity", req.Complexity)

	// Generate beverages using AI
	aiManager := ai.GetGlobalManager()
	if aiManager == nil {
		h.tracing.RecordError(span, nil, "AI manager not available")
		httputils.WriteErrorResponse(w, http.StatusServiceUnavailable, "ai_unavailable", "AI service is not available")
		return
	}

	aiRequest := &common.BeverageGenerationRequest{
		Theme:       req.Theme,
		Ingredients: req.Ingredients,
		Dietary:     req.Dietary,
		Complexity:  req.Complexity,
		Style:       req.Style,
		Temperature: req.Temperature,
		Count:       req.Count,
	}

	aiResponse, err := aiManager.GenerateBeverage(ctx, aiRequest)
	if err != nil {
		h.tracing.RecordError(span, err, "AI beverage generation failed")
		h.logger.ErrorContext(ctx, "AI beverage generation failed", err)
		httputils.WriteErrorResponse(w, http.StatusInternalServerError, "generation_failed", "Failed to generate beverages")
		return
	}

	// Convert AI response to API response
	beverages := make([]BeverageResponse, len(aiResponse.Beverages))
	for i, beverage := range aiResponse.Beverages {
		beverageID := uuid.New()
		beverages[i] = BeverageResponse{
			ID:          beverageID,
			Name:        beverage.Name,
			Description: beverage.Description,
			Theme:       req.Theme,
			Ingredients: convertAIIngredients(beverage.Ingredients),
			Instructions: beverage.Instructions,
			PrepTime:     beverage.PrepTime,
			Servings:     beverage.Servings,
			Difficulty:   beverage.Difficulty,
			Tags:         beverage.Tags,
			Metadata:     beverage.Metadata,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			CreatedBy:    "ai-generator",
			URL:          httputils.BuildResourceURL(r, "beverages", beverageID),
		}
	}

	response := map[string]interface{}{
		"beverages": beverages,
		"metadata": map[string]interface{}{
			"ai_provider": aiResponse.Provider,
			"ai_model":    aiResponse.Model,
			"tokens_used": aiResponse.Usage.TotalTokens,
			"generated_at": aiResponse.CreatedAt,
		},
	}

	h.tracing.RecordSuccess(span, "Beverages generated successfully")
	h.logger.InfoContext(ctx, "Beverages generated successfully",
		"count", len(beverages),
		"theme", req.Theme,
		"ai_provider", aiResponse.Provider,
		"tokens_used", aiResponse.Usage.TotalTokens)

	httputils.WriteJSONResponse(w, http.StatusOK, response)
}

// Search handles GET /api/v1/beverages/search
func (h *BeverageHandler) Search(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "GET", "/api/v1/beverages/search", r.UserAgent())
	defer span.End()

	h.logger.InfoContext(ctx, "Searching beverages")

	// Parse search parameters
	pagination := httputils.ParsePaginationParams(r)
	filter := httputils.ParseFilterParams(r)

	h.logger.DebugContext(ctx, "Search parameters",
		"search", filter.Search,
		"filters", filter.Fields)

	// TODO: Implement actual search
	// For now, return empty results
	beverages := []BeverageResponse{}
	total := 0

	h.tracing.RecordSuccess(span, "Beverage search completed")
	h.logger.InfoContext(ctx, "Beverage search completed",
		"results_count", len(beverages),
		"search_query", filter.Search)

	httputils.WritePaginatedResponse(w, http.StatusOK, beverages, pagination, total)
}

// Statistics handles GET /api/v1/beverages/stats
func (h *BeverageHandler) Statistics(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "GET", "/api/v1/beverages/stats", r.UserAgent())
	defer span.End()

	h.logger.InfoContext(ctx, "Getting beverage statistics")

	// TODO: Implement actual statistics
	stats := map[string]interface{}{
		"total_beverages":     42,
		"themes":             []string{"Mars Base", "Space Station", "Earth Classic"},
		"popular_ingredients": []string{"Coffee", "Tea", "Protein Powder"},
		"difficulty_distribution": map[string]int{
			"easy":   15,
			"medium": 20,
			"hard":   7,
		},
		"generated_today": 5,
		"last_updated":   time.Now(),
	}

	h.tracing.RecordSuccess(span, "Beverage statistics retrieved")
	h.logger.InfoContext(ctx, "Beverage statistics retrieved")

	httputils.WriteJSONResponse(w, http.StatusOK, stats)
}

// Helper functions

func convertIngredients(ingredients []IngredientRequest) []IngredientResponse {
	result := make([]IngredientResponse, len(ingredients))
	for i, ing := range ingredients {
		result[i] = IngredientResponse{
			Name:       ing.Name,
			Quantity:   ing.Quantity,
			Unit:       ing.Unit,
			Type:       ing.Type,
			Optional:   ing.Optional,
			Substitute: ing.Substitute,
		}
	}
	return result
}

func convertAIIngredients(ingredients []common.BeverageIngredient) []IngredientResponse {
	result := make([]IngredientResponse, len(ingredients))
	for i, ing := range ingredients {
		result[i] = IngredientResponse{
			Name:       ing.Name,
			Quantity:   ing.Quantity,
			Unit:       ing.Unit,
			Type:       ing.Type,
			Optional:   ing.Optional,
			Substitute: ing.Substitute,
		}
	}
	return result
}
