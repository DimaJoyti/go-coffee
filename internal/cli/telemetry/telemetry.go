package telemetry

import (
"context"
"time"

"github.com/DimaJoyti/go-coffee/internal/cli/config"
"go.uber.org/zap"
)

// Metrics holds all CLI metrics
type Metrics struct {
CommandCount map[string]int64
ErrorCount   map[string]int64
Durations    map[string][]time.Duration
}

var (
logger  *zap.Logger
metrics *Metrics
)

// Init initializes telemetry for the CLI
func Init(ctx context.Context, cfg config.TelemetryConfig) (func(context.Context) error, error) {
if !cfg.Enabled {
return func(context.Context) error { return nil }, nil
}

// Initialize logger
var err error
logger, err = zap.NewProduction()
if err != nil {
return nil, err
}

// Initialize metrics
metrics = &Metrics{
CommandCount: make(map[string]int64),
ErrorCount:   make(map[string]int64),
Durations:    make(map[string][]time.Duration),
}

logger.Info("Telemetry initialized", 
zap.String("service", cfg.ServiceName),
zap.String("endpoint", cfg.Endpoint),
)

// Return shutdown function
return func(ctx context.Context) error {
if logger != nil {
logger.Sync()
}
return nil
}, nil
}

// RecordCommandExecution records metrics for command execution
func RecordCommandExecution(ctx context.Context, command string, duration time.Duration, err error) {
if metrics == nil {
return
}

// Record command execution
metrics.CommandCount[command]++

// Record duration
metrics.Durations[command] = append(metrics.Durations[command], duration)

// Record error if any
if err != nil {
metrics.ErrorCount[command]++
if logger != nil {
logger.Error("Command execution failed",
zap.String("command", command),
zap.Duration("duration", duration),
zap.Error(err),
)
}
} else if logger != nil {
logger.Info("Command executed successfully",
zap.String("command", command),
zap.Duration("duration", duration),
)
}
}

// GetMetrics returns current metrics
func GetMetrics() *Metrics {
return metrics
}

// GetLogger returns the telemetry logger
func GetLogger() *zap.Logger {
return logger
}
