package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// APIResponse represents a standard API response
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// APIError represents an API error
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Meta represents response metadata
type Meta struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Offset  int `json:"offset"`
}

// SortParams represents sorting parameters
type SortParams struct {
	Field string `json:"field"`
	Order string `json:"order"` // asc, desc
}

// FilterParams represents filtering parameters
type FilterParams struct {
	Search string            `json:"search,omitempty"`
	Fields map[string]string `json:"fields,omitempty"`
}

// WriteJSONResponse writes a JSON response
func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	response := APIResponse{
		Success:   statusCode < 400,
		Data:      data,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Fallback error response
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// WriteErrorResponse writes an error response
func WriteErrorResponse(w http.ResponseWriter, statusCode int, code, message string) {
	response := APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Fallback error response
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}

// WriteErrorResponseWithDetails writes an error response with details
func WriteErrorResponseWithDetails(w http.ResponseWriter, statusCode int, code, message, details string) {
	response := APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Fallback error response
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}

// WritePaginatedResponse writes a paginated response
func WritePaginatedResponse(w http.ResponseWriter, statusCode int, data interface{}, pagination PaginationParams, total int) {
	totalPages := (total + pagination.PerPage - 1) / pagination.PerPage
	
	response := APIResponse{
		Success: statusCode < 400,
		Data:    data,
		Meta: &Meta{
			Page:       pagination.Page,
			PerPage:    pagination.PerPage,
			Total:      total,
			TotalPages: totalPages,
		},
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Fallback error response
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// ParsePaginationParams parses pagination parameters from request
func ParsePaginationParams(r *http.Request) PaginationParams {
	page := parseIntParam(r, "page", 1)
	perPage := parseIntParam(r, "per_page", 20)
	
	// Validate and limit parameters
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	
	offset := (page - 1) * perPage
	
	return PaginationParams{
		Page:    page,
		PerPage: perPage,
		Offset:  offset,
	}
}

// ParseSortParams parses sorting parameters from request
func ParseSortParams(r *http.Request, defaultField string) SortParams {
	field := r.URL.Query().Get("sort_field")
	if field == "" {
		field = defaultField
	}
	
	order := strings.ToLower(r.URL.Query().Get("sort_order"))
	if order != "asc" && order != "desc" {
		order = "desc"
	}
	
	return SortParams{
		Field: field,
		Order: order,
	}
}

// ParseFilterParams parses filtering parameters from request
func ParseFilterParams(r *http.Request) FilterParams {
	search := r.URL.Query().Get("search")
	
	fields := make(map[string]string)
	for key, values := range r.URL.Query() {
		if strings.HasPrefix(key, "filter_") && len(values) > 0 {
			fieldName := strings.TrimPrefix(key, "filter_")
			fields[fieldName] = values[0]
		}
	}
	
	return FilterParams{
		Search: search,
		Fields: fields,
	}
}

// parseIntParam parses an integer parameter from request with default value
func parseIntParam(r *http.Request, param string, defaultValue int) int {
	value := r.URL.Query().Get(param)
	if value == "" {
		return defaultValue
	}
	
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	
	return parsed
}

// ParseUUIDParam parses a UUID parameter from URL path
func ParseUUIDParam(r *http.Request, param string) (uuid.UUID, error) {
	value := r.PathValue(param)
	if value == "" {
		return uuid.Nil, fmt.Errorf("missing %s parameter", param)
	}
	
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid %s format: %w", param, err)
	}
	
	return id, nil
}

// ParseStringParam parses a string parameter from URL path
func ParseStringParam(r *http.Request, param string) (string, error) {
	value := r.PathValue(param)
	if value == "" {
		return "", fmt.Errorf("missing %s parameter", param)
	}
	
	return value, nil
}

// ValidateContentType validates the request content type
func ValidateContentType(r *http.Request, expectedType string) error {
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		return fmt.Errorf("missing Content-Type header")
	}
	
	// Handle content type with charset
	if strings.Contains(contentType, ";") {
		contentType = strings.Split(contentType, ";")[0]
	}
	
	if strings.TrimSpace(contentType) != expectedType {
		return fmt.Errorf("invalid Content-Type: expected %s, got %s", expectedType, contentType)
	}
	
	return nil
}

// DecodeJSONBody decodes JSON request body
func DecodeJSONBody(r *http.Request, dest interface{}) error {
	if err := ValidateContentType(r, "application/json"); err != nil {
		return err
	}
	
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	
	if err := decoder.Decode(dest); err != nil {
		return fmt.Errorf("failed to decode JSON body: %w", err)
	}
	
	return nil
}

// GetRequestID extracts request ID from context or headers
func GetRequestID(r *http.Request) string {
	// Try to get from context first
	if requestID := r.Context().Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	
	// Fallback to header
	return r.Header.Get("X-Request-ID")
}

// GetUserID extracts user ID from context
func GetUserID(r *http.Request) string {
	if userID := r.Context().Value("user_id"); userID != nil {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	
	return ""
}

// BuildResourceURL builds a URL for a resource
func BuildResourceURL(r *http.Request, resourceType string, id interface{}) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	
	return fmt.Sprintf("%s://%s/api/v1/%s/%v", scheme, r.Host, resourceType, id)
}

// ValidateRequiredFields validates that required fields are present
func ValidateRequiredFields(data map[string]interface{}, requiredFields []string) error {
	var missingFields []string
	
	for _, field := range requiredFields {
		if value, exists := data[field]; !exists || value == nil || value == "" {
			missingFields = append(missingFields, field)
		}
	}
	
	if len(missingFields) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missingFields, ", "))
	}
	
	return nil
}

