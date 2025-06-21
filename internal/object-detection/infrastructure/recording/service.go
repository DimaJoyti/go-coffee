package recording

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service implements the recording service
type Service struct {
	logger     *zap.Logger
	repository domain.RecordingRepository
	config     ServiceConfig
	
	// Active recordings
	activeRecordings map[string]*ActiveRecording
	recorders        map[string]domain.VideoRecorder
	
	// Storage management
	storageManager *StorageManager
	
	mutex sync.RWMutex
}

// ServiceConfig configures the recording service
type ServiceConfig struct {
	StorageBasePath      string                              `yaml:"storage_base_path"`
	MaxConcurrentRecordings int                            `yaml:"max_concurrent_recordings"`
	DefaultRetentionPolicy  domain.RetentionPolicy         `yaml:"default_retention_policy"`
	QualitySettings         map[domain.RecordingQuality]QualitySettings `yaml:"quality_settings"`
	EnableCompression       bool                            `yaml:"enable_compression"`
	CompressionLevel        int                             `yaml:"compression_level"`
	ThumbnailGeneration     bool                            `yaml:"thumbnail_generation"`
	CleanupInterval         time.Duration                   `yaml:"cleanup_interval"`
	ArchiveInterval         time.Duration                   `yaml:"archive_interval"`
	MaxStorageSize          int64                           `yaml:"max_storage_size"`
	StorageWarningThreshold float64                         `yaml:"storage_warning_threshold"`
}

// QualitySettings defines video quality parameters
type QualitySettings struct {
	Width       int     `yaml:"width"`
	Height      int     `yaml:"height"`
	FrameRate   float64 `yaml:"frame_rate"`
	Bitrate     int     `yaml:"bitrate"`
	Quality     int     `yaml:"quality"`
	Compression string  `yaml:"compression"`
}

// ActiveRecording tracks an active recording session
type ActiveRecording struct {
	Recording *domain.Recording
	Recorder  domain.VideoRecorder
	StartTime time.Time
	Config    domain.RecordingConfig
	Stats     *domain.RecordingStats
	mutex     sync.RWMutex
}

// DefaultServiceConfig returns default configuration
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		StorageBasePath:         "/var/lib/object-detection/recordings",
		MaxConcurrentRecordings: 10,
		DefaultRetentionPolicy: domain.RetentionPolicy{
			MaxAge:          30 * 24 * time.Hour, // 30 days
			MaxSize:         100 * 1024 * 1024 * 1024, // 100 GB
			MaxCount:        10000,
			AutoArchive:     true,
			ArchiveAfter:    7 * 24 * time.Hour, // 7 days
			CompressionType: "gzip",
			DeleteAfter:     90 * 24 * time.Hour, // 90 days
		},
		QualitySettings: map[domain.RecordingQuality]QualitySettings{
			domain.RecordingQualityLow: {
				Width: 640, Height: 480, FrameRate: 15, Bitrate: 500000, Quality: 50,
			},
			domain.RecordingQualityMedium: {
				Width: 1280, Height: 720, FrameRate: 25, Bitrate: 1500000, Quality: 70,
			},
			domain.RecordingQualityHigh: {
				Width: 1920, Height: 1080, FrameRate: 30, Bitrate: 3000000, Quality: 85,
			},
			domain.RecordingQualityUltra: {
				Width: 3840, Height: 2160, FrameRate: 30, Bitrate: 8000000, Quality: 95,
			},
		},
		EnableCompression:       true,
		CompressionLevel:        6,
		ThumbnailGeneration:     true,
		CleanupInterval:         1 * time.Hour,
		ArchiveInterval:         24 * time.Hour,
		MaxStorageSize:          500 * 1024 * 1024 * 1024, // 500 GB
		StorageWarningThreshold: 0.8, // 80%
	}
}

// NewService creates a new recording service
func NewService(logger *zap.Logger, repository domain.RecordingRepository, config ServiceConfig) *Service {
	service := &Service{
		logger:           logger.With(zap.String("component", "recording_service")),
		repository:       repository,
		config:           config,
		activeRecordings: make(map[string]*ActiveRecording),
		recorders:        make(map[string]domain.VideoRecorder),
		storageManager:   NewStorageManager(logger, config),
	}

	// Start background tasks
	go service.cleanupLoop()
	go service.archiveLoop()

	return service
}

