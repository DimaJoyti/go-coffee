package marketplace

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// HTTP Handlers for Marketplace operations

// GetSolutionsHandler handles GET /api/v1/solutions
func (s *Service) GetSolutionsHandler(c *gin.Context) {
	// Parse query parameters
	var req GetSolutionsRequest

	if category := c.Query("category"); category != "" {
		if categoryInt, err := strconv.Atoi(category); err == nil {
			solutionCategory := SolutionCategory(categoryInt)
			req.Category = &solutionCategory
		}
	}

	if status := c.Query("status"); status != "" {
		if statusInt, err := strconv.Atoi(status); err == nil {
			solutionStatus := SolutionStatus(statusInt)
			req.Status = &solutionStatus
		}
	}

	req.DeveloperAddress = c.Query("developer")

	if minRating := c.Query("min_rating"); minRating != "" {
		if rating, err := strconv.ParseFloat(minRating, 64); err == nil {
			req.MinRating = rating
		}
	}

	// Parse pagination
	req.Limit = 20 // default
	if limit := c.Query("limit"); limit != "" {
		if limitInt, err := strconv.Atoi(limit); err == nil && limitInt > 0 && limitInt <= 100 {
			req.Limit = limitInt
		}
	}

	req.Offset = 0 // default
	if offset := c.Query("offset"); offset != "" {
		if offsetInt, err := strconv.Atoi(offset); err == nil && offsetInt >= 0 {
			req.Offset = offsetInt
		}
	}

	req.SortBy = c.Query("sort_by")
	req.SortOrder = c.Query("sort_order")

	// Get solutions
	solutions, total, err := s.solutionRepo.List(c.Request.Context(), &SolutionFilter{
		Category:         req.Category,
		Status:           req.Status,
		DeveloperAddress: req.DeveloperAddress,
		MinRating:        req.MinRating,
		Limit:            req.Limit,
		Offset:           req.Offset,
		SortBy:           req.SortBy,
		SortOrder:        req.SortOrder,
	})
	if err != nil {
		s.logger.Error("Failed to get solutions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get solutions"})
		return
	}

	response := &GetSolutionsResponse{
		Solutions: solutions,
		Total:     total,
		Limit:     req.Limit,
		Offset:    req.Offset,
	}

	c.JSON(http.StatusOK, response)
}

