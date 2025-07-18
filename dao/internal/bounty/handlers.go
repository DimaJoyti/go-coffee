package bounty

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// HTTP Handlers for Bounty operations

// GetBountiesHandler handles GET /api/v1/bounties
func (s *Service) GetBountiesHandler(c *gin.Context) {
	// Parse query parameters
	var req GetBountiesRequest

	if status := c.Query("status"); status != "" {
		if statusInt, err := strconv.Atoi(status); err == nil {
			bountyStatus := BountyStatus(statusInt)
			req.Status = &bountyStatus
		}
	}

	if category := c.Query("category"); category != "" {
		if categoryInt, err := strconv.Atoi(category); err == nil {
			bountyCategory := BountyCategory(categoryInt)
			req.Category = &bountyCategory
		}
	}

	req.CreatorAddress = c.Query("creator")
	req.AssigneeAddress = c.Query("assignee")

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

	// Get bounties
	bounties, total, err := s.bountyRepo.List(c.Request.Context(), &BountyFilter{
		Status:          req.Status,
		Category:        req.Category,
		CreatorAddress:  req.CreatorAddress,
		AssigneeAddress: req.AssigneeAddress,
		Limit:           req.Limit,
		Offset:          req.Offset,
	})
	if err != nil {
		s.logger.Error("Failed to get bounties", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get bounties"})
		return
	}

	response := &GetBountiesResponse{
		Bounties: bounties,
		Total:    total,
		Limit:    req.Limit,
		Offset:   req.Offset,
	}

	c.JSON(http.StatusOK, response)
}

// CreateBountyHandler handles POST /api/v1/bounties
func (s *Service) CreateBountyHandler(c *gin.Context) {
	var req CreateBountyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate required fields
	if req.Title == "" || req.Description == "" || req.CreatorAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	// Create bounty
	response, err := s.CreateBounty(c.Request.Context(), &req)
	if err != nil {
		s.logger.Error("Failed to create bounty", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetBountyHandler handles GET /api/v1/bounties/:id
func (s *Service) GetBountyHandler(c *gin.Context) {
	bountyIDStr := c.Param("id")
	if bountyIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing bounty ID"})
		return
	}

	bountyID, err := strconv.ParseUint(bountyIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bounty ID"})
		return
	}

	bountyDetails, err := s.GetBounty(c.Request.Context(), bountyID)
	if err != nil {
		s.logger.Error("Failed to get bounty", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Bounty not found"})
		return
	}

	c.JSON(http.StatusOK, bountyDetails)
}

// ApplyForBountyHandler handles POST /api/v1/bounties/:id/apply
func (s *Service) ApplyForBountyHandler(c *gin.Context) {
	bountyIDStr := c.Param("id")
	if bountyIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing bounty ID"})
		return
	}

	bountyID, err := strconv.ParseUint(bountyIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bounty ID"})
		return
	}

	var req ApplyBountyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	req.BountyID = bountyID

	// Validate required fields
	if req.ApplicantAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing applicant address"})
		return
	}

	// Apply for bounty
	response, err := s.ApplyForBounty(c.Request.Context(), &req)
	if err != nil {
		s.logger.Error("Failed to apply for bounty", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// AssignBountyHandler handles POST /api/v1/bounties/:id/assign
func (s *Service) AssignBountyHandler(c *gin.Context) {
	bountyIDStr := c.Param("id")
	if bountyIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing bounty ID"})
		return
	}

	bountyID, err := strconv.ParseUint(bountyIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bounty ID"})
		return
	}

	var req AssignBountyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	req.BountyID = bountyID

	// Validate required fields
	if req.AssigneeAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing assignee address"})
		return
	}

	// Assign bounty
	response, err := s.AssignBounty(c.Request.Context(), &req)
	if err != nil {
		s.logger.Error("Failed to assign bounty", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// StartBountyHandler handles POST /api/v1/bounties/:id/start
func (s *Service) StartBountyHandler(c *gin.Context) {
	bountyIDStr := c.Param("id")
	if bountyIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing bounty ID"})
		return
	}

	bountyID, err := strconv.ParseUint(bountyIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bounty ID"})
		return
	}

	// Get bounty and update status to IN_PROGRESS
	bounty, err := s.bountyRepo.GetByID(c.Request.Context(), bountyID)
	if err != nil {
		s.logger.Error("Failed to get bounty", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Bounty not found"})
		return
	}

	if bounty.Status != BountyStatusAssigned {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bounty is not assigned"})
		return
	}

	bounty.Status = BountyStatusInProgress
	if err := s.bountyRepo.Update(c.Request.Context(), bounty); err != nil {
		s.logger.Error("Failed to update bounty status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start bounty"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bounty_id": bountyID,
		"status":    bounty.Status.String(),
		"message":   "Bounty started successfully",
	})
}

// SubmitBountyHandler handles POST /api/v1/bounties/:id/submit
func (s *Service) SubmitBountyHandler(c *gin.Context) {
	bountyIDStr := c.Param("id")
	if bountyIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing bounty ID"})
		return
	}

	bountyID, err := strconv.ParseUint(bountyIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bounty ID"})
		return
	}

	// Get bounty and update status to SUBMITTED
	bounty, err := s.bountyRepo.GetByID(c.Request.Context(), bountyID)
	if err != nil {
		s.logger.Error("Failed to get bounty", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Bounty not found"})
		return
	}

	if bounty.Status != BountyStatusInProgress {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bounty is not in progress"})
		return
	}

	bounty.Status = BountyStatusSubmitted
	if err := s.bountyRepo.Update(c.Request.Context(), bounty); err != nil {
		s.logger.Error("Failed to update bounty status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit bounty"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bounty_id": bountyID,
		"status":    bounty.Status.String(),
		"message":   "Bounty submitted successfully",
	})
}

// CompleteMilestoneHandler handles POST /api/v1/bounties/:id/milestones/:milestone_id/complete
func (s *Service) CompleteMilestoneHandler(c *gin.Context) {
	bountyIDStr := c.Param("id")
	milestoneIDStr := c.Param("milestone_id")

	if bountyIDStr == "" || milestoneIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing bounty ID or milestone ID"})
		return
	}

	bountyID, err := strconv.ParseUint(bountyIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bounty ID"})
		return
	}

	milestoneIndex, err := strconv.ParseUint(milestoneIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid milestone ID"})
		return
	}

	req := &CompleteMilestoneRequest{
		BountyID:       bountyID,
		MilestoneIndex: milestoneIndex,
	}

	// Complete milestone
	response, err := s.CompleteMilestone(c.Request.Context(), req)
	if err != nil {
		s.logger.Error("Failed to complete milestone", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetBountyApplicationsHandler handles GET /api/v1/bounties/:id/applications
func (s *Service) GetBountyApplicationsHandler(c *gin.Context) {
	bountyIDStr := c.Param("id")
	if bountyIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing bounty ID"})
		return
	}

	bountyID, err := strconv.ParseUint(bountyIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bounty ID"})
		return
	}

	applications, err := s.applicationRepo.GetByBountyID(c.Request.Context(), bountyID)
	if err != nil {
		s.logger.Error("Failed to get bounty applications", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get applications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bounty_id":    bountyID,
		"applications": applications,
	})
}

// VerifyPerformanceHandler handles POST /api/v1/performance/verify
func (s *Service) VerifyPerformanceHandler(c *gin.Context) {
	var req VerifyPerformanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get bounty
	bounty, err := s.bountyRepo.GetByID(c.Request.Context(), req.BountyID)
	if err != nil {
		s.logger.Error("Failed to get bounty", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Bounty not found"})
		return
	}

	// Update performance metrics
	bounty.TVLImpact = req.TVLImpact
	bounty.MAUImpact = req.MAUImpact
	bounty.PerformanceVerified = true

	if err := s.bountyRepo.Update(c.Request.Context(), bounty); err != nil {
		s.logger.Error("Failed to update bounty performance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify performance"})
		return
	}

	// Calculate bonus based on performance
	bonusEarned := s.calculatePerformanceBonus(req.TVLImpact, req.MAUImpact)
	reputationBonus := s.calculateReputationBonus(req.TVLImpact, req.MAUImpact)

	response := &VerifyPerformanceResponse{
		BountyID:        req.BountyID,
		TVLImpact:       req.TVLImpact,
		MAUImpact:       req.MAUImpact,
		BonusEarned:     bonusEarned,
		ReputationBonus: reputationBonus,
	}

	c.JSON(http.StatusOK, response)
}

// GetPerformanceStatsHandler handles GET /api/v1/performance/stats
func (s *Service) GetPerformanceStatsHandler(c *gin.Context) {
	// Mock performance stats
	stats := gin.H{
		"total_bounties_completed": 25,
		"total_tvl_impact":         "5000000.00", // $5M
		"total_mau_impact":         15000,
		"average_completion_time":  14, // days
		"top_performers": []gin.H{
			{
				"address":            "0x1234567890123456789012345678901234567890",
				"bounties_completed": 5,
				"tvl_contributed":    "1000000.00",
				"reputation_score":   95,
			},
		},
	}

	c.JSON(http.StatusOK, stats)
}

// GetDeveloperBountiesHandler handles GET /api/v1/developers/:address/bounties
func (s *Service) GetDeveloperBountiesHandler(c *gin.Context) {
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

	bounties, err := s.bountyRepo.GetByAssignee(c.Request.Context(), address, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get developer bounties", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get bounties"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address":  address,
		"bounties": bounties,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetDeveloperReputationHandler handles GET /api/v1/developers/:address/reputation
func (s *Service) GetDeveloperReputationHandler(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing address"})
		return
	}

	// Get reputation from cache or calculate
	reputation := 0
	if cached, exists := s.reputationCache[address]; exists {
		reputation = cached
	}

	c.JSON(http.StatusOK, gin.H{
		"address":          address,
		"reputation_score": reputation,
		"last_updated":     time.Now(),
	})
}

// GetDeveloperLeaderboardHandler handles GET /api/v1/developers/leaderboard
func (s *Service) GetDeveloperLeaderboardHandler(c *gin.Context) {
	// Parse limit
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if limitInt, err := strconv.Atoi(limitStr); err == nil && limitInt > 0 && limitInt <= 50 {
			limit = limitInt
		}
	}

	developers, err := s.developerRepo.GetTopByReputation(c.Request.Context(), limit)
	if err != nil {
		s.logger.Error("Failed to get developer leaderboard", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get leaderboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": developers,
		"limit":       limit,
	})
}

// Helper methods

func (s *Service) calculatePerformanceBonus(tvlImpact decimal.Decimal, mauImpact int) decimal.Decimal {
	// Simple bonus calculation based on impact
	bonus := decimal.Zero
	if tvlImpact.GreaterThan(decimal.NewFromFloat(1000000)) { // > $1M TVL
		bonus = bonus.Add(decimal.NewFromFloat(100))
	}
	if mauImpact > 1000 { // > 1000 MAU
		bonus = bonus.Add(decimal.NewFromFloat(50))
	}
	return bonus
}

func (s *Service) calculateReputationBonus(tvlImpact decimal.Decimal, mauImpact int) int {
	// Simple reputation bonus calculation
	bonus := 0
	if tvlImpact.GreaterThan(decimal.NewFromFloat(1000000)) {
		bonus += 20
	}
	if mauImpact > 1000 {
		bonus += 10
	}
	return bonus
}