// StartRecording starts a new recording
func (s *Service) StartRecording(config domain.RecordingConfig) (*domain.Recording, error) {
	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid recording config: %w", err)
	}

	// Check concurrent recording limit
	s.mutex.RLock()
	activeCount := len(s.activeRecordings)
	s.mutex.RUnlock()

	if activeCount >= s.config.MaxConcurrentRecordings {
		return nil, fmt.Errorf("maximum concurrent recordings limit reached: %d", s.config.MaxConcurrentRecordings)
	}

	// Create recording record
	recording := &domain.Recording{
		ID:        uuid.New().String(),
		StreamID:  config.StreamID,
		Type:      config.Type,
		Status:    domain.RecordingStatusPending,
		Quality:   config.Quality,
		Format:    config.Format,
		StartTime: time.Now(),
		PreBuffer: config.PreBuffer,
		PostBuffer: config.PostBuffer,
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Set expiration based on retention policy
	if config.RetentionPolicy.MaxAge > 0 {
		expiresAt := recording.StartTime.Add(config.RetentionPolicy.MaxAge)
		recording.ExpiresAt = &expiresAt
	}

	// Generate filename and path
	recording.Filename = s.generateFilename(recording)
	recording.FilePath = filepath.Join(s.config.StorageBasePath, config.StreamID, recording.Filename)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(recording.FilePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create recording directory: %w", err)
	}

	// Create video recorder
	recorder, err := s.createVideoRecorder(config, recording.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create video recorder: %w", err)
	}

	// Start recording
	if err := recorder.StartRecording(config); err != nil {
		return nil, fmt.Errorf("failed to start recording: %w", err)
	}

	// Update recording status
	recording.Status = domain.RecordingStatusRecording
	recording.UpdatedAt = time.Now()

	// Save to repository
	if err := s.repository.CreateRecording(recording); err != nil {
		recorder.StopRecording()
		return nil, fmt.Errorf("failed to save recording: %w", err)
	}

	// Track active recording
	activeRecording := &ActiveRecording{
		Recording: recording,
		Recorder:  recorder,
		StartTime: time.Now(),
		Config:    config,
		Stats:     &domain.RecordingStats{},
	}

	s.mutex.Lock()
	s.activeRecordings[recording.ID] = activeRecording
	s.recorders[config.StreamID] = recorder
	s.mutex.Unlock()

	s.logger.Info("Recording started",
		zap.String("recording_id", recording.ID),
		zap.String("stream_id", config.StreamID),
		zap.String("type", string(config.Type)),
		zap.String("quality", string(config.Quality)))

	return recording, nil
}

// StopRecording stops an active recording
func (s *Service) StopRecording(id string) error {
	s.mutex.Lock()
	activeRecording, exists := s.activeRecordings[id]
	if !exists {
		s.mutex.Unlock()
		return fmt.Errorf("recording not found or not active: %s", id)
	}
	delete(s.activeRecordings, id)
	delete(s.recorders, activeRecording.Recording.StreamID)
	s.mutex.Unlock()

	// Stop the recorder
	if err := activeRecording.Recorder.StopRecording(); err != nil {
		s.logger.Error("Failed to stop recorder", zap.String("recording_id", id), zap.Error(err))
	}

	// Update recording record
	now := time.Now()
	activeRecording.Recording.EndTime = &now
	activeRecording.Recording.Status = domain.RecordingStatusCompleted
	activeRecording.Recording.Duration = now.Sub(activeRecording.Recording.StartTime)
	activeRecording.Recording.UpdatedAt = now

	// Get file size
	if fileInfo, err := os.Stat(activeRecording.Recording.FilePath); err == nil {
		activeRecording.Recording.FileSize = fileInfo.Size()
	}

	// Update in repository
	if err := s.repository.UpdateRecording(activeRecording.Recording); err != nil {
		s.logger.Error("Failed to update recording", zap.String("recording_id", id), zap.Error(err))
	}

	// Generate thumbnail if enabled
	if s.config.ThumbnailGeneration {
		go s.generateThumbnail(activeRecording.Recording)
	}

	s.logger.Info("Recording stopped",
		zap.String("recording_id", id),
		zap.Duration("duration", activeRecording.Recording.Duration),
		zap.Int64("file_size", activeRecording.Recording.FileSize))

	return nil
}

// GetRecording retrieves a recording by ID
func (s *Service) GetRecording(id string) (*domain.Recording, error) {
	return s.repository.GetRecording(id)
}

