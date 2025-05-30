package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Level represents logging level
type Level int

const (
	// DebugLevel level for detailed logging
	DebugLevel Level = iota
	// InfoLevel level for informational messages
	InfoLevel
	// WarnLevel level for warnings
	WarnLevel
	// ErrorLevel level for errors
	ErrorLevel
	// FatalLevel level for critical errors
	FatalLevel
)

// Logger struct for logging - compatible with zap interface
type Logger struct {
	logger     *log.Logger
	level      Level
	timeFormat string
	fields     map[string]interface{}
}

// Config configuration for logger
type Config struct {
	Level      Level
	TimeFormat string
	Output     *os.File
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Level:      InfoLevel,
		TimeFormat: time.RFC3339,
		Output:     os.Stdout,
	}
}

// NewLogger creates a new logger
func NewLogger(config *Config) *Logger {
	if config == nil {
		config = DefaultConfig()
	}

	logger := log.New(config.Output, "", 0)
	return &Logger{
		logger:     logger,
		level:      config.Level,
		timeFormat: config.TimeFormat,
		fields:     make(map[string]interface{}),
	}
}

// formatFields formats fields for output
func (l *Logger) formatFields() string {
	if len(l.fields) == 0 {
		return ""
	}

	result := "{"
	first := true
	for k, v := range l.fields {
		if !first {
			result += ", "
		}
		result += fmt.Sprintf("%s: %v", k, v)
		first = false
	}
	result += "}"
	return result
}

// log performs logging
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	levelStr := ""
	switch level {
	case DebugLevel:
		levelStr = "DEBUG"
	case InfoLevel:
		levelStr = "INFO"
	case WarnLevel:
		levelStr = "WARN"
	case ErrorLevel:
		levelStr = "ERROR"
	case FatalLevel:
		levelStr = "FATAL"
	}

	timestamp := time.Now().Format(l.timeFormat)
	fields := l.formatFields()
	message := fmt.Sprintf(format, args...)

	if fields != "" {
		l.logger.Printf("%s [%s] %s %s\n", timestamp, levelStr, message, fields)
	} else {
		l.logger.Printf("%s [%s] %s\n", timestamp, levelStr, message)
	}

	if level == FatalLevel {
		os.Exit(1)
	}
}

// Debug logs a message with Debug level
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DebugLevel, format, args...)
}

// Info logs a message with Info level
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(InfoLevel, format, args...)
}

// Warn logs a message with Warn level
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WarnLevel, format, args...)
}

// Error logs a message with Error level
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ErrorLevel, format, args...)
}

// Fatal logs a message with Fatal level
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FatalLevel, format, args...)
}

// WithField returns a new logger with added field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	newLogger := &Logger{
		logger:     l.logger,
		level:      l.level,
		timeFormat: l.timeFormat,
		fields:     make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new field
	newLogger.fields[key] = value

	return newLogger
}

// WithFields returns a new logger with added fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newLogger := &Logger{
		logger:     l.logger,
		level:      l.level,
		timeFormat: l.timeFormat,
		fields:     make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new fields
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// New creates a new logger with service name (for compatibility)
func New(serviceName string) *Logger {
	config := DefaultConfig()
	logger := NewLogger(config)
	return logger.WithField("service", serviceName)
}

// Zap-compatible methods for AI processor compatibility

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}

// String creates a string field
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an int field
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Float64 creates a float64 field
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool creates a bool field
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Any creates a field with any value
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Error creates an error field
func Error(err error) Field {
	return Field{Key: "error", Value: err.Error()}
}

// With creates a new logger with the given fields (zap-compatible)
func (l *Logger) With(fields ...Field) *Logger {
	newFields := make(map[string]interface{})

	// Copy existing fields
	for k, v := range l.fields {
		newFields[k] = v
	}

	// Add new fields
	for _, field := range fields {
		newFields[field.Key] = field.Value
	}

	return &Logger{
		logger:     l.logger,
		level:      l.level,
		timeFormat: l.timeFormat,
		fields:     newFields,
	}
}

// Sugar returns a sugared logger (for zap compatibility)
func (l *Logger) Sugar() *Logger {
	return l
}

// Sync flushes any buffered log entries (for zap compatibility)
func (l *Logger) Sync() error {
	// Standard logger doesn't need syncing
	return nil
}
