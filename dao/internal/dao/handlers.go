package dao

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HTTP Handlers for DAO operations

// GetProposalsHandler handles GET /api/v1/proposals
func (s *Service) GetProposalsHandler(c *gin.Context) {
	// Parse query parameters
	var req GetProposalsRequest
	
	if status := c.Query("status"); status != "" {
		if statusInt, err := strconv.Atoi(status); err == nil {
			proposalStatus := ProposalStatus(statusInt)
			req.Status = &proposalStatus
		}
	}
	
	if category := c.Query("category"); category != "" {
		if categoryInt, err := strconv.Atoi(category); err == nil {
			proposalCategory := ProposalCategory(categoryInt)
			req.Category = &proposalCategory
		}
	}
	
	req.ProposerAddress = c.Query("proposer")
	
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

	// Get proposals
	response, err := s.GetProposals(c.Request.Context(), &req)
	if err != nil {
		s.logger.Error("Failed to get proposals", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get proposals"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CreateProposalHandler handles POST /api/v1/proposals
func (s *Service) CreateProposalHandler(c *gin.Context) {
	var req CreateProposalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate required fields
	if req.Title == "" || req.Description == "" || req.ProposerAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	// Create proposal
	response, err := s.CreateProposal(c.Request.Context(), &req)
	if err != nil {
		s.logger.Error("Failed to create proposal", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetProposalHandler handles GET /api/v1/proposals/:id
func (s *Service) GetProposalHandler(c *gin.Context) {
	proposalID := c.Param("id")
	if proposalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing proposal ID"})
		return
	}

	proposal, err := s.GetProposal(c.Request.Context(), proposalID)
	if err != nil {
		s.logger.Error("Failed to get proposal", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Proposal not found"})
		return
	}

	c.JSON(http.StatusOK, proposal)
}

// VoteOnProposalHandler handles POST /api/v1/proposals/:id/vote
func (s *Service) VoteOnProposalHandler(c *gin.Context) {
	proposalID := c.Param("id")
	if proposalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing proposal ID"})
		return
	}

	var req VoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	req.ProposalID = proposalID

	// Validate required fields
	if req.VoterAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing voter address"})
		return
	}

	// Cast vote
	response, err := s.VoteOnProposal(c.Request.Context(), &req)
	if err != nil {
		s.logger.Error("Failed to cast vote", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetProposalVotesHandler handles GET /api/v1/proposals/:id/votes
func (s *Service) GetProposalVotesHandler(c *gin.Context) {
	proposalID := c.Param("id")
	if proposalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing proposal ID"})
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

	votes, err := s.voteRepo.GetByProposal(c.Request.Context(), proposalID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get proposal votes", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get votes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"votes":  votes,
		"limit":  limit,
		"offset": offset,
	})
}

// GetGovernanceStatsHandler handles GET /api/v1/governance/stats
func (s *Service) GetGovernanceStatsHandler(c *gin.Context) {
	stats, err := s.GetGovernanceStats(c.Request.Context())
	if err != nil {
		s.logger.Error("Failed to get governance stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetVotingPowerHandler handles GET /api/v1/governance/voting-power/:address
func (s *Service) GetVotingPowerHandler(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing address"})
		return
	}

	power, err := s.GetVotingPower(c.Request.Context(), address)
	if err != nil {
		s.logger.Error("Failed to get voting power", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get voting power"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address":      address,
		"voting_power": power,
	})
}

// GetDelegateHandler handles GET /api/v1/governance/delegate/:address
func (s *Service) GetDelegateHandler(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing address"})
		return
	}

	// Mock delegate info - in real implementation, query from contract
	c.JSON(http.StatusOK, gin.H{
		"address":  address,
		"delegate": address, // Self-delegated by default
	})
}

// DelegateVotesHandler handles POST /api/v1/governance/delegate
func (s *Service) DelegateVotesHandler(c *gin.Context) {
	var req struct {
		DelegatorAddress string `json:"delegator_address" binding:"required"`
		DelegateAddress  string `json:"delegate_address" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Mock delegation - in real implementation, call contract
	s.logger.Info("Delegating votes",
		zap.String("delegator", req.DelegatorAddress),
		zap.String("delegate", req.DelegateAddress))

	c.JSON(http.StatusOK, gin.H{
		"transaction_hash": "0xdelegate1234567890abcdef1234567890abcdef1234567890abcdef123456",
		"delegator":        req.DelegatorAddress,
		"delegate":         req.DelegateAddress,
	})
}

// Developer-related handlers

// GetDevelopersHandler handles GET /api/v1/developers
func (s *Service) GetDevelopersHandler(c *gin.Context) {
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

	developers, err := s.developerRepo.GetActive(c.Request.Context(), limit, offset)
	if err != nil {
		s.logger.Error("Failed to get developers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get developers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"developers": developers,
		"limit":      limit,
		"offset":     offset,
	})
}

// GetDeveloperHandler handles GET /api/v1/developers/:address
func (s *Service) GetDeveloperHandler(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing address"})
		return
	}

	developer, err := s.developerRepo.GetByAddress(c.Request.Context(), address)
	if err != nil {
		s.logger.Error("Failed to get developer", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Developer not found"})
		return
	}

	c.JSON(http.StatusOK, developer)
}

// UpdateDeveloperProfileHandler handles POST /api/v1/developers/:address/profile
func (s *Service) UpdateDeveloperProfileHandler(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing address"})
		return
	}

	var req struct {
		Username        string   `json:"username"`
		Email           string   `json:"email"`
		GithubUsername  string   `json:"github_username"`
		DiscordUsername string   `json:"discord_username"`
		Bio             string   `json:"bio"`
		Skills          []string `json:"skills"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get existing profile or create new one
	profile, err := s.developerRepo.GetByAddress(c.Request.Context(), address)
	if err != nil {
		// Create new profile
		profile = &DeveloperProfile{
			WalletAddress: address,
			IsActive:      true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			LastActivity:  time.Now(),
		}
	}

	// Update profile fields
	profile.Username = req.Username
	profile.Email = req.Email
	profile.GithubUsername = req.GithubUsername
	profile.DiscordUsername = req.DiscordUsername
	profile.Bio = req.Bio
	profile.Skills = req.Skills

	// Save profile
	if profile.ID == 0 {
		err = s.developerRepo.Create(c.Request.Context(), profile)
	} else {
		err = s.developerRepo.Update(c.Request.Context(), profile)
	}

	if err != nil {
		s.logger.Error("Failed to update developer profile", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// GetDeveloperProposalsHandler handles GET /api/v1/developers/:address/proposals
func (s *Service) GetDeveloperProposalsHandler(c *gin.Context) {
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

	proposals, total, err := s.proposalRepo.List(c.Request.Context(), &ProposalFilter{
		Proposer: address,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		s.logger.Error("Failed to get developer proposals", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get proposals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"proposals": proposals,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// GetDeveloperVotesHandler handles GET /api/v1/developers/:address/votes
func (s *Service) GetDeveloperVotesHandler(c *gin.Context) {
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

	votes, err := s.voteRepo.GetByVoter(c.Request.Context(), address, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get developer votes", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get votes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"votes":  votes,
		"limit":  limit,
		"offset": offset,
	})
}
