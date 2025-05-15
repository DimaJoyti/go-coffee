package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a wrapper around zap.Logger
type Logger struct {
	*zap.Logger
}

// Config represents the logger configuration
type Config struct {
	// Level is the minimum enabled logging level
	Level string `json:"level"`
	// Development puts the logger in development mode
	Development bool `json:"development"`
	// Encoding sets the logger's encoding (json or console)
	Encoding string `json:"encoding"`
}

// NewLogger creates a new logger
func NewLogger(cfg Config) (*Logger, error) {
	// Set default values
	if cfg.Level == "" {
		cfg.Level = "info"
	}
	if cfg.Encoding == "" {
		cfg.Encoding = "json"
	}

	// Parse log level
	level := zap.NewAtomicLevel()
	err := level.UnmarshalText([]byte(cfg.Level))
	if err != nil {
		return nil, err
	}

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create zap config
	zapConfig := zap.Config{
		Level:             level,
		Development:       cfg.Development,
		Encoding:          cfg.Encoding,
		EncoderConfig:     encoderConfig,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		DisableCaller:     false,
		DisableStacktrace: false,
	}

	// Build logger
	logger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{logger}, nil
}

// NewDefaultLogger creates a new logger with default configuration
func NewDefaultLogger() *Logger {
	// Create default config
	cfg := Config{
		Level:       "info",
		Development: false,
		Encoding:    "json",
	}

	// Check environment
	if os.Getenv("ENVIRONMENT") == "development" {
		cfg.Development = true
		cfg.Encoding = "console"
	}

	// Create logger
	logger, err := NewLogger(cfg)
	if err != nil {
		// If we can't create a logger, use a default one
		zapLogger, _ := zap.NewProduction()
		return &Logger{zapLogger}
	}

	return logger
}

// With adds a variadic number of fields to the logging context
func (l *Logger) With(fields ...zapcore.Field) *Logger {
	return &Logger{l.Logger.With(fields...)}
}

// Named adds a sub-logger with the specified name
func (l *Logger) Named(name string) *Logger {
	return &Logger{l.Logger.Named(name)}
}

// Sugar returns a sugared logger
func (l *Logger) Sugar() *zap.SugaredLogger {
	return l.Logger.Sugar()
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// Fields creates a field for structured logging
var Fields = zap.Fields

// String creates a field with the given key and string value
func String(key, value string) zapcore.Field {
	return zap.String(key, value)
}

// Int creates a field with the given key and int value
func Int(key string, value int) zapcore.Field {
	return zap.Int(key, value)
}

// Int64 creates a field with the given key and int64 value
func Int64(key string, value int64) zapcore.Field {
	return zap.Int64(key, value)
}

// Float64 creates a field with the given key and float64 value
func Float64(key string, value float64) zapcore.Field {
	return zap.Float64(key, value)
}

// Bool creates a field with the given key and bool value
func Bool(key string, value bool) zapcore.Field {
	return zap.Bool(key, value)
}

// Error creates a field with the given key and error value
func Error(err error) zapcore.Field {
	return zap.Error(err)
}

// Any creates a field with the given key and arbitrary value
func Any(key string, value interface{}) zapcore.Field {
	return zap.Any(key, value)
}