// GetRecordings retrieves recordings with filters
func (s *Service) GetRecordings(filters domain.RecordingFilters) ([]*domain.Recording, error) {
	return s.repository.GetRecordings(filters)
}

// DeleteRecording deletes a recording and its file
func (s *Service) DeleteRecording(id string) error {
	// Get recording
	recording, err := s.repository.GetRecording(id)
	if err != nil {
		return fmt.Errorf("failed to get recording: %w", err)
	}

	// Stop if currently recording
	s.mutex.RLock()
	if _, isActive := s.activeRecordings[id]; isActive {
		s.mutex.RUnlock()
		if err := s.StopRecording(id); err != nil {
			s.logger.Error("Failed to stop active recording before deletion", zap.Error(err))
		}
	} else {
		s.mutex.RUnlock()
	}

	// Delete file
	if err := os.Remove(recording.FilePath); err != nil && !os.IsNotExist(err) {
		s.logger.Error("Failed to delete recording file", 
			zap.String("file_path", recording.FilePath), 
			zap.Error(err))
	}

	// Delete from repository
	if err := s.repository.DeleteRecording(id); err != nil {
		return fmt.Errorf("failed to delete recording from repository: %w", err)
	}

	s.logger.Info("Recording deleted",
		zap.String("recording_id", id),
		zap.String("file_path", recording.FilePath))

	return nil
}

// TriggerRecording triggers a recording based on an event
func (s *Service) TriggerRecording(trigger domain.RecordingTrigger, context domain.RecordingContext) (*domain.Recording, error) {
	// Get recording configuration for the stream
	config, err := s.getRecordingConfigForStream(context.StreamID, trigger)
	if err != nil {
		return nil, fmt.Errorf("failed to get recording config: %w", err)
	}

	// Check if trigger conditions are met
	if !s.shouldTriggerRecording(config, context) {
		return nil, nil // No recording needed
	}

	// Set trigger-specific properties
	config.Type = domain.RecordingTypeEventTriggered
	
	// Find matching trigger config
	for _, triggerConfig := range config.Triggers {
		if triggerConfig.Type == trigger && triggerConfig.IsEnabled {
			config.PreBuffer = triggerConfig.PreBuffer
			config.PostBuffer = triggerConfig.PostBuffer
			if triggerConfig.MaxDuration > 0 {
				config.MaxDuration = triggerConfig.MaxDuration
			}
			break
		}
	}

	// Start recording
	recording, err := s.StartRecording(config)
	if err != nil {
		return nil, fmt.Errorf("failed to start triggered recording: %w", err)
	}

	// Set trigger-specific metadata
	recording.Trigger = trigger
	recording.AlertID = context.AlertID
	recording.ZoneID = context.ZoneID
	recording.ObjectID = context.ObjectID
	recording.Metadata["trigger_context"] = context
	recording.UpdatedAt = time.Now()

	// Update in repository
	if err := s.repository.UpdateRecording(recording); err != nil {
		s.logger.Error("Failed to update recording with trigger metadata", zap.Error(err))
	}

	s.logger.Info("Recording triggered",
		zap.String("recording_id", recording.ID),
		zap.String("trigger", string(trigger)),
		zap.String("stream_id", context.StreamID),
		zap.String("alert_id", context.AlertID))

	return recording, nil
}

// ProcessAlert processes an alert and triggers recording if needed
func (s *Service) ProcessAlert(alert *domain.Alert) error {
	context := domain.RecordingContext{
		StreamID:    alert.StreamID,
		AlertID:     alert.ID,
		ZoneID:      alert.ZoneID,
		ObjectID:    alert.ObjectID,
		ObjectClass: alert.ObjectClass,
		Confidence:  alert.Confidence,
		Position:    alert.Position,
		Metadata:    alert.Metadata,
		Timestamp:   alert.CreatedAt,
	}

	_, err := s.TriggerRecording(domain.RecordingTriggerAlert, context)
	return err
}

// ProcessZoneEvent processes a zone event and triggers recording if needed
func (s *Service) ProcessZoneEvent(event *domain.ZoneEvent) error {
	context := domain.RecordingContext{
		StreamID:    event.StreamID,
		ZoneID:      event.ZoneID,
		ObjectID:    event.ObjectID,
		ObjectClass: event.ObjectClass,
		Confidence:  event.Confidence,
		Position:    &event.Position,
		Metadata:    event.Metadata,
		Timestamp:   event.Timestamp,
	}

	_, err := s.TriggerRecording(domain.RecordingTriggerZoneViolation, context)
	return err
}

