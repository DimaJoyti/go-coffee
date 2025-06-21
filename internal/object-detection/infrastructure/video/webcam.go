package video

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
	"gocv.io/x/gocv"
)

// WebcamInput handles webcam video input
type WebcamInput struct {
	logger    *zap.Logger
	deviceID  int
	capture   *gocv.VideoCapture
	isActive  bool
	frameRate float64
	width     int
	height    int
}

// NewWebcamInput creates a new webcam input handler
func NewWebcamInput(logger *zap.Logger, deviceID int) *WebcamInput {
	return &WebcamInput{
		logger:   logger.With(zap.String("component", "webcam_input"), zap.Int("device_id", deviceID)),
		deviceID: deviceID,
		isActive: false,
	}
}

// NewWebcamInputFromSource creates a webcam input from a source string
func NewWebcamInputFromSource(logger *zap.Logger, source string) (*WebcamInput, error) {
	deviceID := 0

	// Parse device ID from source
	if source != "" && source != "/dev/video0" {
		// Try to parse as integer
		if id, err := strconv.Atoi(source); err == nil {
			deviceID = id
		} else {
			// Try to parse device path like /dev/video1
			if len(source) > 10 && source[:10] == "/dev/video" {
				if id, err := strconv.Atoi(source[10:]); err == nil {
					deviceID = id
				}
			}
		}
	}

	return NewWebcamInput(logger, deviceID), nil
}

// Open opens the webcam for capture
func (w *WebcamInput) Open() error {
	w.logger.Info("Opening webcam", zap.Int("device_id", w.deviceID))

	// Open video capture
	capture, err := gocv.OpenVideoCapture(w.deviceID)
	if err != nil {
		return fmt.Errorf("failed to open webcam device %d: %w", w.deviceID, err)
	}

	if !capture.IsOpened() {
		capture.Close()
		return fmt.Errorf("webcam device %d is not available", w.deviceID)
	}

	w.capture = capture
	w.isActive = true

	// Get webcam properties
	if err := w.getProperties(); err != nil {
		w.Close()
		return fmt.Errorf("failed to get webcam properties: %w", err)
	}

	w.logger.Info("Webcam opened successfully",
		zap.Float64("fps", w.frameRate),
		zap.Int("width", w.width),
		zap.Int("height", w.height))

	return nil
}

// Close closes the webcam
func (w *WebcamInput) Close() error {
	w.logger.Info("Closing webcam")

	w.isActive = false

	if w.capture != nil {
		w.capture.Close()
		w.capture = nil
	}

	w.logger.Info("Webcam closed")
	return nil
}

// IsActive returns whether the webcam is active
func (w *WebcamInput) IsActive() bool {
	return w.isActive && w.capture != nil && w.capture.IsOpened()
}

