package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/application"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/gorilla/mux"
)

// Handler represents HTTP handlers for auth service
type Handler struct {
	authService application.AuthService
	mfaService  application.MFAService
	logger      *logger.Logger
}

// NewHandler creates a new HTTP handler
func NewHandler(
	authService application.AuthService,
	mfaService application.MFAService,
	logger *logger.Logger,
) *Handler {
	return &Handler{
		authService: authService,
		mfaService:  mfaService,
		logger:      logger,
	}
}

// RegisterRoutes registers HTTP routes
func (h *Handler) RegisterRoutes(router *mux.Router) {
	// Health check
	router.HandleFunc("/health", h.HealthCheck).Methods("GET")

	// API routes
	api := router.PathPrefix("/api/v1/auth").Subrouter()

	// Public routes (no authentication required)
	api.HandleFunc("/register", h.Register).Methods("POST")
	api.HandleFunc("/login", h.Login).Methods("POST")
	api.HandleFunc("/refresh", h.RefreshToken).Methods("POST")
	api.HandleFunc("/validate", h.ValidateToken).Methods("POST")
	api.HandleFunc("/forgot-password", h.ForgotPassword).Methods("POST")
	api.HandleFunc("/reset-password", h.ResetPassword).Methods("POST")

	// Protected routes (authentication required)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(h.authMiddleware)

	protected.HandleFunc("/logout", h.Logout).Methods("POST")
	protected.HandleFunc("/me", h.GetUserInfo).Methods("GET")
	protected.HandleFunc("/change-password", h.ChangePassword).Methods("POST")
	protected.HandleFunc("/sessions", h.GetUserSessions).Methods("GET")
	protected.HandleFunc("/sessions/{sessionId}", h.RevokeSession).Methods("DELETE")

	// MFA routes
	mfa := protected.PathPrefix("/mfa").Subrouter()
	mfa.HandleFunc("/enable", h.EnableMFA).Methods("POST")
	mfa.HandleFunc("/disable", h.DisableMFA).Methods("POST")
	mfa.HandleFunc("/verify", h.VerifyMFA).Methods("POST")
	mfa.HandleFunc("/backup-codes", h.GenerateBackupCodes).Methods("POST")
	mfa.HandleFunc("/backup-codes", h.GetBackupCodes).Methods("GET")

	// Security routes
	security := protected.PathPrefix("/security").Subrouter()
	security.HandleFunc("/events", h.GetSecurityEvents).Methods("GET")
	security.HandleFunc("/devices", h.GetTrustedDevices).Methods("GET")
	security.HandleFunc("/devices/{deviceId}/trust", h.TrustDevice).Methods("POST")
	security.HandleFunc("/devices/{deviceId}", h.RemoveDevice).Methods("DELETE")
}

// Authentication Handlers

// Register handles user registration
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req application.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Registration failed")
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusCreated, resp)
}

// Login handles user login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req application.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Extract client information
	req.IPAddress = h.getClientIP(r)
	req.UserAgent = r.UserAgent()

	resp, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).WithFields(map[string]interface{}{
			"email":      req.Email,
			"ip_address": req.IPAddress,
		}).Error("Login failed")
		h.respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, resp)
}

// Logout handles user logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r.Context())
	if userID == "" {
		h.respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req application.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If no body provided, logout from current session only
		req.LogoutAll = false
	}

	resp, err := h.authService.Logout(r.Context(), userID, &req)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Logout failed")
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, resp)
}

// RefreshToken handles token refresh
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req application.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.authService.RefreshToken(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Token refresh failed")
		h.respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, resp)
}

// ValidateToken handles token validation
func (h *Handler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	var req application.ValidateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.authService.ValidateToken(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Token validation failed")
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, resp)
}

// GetUserInfo handles getting user information
func (h *Handler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r.Context())
	if userID == "" {
		h.respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	req := &application.GetUserInfoRequest{UserID: userID}
	resp, err := h.authService.GetUserInfo(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Get user info failed")
		h.respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, resp)
}

// ChangePassword handles password change
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r.Context())
	if userID == "" {
		h.respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req application.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.authService.ChangePassword(r.Context(), userID, &req)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Password change failed")
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, resp)
}

// ForgotPassword handles forgot password requests
func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req application.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.authService.ForgotPassword(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Forgot password failed")
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, resp)
}

