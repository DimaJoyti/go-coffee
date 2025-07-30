package commands

import (
	"fmt"
	"time"
)

// formatDuration formats a time.Duration into a human-readable string
// This is a shared utility function used across multiple CLI commands
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
	return fmt.Sprintf("%.1fd", d.Hours()/24)
}

// formatSize formats bytes into a human-readable string
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// getHealthEmoji returns an emoji representation of health status
func getHealthEmoji(health string) string {
	switch health {
	case "healthy":
		return "✅ " + health
	case "degraded":
		return "⚠️ " + health
	case "unhealthy":
		return "❌ " + health
	default:
		return "❓ " + health
	}
}

// getStatusEmoji returns an emoji representation of service status
func getStatusEmoji(status string) string {
	switch status {
	case "running":
		return "🟢 " + status
	case "stopped":
		return "🔴 " + status
	case "starting":
		return "🟡 " + status
	default:
		return "⚪ " + status
	}
}
