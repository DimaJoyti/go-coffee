package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/application/bus"
	"github.com/DimaJoyti/go-coffee/internal/auth/application/commands"
	"github.com/DimaJoyti/go-coffee/internal/auth/application/queries"
	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/gorilla/mux"
)

// CleanHandler implements clean HTTP handlers using CQRS pattern
type CleanHandler struct {
	commandBus *bus.CommandBusImpl
	queryBus   *bus.QueryBusImpl
	logger     *logger.Logger
}

// NewCleanHandler creates a new clean HTTP handler
func NewCleanHandler(
	commandBus *bus.CommandBusImpl,
	queryBus *bus.QueryBusImpl,
	logger *logger.Logger,
) *CleanHandler {
	return &CleanHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
		logger:     logger,
	}
}

// RegisterRoutes registers clean HTTP routes
func (h *CleanHandler) RegisterRoutes(router *mux.Router) {
	// Health check
	router.HandleFunc("/health", h.HealthCheck).Methods("GET")

	// API v2 routes (clean architecture)
	api := router.PathPrefix("/api/v2/auth").Subrouter()

	// Public routes
	api.HandleFunc("/register", h.Register).Methods("POST")
	api.HandleFunc("/login", h.Login).Methods("POST")
	api.HandleFunc("/refresh", h.RefreshToken).Methods("POST")
	api.HandleFunc("/validate", h.ValidateToken).Methods("POST")

	// Protected routes would use middleware here
	// For now, implementing without middleware for simplicity
	api.HandleFunc("/logout", h.Logout).Methods("POST")
	api.HandleFunc("/me", h.GetUserProfile).Methods("GET")
	api.HandleFunc("/users/{id}", h.GetUserByID).Methods("GET")
	api.HandleFunc("/users", h.GetUsers).Methods("GET")
	api.HandleFunc("/change-password", h.ChangePassword).Methods("POST")
	api.HandleFunc("/sessions", h.GetUserSessions).Methods("GET")
	api.HandleFunc("/sessions/{id}", h.GetSessionByID).Methods("GET")
}

