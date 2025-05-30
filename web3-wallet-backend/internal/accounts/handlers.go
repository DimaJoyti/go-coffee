package accounts

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
)

// Handler handles HTTP requests for accounts
type Handler struct {
	service   Service
	logger    *logger.Logger
	validator *validator.Validate
}

// NewHandler creates a new accounts handler
func NewHandler(service Service, logger *logger.Logger) *Handler {
	return &Handler{
		service:   service,
		logger:    logger,
		validator: validator.New(),
	}
}

// RegisterRoutes registers account routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	accounts := router.Group("/accounts")
	{
		// Public routes
		accounts.POST("/register", h.CreateAccount)
		accounts.POST("/login", h.Login)
		accounts.POST("/logout", h.Logout)
		accounts.POST("/refresh", h.RefreshToken)
		accounts.POST("/password/reset", h.ResetPassword)
		accounts.POST("/password/confirm", h.ConfirmPasswordReset)

		// Protected routes (require authentication)
		protected := accounts.Group("")
		protected.Use(h.AuthMiddleware())
		{
			protected.GET("/me", h.GetCurrentAccount)
			protected.PUT("/me", h.UpdateCurrentAccount)
			protected.DELETE("/me", h.DeleteCurrentAccount)
			protected.POST("/password/change", h.ChangePassword)

			// Two-factor authentication
			protected.POST("/2fa/enable", h.EnableTwoFactor)
			protected.POST("/2fa/disable", h.DisableTwoFactor)
			protected.POST("/2fa/verify", h.VerifyTwoFactor)

			// KYC
			protected.POST("/kyc/documents", h.SubmitKYCDocument)
			protected.GET("/kyc/documents", h.GetKYCDocuments)

			// Security
			protected.GET("/security/events", h.GetSecurityEvents)
		}

		// Admin routes
		admin := accounts.Group("/admin")
		admin.Use(h.AuthMiddleware(), h.AdminMiddleware())
		{
			admin.GET("", h.ListAccounts)
			admin.GET("/:id", h.GetAccount)
			admin.PUT("/:id", h.UpdateAccount)
			admin.DELETE("/:id", h.DeleteAccount)
			admin.PUT("/:id/kyc/status", h.UpdateKYCStatus)
		}
	}
}

// CreateAccount handles account creation
func (h *Handler) CreateAccount(c *gin.Context) {
	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error("Validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	account, err := h.service.CreateAccount(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create account", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": account})
}

// Login handles user authentication
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error("Validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed"})
		return
	}

	response, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Login failed", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// Logout handles user logout
func (h *Handler) Logout(c *gin.Context) {
	token := h.extractToken(c)
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing authorization token"})
		return
	}

	err := h.service.Logout(c.Request.Context(), token)
	if err != nil {
		h.logger.Error("Logout failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// GetCurrentAccount gets the current authenticated user's account
func (h *Handler) GetCurrentAccount(c *gin.Context) {
	accountID := h.getAccountIDFromContext(c)
	if accountID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	account, err := h.service.GetAccount(c.Request.Context(), accountID)
	if err != nil {
		h.logger.Error("Failed to get account", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": account})
}

// UpdateCurrentAccount updates the current authenticated user's account
func (h *Handler) UpdateCurrentAccount(c *gin.Context) {
	accountID := h.getAccountIDFromContext(c)
	if accountID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	account, err := h.service.UpdateAccount(c.Request.Context(), accountID, &req)
	if err != nil {
		h.logger.Error("Failed to update account", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": account})
}

// DeleteCurrentAccount deletes the current authenticated user's account
func (h *Handler) DeleteCurrentAccount(c *gin.Context) {
	accountID := h.getAccountIDFromContext(c)
	if accountID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	err := h.service.DeleteAccount(c.Request.Context(), accountID)
	if err != nil {
		h.logger.Error("Failed to delete account", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}

// ListAccounts lists accounts (admin only)
func (h *Handler) ListAccounts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	req := &AccountListRequest{
		Page:       page,
		Limit:      limit,
		Status:     AccountStatus(c.Query("status")),
		KYCStatus:  KYCStatus(c.Query("kyc_status")),
		Country:    c.Query("country"),
		SearchTerm: c.Query("search"),
	}

	response, err := h.service.ListAccounts(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to list accounts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list accounts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// GetAccount gets an account by ID (admin only)
func (h *Handler) GetAccount(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Account ID is required"})
		return
	}

	account, err := h.service.GetAccount(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get account", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": account})
}

// UpdateAccount updates an account (admin only)
func (h *Handler) UpdateAccount(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Account ID is required"})
		return
	}

	var req UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	account, err := h.service.UpdateAccount(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.Error("Failed to update account", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": account})
}

// DeleteAccount deletes an account (admin only)
func (h *Handler) DeleteAccount(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Account ID is required"})
		return
	}

	err := h.service.DeleteAccount(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to delete account", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}

// GetSecurityEvents gets security events for the current user
func (h *Handler) GetSecurityEvents(c *gin.Context) {
	accountID := h.getAccountIDFromContext(c)
	if accountID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	events, err := h.service.GetSecurityEvents(c.Request.Context(), accountID)
	if err != nil {
		h.logger.Error("Failed to get security events", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get security events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": events})
}

// Middleware functions

// AuthMiddleware validates authentication
func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := h.extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization token"})
			c.Abort()
			return
		}

		session, err := h.service.ValidateSession(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Store account ID in context
		c.Set("account_id", session.AccountID)
		c.Next()
	}
}

// AdminMiddleware validates admin access
func (h *Handler) AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID := h.getAccountIDFromContext(c)
		if accountID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// TODO: Check if user has admin role
		// For now, allow all authenticated users
		c.Next()
	}
}

// Helper functions

// extractToken extracts the bearer token from the Authorization header
func (h *Handler) extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Remove "Bearer " prefix
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}

	return ""
}

// getAccountIDFromContext gets the account ID from the gin context
func (h *Handler) getAccountIDFromContext(c *gin.Context) string {
	accountID, exists := c.Get("account_id")
	if !exists {
		return ""
	}

	accountIDStr, ok := accountID.(string)
	if !ok {
		return ""
	}

	return accountIDStr
}

// Placeholder handlers for remaining endpoints
func (h *Handler) RefreshToken(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *Handler) ChangePassword(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *Handler) ResetPassword(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *Handler) ConfirmPasswordReset(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *Handler) EnableTwoFactor(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *Handler) DisableTwoFactor(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *Handler) VerifyTwoFactor(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *Handler) SubmitKYCDocument(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *Handler) GetKYCDocuments(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *Handler) UpdateKYCStatus(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}