// CreateClip creates a clip from an existing recording
func (s *Service) CreateClip(recordingID string, startOffset, duration time.Duration, title string) (*domain.RecordingClip, error) {
	// Get the source recording
	recording, err := s.repository.GetRecording(recordingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recording: %w", err)
	}

	// Validate clip parameters
	if startOffset < 0 {
		return nil, fmt.Errorf("start offset cannot be negative")
	}
	if duration <= 0 {
		return nil, fmt.Errorf("duration must be positive")
	}
	if startOffset+duration > recording.GetDuration() {
		return nil, fmt.Errorf("clip extends beyond recording duration")
	}

	// Create clip record
	clip := &domain.RecordingClip{
		ID:          uuid.New().String(),
		RecordingID: recordingID,
		StreamID:    recording.StreamID,
		AlertID:     recording.AlertID,
		Title:       title,
		StartOffset: startOffset,
		Duration:    duration,
		Quality:     recording.Quality,
		Format:      recording.Format,
		Metadata:    make(map[string]interface{}),
		CreatedAt:   time.Now(),
	}

	// Generate clip filename
	clip.Filename = s.generateClipFilename(clip)
	clip.FilePath = filepath.Join(filepath.Dir(recording.FilePath), "clips", clip.Filename)

	// Ensure clips directory exists
	if err := os.MkdirAll(filepath.Dir(clip.FilePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create clips directory: %w", err)
	}

	// Extract clip from recording (simplified - would use FFmpeg in production)
	if err := s.extractClip(recording.FilePath, clip.FilePath, startOffset, duration); err != nil {
		return nil, fmt.Errorf("failed to extract clip: %w", err)
	}

	// Get clip file size
	if fileInfo, err := os.Stat(clip.FilePath); err == nil {
		clip.FileSize = fileInfo.Size()
	}

	// Save clip to repository
	if err := s.repository.CreateRecordingClip(clip); err != nil {
		os.Remove(clip.FilePath) // Clean up file on error
		return nil, fmt.Errorf("failed to save clip: %w", err)
	}

	s.logger.Info("Recording clip created",
		zap.String("clip_id", clip.ID),
		zap.String("recording_id", recordingID),
		zap.Duration("start_offset", startOffset),
		zap.Duration("duration", duration))

	return clip, nil
}

// GetRecordingClips retrieves clips for a recording
func (s *Service) GetRecordingClips(recordingID string) ([]*domain.RecordingClip, error) {
	return s.repository.GetRecordingClips(recordingID)
}

// DeleteClip deletes a recording clip
func (s *Service) DeleteClip(id string) error {
	// Get clip
	clip, err := s.repository.GetRecordingClip(id)
	if err != nil {
		return fmt.Errorf("failed to get clip: %w", err)
	}

	// Delete file
	if err := os.Remove(clip.FilePath); err != nil && !os.IsNotExist(err) {
		s.logger.Error("Failed to delete clip file", 
			zap.String("file_path", clip.FilePath), 
			zap.Error(err))
	}

	// Delete from repository
	if err := s.repository.DeleteRecordingClip(id); err != nil {
		return fmt.Errorf("failed to delete clip from repository: %w", err)
	}

	s.logger.Info("Recording clip deleted",
		zap.String("clip_id", id),
		zap.String("file_path", clip.FilePath))

	return nil
}

// Helper methods

func (s *Service) generateFilename(recording *domain.Recording) string {
	timestamp := recording.StartTime.Format("20060102_150405")
	return fmt.Sprintf("%s_%s_%s.%s", 
		recording.StreamID, 
		timestamp, 
		recording.ID[:8], 
		recording.Format)
}

func (s *Service) generateClipFilename(clip *domain.RecordingClip) string {
	timestamp := clip.CreatedAt.Format("20060102_150405")
	return fmt.Sprintf("clip_%s_%s_%s.%s", 
		clip.StreamID, 
		timestamp, 
		clip.ID[:8], 
		clip.Format)
}

func (s *Service) createVideoRecorder(config domain.RecordingConfig, filePath string) (domain.VideoRecorder, error) {
	// This would create an actual video recorder implementation
	// For now, return a mock recorder
	return NewMockVideoRecorder(s.logger, filePath), nil
}

func (s *Service) getRecordingConfigForStream(streamID string, trigger domain.RecordingTrigger) (domain.RecordingConfig, error) {
	// This would typically load configuration from database or config file
	// For now, return a default config
	return domain.RecordingConfig{
		StreamID:    streamID,
		Type:        domain.RecordingTypeEventTriggered,
		Quality:     domain.RecordingQualityMedium,
		Format:      domain.VideoFormatMP4,
		MaxDuration: 5 * time.Minute,
		PreBuffer:   10 * time.Second,
		PostBuffer:  10 * time.Second,
		StoragePath: s.config.StorageBasePath,
		RetentionPolicy: s.config.DefaultRetentionPolicy,
		FrameRate:   25.0,
		Resolution:  domain.Resolution{Width: 1280, Height: 720},
		Bitrate:     1500000,
		Triggers: []domain.TriggerConfig{
			{
				Type:        trigger,
				IsEnabled:   true,
				PreBuffer:   10 * time.Second,
				PostBuffer:  10 * time.Second,
				MaxDuration: 5 * time.Minute,
				Cooldown:    30 * time.Second,
			},
		},
	}, nil
}

func (s *Service) shouldTriggerRecording(config domain.RecordingConfig, context domain.RecordingContext) bool {
	// Check if there's already an active recording for this stream
	s.mutex.RLock()
	for _, activeRecording := range s.activeRecordings {
		if activeRecording.Recording.StreamID == context.StreamID {
			s.mutex.RUnlock()
			return false // Already recording
		}
	}
	s.mutex.RUnlock()

	// Check trigger conditions
	for _, triggerConfig := range config.Triggers {
		if triggerConfig.IsEnabled && triggerConfig.Conditions.MatchesTriggerConditions(context) {
			return true
		}
	}

	return false
}

func (s *Service) extractClip(sourcePath, clipPath string, startOffset, duration time.Duration) error {
	// This would use FFmpeg to extract the clip
	// For now, just copy the file (simplified implementation)
	s.logger.Info("Extracting clip (simplified implementation)",
		zap.String("source", sourcePath),
		zap.String("clip", clipPath),
		zap.Duration("start", startOffset),
		zap.Duration("duration", duration))
	
	// In production, this would be:
	// ffmpeg -i sourcePath -ss startOffset -t duration -c copy clipPath
	return nil
}

func (s *Service) generateThumbnail(recording *domain.Recording) {
	// This would generate a thumbnail from the video
	s.logger.Info("Generating thumbnail (not implemented)",
		zap.String("recording_id", recording.ID),
		zap.String("file_path", recording.FilePath))
}

func (s *Service) cleanupLoop() {
	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := s.CleanupExpiredRecordings(); err != nil {
			s.logger.Error("Failed to cleanup expired recordings", zap.Error(err))
		}
	}
}

