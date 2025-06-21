package recording

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"go.uber.org/zap"
)

// StorageManager manages recording storage and cleanup
type StorageManager struct {
	logger *zap.Logger
	config ServiceConfig
}

// NewStorageManager creates a new storage manager
func NewStorageManager(logger *zap.Logger, config ServiceConfig) *StorageManager {
	return &StorageManager{
		logger: logger.With(zap.String("component", "storage_manager")),
		config: config,
	}
}

// GetStorageUsage returns storage usage statistics for a stream
func (sm *StorageManager) GetStorageUsage(streamID string) (*domain.StorageUsage, error) {
	streamPath := filepath.Join(sm.config.StorageBasePath, streamID)
	
	// Get disk usage
	totalSize, usedSize, err := sm.getDiskUsage(sm.config.StorageBasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk usage: %w", err)
	}

	// Get stream-specific usage
	streamUsage, recordingCount, oldestRecording, newestRecording, err := sm.getStreamUsage(streamPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get stream usage: %w", err)
	}

	usage := &domain.StorageUsage{
		StreamID:        streamID,
		TotalSize:       totalSize,
		UsedSize:        usedSize,
		AvailableSize:   totalSize - usedSize,
		RecordingCount:  recordingCount,
		OldestRecording: oldestRecording,
		NewestRecording: newestRecording,
	}

	if recordingCount > 0 {
		usage.AverageFileSize = streamUsage / recordingCount
	}

	// Calculate compression ratio (simplified)
	usage.CompressionRatio = 0.7 // Assume 30% compression

	return usage, nil
}

// OptimizeStorage optimizes storage usage by cleaning up and compressing files
func (sm *StorageManager) OptimizeStorage() error {
	sm.logger.Info("Starting storage optimization")

	// Check if storage is getting full
	totalSize, usedSize, err := sm.getDiskUsage(sm.config.StorageBasePath)
	if err != nil {
		return fmt.Errorf("failed to get disk usage: %w", err)
	}

	usagePercentage := float64(usedSize) / float64(totalSize)
	
	if usagePercentage > sm.config.StorageWarningThreshold {
		sm.logger.Warn("Storage usage is high",
			zap.Float64("usage_percentage", usagePercentage*100),
			zap.Float64("threshold", sm.config.StorageWarningThreshold*100))

		// Perform aggressive cleanup
		if err := sm.performAggressiveCleanup(); err != nil {
			return fmt.Errorf("failed to perform aggressive cleanup: %w", err)
		}
	}

	// Compress old files
	if sm.config.EnableCompression {
		if err := sm.compressOldFiles(); err != nil {
			sm.logger.Error("Failed to compress old files", zap.Error(err))
		}
	}

	// Remove empty directories
	if err := sm.removeEmptyDirectories(); err != nil {
		sm.logger.Error("Failed to remove empty directories", zap.Error(err))
	}

	sm.logger.Info("Storage optimization completed")
	return nil
}

// getDiskUsage returns total and used disk space
func (sm *StorageManager) getDiskUsage(path string) (total, used int64, err error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, 0, err
	}

	// Available blocks * size per block = available space in bytes
	total = int64(stat.Blocks) * int64(stat.Bsize)
	available := int64(stat.Bavail) * int64(stat.Bsize)
	used = total - available

	return total, used, nil
}

// getStreamUsage calculates storage usage for a specific stream
func (sm *StorageManager) getStreamUsage(streamPath string) (totalSize, recordingCount int64, oldestRecording, newestRecording *time.Time, err error) {
	if _, err := os.Stat(streamPath); os.IsNotExist(err) {
		return 0, 0, nil, nil, nil
	}

	err = filepath.Walk(streamPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && sm.isVideoFile(path) {
			totalSize += info.Size()
			recordingCount++

			modTime := info.ModTime()
			if oldestRecording == nil || modTime.Before(*oldestRecording) {
				oldestRecording = &modTime
			}
			if newestRecording == nil || modTime.After(*newestRecording) {
				newestRecording = &modTime
			}
		}

		return nil
	})

	return totalSize, recordingCount, oldestRecording, newestRecording, err
}

// isVideoFile checks if a file is a video file based on extension
func (sm *StorageManager) isVideoFile(filename string) bool {
	ext := filepath.Ext(filename)
	videoExtensions := []string{".mp4", ".avi", ".mov", ".mkv", ".webm"}
	
	for _, videoExt := range videoExtensions {
		if ext == videoExt {
			return true
		}
	}
	
	return false
}

