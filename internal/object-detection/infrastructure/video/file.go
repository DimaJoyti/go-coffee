package video

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"gocv.io/x/gocv"
)

// FileInput handles video file input
type FileInput struct {
	logger       *zap.Logger
	filePath     string
	capture      *gocv.VideoCapture
	isActive     bool
	frameRate    float64
	width        int
	height       int
	totalFrames  int
	currentFrame int
	duration     time.Duration
}

// SupportedVideoFormats lists supported video file formats
var SupportedVideoFormats = []string{
	".mp4", ".avi", ".mov", ".mkv", ".wmv", ".flv", ".webm", ".m4v", ".3gp", ".mpg", ".mpeg",
}

// NewFileInput creates a new file input handler
func NewFileInput(logger *zap.Logger, filePath string) *FileInput {
	return &FileInput{
		logger:   logger.With(zap.String("component", "file_input"), zap.String("file", filePath)),
		filePath: filePath,
		isActive: false,
	}
}

// Open opens the video file for reading
func (f *FileInput) Open() error {
	f.logger.Info("Opening video file", zap.String("path", f.filePath))

	// Check if file exists
	if _, err := os.Stat(f.filePath); os.IsNotExist(err) {
		return fmt.Errorf("video file does not exist: %s", f.filePath)
	}

	// Check if file format is supported
	if !f.isFormatSupported() {
		return fmt.Errorf("unsupported video format: %s", filepath.Ext(f.filePath))
	}

	// Open video file
	capture, err := gocv.OpenVideoCapture(f.filePath)
	if err != nil {
		return fmt.Errorf("failed to open video file %s: %w", f.filePath, err)
	}

	if !capture.IsOpened() {
		capture.Close()
		return fmt.Errorf("video file %s cannot be opened", f.filePath)
	}

	f.capture = capture
	f.isActive = true

	// Get file properties
	if err := f.getProperties(); err != nil {
		f.Close()
		return fmt.Errorf("failed to get video file properties: %w", err)
	}

	f.logger.Info("Video file opened successfully",
		zap.Float64("fps", f.frameRate),
		zap.Int("width", f.width),
		zap.Int("height", f.height),
		zap.Int("total_frames", f.totalFrames),
		zap.Duration("duration", f.duration))

	return nil
}

// Close closes the video file
func (f *FileInput) Close() error {
	f.logger.Info("Closing video file")

	f.isActive = false

	if f.capture != nil {
		f.capture.Close()
		f.capture = nil
	}

	f.logger.Info("Video file closed")
	return nil
}

// IsActive returns whether the file input is active
func (f *FileInput) IsActive() bool {
	return f.isActive && f.capture != nil && f.capture.IsOpened()
}

