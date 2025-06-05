package infrastructure

import (
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

type ThreatDetectorConfig struct {
	SuspiciousIPThreshold int
	BlockDuration         time.Duration
	AllowedUserAgents    []string
	BlockedIPRanges      []string
}

type BasicThreatDetector struct {
	config *ThreatDetectorConfig
	logger *logger.Logger
}

func NewBasicThreatDetector(config *ThreatDetectorConfig, logger *logger.Logger) *BasicThreatDetector {
	return &BasicThreatDetector{
		config: config,
		logger: logger,
	}
}
