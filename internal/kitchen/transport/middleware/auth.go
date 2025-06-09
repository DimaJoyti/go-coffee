package middleware

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// AuthInterceptor provides authentication for gRPC calls
type AuthInterceptor struct {
	logger    *logger.Logger
	jwtSecret string
}

// NewAuthInterceptor creates a new authentication interceptor
func NewAuthInterceptor(logger *logger.Logger, jwtSecret string) *AuthInterceptor {
	return &AuthInterceptor{
		logger:    logger,
		jwtSecret: jwtSecret,
	}
}

// UnaryInterceptor provides authentication for unary gRPC calls
func (a *AuthInterceptor) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Skip authentication for health checks and public methods
	if a.isPublicMethod(info.FullMethod) {
		return handler(ctx, req)
	}

	// Extract and validate token
	userInfo, err := a.authenticate(ctx)
	if err != nil {
		a.logger.WithError(err).WithField("method", info.FullMethod).Error("Authentication failed")
		return nil, status.Errorf(codes.Unauthenticated, "authentication failed: %v", err)
	}

	// Add user info to context
	ctx = a.addUserToContext(ctx, userInfo)

	return handler(ctx, req)
}

// StreamInterceptor provides authentication for streaming gRPC calls
func (a *AuthInterceptor) StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// Skip authentication for health checks and public methods
	if a.isPublicMethod(info.FullMethod) {
		return handler(srv, ss)
	}

	// Extract and validate token
	userInfo, err := a.authenticate(ss.Context())
	if err != nil {
		a.logger.WithError(err).WithField("method", info.FullMethod).Error("Authentication failed")
		return status.Errorf(codes.Unauthenticated, "authentication failed: %v", err)
	}

	// Create new context with user info
	ctx := a.addUserToContext(ss.Context(), userInfo)
	wrappedStream := &wrappedServerStream{
		ServerStream: ss,
		ctx:          ctx,
	}

	return handler(srv, wrappedStream)
}

// authenticate extracts and validates the authentication token
func (a *AuthInterceptor) authenticate(ctx context.Context) (*UserInfo, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}

	// Extract authorization header
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return nil, fmt.Errorf("missing authorization header")
	}

	authHeader := authHeaders[0]
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return nil, fmt.Errorf("empty token")
	}

	// Validate token (simplified - in production, use proper JWT validation)
	userInfo, err := a.validateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return userInfo, nil
}

// validateToken validates the JWT token and extracts user information
func (a *AuthInterceptor) validateToken(token string) (*UserInfo, error) {
	// Simplified token validation - in production, use proper JWT library
	// For now, we'll accept any non-empty token and extract basic info
	
	// In a real implementation, you would:
	// 1. Parse the JWT token
	// 2. Verify the signature using the secret
	// 3. Check expiration
	// 4. Extract claims
	
	if token == "invalid" {
		return nil, fmt.Errorf("token is invalid")
	}

	// Mock user info extraction
	userInfo := &UserInfo{
		UserID:   "user_123",
		Username: "kitchen_staff",
		Role:     "staff",
		Permissions: []string{
			"kitchen:read",
			"kitchen:write",
			"orders:read",
			"orders:update",
		},
	}

	// Extract role from token (simplified)
	if strings.Contains(token, "manager") {
		userInfo.Role = "manager"
		userInfo.Permissions = append(userInfo.Permissions, 
			"staff:read", 
			"staff:write", 
			"equipment:read", 
			"equipment:write",
			"analytics:read",
		)
	}

	return userInfo, nil
}

// isPublicMethod checks if a method is public and doesn't require authentication
func (a *AuthInterceptor) isPublicMethod(method string) bool {
	publicMethods := []string{
		"/grpc.health.v1.Health/Check",
		"/grpc.health.v1.Health/Watch",
		"/kitchen.KitchenService/GetQueueStatus", // Public queue status
	}

	for _, publicMethod := range publicMethods {
		if method == publicMethod {
			return true
		}
	}

	return false
}

// addUserToContext adds user information to the context
func (a *AuthInterceptor) addUserToContext(ctx context.Context, userInfo *UserInfo) context.Context {
	return context.WithValue(ctx, userContextKey, userInfo)
}

// GetUserFromContext extracts user information from context
func GetUserFromContext(ctx context.Context) (*UserInfo, bool) {
	userInfo, ok := ctx.Value(userContextKey).(*UserInfo)
	return userInfo, ok
}

