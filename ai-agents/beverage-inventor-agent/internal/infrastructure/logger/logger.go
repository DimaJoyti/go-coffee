package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// Logger defines the logging interface
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, err error, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	With(fields ...interface{}) Logger
}

// StructuredLogger implements structured logging
type StructuredLogger struct {
	level      LogLevel
	format     LogFormat
	output     io.Writer
	fields     map[string]interface{}
	serviceName string
}

// LogLevel represents the logging level
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// LogFormat represents the logging format
type LogFormat int

const (
	JSONFormat LogFormat = iota
	TextFormat
)

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	Service     string                 `json:"service"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Caller      string                 `json:"caller,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
	SpanID      string                 `json:"span_id,omitempty"`
}

// New creates a new structured logger
func New() Logger {
	return &StructuredLogger{
		level:       InfoLevel,
		format:      JSONFormat,
		output:      os.Stdout,
		fields:      make(map[string]interface{}),
		serviceName: "beverage-inventor-agent",
	}
}

// NewWithConfig creates a new logger with configuration
func NewWithConfig(level string, format string, output io.Writer) Logger {
	logger := &StructuredLogger{
		level:       parseLogLevel(level),
		format:      parseLogFormat(format),
		output:      output,
		fields:      make(map[string]interface{}),
		serviceName: "beverage-inventor-agent",
	}
	return logger
}

// Info logs an info message
func (l *StructuredLogger) Info(msg string, fields ...interface{}) {
	if l.level <= InfoLevel {
		l.log(InfoLevel, msg, nil, fields...)
	}
}

// Error logs an error message
func (l *StructuredLogger) Error(msg string, err error, fields ...interface{}) {
	if l.level <= ErrorLevel {
		l.log(ErrorLevel, msg, err, fields...)
	}
}

// Debug logs a debug message
func (l *StructuredLogger) Debug(msg string, fields ...interface{}) {
	if l.level <= DebugLevel {
		l.log(DebugLevel, msg, nil, fields...)
	}
}

// Warn logs a warning message
func (l *StructuredLogger) Warn(msg string, fields ...interface{}) {
	if l.level <= WarnLevel {
		l.log(WarnLevel, msg, nil, fields...)
	}
}

// With creates a new logger with additional fields
func (l *StructuredLogger) With(fields ...interface{}) Logger {
	newFields := make(map[string]interface{})
	
	// Copy existing fields
	for k, v := range l.fields {
		newFields[k] = v
	}
	
	// Add new fields
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key := fmt.Sprintf("%v", fields[i])
			value := fields[i+1]
			newFields[key] = value
		}
	}
	
	return &StructuredLogger{
		level:       l.level,
		format:      l.format,
		output:      l.output,
		fields:      newFields,
		serviceName: l.serviceName,
	}
}

// log performs the actual logging
func (l *StructuredLogger) log(level LogLevel, msg string, err error, fields ...interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     levelToString(level),
		Message:   msg,
		Service:   l.serviceName,
		Fields:    make(map[string]interface{}),
		Caller:    getCaller(),
	}

	// Add persistent fields
	for k, v := range l.fields {
		entry.Fields[k] = v
	}

	// Add provided fields
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key := fmt.Sprintf("%v", fields[i])
			value := fields[i+1]
			entry.Fields[key] = value
		}
	}

	// Add error if present
	if err != nil {
		entry.Error = err.Error()
	}

	// Format and write the log entry
	var output string
	switch l.format {
	case JSONFormat:
		output = l.formatJSON(entry)
	case TextFormat:
		output = l.formatText(entry)
	default:
		output = l.formatJSON(entry)
	}

	fmt.Fprintln(l.output, output)
}

// formatJSON formats the log entry as JSON
func (l *StructuredLogger) formatJSON(entry LogEntry) string {
	data, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple format if JSON marshaling fails
		return fmt.Sprintf(`{"timestamp":"%s","level":"%s","message":"%s","error":"failed to marshal log entry: %v"}`,
			entry.Timestamp.Format(time.RFC3339), entry.Level, entry.Message, err)
	}
	return string(data)
}

// formatText formats the log entry as human-readable text
func (l *StructuredLogger) formatText(entry LogEntry) string {
	var parts []string
	
	// Timestamp and level
	parts = append(parts, fmt.Sprintf("[%s] %s", 
		entry.Timestamp.Format("2006-01-02 15:04:05"), 
		strings.ToUpper(entry.Level)))
	
	// Message
	parts = append(parts, entry.Message)
	
	// Error
	if entry.Error != "" {
		parts = append(parts, fmt.Sprintf("error=%s", entry.Error))
	}
	
	// Fields
	for k, v := range entry.Fields {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	
	// Caller
	if entry.Caller != "" {
		parts = append(parts, fmt.Sprintf("caller=%s", entry.Caller))
	}
	
	return strings.Join(parts, " ")
}

// getCaller returns the caller information
func getCaller() string {
	_, file, line, ok := runtime.Caller(3) // Skip log, log method, and public method
	if !ok {
		return ""
	}
	
	// Get just the filename, not the full path
	parts := strings.Split(file, "/")
	if len(parts) > 0 {
		file = parts[len(parts)-1]
	}
	
	return fmt.Sprintf("%s:%d", file, line)
}

// parseLogLevel parses a string log level
func parseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

// parseLogFormat parses a string log format
func parseLogFormat(format string) LogFormat {
	switch strings.ToLower(format) {
	case "json":
		return JSONFormat
	case "text":
		return TextFormat
	default:
		return JSONFormat
	}
}

// levelToString converts a log level to string
func levelToString(level LogLevel) string {
	switch level {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	default:
		return "info"
	}
}

// StandardLogger provides a simple wrapper around the standard log package
type StandardLogger struct {
	logger *log.Logger
}

// NewStandardLogger creates a logger that wraps the standard log package
func NewStandardLogger() Logger {
	return &StandardLogger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// Info logs an info message using standard logger
func (s *StandardLogger) Info(msg string, fields ...interface{}) {
	s.logger.Printf("[INFO] %s %s", msg, s.formatFields(fields...))
}

// Error logs an error message using standard logger
func (s *StandardLogger) Error(msg string, err error, fields ...interface{}) {
	errMsg := ""
	if err != nil {
		errMsg = fmt.Sprintf(" error=%v", err)
	}
	s.logger.Printf("[ERROR] %s%s %s", msg, errMsg, s.formatFields(fields...))
}

// Debug logs a debug message using standard logger
func (s *StandardLogger) Debug(msg string, fields ...interface{}) {
	s.logger.Printf("[DEBUG] %s %s", msg, s.formatFields(fields...))
}

// Warn logs a warning message using standard logger
func (s *StandardLogger) Warn(msg string, fields ...interface{}) {
	s.logger.Printf("[WARN] %s %s", msg, s.formatFields(fields...))
}

// With creates a new logger with additional fields (no-op for standard logger)
func (s *StandardLogger) With(fields ...interface{}) Logger {
	return s
}

// formatFields formats key-value pairs for standard logger
func (s *StandardLogger) formatFields(fields ...interface{}) string {
	if len(fields) == 0 {
		return ""
	}
	
	var parts []string
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key := fmt.Sprintf("%v", fields[i])
			value := fields[i+1]
			parts = append(parts, fmt.Sprintf("%s=%v", key, value))
		}
	}
	
	return strings.Join(parts, " ")
}