// Register handles user registration
func (h *CleanHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if err := h.validateRegisterRequest(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create command
	cmd := commands.RegisterUserCommand{
		Email:       req.Email,
		Password:    req.Password,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		PhoneNumber: req.PhoneNumber,
		Role:        domain.UserRole(req.Role),
		Metadata:    req.Metadata,
	}

	// Execute command
	result, err := h.commandBus.Handle(r.Context(), cmd)
	if err != nil {
		h.logger.ErrorWithFields("Register command failed", 
			logger.Error(err),
			logger.String("email", req.Email))
		h.respondWithError(w, http.StatusInternalServerError, "Registration failed")
		return
	}

	if !result.Success {
		h.respondWithError(w, http.StatusBadRequest, result.Message)
		return
	}

	h.respondWithJSON(w, http.StatusCreated, RegisterResponse{
		Success: true,
		Message: result.Message,
		Data:    result.Data,
	})
}

// Login handles user login
func (h *CleanHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if err := h.validateLoginRequest(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create command
	cmd := commands.LoginUserCommand{
		Email:      req.Email,
		Password:   req.Password,
		RememberMe: req.RememberMe,
		IPAddress:  h.getClientIP(r),
		UserAgent:  r.UserAgent(),
	}

	// Execute command
	result, err := h.commandBus.Handle(r.Context(), cmd)
	if err != nil {
		h.logger.ErrorWithFields("Login command failed", 
			logger.Error(err),
			logger.String("email", req.Email))
		h.respondWithError(w, http.StatusUnauthorized, "Login failed")
		return
	}

	if !result.Success {
		h.respondWithError(w, http.StatusUnauthorized, result.Message)
		return
	}

	h.respondWithJSON(w, http.StatusOK, LoginResponse{
		Success: true,
		Message: result.Message,
		Data:    result.Data,
	})
}

// Logout handles user logout
func (h *CleanHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Extract user ID and session ID from context (would be set by middleware)
	userID := h.getUserIDFromContext(r.Context())
	sessionID := h.getSessionIDFromContext(r.Context())

	if userID == "" || sessionID == "" {
		h.respondWithError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Create command
	cmd := commands.LogoutUserCommand{
		UserID:    userID,
		SessionID: sessionID,
		Reason:    "user_logout",
	}

	// Execute command
	result, err := h.commandBus.Handle(r.Context(), cmd)
	if err != nil {
		h.logger.ErrorWithFields("Logout command failed", 
			logger.Error(err),
			logger.String("user_id", userID))
		h.respondWithError(w, http.StatusInternalServerError, "Logout failed")
		return
	}

	h.respondWithJSON(w, http.StatusOK, LogoutResponse{
		Success: true,
		Message: result.Message,
	})
}

// ChangePassword handles password change
func (h *CleanHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	userID := h.getUserIDFromContext(r.Context())
	if userID == "" {
		h.respondWithError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Create command
	cmd := commands.ChangePasswordCommand{
		UserID:      userID,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
		Forced:      false,
	}

	// Execute command
	result, err := h.commandBus.Handle(r.Context(), cmd)
	if err != nil {
		h.logger.ErrorWithFields("Change password command failed", 
			logger.Error(err),
			logger.String("user_id", userID))
		h.respondWithError(w, http.StatusInternalServerError, "Password change failed")
		return
	}

	if !result.Success {
		h.respondWithError(w, http.StatusBadRequest, result.Message)
		return
	}

	h.respondWithJSON(w, http.StatusOK, ChangePasswordResponse{
		Success: true,
		Message: result.Message,
	})
}

// GetUserProfile handles getting user profile
func (h *CleanHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r.Context())
	if userID == "" {
		h.respondWithError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Create query
	query := queries.GetUserProfileQuery{
		UserID: userID,
	}

	// Execute query
	result, err := h.queryBus.Handle(r.Context(), query)
	if err != nil {
		h.logger.ErrorWithFields("Get user profile query failed", 
			logger.Error(err),
			logger.String("user_id", userID))
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get user profile")
		return
	}

	h.respondWithJSON(w, http.StatusOK, result)
}

// GetUserByID handles getting user by ID
func (h *CleanHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		h.respondWithError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Create query
	query := queries.GetUserByIDQuery{
		UserID: userID,
	}

	// Execute query
	result, err := h.queryBus.Handle(r.Context(), query)
	if err != nil {
		h.logger.ErrorWithFields("Get user by ID query failed", 
			logger.Error(err),
			logger.String("user_id", userID))
		h.respondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, result)
}

// GetUsers handles getting users with pagination
func (h *CleanHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 20
	}

	// Create query
	query := queries.GetUsersQuery{
		Page:     page,
		PageSize: pageSize,
		Search:   r.URL.Query().Get("search"),
		SortBy:   r.URL.Query().Get("sort_by"),
		SortDir:  r.URL.Query().Get("sort_dir"),
	}

	// Execute query
	result, err := h.queryBus.Handle(r.Context(), query)
	if err != nil {
		h.logger.ErrorWithFields("Get users query failed", logger.Error(err))
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get users")
		return
	}

	h.respondWithJSON(w, http.StatusOK, result)
}

// GetUserSessions handles getting user sessions
func (h *CleanHandler) GetUserSessions(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r.Context())
	if userID == "" {
		h.respondWithError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 20
	}

	// Create query
	query := queries.GetUserSessionsQuery{
		UserID:   userID,
		Page:     page,
		PageSize: pageSize,
	}

	// Execute query
	result, err := h.queryBus.Handle(r.Context(), query)
	if err != nil {
		h.logger.ErrorWithFields("Get user sessions query failed", 
			logger.Error(err),
			logger.String("user_id", userID))
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get user sessions")
		return
	}

	h.respondWithJSON(w, http.StatusOK, result)
}

// RefreshToken handles token refresh (placeholder)
func (h *CleanHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	h.respondWithError(w, http.StatusNotImplemented, "Not implemented yet")
}

// ValidateToken handles token validation (placeholder)
func (h *CleanHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	h.respondWithError(w, http.StatusNotImplemented, "Not implemented yet")
}

// GetSessionByID handles getting session by ID (placeholder)
func (h *CleanHandler) GetSessionByID(w http.ResponseWriter, r *http.Request) {
	h.respondWithError(w, http.StatusNotImplemented, "Not implemented yet")
}

// HealthCheck returns service health status
func (h *CleanHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"service":   "auth-service",
		"version":   "v2",
		"timestamp": time.Now(),
	})
}

// Helper methods

func (h *CleanHandler) getUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}

func (h *CleanHandler) getSessionIDFromContext(ctx context.Context) string {
	if sessionID, ok := ctx.Value("session_id").(string); ok {
		return sessionID
	}
	return ""
}

func (h *CleanHandler) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fall back to RemoteAddr
	return r.RemoteAddr
}

func (h *CleanHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]interface{}{
		"success": false,
		"error":   message,
	})
}

func (h *CleanHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