// CheckPermission checks if the user has the required permission
func CheckPermission(ctx context.Context, permission string) error {
	userInfo, ok := GetUserFromContext(ctx)
	if !ok {
		return fmt.Errorf("user not found in context")
	}

	for _, perm := range userInfo.Permissions {
		if perm == permission {
			return nil
		}
	}

	return fmt.Errorf("insufficient permissions: required %s", permission)
}

// UserInfo represents authenticated user information
type UserInfo struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

// HasPermission checks if user has a specific permission
func (u *UserInfo) HasPermission(permission string) bool {
	for _, perm := range u.Permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// IsManager checks if user is a manager
func (u *UserInfo) IsManager() bool {
	return u.Role == "manager"
}

// IsStaff checks if user is kitchen staff
func (u *UserInfo) IsStaff() bool {
	return u.Role == "staff"
}

// Context key for user information
type contextKey string

const userContextKey contextKey = "user"

// wrappedServerStream wraps grpc.ServerStream with a custom context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the custom context
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// RoleBasedAccessControl provides role-based access control
type RoleBasedAccessControl struct {
	logger *logger.Logger
}

// NewRoleBasedAccessControl creates a new RBAC instance
func NewRoleBasedAccessControl(logger *logger.Logger) *RoleBasedAccessControl {
	return &RoleBasedAccessControl{
		logger: logger,
	}
}

// CheckAccess checks if user has access to a specific resource and action
func (rbac *RoleBasedAccessControl) CheckAccess(ctx context.Context, resource, action string) error {
	userInfo, ok := GetUserFromContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	permission := fmt.Sprintf("%s:%s", resource, action)
	
	if !userInfo.HasPermission(permission) {
		rbac.logger.WithFields(map[string]interface{}{
			"user_id":    userInfo.UserID,
			"role":       userInfo.Role,
			"resource":   resource,
			"action":     action,
			"permission": permission,
		}).Warn("Access denied")
		
		return status.Errorf(codes.PermissionDenied, "access denied: insufficient permissions for %s", permission)
	}

	rbac.logger.WithFields(map[string]interface{}{
		"user_id":    userInfo.UserID,
		"role":       userInfo.Role,
		"resource":   resource,
		"action":     action,
		"permission": permission,
	}).Debug("Access granted")

	return nil
}

// RequireManagerRole ensures the user is a manager
func (rbac *RoleBasedAccessControl) RequireManagerRole(ctx context.Context) error {
	userInfo, ok := GetUserFromContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	if !userInfo.IsManager() {
		rbac.logger.WithFields(map[string]interface{}{
			"user_id": userInfo.UserID,
			"role":    userInfo.Role,
		}).Warn("Manager role required")
		
		return status.Errorf(codes.PermissionDenied, "manager role required")
	}

	return nil
}

// RequireStaffRole ensures the user is kitchen staff
func (rbac *RoleBasedAccessControl) RequireStaffRole(ctx context.Context) error {
	userInfo, ok := GetUserFromContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	if !userInfo.IsStaff() && !userInfo.IsManager() {
		rbac.logger.WithFields(map[string]interface{}{
			"user_id": userInfo.UserID,
			"role":    userInfo.Role,
		}).Warn("Staff role required")
		
		return status.Errorf(codes.PermissionDenied, "kitchen staff role required")
	}

	return nil
}

// AuditLog logs access attempts for security auditing
type AuditLog struct {
	logger *logger.Logger
}

// NewAuditLog creates a new audit log instance
func NewAuditLog(logger *logger.Logger) *AuditLog {
	return &AuditLog{
		logger: logger,
	}
}

// LogAccess logs an access attempt
func (al *AuditLog) LogAccess(ctx context.Context, method, resource, action string, success bool) {
	userInfo, _ := GetUserFromContext(ctx)
	
	logFields := map[string]interface{}{
		"method":   method,
		"resource": resource,
		"action":   action,
		"success":  success,
	}

	if userInfo != nil {
		logFields["user_id"] = userInfo.UserID
		logFields["role"] = userInfo.Role
	}

	if success {
		al.logger.WithFields(logFields).Info("Access granted")
	} else {
		al.logger.WithFields(logFields).Warn("Access denied")
	}
}
