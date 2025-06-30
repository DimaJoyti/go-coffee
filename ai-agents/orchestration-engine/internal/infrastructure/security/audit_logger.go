package security

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// AuditLogger provides comprehensive audit logging functionality
type AuditLoggerImpl struct {
	storage    AuditStorage
	config     *AuditConfig
	logger     Logger
	buffer     []*AuditEntry
	bufferSize int
	mutex      sync.RWMutex
	stopCh     chan struct{}
}

// AuditConfig contains audit logging configuration
type AuditConfig struct {
	Enabled             bool          `json:"enabled"`
	BufferSize          int           `json:"buffer_size"`
	FlushInterval       time.Duration `json:"flush_interval"`
	RetentionPeriod     time.Duration `json:"retention_period"`
	IncludeSensitiveData bool         `json:"include_sensitive_data"`
	LogLevels           []string      `json:"log_levels"`
	Categories          []string      `json:"categories"`
	MaxEntrySize        int           `json:"max_entry_size"`
	CompressEntries     bool          `json:"compress_entries"`
}

// AuditEntry represents a single audit log entry
type AuditEntry struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"`
	Category    string                 `json:"category"`
	Action      string                 `json:"action"`
	UserID      string                 `json:"user_id"`
	Username    string                 `json:"username"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	Resource    string                 `json:"resource"`
	ResourceID  string                 `json:"resource_id"`
	Success     bool                   `json:"success"`
	ErrorCode   string                 `json:"error_code"`
	ErrorMessage string                `json:"error_message"`
	Details     map[string]interface{} `json:"details"`
	Metadata    map[string]string      `json:"metadata"`
	RequestID   string                 `json:"request_id"`
	SessionID   string                 `json:"session_id"`
	Duration    time.Duration          `json:"duration"`
	Size        int64                  `json:"size"`
}

// AuditStorage interface for audit log storage
type AuditStorage interface {
	Store(ctx context.Context, entries []*AuditEntry) error
	Query(ctx context.Context, filter *AuditFilter) ([]*AuditEntry, error)
	Delete(ctx context.Context, filter *AuditFilter) error
	GetStats(ctx context.Context) (*AuditStats, error)
}

// AuditFilter represents audit log filtering options
type AuditFilter struct {
	StartTime   *time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	UserID      string     `json:"user_id"`
	Username    string     `json:"username"`
	IPAddress   string     `json:"ip_address"`
	Category    string     `json:"category"`
	Action      string     `json:"action"`
	Resource    string     `json:"resource"`
	Success     *bool      `json:"success"`
	Level       string     `json:"level"`
	Limit       int        `json:"limit"`
	Offset      int        `json:"offset"`
}

// AuditStats represents audit log statistics
type AuditStats struct {
	TotalEntries    int64     `json:"total_entries"`
	EntriesPerLevel map[string]int64 `json:"entries_per_level"`
	EntriesPerCategory map[string]int64 `json:"entries_per_category"`
	SuccessRate     float64   `json:"success_rate"`
	LastEntry       time.Time `json:"last_entry"`
	OldestEntry     time.Time `json:"oldest_entry"`
	StorageSize     int64     `json:"storage_size"`
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(storage AuditStorage, config *AuditConfig, logger Logger) *AuditLoggerImpl {
	if config == nil {
		config = DefaultAuditConfig()
	}

	auditLogger := &AuditLoggerImpl{
		storage:    storage,
		config:     config,
		logger:     logger,
		buffer:     make([]*AuditEntry, 0, config.BufferSize),
		bufferSize: config.BufferSize,
		stopCh:     make(chan struct{}),
	}

	// Start background flush routine
	if config.Enabled {
		go auditLogger.flushRoutine()
	}

	return auditLogger
}

// DefaultAuditConfig returns default audit configuration
func DefaultAuditConfig() *AuditConfig {
	return &AuditConfig{
		Enabled:              true,
		BufferSize:           1000,
		FlushInterval:        30 * time.Second,
		RetentionPeriod:      90 * 24 * time.Hour, // 90 days
		IncludeSensitiveData: false,
		LogLevels:           []string{"info", "warn", "error", "critical"},
		Categories:          []string{"auth", "workflow", "agent", "system", "security"},
		MaxEntrySize:        10 * 1024, // 10KB
		CompressEntries:     true,
	}
}

// LogAuthEvent logs an authentication event
func (al *AuditLoggerImpl) LogAuthEvent(ctx context.Context, event *AuthEvent) error {
	if !al.config.Enabled {
		return nil
	}

	entry := &AuditEntry{
		ID:        generateAuditID(),
		Timestamp: time.Now(),
		Level:     al.getEventLevel(event.Type),
		Category:  "auth",
		Action:    event.Type,
		UserID:    event.UserID,
		Username:  event.Username,
		IPAddress: event.IPAddress,
		Success:   al.isAuthEventSuccess(event.Type),
		Details:   event.Details,
		RequestID: getRequestID(ctx),
		SessionID: event.SessionID,
		Metadata:  make(map[string]string),
	}

	if event.Reason != "" {
		entry.ErrorMessage = event.Reason
	}

	return al.logEntry(entry)
}

// LogSecurityEvent logs a security event
func (al *AuditLoggerImpl) LogSecurityEvent(ctx context.Context, event *SecurityEvent) error {
	if !al.config.Enabled {
		return nil
	}

	entry := &AuditEntry{
		ID:        generateAuditID(),
		Timestamp: time.Now(),
		Level:     "warn",
		Category:  "security",
		Action:    event.Type,
		UserID:    event.UserID,
		IPAddress: event.IPAddress,
		Success:   false,
		Details:   event.Details,
		RequestID: getRequestID(ctx),
		Metadata:  make(map[string]string),
	}

	return al.logEntry(entry)
}

// LogWorkflowEvent logs a workflow event
func (al *AuditLoggerImpl) LogWorkflowEvent(ctx context.Context, workflowID, action string, success bool, details map[string]interface{}) error {
	if !al.config.Enabled {
		return nil
	}

	securityCtx, _ := GetSecurityContext(ctx)
	
	entry := &AuditEntry{
		ID:         generateAuditID(),
		Timestamp:  time.Now(),
		Level:      al.getActionLevel(action, success),
		Category:   "workflow",
		Action:     action,
		Resource:   "workflow",
		ResourceID: workflowID,
		Success:    success,
		Details:    details,
		RequestID:  getRequestID(ctx),
		Metadata:   make(map[string]string),
	}

	if securityCtx != nil {
		entry.UserID = securityCtx.UserID
		entry.Username = securityCtx.Username
		entry.IPAddress = securityCtx.IPAddress
		entry.UserAgent = securityCtx.UserAgent
	}

	return al.logEntry(entry)
}

// LogAgentEvent logs an agent event
func (al *AuditLoggerImpl) LogAgentEvent(ctx context.Context, agentType, action string, success bool, duration time.Duration, details map[string]interface{}) error {
	if !al.config.Enabled {
		return nil
	}

	securityCtx, _ := GetSecurityContext(ctx)
	
	entry := &AuditEntry{
		ID:         generateAuditID(),
		Timestamp:  time.Now(),
		Level:      al.getActionLevel(action, success),
		Category:   "agent",
		Action:     action,
		Resource:   "agent",
		ResourceID: agentType,
		Success:    success,
		Duration:   duration,
		Details:    details,
		RequestID:  getRequestID(ctx),
		Metadata:   make(map[string]string),
	}

	if securityCtx != nil {
		entry.UserID = securityCtx.UserID
		entry.Username = securityCtx.Username
		entry.IPAddress = securityCtx.IPAddress
		entry.UserAgent = securityCtx.UserAgent
	}

	return al.logEntry(entry)
}

// LogSystemEvent logs a system event
func (al *AuditLoggerImpl) LogSystemEvent(ctx context.Context, action string, success bool, details map[string]interface{}) error {
	if !al.config.Enabled {
		return nil
	}

	entry := &AuditEntry{
		ID:        generateAuditID(),
		Timestamp: time.Now(),
		Level:     al.getActionLevel(action, success),
		Category:  "system",
		Action:    action,
		Resource:  "system",
		Success:   success,
		Details:   details,
		RequestID: getRequestID(ctx),
		Metadata:  make(map[string]string),
	}

	return al.logEntry(entry)
}

// LogDataAccess logs a data access event
func (al *AuditLoggerImpl) LogDataAccess(ctx context.Context, resource, resourceID, action string, success bool, size int64) error {
	if !al.config.Enabled {
		return nil
	}

	securityCtx, _ := GetSecurityContext(ctx)
	
	entry := &AuditEntry{
		ID:         generateAuditID(),
		Timestamp:  time.Now(),
		Level:      "info",
		Category:   "data_access",
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Success:    success,
		Size:       size,
		RequestID:  getRequestID(ctx),
		Metadata:   make(map[string]string),
	}

	if securityCtx != nil {
		entry.UserID = securityCtx.UserID
		entry.Username = securityCtx.Username
		entry.IPAddress = securityCtx.IPAddress
		entry.UserAgent = securityCtx.UserAgent
	}

	return al.logEntry(entry)
}

// logEntry adds an entry to the audit log
func (al *AuditLoggerImpl) logEntry(entry *AuditEntry) error {
	// Check if category is enabled
	if !al.isCategoryEnabled(entry.Category) {
		return nil
	}

	// Check if level is enabled
	if !al.isLevelEnabled(entry.Level) {
		return nil
	}

	// Sanitize sensitive data if needed
	if !al.config.IncludeSensitiveData {
		al.sanitizeEntry(entry)
	}

	// Check entry size
	if al.getEntrySize(entry) > al.config.MaxEntrySize {
		al.logger.Warn("Audit entry too large, truncating", "entry_id", entry.ID, "size", al.getEntrySize(entry))
		al.truncateEntry(entry)
	}

	al.mutex.Lock()
	defer al.mutex.Unlock()

	// Add to buffer
	al.buffer = append(al.buffer, entry)

	// Flush if buffer is full
	if len(al.buffer) >= al.bufferSize {
		go al.flush()
	}

	return nil
}

// flush writes buffered entries to storage
func (al *AuditLoggerImpl) flush() {
	al.mutex.Lock()
	if len(al.buffer) == 0 {
		al.mutex.Unlock()
		return
	}

	entries := make([]*AuditEntry, len(al.buffer))
	copy(entries, al.buffer)
	al.buffer = al.buffer[:0]
	al.mutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := al.storage.Store(ctx, entries); err != nil {
		al.logger.Error("Failed to store audit entries", err, "count", len(entries))
		
		// Re-add entries to buffer on failure
		al.mutex.Lock()
		al.buffer = append(entries, al.buffer...)
		al.mutex.Unlock()
	} else {
		al.logger.Debug("Stored audit entries", "count", len(entries))
	}
}

// flushRoutine periodically flushes the buffer
func (al *AuditLoggerImpl) flushRoutine() {
	ticker := time.NewTicker(al.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			al.flush()
		case <-al.stopCh:
			al.flush() // Final flush
			return
		}
	}
}

// Query queries audit logs
func (al *AuditLoggerImpl) Query(ctx context.Context, filter *AuditFilter) ([]*AuditEntry, error) {
	if !al.config.Enabled {
		return nil, fmt.Errorf("audit logging is disabled")
	}

	return al.storage.Query(ctx, filter)
}

// GetStats returns audit log statistics
func (al *AuditLoggerImpl) GetStats(ctx context.Context) (*AuditStats, error) {
	if !al.config.Enabled {
		return nil, fmt.Errorf("audit logging is disabled")
	}

	return al.storage.GetStats(ctx)
}

// Cleanup removes old audit entries
func (al *AuditLoggerImpl) Cleanup(ctx context.Context) error {
	if !al.config.Enabled {
		return nil
	}

	cutoff := time.Now().Add(-al.config.RetentionPeriod)
	filter := &AuditFilter{
		EndTime: &cutoff,
	}

	return al.storage.Delete(ctx, filter)
}

// Stop stops the audit logger
func (al *AuditLoggerImpl) Stop() {
	close(al.stopCh)
}

// Helper methods

func (al *AuditLoggerImpl) getEventLevel(eventType string) string {
	switch eventType {
	case "login_failed", "account_locked", "password_changed":
		return "warn"
	case "login_success", "logout":
		return "info"
	case "user_created", "user_deleted":
		return "info"
	default:
		return "info"
	}
}

func (al *AuditLoggerImpl) getActionLevel(action string, success bool) string {
	if !success {
		return "error"
	}

	switch action {
	case "create", "update", "delete":
		return "info"
	case "execute", "start", "stop":
		return "info"
	default:
		return "info"
	}
}

func (al *AuditLoggerImpl) isAuthEventSuccess(eventType string) bool {
	return eventType == "login_success" || eventType == "logout" || eventType == "user_created"
}

func (al *AuditLoggerImpl) isCategoryEnabled(category string) bool {
	if len(al.config.Categories) == 0 {
		return true
	}

	for _, enabled := range al.config.Categories {
		if enabled == category {
			return true
		}
	}

	return false
}

func (al *AuditLoggerImpl) isLevelEnabled(level string) bool {
	if len(al.config.LogLevels) == 0 {
		return true
	}

	for _, enabled := range al.config.LogLevels {
		if enabled == level {
			return true
		}
	}

	return false
}

func (al *AuditLoggerImpl) sanitizeEntry(entry *AuditEntry) {
	// Remove sensitive fields
	if entry.Details != nil {
		delete(entry.Details, "password")
		delete(entry.Details, "token")
		delete(entry.Details, "secret")
		delete(entry.Details, "key")
		delete(entry.Details, "authorization")
	}

	// Mask IP address partially
	if entry.IPAddress != "" {
		parts := strings.Split(entry.IPAddress, ".")
		if len(parts) == 4 {
			entry.IPAddress = fmt.Sprintf("%s.%s.xxx.xxx", parts[0], parts[1])
		}
	}
}

func (al *AuditLoggerImpl) getEntrySize(entry *AuditEntry) int {
	data, _ := json.Marshal(entry)
	return len(data)
}

func (al *AuditLoggerImpl) truncateEntry(entry *AuditEntry) {
	// Truncate details if too large
	if entry.Details != nil {
		detailsData, _ := json.Marshal(entry.Details)
		if len(detailsData) > al.config.MaxEntrySize/2 {
			entry.Details = map[string]interface{}{
				"truncated": true,
				"original_size": len(detailsData),
			}
		}
	}
}

func generateAuditID() string {
	return fmt.Sprintf("audit_%d", time.Now().UnixNano())
}

func getRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}