// ReadFrame reads the next frame from the webcam
func (w *WebcamInput) ReadFrame(ctx context.Context) ([]byte, error) {
	if !w.IsActive() {
		return nil, fmt.Errorf("webcam is not active")
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

	if !w.capture.Read(&img) {
		return nil, fmt.Errorf("failed to read frame from webcam")
	}

	if img.Empty() {
		return nil, fmt.Errorf("empty frame received from webcam")
	}

	// Convert to JPEG bytes
	buf, err := gocv.IMEncode(gocv.JPEGFileExt, img)
	if err != nil {
		return nil, fmt.Errorf("failed to encode frame: %w", err)
	}
	defer buf.Close()

	return buf.GetBytes(), nil
}

// ReadFrameMat reads the next frame as a gocv.Mat
func (w *WebcamInput) ReadFrameMat(ctx context.Context) (*gocv.Mat, error) {
	if !w.IsActive() {
		return nil, fmt.Errorf("webcam is not active")
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Read frame
	img := gocv.NewMat()

	if !w.capture.Read(&img) {
		img.Close()
		return nil, fmt.Errorf("failed to read frame from webcam")
	}

	if img.Empty() {
		img.Close()
		return nil, fmt.Errorf("empty frame received from webcam")
	}

	return &img, nil
}

// GetFrameRate returns the webcam frame rate
func (w *WebcamInput) GetFrameRate() float64 {
	return w.frameRate
}

// GetResolution returns the webcam resolution
func (w *WebcamInput) GetResolution() (width, height int) {
	return w.width, w.height
}

// SetResolution sets the webcam resolution
func (w *WebcamInput) SetResolution(width, height int) error {
	if !w.IsActive() {
		return fmt.Errorf("webcam is not active")
	}

	w.logger.Info("Setting webcam resolution",
		zap.Int("width", width),
		zap.Int("height", height))

	// Set resolution
	w.capture.Set(gocv.VideoCaptureFrameWidth, float64(width))
	w.capture.Set(gocv.VideoCaptureFrameHeight, float64(height))

	// Update stored values
	w.width = int(w.capture.Get(gocv.VideoCaptureFrameWidth))
	w.height = int(w.capture.Get(gocv.VideoCaptureFrameHeight))

	w.logger.Info("Webcam resolution updated",
		zap.Int("actual_width", w.width),
		zap.Int("actual_height", w.height))

	return nil
}

// SetFrameRate sets the webcam frame rate
func (w *WebcamInput) SetFrameRate(fps float64) error {
	if !w.IsActive() {
		return fmt.Errorf("webcam is not active")
	}

	w.logger.Info("Setting webcam frame rate", zap.Float64("fps", fps))

	// Set frame rate
	w.capture.Set(gocv.VideoCaptureFPS, fps)

	// Update stored value
	w.frameRate = w.capture.Get(gocv.VideoCaptureFPS)

	w.logger.Info("Webcam frame rate updated", zap.Float64("actual_fps", w.frameRate))

	return nil
}

// getProperties retrieves webcam properties
func (w *WebcamInput) getProperties() error {
	if w.capture == nil {
		return fmt.Errorf("capture is not initialized")
	}

	// Get frame rate
	fps := w.capture.Get(gocv.VideoCaptureFPS)
	if fps <= 0 {
		// Default to 30 FPS if unable to detect
		fps = 30.0
		w.logger.Warn("Unable to detect webcam frame rate, using default", zap.Float64("fps", fps))
	}
	w.frameRate = fps

	// Get resolution
	width := int(w.capture.Get(gocv.VideoCaptureFrameWidth))
	height := int(w.capture.Get(gocv.VideoCaptureFrameHeight))

	if width <= 0 || height <= 0 {
		// Try to read a frame to get dimensions
		img := gocv.NewMat()
		defer img.Close()

		if w.capture.Read(&img) {
			width = img.Cols()
			height = img.Rows()
		}

		if width <= 0 || height <= 0 {
			return fmt.Errorf("unable to determine webcam resolution")
		}
	}

	w.width = width
	w.height = height

	return nil
}

// GetStats returns webcam statistics
func (w *WebcamInput) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"device_id":   w.deviceID,
		"is_active":   w.IsActive(),
		"frame_rate":  w.frameRate,
		"width":       w.width,
		"height":      w.height,
		"type":        "webcam",
	}
}

// StartCapture starts continuous frame capture with a callback
func (w *WebcamInput) StartCapture(ctx context.Context, frameCallback func([]byte) error) error {
	if !w.IsActive() {
		return fmt.Errorf("webcam is not active")
	}

	w.logger.Info("Starting webcam capture")

	frameInterval := time.Duration(1000.0/w.frameRate) * time.Millisecond
	ticker := time.NewTicker(frameInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Webcam capture stopped by context")
			return ctx.Err()

		case <-ticker.C:
			frameData, err := w.ReadFrame(ctx)
			if err != nil {
				w.logger.Error("Failed to read frame", zap.Error(err))
				continue
			}

			if err := frameCallback(frameData); err != nil {
				w.logger.Error("Frame callback error", zap.Error(err))
				continue
			}
		}
	}
}

// ListAvailableDevices lists available webcam devices
func ListAvailableDevices(logger *zap.Logger) []int {
	var devices []int

	// Try to open devices 0-9
	for i := 0; i < 10; i++ {
		capture, err := gocv.OpenVideoCapture(i)
		if err != nil {
			continue
		}

		if capture.IsOpened() {
			devices = append(devices, i)
			logger.Debug("Found webcam device", zap.Int("device_id", i))
		}

		capture.Close()
	}

	logger.Info("Available webcam devices", zap.Ints("devices", devices))
	return devices
}