// ResetPassword handles password reset
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req application.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.authService.ResetPassword(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Password reset failed")
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, resp)
}

// Session Management Handlers

// GetUserSessions gets all user sessions
func (h *Handler) GetUserSessions(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r.Context())
	if userID == "" {
		h.respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	req := &application.GetUserSessionsRequest{UserID: userID}
	resp, err := h.authService.GetUserSessions(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Get user sessions failed")
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, resp)
}

// RevokeSession revokes a specific session
func (h *Handler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r.Context())
	if userID == "" {
		h.respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	req := &application.RevokeSessionRequest{
		UserID:    userID,
		SessionID: sessionID,
	}

	resp, err := h.authService.RevokeSession(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).WithFields(map[string]interface{}{
			"user_id":    userID,
			"session_id": sessionID,
		}).Error("Revoke session failed")
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, resp)
}

// MFA Handlers

// EnableMFA enables multi-factor authentication
func (h *Handler) EnableMFA(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r.Context())
	if userID == "" {
		h.respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req application.EnableMFARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.UserID = userID
	resp, err := h.mfaService.EnableMFA(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Enable MFA failed")
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, resp)
}

// DisableMFA disables multi-factor authentication
func (h *Handler) DisableMFA(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r.Context())
	if userID == "" {
		h.respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req application.DisableMFARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.UserID = userID
	resp, err := h.mfaService.DisableMFA(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Disable MFA failed")
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, resp)
}

// VerifyMFA verifies MFA code
func (h *Handler) VerifyMFA(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r.Context())
	if userID == "" {
		h.respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req application.VerifyMFARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.UserID = userID
	resp, err := h.mfaService.VerifyMFA(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Verify MFA failed")
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, resp)
}

// GenerateBackupCodes generates new MFA backup codes
func (h *Handler) GenerateBackupCodes(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r.Context())
	if userID == "" {
		h.respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	req := &application.GenerateBackupCodesRequest{UserID: userID}
	resp, err := h.mfaService.GenerateBackupCodes(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Generate backup codes failed")
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, resp)
}

// GetBackupCodes gets remaining MFA backup codes
func (h *Handler) GetBackupCodes(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r.Context())
	if userID == "" {
		h.respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	req := &application.GetBackupCodesRequest{UserID: userID}
	resp, err := h.mfaService.GetBackupCodes(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Get backup codes failed")
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, resp)
}

// Security Handlers

// GetSecurityEvents gets user security events
func (h *Handler) GetSecurityEvents(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r.Context())
	if userID == "" {
		h.respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse query parameters
	limit := 50 // default
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	req := &application.GetSecurityEventsRequest{
		UserID: userID,
		Limit:  limit,
	}

	resp, err := h.authService.GetSecurityEvents(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Get security events failed")
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, resp)
}

// GetTrustedDevices gets user's trusted devices
func (h *Handler) GetTrustedDevices(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r.Context())
	if userID == "" {
		h.respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	req := &application.GetTrustedDevicesRequest{UserID: userID}
	resp, err := h.authService.GetTrustedDevices(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Get trusted devices failed")
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, resp)
}

// TrustDevice marks a device as trusted
func (h *Handler) TrustDevice(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r.Context())
	if userID == "" {
		h.respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	deviceID := vars["deviceId"]

	req := &application.TrustDeviceRequest{
		UserID:   userID,
		DeviceID: deviceID,
	}

	resp, err := h.authService.TrustDevice(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).WithFields(map[string]interface{}{
			"user_id":   userID,
			"device_id": deviceID,
		}).Error("Trust device failed")
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, resp)
}

// RemoveDevice removes a trusted device
func (h *Handler) RemoveDevice(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r.Context())
	if userID == "" {
		h.respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	deviceID := vars["deviceId"]

	req := &application.RemoveDeviceRequest{
		UserID:   userID,
		DeviceID: deviceID,
	}

	resp, err := h.authService.RemoveDevice(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).WithFields(map[string]interface{}{
			"user_id":   userID,
			"device_id": deviceID,
		}).Error("Remove device failed")
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, resp)
}

// HealthCheck returns service health status
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"service":   "auth-service",
		"timestamp": time.Now(),
	})
}

// Helper methods

func (h *Handler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

func (h *Handler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (h *Handler) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

func (h *Handler) getUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}