// performAggressiveCleanup removes old files to free up space
func (sm *StorageManager) performAggressiveCleanup() error {
	sm.logger.Info("Performing aggressive cleanup")

	// Find files older than retention policy
	cutoffTime := time.Now().Add(-sm.config.DefaultRetentionPolicy.MaxAge)
	
	var filesToDelete []string
	var totalSizeToDelete int64

	err := filepath.Walk(sm.config.StorageBasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && sm.isVideoFile(path) && info.ModTime().Before(cutoffTime) {
			filesToDelete = append(filesToDelete, path)
			totalSizeToDelete += info.Size()
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk storage directory: %w", err)
	}

	// Delete files
	var deletedCount int
	var deletedSize int64

	for _, filePath := range filesToDelete {
		if info, err := os.Stat(filePath); err == nil {
			if err := os.Remove(filePath); err != nil {
				sm.logger.Error("Failed to delete file", 
					zap.String("file", filePath), 
					zap.Error(err))
			} else {
				deletedCount++
				deletedSize += info.Size()
			}
		}
	}

	sm.logger.Info("Aggressive cleanup completed",
		zap.Int("files_deleted", deletedCount),
		zap.Int64("size_freed_mb", deletedSize/(1024*1024)))

	return nil
}

// compressOldFiles compresses files older than a certain age
func (sm *StorageManager) compressOldFiles() error {
	sm.logger.Info("Compressing old files")

	compressAfter := time.Now().Add(-24 * time.Hour) // Compress files older than 24 hours
	var compressedCount int

	err := filepath.Walk(sm.config.StorageBasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && sm.isVideoFile(path) && info.ModTime().Before(compressAfter) {
			// Check if already compressed (simplified check)
			if filepath.Ext(path) == ".mp4" && !sm.isCompressed(path) {
				if err := sm.compressFile(path); err != nil {
					sm.logger.Error("Failed to compress file", 
						zap.String("file", path), 
						zap.Error(err))
				} else {
					compressedCount++
				}
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk storage directory: %w", err)
	}

	if compressedCount > 0 {
		sm.logger.Info("File compression completed", zap.Int("files_compressed", compressedCount))
	}

	return nil
}

// isCompressed checks if a file is already compressed (simplified check)
func (sm *StorageManager) isCompressed(filePath string) bool {
	// This is a simplified check - in practice, you'd check video metadata
	// or use a naming convention to track compressed files
	return false
}

// compressFile compresses a video file
func (sm *StorageManager) compressFile(filePath string) error {
	// This would use FFmpeg to compress the video file
	// For now, just log the operation
	sm.logger.Info("Compressing file (not implemented)", zap.String("file", filePath))
	
	// In production, this would be:
	// ffmpeg -i input.mp4 -c:v libx264 -crf 23 -c:a aac -b:a 128k output_compressed.mp4
	// Then replace the original file with the compressed version
	
	return nil
}

// removeEmptyDirectories removes empty directories in the storage path
func (sm *StorageManager) removeEmptyDirectories() error {
	var removedCount int

	// Walk the directory tree in reverse order (deepest first)
	err := filepath.Walk(sm.config.StorageBasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && path != sm.config.StorageBasePath {
			// Check if directory is empty
			if sm.isDirectoryEmpty(path) {
				if err := os.Remove(path); err != nil {
					sm.logger.Error("Failed to remove empty directory", 
						zap.String("directory", path), 
						zap.Error(err))
				} else {
					removedCount++
					sm.logger.Debug("Removed empty directory", zap.String("directory", path))
				}
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk storage directory: %w", err)
	}

	if removedCount > 0 {
		sm.logger.Info("Empty directory cleanup completed", zap.Int("directories_removed", removedCount))
	}

	return nil
}

// isDirectoryEmpty checks if a directory is empty
func (sm *StorageManager) isDirectoryEmpty(dirPath string) bool {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false
	}
	return len(entries) == 0
}

// GetStorageHealth returns storage health metrics
func (sm *StorageManager) GetStorageHealth() (*StorageHealth, error) {
	totalSize, usedSize, err := sm.getDiskUsage(sm.config.StorageBasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk usage: %w", err)
	}

	usagePercentage := float64(usedSize) / float64(totalSize)
	
	health := &StorageHealth{
		TotalSize:       totalSize,
		UsedSize:        usedSize,
		AvailableSize:   totalSize - usedSize,
		UsagePercentage: usagePercentage,
		Status:          sm.getHealthStatus(usagePercentage),
		LastCheck:       time.Now(),
	}

	return health, nil
}

// getHealthStatus determines storage health status based on usage
func (sm *StorageManager) getHealthStatus(usagePercentage float64) StorageHealthStatus {
	switch {
	case usagePercentage >= 0.95:
		return StorageHealthCritical
	case usagePercentage >= sm.config.StorageWarningThreshold:
		return StorageHealthWarning
	case usagePercentage >= 0.5:
		return StorageHealthGood
	default:
		return StorageHealthExcellent
	}
}

// Supporting types for storage management

// StorageHealth represents storage health metrics
type StorageHealth struct {
	TotalSize       int64                `json:"total_size"`
	UsedSize        int64                `json:"used_size"`
	AvailableSize   int64                `json:"available_size"`
	UsagePercentage float64              `json:"usage_percentage"`
	Status          StorageHealthStatus  `json:"status"`
	LastCheck       time.Time            `json:"last_check"`
}

// StorageHealthStatus represents storage health status
type StorageHealthStatus string

const (
	StorageHealthExcellent StorageHealthStatus = "excellent"
	StorageHealthGood      StorageHealthStatus = "good"
	StorageHealthWarning   StorageHealthStatus = "warning"
	StorageHealthCritical  StorageHealthStatus = "critical"
)
