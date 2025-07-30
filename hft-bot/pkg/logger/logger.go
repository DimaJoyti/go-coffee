package logger

import (
	"fmt"
	"log"
	"time"
)

// SimpleLogger implements a basic logger interface without external dependencies
type SimpleLogger struct {
	prefix string
}

// Config holds simple logger configuration
type SimpleConfig struct {
	Level       string `yaml:"level" json:"level"`
	Format      string `yaml:"format" json:"format"`
	ServiceName string `yaml:"service_name" json:"service_name"`
	Environment string `yaml:"environment" json:"environment"`
	Version     string `yaml:"version" json:"version"`
}

// NewSimple creates a new simple logger instance
func NewSimple(cfg SimpleConfig) (*SimpleLogger, error) {
	prefix := fmt.Sprintf("[%s]", cfg.ServiceName)
	if cfg.Environment != "" {
		prefix = fmt.Sprintf("[%s-%s]", cfg.ServiceName, cfg.Environment)
	}
	
	return &SimpleLogger{
		prefix: prefix,
	}, nil
}

// NewSimpleWithPrefix creates a new simple logger with a custom prefix
func NewSimpleWithPrefix(prefix string) *SimpleLogger {
	return &SimpleLogger{
		prefix: prefix,
	}
}

// Info logs an info message with key-value pairs
func (l *SimpleLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logWithLevel("INFO", msg, keysAndValues...)
}

// Error logs an error message with key-value pairs
func (l *SimpleLogger) Error(msg string, keysAndValues ...interface{}) {
	l.logWithLevel("ERROR", msg, keysAndValues...)
}

// Warn logs a warning message with key-value pairs
func (l *SimpleLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.logWithLevel("WARN", msg, keysAndValues...)
}

// Debug logs a debug message with key-value pairs
func (l *SimpleLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.logWithLevel("DEBUG", msg, keysAndValues...)
}

// logWithLevel logs a message with the specified level and key-value pairs
func (l *SimpleLogger) logWithLevel(level, msg string, keysAndValues ...interface{}) {
	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z07:00")
	logMsg := fmt.Sprintf("%s [%s] %s: %s", timestamp, level, l.prefix, msg)
	
	// Add key-value pairs
	if len(keysAndValues) > 0 {
		logMsg += " |"
		for i := 0; i < len(keysAndValues); i += 2 {
			if i+1 < len(keysAndValues) {
				logMsg += fmt.Sprintf(" %v=%v", keysAndValues[i], keysAndValues[i+1])
			}
		}
	}
	
	log.Println(logMsg)
}

// Sync flushes any buffered log entries (no-op for simple logger)
func (l *SimpleLogger) Sync() error {
	return nil
}

// WithPrefix creates a new logger with an additional prefix
func (l *SimpleLogger) WithPrefix(prefix string) *SimpleLogger {
	newPrefix := fmt.Sprintf("%s[%s]", l.prefix, prefix)
	return &SimpleLogger{
		prefix: newPrefix,
	}
}

// SetLevel sets the log level (no-op for simple logger - logs everything)
func (l *SimpleLogger) SetLevel(level string) {
	// Simple logger logs everything - level filtering can be added if needed
}

// GetLevel returns the current log level
func (l *SimpleLogger) GetLevel() string {
	return "info" // Simple logger default
}

// IsDebugEnabled returns whether debug logging is enabled
func (l *SimpleLogger) IsDebugEnabled() bool {
	return true // Simple logger logs everything
}

// IsInfoEnabled returns whether info logging is enabled
func (l *SimpleLogger) IsInfoEnabled() bool {
	return true
}

// IsWarnEnabled returns whether warn logging is enabled
func (l *SimpleLogger) IsWarnEnabled() bool {
	return true
}

// IsErrorEnabled returns whether error logging is enabled
func (l *SimpleLogger) IsErrorEnabled() bool {
	return true
}

// LoggerInterface defines the interface that our simple logger implements
type LoggerInterface interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
	Sync() error
}

// Ensure SimpleLogger implements LoggerInterface
var _ LoggerInterface = (*SimpleLogger)(nil)

// Helper functions for creating structured fields (for compatibility)

// String creates a string field
func String(key, value string) interface{} {
	return []interface{}{key, value}
}

// Int creates an int field
func Int(key string, value int) interface{} {
	return []interface{}{key, value}
}

// Int64 creates an int64 field
func Int64(key string, value int64) interface{} {
	return []interface{}{key, value}
}

// Float64 creates a float64 field
func Float64(key string, value float64) interface{} {
	return []interface{}{key, value}
}

// Bool creates a bool field
func Bool(key string, value bool) interface{} {
	return []interface{}{key, value}
}

// Duration creates a duration field
func Duration(key string, value time.Duration) interface{} {
	return []interface{}{key, value}
}

// Time creates a time field
func Time(key string, value time.Time) interface{} {
	return []interface{}{key, value.Format(time.RFC3339)}
}

// Any creates a field for any type
func Any(key string, value interface{}) interface{} {
	return []interface{}{key, value}
}

// Error creates an error field
func ErrorField(err error) interface{} {
	if err == nil {
		return []interface{}{"error", nil}
	}
	return []interface{}{"error", err.Error()}
}

// FlattenFields flattens field helper results into a single slice
func FlattenFields(fields ...interface{}) []interface{} {
	var result []interface{}
	for _, field := range fields {
		if slice, ok := field.([]interface{}); ok {
			result = append(result, slice...)
		} else {
			result = append(result, field)
		}
	}
	return result
}

// Example usage:
// logger.Info("User logged in", 
//     String("user_id", "123"), 
//     Int("attempts", 3),
//     Duration("response_time", time.Millisecond*150))
//
// Or with flattening:
// fields := []interface{}{
//     String("user_id", "123"),
//     Int("attempts", 3),
// }
// logger.Info("User logged in", FlattenFields(fields...)...)
