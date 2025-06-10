package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// Level represents logging level with enhanced type safety
type Level int

const (
	// DebugLevel for detailed debugging information
	DebugLevel Level = iota
	// InfoLevel for general informational messages
	InfoLevel
	// WarnLevel for warning conditions
	WarnLevel
	// ErrorLevel for error conditions
	ErrorLevel
	// FatalLevel for critical errors that cause program termination
	FatalLevel
)

// String returns the string representation of the log level
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Color returns ANSI color code for the log level
func (l Level) Color() string {
	switch l {
	case DebugLevel:
		return "\033[36m" // Cyan
	case InfoLevel:
		return "\033[32m" // Green
	case WarnLevel:
		return "\033[33m" // Yellow
	case ErrorLevel:
		return "\033[31m" // Red
	case FatalLevel:
		return "\033[35m" // Magenta
	default:
		return "\033[0m" // Reset
	}
}

// Logger provides structured, colorized logging with enhanced features
type Logger struct {
	logger     *log.Logger
	level      Level
	timeFormat string
	fields     map[string]interface{}
	colorized  bool
	jsonFormat bool
	service    string
}

// Config provides comprehensive logger configuration
type Config struct {
	Level      Level                  `json:"level" yaml:"level"`
	TimeFormat string                 `json:"time_format" yaml:"time_format"`
	Output     *os.File               `json:"-" yaml:"-"`
	Colorized  bool                   `json:"colorized" yaml:"colorized"`
	JSONFormat bool                   `json:"json_format" yaml:"json_format"`
	Service    string                 `json:"service" yaml:"service"`
	Fields     map[string]interface{} `json:"fields" yaml:"fields"`
}

// DefaultConfig returns optimized default configuration
func DefaultConfig() *Config {
	return &Config{
		Level:      InfoLevel,
		TimeFormat: time.RFC3339,
		Output:     os.Stdout,
		Colorized:  true,
		JSONFormat: false,
		Service:    "go-coffee",
		Fields:     make(map[string]interface{}),
	}
}

// ProductionConfig returns production-optimized configuration
func ProductionConfig() *Config {
	return &Config{
		Level:      InfoLevel,
		TimeFormat: time.RFC3339,
		Output:     os.Stdout,
		Colorized:  false,
		JSONFormat: true,
		Service:    "go-coffee",
		Fields:     make(map[string]interface{}),
	}
}

// DevelopmentConfig returns development-optimized configuration
func DevelopmentConfig() *Config {
	return &Config{
		Level:      DebugLevel,
		TimeFormat: "15:04:05",
		Output:     os.Stdout,
		Colorized:  true,
		JSONFormat: false,
		Service:    "go-coffee-dev",
		Fields:     make(map[string]interface{}),
	}
}

// NewLogger creates an enhanced logger with comprehensive configuration
func NewLogger(config *Config) *Logger {
	if config == nil {
		config = DefaultConfig()
	}

	logger := log.New(config.Output, "", 0)

	// Copy fields from config
	fields := make(map[string]interface{})
	for k, v := range config.Fields {
		fields[k] = v
	}

	// Add service field if provided
	if config.Service != "" {
		fields["service"] = config.Service
	}

	return &Logger{
		logger:     logger,
		level:      config.Level,
		timeFormat: config.TimeFormat,
		fields:     fields,
		colorized:  config.Colorized,
		jsonFormat: config.JSONFormat,
		service:    config.Service,
	}
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Service   string                 `json:"service,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Caller    string                 `json:"caller,omitempty"`
}

// formatFields formats fields for console output
func (l *Logger) formatFields() string {
	if len(l.fields) == 0 {
		return ""
	}

	var parts []string
	for k, v := range l.fields {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	return strings.Join(parts, " ")
}

// formatJSON formats log entry as JSON
func (l *Logger) formatJSON(level Level, message string) string {
	entry := LogEntry{
		Timestamp: time.Now().Format(l.timeFormat),
		Level:     level.String(),
		Message:   message,
		Service:   l.service,
		Fields:    l.fields,
	}

	// Add caller information in debug mode
	if level == DebugLevel {
		if _, file, line, ok := runtime.Caller(3); ok {
			entry.Caller = fmt.Sprintf("%s:%d", file, line)
		}
	}

	data, _ := json.Marshal(entry)
	return string(data)
}

// formatConsole formats log entry for console output
func (l *Logger) formatConsole(level Level, message string) string {
	timestamp := time.Now().Format(l.timeFormat)
	levelStr := level.String()

	if l.colorized {
		levelStr = fmt.Sprintf("%s%s\033[0m", level.Color(), levelStr)
	}

	fields := l.formatFields()
	if fields != "" {
		return fmt.Sprintf("%s [%s] %s | %s", timestamp, levelStr, message, fields)
	}
	return fmt.Sprintf("%s [%s] %s", timestamp, levelStr, message)
}

// log performs enhanced logging with JSON/console formatting
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	message := fmt.Sprintf(format, args...)

	var output string
	if l.jsonFormat {
		output = l.formatJSON(level, message)
	} else {
		output = l.formatConsole(level, message)
	}

	l.logger.Println(output)

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

// WithError returns a new logger with an error field
func (l *Logger) WithError(err error) *Logger {
	return l.WithField("error", err.Error())
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

// Duration creates a duration field
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value.String()}
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

// InfoWithFields logs an info message with fields
func (l *Logger) InfoWithFields(msg string, fields ...Field) {
	logger := l.With(fields...)
	logger.Info(msg)
}

// ErrorWithFields logs an error message with fields
func (l *Logger) ErrorWithFields(msg string, fields ...Field) {
	logger := l.With(fields...)
	logger.Error(msg)
}

// WarnWithFields logs a warn message with fields
func (l *Logger) WarnWithFields(msg string, fields ...Field) {
	logger := l.With(fields...)
	logger.Warn(msg)
}

// Named creates a new logger with a name field (for zap compatibility)
func (l *Logger) Named(name string) *Logger {
	return l.WithField("component", name)
}