// SanitizeString sanitizes a string input
func SanitizeString(input string) string {
	// Basic sanitization - remove control characters
	var result strings.Builder
	for _, r := range input {
		if r >= 32 && r != 127 { // Printable ASCII characters
			result.WriteRune(r)
		}
	}
	return strings.TrimSpace(result.String())
}

// LimitString limits string length
func LimitString(input string, maxLength int) string {
	if len(input) <= maxLength {
		return input
	}
	return input[:maxLength]
}

// HTTPError represents an HTTP error with status code
type HTTPError struct {
	StatusCode int
	Code       string
	Message    string
	Details    string
}

// Error implements the error interface
func (e HTTPError) Error() string {
	return e.Message
}

// NewHTTPError creates a new HTTP error
func NewHTTPError(statusCode int, code, message string) HTTPError {
	return HTTPError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}

// NewHTTPErrorWithDetails creates a new HTTP error with details
func NewHTTPErrorWithDetails(statusCode int, code, message, details string) HTTPError {
	return HTTPError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Details:    details,
	}
}

// Common HTTP errors
var (
	ErrBadRequest          = NewHTTPError(http.StatusBadRequest, "bad_request", "Bad request")
	ErrUnauthorized        = NewHTTPError(http.StatusUnauthorized, "unauthorized", "Unauthorized")
	ErrForbidden           = NewHTTPError(http.StatusForbidden, "forbidden", "Forbidden")
	ErrNotFound            = NewHTTPError(http.StatusNotFound, "not_found", "Resource not found")
	ErrMethodNotAllowed    = NewHTTPError(http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
	ErrConflict            = NewHTTPError(http.StatusConflict, "conflict", "Resource conflict")
	ErrUnprocessableEntity = NewHTTPError(http.StatusUnprocessableEntity, "unprocessable_entity", "Unprocessable entity")
	ErrInternalServer      = NewHTTPError(http.StatusInternalServerError, "internal_server_error", "Internal server error")
	ErrServiceUnavailable  = NewHTTPError(http.StatusServiceUnavailable, "service_unavailable", "Service unavailable")
)

// WriteHTTPError writes an HTTP error response
func WriteHTTPError(w http.ResponseWriter, err HTTPError) {
	if err.Details != "" {
		WriteErrorResponseWithDetails(w, err.StatusCode, err.Code, err.Message, err.Details)
	} else {
		WriteErrorResponse(w, err.StatusCode, err.Code, err.Message)
	}
}

// HandleError handles different types of errors and writes appropriate response
func HandleError(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case HTTPError:
		WriteHTTPError(w, e)
	default:
		WriteErrorResponse(w, http.StatusInternalServerError, "internal_server_error", "Internal server error")
	}
}
