package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Level представляє рівень логування
type Level int

const (
	// DebugLevel рівень для детального логування
	DebugLevel Level = iota
	// InfoLevel рівень для інформаційних повідомлень
	InfoLevel
	// WarnLevel рівень для попереджень
	WarnLevel
	// ErrorLevel рівень для помилок
	ErrorLevel
	// FatalLevel рівень для критичних помилок
	FatalLevel
)

// Logger інтерфейс для логування
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
}

// Config конфігурація для логера
type Config struct {
	Level      Level
	TimeFormat string
	Output     *os.File
}

// DefaultConfig повертає типову конфігурацію
func DefaultConfig() *Config {
	return &Config{
		Level:      InfoLevel,
		TimeFormat: time.RFC3339,
		Output:     os.Stdout,
	}
}

// standardLogger реалізує інтерфейс Logger
type standardLogger struct {
	logger     *log.Logger
	level      Level
	timeFormat string
	fields     map[string]interface{}
}

// NewLogger створює новий логер
func NewLogger(config *Config) Logger {
	if config == nil {
		config = DefaultConfig()
	}

	logger := log.New(config.Output, "", 0)
	return &standardLogger{
		logger:     logger,
		level:      config.Level,
		timeFormat: config.TimeFormat,
		fields:     make(map[string]interface{}),
	}
}

// formatFields форматує поля для виведення
func (l *standardLogger) formatFields() string {
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

// log виконує логування
func (l *standardLogger) log(level Level, format string, args ...interface{}) {
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

// Debug логує повідомлення з рівнем Debug
func (l *standardLogger) Debug(format string, args ...interface{}) {
	l.log(DebugLevel, format, args...)
}

// Info логує повідомлення з рівнем Info
func (l *standardLogger) Info(format string, args ...interface{}) {
	l.log(InfoLevel, format, args...)
}

// Warn логує повідомлення з рівнем Warn
func (l *standardLogger) Warn(format string, args ...interface{}) {
	l.log(WarnLevel, format, args...)
}

// Error логує повідомлення з рівнем Error
func (l *standardLogger) Error(format string, args ...interface{}) {
	l.log(ErrorLevel, format, args...)
}

// Fatal логує повідомлення з рівнем Fatal
func (l *standardLogger) Fatal(format string, args ...interface{}) {
	l.log(FatalLevel, format, args...)
}

// WithField повертає новий логер з доданим полем
func (l *standardLogger) WithField(key string, value interface{}) Logger {
	newLogger := &standardLogger{
		logger:     l.logger,
		level:      l.level,
		timeFormat: l.timeFormat,
		fields:     make(map[string]interface{}),
	}

	// Копіюємо існуючі поля
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Додаємо нове поле
	newLogger.fields[key] = value

	return newLogger
}

// WithFields повертає новий логер з доданими полями
func (l *standardLogger) WithFields(fields map[string]interface{}) Logger {
	newLogger := &standardLogger{
		logger:     l.logger,
		level:      l.level,
		timeFormat: l.timeFormat,
		fields:     make(map[string]interface{}),
	}

	// Копіюємо існуючі поля
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Додаємо нові поля
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}
