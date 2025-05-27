package config

import (
	"time"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/kafka"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/redis"
)

// Region represents a region configuration
type Region struct {
	Name     string
	Priority int
	Redis    redis.Config
	Kafka    kafka.Config
}

// MultiRegionConfig represents a multi-region configuration
type MultiRegionConfig struct {
	Regions       []Region
	CurrentRegion string
	Failover      FailoverConfig
}

// FailoverConfig represents failover configuration
type FailoverConfig struct {
	Enabled           bool
	CheckInterval     time.Duration
	FailureThreshold  int
	RecoveryThreshold int
	Timeout           time.Duration
}

// GetCurrentRegion gets the current region configuration
func (c *MultiRegionConfig) GetCurrentRegion() *Region {
	for _, region := range c.Regions {
		if region.Name == c.CurrentRegion {
			return &region
		}
	}
	return nil
}

// GetRegionByName gets a region configuration by name
func (c *MultiRegionConfig) GetRegionByName(name string) *Region {
	for _, region := range c.Regions {
		if region.Name == name {
			return &region
		}
	}
	return nil
}

// GetPrimaryRegion gets the primary region configuration
func (c *MultiRegionConfig) GetPrimaryRegion() *Region {
	var primary *Region
	for _, region := range c.Regions {
		if primary == nil || region.Priority < primary.Priority {
			primary = &region
		}
	}
	return primary
}

// GetBackupRegions gets backup region configurations
func (c *MultiRegionConfig) GetBackupRegions() []*Region {
	primary := c.GetPrimaryRegion()
	if primary == nil {
		return nil
	}

	backups := make([]*Region, 0, len(c.Regions)-1)
	for _, region := range c.Regions {
		if region.Name != primary.Name {
			backups = append(backups, &region)
		}
	}
	return backups
}