// CreateSolutionHandler handles POST /api/v1/solutions
func (s *Service) CreateSolutionHandler(c *gin.Context) {
	var req CreateSolutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate required fields
	if req.Name == "" || req.Description == "" || req.DeveloperAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	// Create solution
	response, err := s.CreateSolution(c.Request.Context(), &req)
	if err != nil {
		s.logger.Error("Failed to create solution", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetSolutionHandler handles GET /api/v1/solutions/:id
func (s *Service) GetSolutionHandler(c *gin.Context) {
	solutionIDStr := c.Param("id")
	if solutionIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing solution ID"})
		return
	}

	solutionID, err := strconv.ParseUint(solutionIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid solution ID"})
		return
	}

	solutionDetails, err := s.GetSolution(c.Request.Context(), solutionID)
	if err != nil {
		s.logger.Error("Failed to get solution", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Solution not found"})
		return
	}

	c.JSON(http.StatusOK, solutionDetails)
}

// UpdateSolutionHandler handles PUT /api/v1/solutions/:id
func (s *Service) UpdateSolutionHandler(c *gin.Context) {
	solutionIDStr := c.Param("id")
	if solutionIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing solution ID"})
		return
	}

	solutionID, err := strconv.ParseUint(solutionIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid solution ID"})
		return
	}

	var req UpdateSolutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	req.SolutionID = solutionID

	// Get existing solution
	solution, err := s.solutionRepo.GetByID(c.Request.Context(), solutionID)
	if err != nil {
		s.logger.Error("Failed to get solution", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Solution not found"})
		return
	}

	// Update fields
	if req.Version != "" {
		solution.Version = req.Version
	}
	if req.Description != "" {
		solution.Description = req.Description
	}
	if req.RepositoryURL != "" {
		solution.RepositoryURL = req.RepositoryURL
	}
	if req.DocumentationURL != "" {
		solution.DocumentationURL = req.DocumentationURL
	}
	if req.DemoURL != "" {
		solution.DemoURL = req.DemoURL
	}
	if len(req.Tags) > 0 {
		solution.Tags = req.Tags
	}

	if err := s.solutionRepo.Update(c.Request.Context(), solution); err != nil {
		s.logger.Error("Failed to update solution", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update solution"})
		return
	}

	// Recalculate quality score
	qualityScore, err := s.calculateQualityScore(c.Request.Context(), solution)
	if err != nil {
		s.logger.Error("Failed to calculate quality score", zap.Error(err))
	} else {
		s.qualityScoreCache[solutionID] = qualityScore
	}

	response := &UpdateSolutionResponse{
		SolutionID:   solutionID,
		Status:       solution.Status,
		QualityScore: qualityScore,
	}

	c.JSON(http.StatusOK, response)
}

// ReviewSolutionHandler handles POST /api/v1/solutions/:id/review
func (s *Service) ReviewSolutionHandler(c *gin.Context) {
	solutionIDStr := c.Param("id")
	if solutionIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing solution ID"})
		return
	}

	solutionID, err := strconv.ParseUint(solutionIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid solution ID"})
		return
	}

	var req ReviewSolutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	req.SolutionID = solutionID

	// Review solution
	response, err := s.ReviewSolution(c.Request.Context(), &req)
	if err != nil {
		s.logger.Error("Failed to review solution", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ApproveSolutionHandler handles POST /api/v1/solutions/:id/approve
func (s *Service) ApproveSolutionHandler(c *gin.Context) {
	solutionIDStr := c.Param("id")
	if solutionIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing solution ID"})
		return
	}

	solutionID, err := strconv.ParseUint(solutionIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid solution ID"})
		return
	}

	var req ApproveSolutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	req.SolutionID = solutionID

	// Approve solution
	response, err := s.ApproveSolution(c.Request.Context(), &req)
	if err != nil {
		s.logger.Error("Failed to approve solution", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// InstallSolutionHandler handles POST /api/v1/solutions/:id/install
func (s *Service) InstallSolutionHandler(c *gin.Context) {
	solutionIDStr := c.Param("id")
	if solutionIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing solution ID"})
		return
	}

	solutionID, err := strconv.ParseUint(solutionIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid solution ID"})
		return
	}

	var req InstallSolutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	req.SolutionID = solutionID

	// Install solution
	response, err := s.InstallSolution(c.Request.Context(), &req)
	if err != nil {
		s.logger.Error("Failed to install solution", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CheckCompatibilityHandler handles GET /api/v1/solutions/:id/compatibility
func (s *Service) CheckCompatibilityHandler(c *gin.Context) {
	solutionIDStr := c.Param("id")
	if solutionIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing solution ID"})
		return
	}

	solutionID, err := strconv.ParseUint(solutionIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid solution ID"})
		return
	}

	environment := c.Query("environment")
	if environment == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing environment parameter"})
		return
	}

	result, err := s.checkCompatibility(c.Request.Context(), solutionID, environment)
	if err != nil {
		s.logger.Error("Failed to check compatibility", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check compatibility"})
		return
	}

	response := &CompatibilityResponse{
		Result: result,
	}

	c.JSON(http.StatusOK, response)
}

// GetSolutionReviewsHandler handles GET /api/v1/solutions/:id/reviews
func (s *Service) GetSolutionReviewsHandler(c *gin.Context) {
	solutionIDStr := c.Param("id")
	if solutionIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing solution ID"})
		return
	}

	solutionID, err := strconv.ParseUint(solutionIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid solution ID"})
		return
	}

	reviews, err := s.reviewRepo.GetBySolutionID(c.Request.Context(), solutionID)
	if err != nil {
		s.logger.Error("Failed to get solution reviews", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get reviews"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"solution_id": solutionID,
		"reviews":     reviews,
	})
}

// GetCategoriesHandler handles GET /api/v1/categories
func (s *Service) GetCategoriesHandler(c *gin.Context) {
	categories, err := s.categoryRepo.GetAll(c.Request.Context())
	if err != nil {
		s.logger.Error("Failed to get categories", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
	})
}

// GetSolutionsByCategoryHandler handles GET /api/v1/categories/:category/solutions
func (s *Service) GetSolutionsByCategoryHandler(c *gin.Context) {
	categoryStr := c.Param("category")
	if categoryStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing category"})
		return
	}

	categoryInt, err := strconv.Atoi(categoryStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
		return
	}

	category := SolutionCategory(categoryInt)

	// Parse pagination
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if limitInt, err := strconv.Atoi(limitStr); err == nil && limitInt > 0 && limitInt <= 100 {
			limit = limitInt
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offsetInt, err := strconv.Atoi(offsetStr); err == nil && offsetInt >= 0 {
			offset = offsetInt
		}
	}

	solutions, err := s.solutionRepo.GetByCategory(c.Request.Context(), category, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get solutions by category", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get solutions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"category":  category.String(),
		"solutions": solutions,
		"limit":     limit,
		"offset":    offset,
	})
}

// CalculateQualityScoreHandler handles POST /api/v1/quality/score
func (s *Service) CalculateQualityScoreHandler(c *gin.Context) {
	var req QualityScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get solution
	solution, err := s.solutionRepo.GetByID(c.Request.Context(), req.SolutionID)
	if err != nil {
		s.logger.Error("Failed to get solution", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Solution not found"})
		return
	}

	// Calculate quality score
	qualityScore, err := s.calculateQualityScore(c.Request.Context(), solution)
	if err != nil {
		s.logger.Error("Failed to calculate quality score", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate quality score"})
		return
	}

	// Update cache
	s.qualityScoreCache[req.SolutionID] = qualityScore

	response := &QualityScoreResponse{
		QualityScore: qualityScore,
	}

	c.JSON(http.StatusOK, response)
}

// GetQualityMetricsHandler handles GET /api/v1/quality/metrics
func (s *Service) GetQualityMetricsHandler(c *gin.Context) {
	// Mock quality metrics
	metrics := gin.H{
		"average_overall_score":       4.2,
		"average_security_score":      4.1,
		"average_performance_score":   4.3,
		"average_usability_score":     4.0,
		"average_documentation_score": 3.8,
		"total_reviews":               156,
		"quality_distribution": gin.H{
			"excellent": 45, // 4.5-5.0
			"good":      67, // 3.5-4.4
			"average":   32, // 2.5-3.4
			"poor":      12, // 1.0-2.4
		},
	}

	c.JSON(http.StatusOK, metrics)
}

// GetDeveloperSolutionsHandler handles GET /api/v1/developers/:address/solutions
func (s *Service) GetDeveloperSolutionsHandler(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing address"})
		return
	}

	// Parse pagination
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if limitInt, err := strconv.Atoi(limitStr); err == nil && limitInt > 0 && limitInt <= 100 {
			limit = limitInt
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offsetInt, err := strconv.Atoi(offsetStr); err == nil && offsetInt >= 0 {
			offset = offsetInt
		}
	}

	solutions, err := s.solutionRepo.GetByDeveloper(c.Request.Context(), address, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get developer solutions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get solutions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address":   address,
		"solutions": solutions,
		"limit":     limit,
		"offset":    offset,
	})
}

// GetDeveloperReviewsHandler handles GET /api/v1/developers/:address/reviews
func (s *Service) GetDeveloperReviewsHandler(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing address"})
		return
	}

	// Parse pagination
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if limitInt, err := strconv.Atoi(limitStr); err == nil && limitInt > 0 && limitInt <= 100 {
			limit = limitInt
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offsetInt, err := strconv.Atoi(offsetStr); err == nil && offsetInt >= 0 {
			offset = offsetInt
		}
	}

	reviews, err := s.reviewRepo.GetByReviewer(c.Request.Context(), address, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get developer reviews", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get reviews"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address": address,
		"reviews": reviews,
		"limit":   limit,
		"offset":  offset,
	})
}

// GetPopularSolutionsHandler handles GET /api/v1/analytics/popular
func (s *Service) GetPopularSolutionsHandler(c *gin.Context) {
	// Parse pagination
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if limitInt, err := strconv.Atoi(limitStr); err == nil && limitInt > 0 && limitInt <= 50 {
			limit = limitInt
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offsetInt, err := strconv.Atoi(offsetStr); err == nil && offsetInt >= 0 {
			offset = offsetInt
		}
	}

	solutions, err := s.solutionRepo.GetPopular(c.Request.Context(), limit, offset)
	if err != nil {
		s.logger.Error("Failed to get popular solutions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get popular solutions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"popular_solutions": solutions,
		"limit":             limit,
		"offset":            offset,
	})
}

// GetTrendingSolutionsHandler handles GET /api/v1/analytics/trending
func (s *Service) GetTrendingSolutionsHandler(c *gin.Context) {
	// Parse pagination
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if limitInt, err := strconv.Atoi(limitStr); err == nil && limitInt > 0 && limitInt <= 50 {
			limit = limitInt
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offsetInt, err := strconv.Atoi(offsetStr); err == nil && offsetInt >= 0 {
			offset = offsetInt
		}
	}

	solutions, err := s.solutionRepo.GetTrending(c.Request.Context(), limit, offset)
	if err != nil {
		s.logger.Error("Failed to get trending solutions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get trending solutions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"trending_solutions": solutions,
		"limit":              limit,
		"offset":             offset,
	})
}

// GetMarketplaceStatsHandler handles GET /api/v1/analytics/stats
func (s *Service) GetMarketplaceStatsHandler(c *gin.Context) {
	// Mock marketplace statistics
	stats := &MarketplaceStats{
		TotalSolutions:     71,
		ApprovedSolutions:  58,
		TotalInstallations: 342,
		TotalReviews:       156,
		AverageRating:      decimal.NewFromFloat(4.2),
		TopCategories: []CategoryStats{
			{Category: SolutionCategoryDeFi, Name: "DeFi", SolutionCount: 15, InstallCount: 89, AverageRating: decimal.NewFromFloat(4.3)},
			{Category: SolutionCategoryDAO, Name: "DAO", SolutionCount: 12, InstallCount: 67, AverageRating: decimal.NewFromFloat(4.1)},
			{Category: SolutionCategoryInfrastructure, Name: "Infrastructure", SolutionCount: 10, InstallCount: 54, AverageRating: decimal.NewFromFloat(4.4)},
		},
		RecentActivity: []ActivityItem{
			{Type: "solution_created", SolutionID: 71, SolutionName: "Advanced DeFi Analytics", UserAddress: "0x1234...5678", Details: "New analytics solution submitted"},
			{Type: "solution_approved", SolutionID: 70, SolutionName: "DAO Voting Widget", UserAddress: "0x9876...4321", Details: "Solution approved for marketplace"},
			{Type: "solution_installed", SolutionID: 69, SolutionName: "Security Audit Tool", UserAddress: "0x5555...9999", Details: "Solution installed in production"},
		},
	}

	c.JSON(http.StatusOK, stats)
}