func (s *Service) archiveLoop() {
	ticker := time.NewTicker(s.config.ArchiveInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := s.ArchiveOldRecordings(); err != nil {
			s.logger.Error("Failed to archive old recordings", zap.Error(err))
		}
	}
}

// CleanupExpiredRecordings removes expired recordings
func (s *Service) CleanupExpiredRecordings() error {
	// Get expired recordings
	filters := domain.RecordingFilters{
		Limit: 1000, // Process in batches
	}

	recordings, err := s.repository.GetRecordings(filters)
	if err != nil {
		return fmt.Errorf("failed to get recordings: %w", err)
	}

	var deletedCount int
	for _, recording := range recordings {
		if recording.IsExpired() {
			if err := s.DeleteRecording(recording.ID); err != nil {
				s.logger.Error("Failed to delete expired recording",
					zap.String("recording_id", recording.ID),
					zap.Error(err))
			} else {
				deletedCount++
			}
		}
	}

	if deletedCount > 0 {
		s.logger.Info("Cleaned up expired recordings", zap.Int("count", deletedCount))
	}

	return nil
}

// ArchiveOldRecordings archives old recordings
func (s *Service) ArchiveOldRecordings() error {
	// Get recordings that need archiving
	archiveThreshold := time.Now().Add(-s.config.DefaultRetentionPolicy.ArchiveAfter)
	filters := domain.RecordingFilters{
		EndTime: &archiveThreshold,
		Limit:   1000,
	}

	recordings, err := s.repository.GetRecordings(filters)
	if err != nil {
		return fmt.Errorf("failed to get recordings for archiving: %w", err)
	}

	var archivedCount int
	for _, recording := range recordings {
		if recording.Status != domain.RecordingStatusArchived {
			if err := s.archiveRecording(recording); err != nil {
				s.logger.Error("Failed to archive recording",
					zap.String("recording_id", recording.ID),
					zap.Error(err))
			} else {
				archivedCount++
			}
		}
	}

	if archivedCount > 0 {
		s.logger.Info("Archived old recordings", zap.Int("count", archivedCount))
	}

	return nil
}