// ReadFrame reads the next frame from the video file
func (f *FileInput) ReadFrame(ctx context.Context) ([]byte, error) {
	if !f.IsActive() {
		return nil, fmt.Errorf("file input is not active")
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Read frame
	img := gocv.NewMat()
	defer img.Close()

	if !f.capture.Read(&img) {
		f.logger.Info("End of video file reached")
		return nil, fmt.Errorf("end of video file")
	}

	if img.Empty() {
		return nil, fmt.Errorf("empty frame received from video file")
	}

	f.currentFrame++

	// Convert to JPEG bytes
	buf, err := gocv.IMEncode(gocv.JPEGFileExt, img)
	if err != nil {
		return nil, fmt.Errorf("failed to encode frame: %w", err)
	}
	defer buf.Close()

	return buf.GetBytes(), nil
}

// ReadFrameMat reads the next frame as a gocv.Mat
func (f *FileInput) ReadFrameMat(ctx context.Context) (*gocv.Mat, error) {
	if !f.IsActive() {
		return nil, fmt.Errorf("file input is not active")
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Read frame
	img := gocv.NewMat()

	if !f.capture.Read(&img) {
		img.Close()
		f.logger.Info("End of video file reached")
		return nil, fmt.Errorf("end of video file")
	}

	if img.Empty() {
		img.Close()
		return nil, fmt.Errorf("empty frame received from video file")
	}

	f.currentFrame++
	return &img, nil
}

// GetFrameRate returns the video file frame rate
func (f *FileInput) GetFrameRate() float64 {
	return f.frameRate
}

// GetResolution returns the video file resolution
func (f *FileInput) GetResolution() (width, height int) {
	return f.width, f.height
}

// GetTotalFrames returns the total number of frames in the video
func (f *FileInput) GetTotalFrames() int {
	return f.totalFrames
}

// GetCurrentFrame returns the current frame number
func (f *FileInput) GetCurrentFrame() int {
	return f.currentFrame
}

// GetDuration returns the video duration
func (f *FileInput) GetDuration() time.Duration {
	return f.duration
}

// GetProgress returns the playback progress as a percentage (0-100)
func (f *FileInput) GetProgress() float64 {
	if f.totalFrames <= 0 {
		return 0
	}
	return (float64(f.currentFrame) / float64(f.totalFrames)) * 100
}

// SeekToFrame seeks to a specific frame number
func (f *FileInput) SeekToFrame(frameNumber int) error {
	if !f.IsActive() {
		return fmt.Errorf("file input is not active")
	}

	if frameNumber < 0 || frameNumber >= f.totalFrames {
		return fmt.Errorf("frame number %d is out of range (0-%d)", frameNumber, f.totalFrames-1)
	}

	f.logger.Info("Seeking to frame", zap.Int("frame", frameNumber))

	// Set frame position
	f.capture.Set(gocv.VideoCapturePosFrames, float64(frameNumber))
	f.currentFrame = frameNumber

	return nil
}

// SeekToTime seeks to a specific time position
func (f *FileInput) SeekToTime(position time.Duration) error {
	if !f.IsActive() {
		return fmt.Errorf("file input is not active")
	}

	if position < 0 || position > f.duration {
		return fmt.Errorf("time position %v is out of range (0-%v)", position, f.duration)
	}

	f.logger.Info("Seeking to time", zap.Duration("position", position))

	// Calculate frame number
	frameNumber := int((position.Seconds() * f.frameRate))
	return f.SeekToFrame(frameNumber)
}

// SeekToPercentage seeks to a specific percentage of the video
func (f *FileInput) SeekToPercentage(percentage float64) error {
	if percentage < 0 || percentage > 100 {
		return fmt.Errorf("percentage %.2f is out of range (0-100)", percentage)
	}

	frameNumber := int((percentage / 100.0) * float64(f.totalFrames))
	return f.SeekToFrame(frameNumber)
}

// isFormatSupported checks if the video format is supported
func (f *FileInput) isFormatSupported() bool {
	ext := strings.ToLower(filepath.Ext(f.filePath))
	for _, format := range SupportedVideoFormats {
		if ext == format {
			return true
		}
	}
	return false
}

// getProperties retrieves video file properties
func (f *FileInput) getProperties() error {
	if f.capture == nil {
		return fmt.Errorf("capture is not initialized")
	}

	// Get frame rate
	fps := f.capture.Get(gocv.VideoCaptureFPS)
	if fps <= 0 {
		fps = 25.0 // Default frame rate
		f.logger.Warn("Unable to detect video frame rate, using default", zap.Float64("fps", fps))
	}
	f.frameRate = fps

	// Get resolution
	width := int(f.capture.Get(gocv.VideoCaptureFrameWidth))
	height := int(f.capture.Get(gocv.VideoCaptureFrameHeight))

	if width <= 0 || height <= 0 {
		return fmt.Errorf("unable to determine video resolution")
	}

	f.width = width
	f.height = height

	// Get total frames
	totalFrames := int(f.capture.Get(gocv.VideoCaptureFrameCount))
	if totalFrames <= 0 {
		f.logger.Warn("Unable to determine total frame count")
		totalFrames = 0
	}
	f.totalFrames = totalFrames

	// Calculate duration
	if totalFrames > 0 && fps > 0 {
		f.duration = time.Duration(float64(totalFrames)/fps) * time.Second
	}

	f.currentFrame = 0

	return nil
}

// GetStats returns video file statistics
func (f *FileInput) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"file_path":      f.filePath,
		"is_active":      f.IsActive(),
		"frame_rate":     f.frameRate,
		"width":          f.width,
		"height":         f.height,
		"total_frames":   f.totalFrames,
		"current_frame":  f.currentFrame,
		"duration":       f.duration.String(),
		"progress":       f.GetProgress(),
		"type":           "file",
	}
}

// StartPlayback starts video playback with a callback
func (f *FileInput) StartPlayback(ctx context.Context, frameCallback func([]byte) error) error {
	if !f.IsActive() {
		return fmt.Errorf("file input is not active")
	}

	f.logger.Info("Starting video playback")

	frameInterval := time.Duration(1000.0/f.frameRate) * time.Millisecond
	ticker := time.NewTicker(frameInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			f.logger.Info("Video playback stopped by context")
			return ctx.Err()

		case <-ticker.C:
			frameData, err := f.ReadFrame(ctx)
			if err != nil {
				if err.Error() == "end of video file" {
					f.logger.Info("Video playback completed")
					return nil
				}
				f.logger.Error("Failed to read frame", zap.Error(err))
				continue
			}

			if err := frameCallback(frameData); err != nil {
				f.logger.Error("Frame callback error", zap.Error(err))
				continue
			}
		}
	}
}

// ValidateVideoFile validates if a file is a valid video file
func ValidateVideoFile(filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	supported := false
	for _, format := range SupportedVideoFormats {
		if ext == format {
			supported = true
			break
		}
	}

	if !supported {
		return fmt.Errorf("unsupported video format: %s", ext)
	}

	// Try to open the file with OpenCV
	capture, err := gocv.OpenVideoCapture(filePath)
	if err != nil {
		return fmt.Errorf("failed to open video file: %w", err)
	}
	defer capture.Close()

	if !capture.IsOpened() {
		return fmt.Errorf("video file cannot be opened")
	}

	// Try to read one frame to validate
	img := gocv.NewMat()
	defer img.Close()

	if !capture.Read(&img) {
		return fmt.Errorf("unable to read frames from video file")
	}

	if img.Empty() {
		return fmt.Errorf("video file contains empty frames")
	}

	return nil
}
