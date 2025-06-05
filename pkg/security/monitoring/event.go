package monitoring

import (
	"context"
)

// This file contains legacy event handling code
// All types are now defined in service.go to avoid duplication

// LogSecurityEvent is a legacy function - use SecurityMonitoringService.LogSecurityEvent instead
func LogSecurityEvent(ctx context.Context, event *SecurityEvent) {
	// Legacy function - implementation moved to service
}