// GetStorageUsage returns storage usage statistics
func (s *Service) GetStorageUsage(streamID string) (*domain.StorageUsage, error) {
	return s.storageManager.GetStorageUsage(streamID)
}

// OptimizeStorage optimizes storage usage
func (s *Service) OptimizeStorage() error {
	return s.storageManager.OptimizeStorage()
}

// GetRecordingStatistics returns recording statistics
func (s *Service) GetRecordingStatistics(streamID string) (*domain.RecordingStatistics, error) {
	return s.repository.GetRecordingStatistics(streamID)
}

// GenerateRecordingReport generates a recording report
func (s *Service) GenerateRecordingReport(filters domain.RecordingFilters, reportType domain.ReportType) (*domain.RecordingReport, error) {
	recordings, err := s.repository.GetRecordings(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get recordings: %w", err)
	}

	report := &domain.RecordingReport{
		ID:          uuid.New().String(),
		ReportType:  reportType,
		Filters:     filters,
		GeneratedAt: time.Now(),
		Data:        make(map[string]interface{}),
	}

	// Generate report data based on type
	switch reportType {
	case domain.ReportTypeSummary:
		report.Data["total_recordings"] = len(recordings)
		report.Data["recordings_by_type"] = s.groupRecordingsByType(recordings)
		report.Data["recordings_by_status"] = s.groupRecordingsByStatus(recordings)
		report.Data["total_duration"] = s.calculateTotalDuration(recordings)
		report.Data["total_size"] = s.calculateTotalSize(recordings)
		report.Data["average_duration"] = s.calculateAverageDuration(recordings)
		report.Data["average_size"] = s.calculateAverageSize(recordings)
	}

	return report, nil
}

// Helper methods for archiving and statistics

func (s *Service) archiveRecording(recording *domain.Recording) error {
	// Compress the recording file
	if s.config.EnableCompression {
		if err := s.compressRecording(recording); err != nil {
			return fmt.Errorf("failed to compress recording: %w", err)
		}
	}

	// Update status
	recording.Status = domain.RecordingStatusArchived
	recording.UpdatedAt = time.Now()

	return s.repository.UpdateRecording(recording)
}

func (s *Service) compressRecording(recording *domain.Recording) error {
	// This would compress the video file using FFmpeg or similar
	s.logger.Info("Compressing recording (not implemented)",
		zap.String("recording_id", recording.ID),
		zap.String("file_path", recording.FilePath))
	return nil
}

func (s *Service) groupRecordingsByType(recordings []*domain.Recording) map[domain.RecordingType]int64 {
	groups := make(map[domain.RecordingType]int64)
	for _, recording := range recordings {
		groups[recording.Type]++
	}
	return groups
}

func (s *Service) groupRecordingsByStatus(recordings []*domain.Recording) map[domain.RecordingStatus]int64 {
	groups := make(map[domain.RecordingStatus]int64)
	for _, recording := range recordings {
		groups[recording.Status]++
	}
	return groups
}

func (s *Service) calculateTotalDuration(recordings []*domain.Recording) time.Duration {
	var total time.Duration
	for _, recording := range recordings {
		total += recording.GetDuration()
	}
	return total
}

func (s *Service) calculateTotalSize(recordings []*domain.Recording) int64 {
	var total int64
	for _, recording := range recordings {
		total += recording.FileSize
	}
	return total
}

func (s *Service) calculateAverageDuration(recordings []*domain.Recording) time.Duration {
	if len(recordings) == 0 {
		return 0
	}
	return s.calculateTotalDuration(recordings) / time.Duration(len(recordings))
}

func (s *Service) calculateAverageSize(recordings []*domain.Recording) int64 {
	if len(recordings) == 0 {
		return 0
	}
	return s.calculateTotalSize(recordings) / int64(len(recordings))
}

// Ensure Service implements domain.RecordingService
var _ domain.RecordingService = (*Service)(nil)
