package video

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"go.uber.org/zap"
	"gocv.io/x/gocv"
)

// openWebcam opens a webcam stream
func (h *StreamHandler) openWebcam() error {
	h.Logger.Info("Opening webcam", zap.String("source", h.Source))

	// Parse webcam device ID
	deviceID := 0
	if h.Source != "" && h.Source != "/dev/video0" {
		if id, err := strconv.Atoi(h.Source); err == nil {
			deviceID = id
		}
	}

	// Open video capture
	capture, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		return fmt.Errorf("failed to open webcam %d: %w", deviceID, err)
	}

	if !capture.IsOpened() {
		capture.Close()
		return fmt.Errorf("webcam %d is not available", deviceID)
	}

	h.Capture = capture
	h.IsActive = true

	h.Logger.Info("Webcam opened successfully", zap.Int("device_id", deviceID))
	return nil
}

// openFile opens a video file stream
func (h *StreamHandler) openFile() error {
	h.Logger.Info("Opening video file", zap.String("source", h.Source))

	// Open video file
	capture, err := gocv.OpenVideoCapture(h.Source)
	if err != nil {
		return fmt.Errorf("failed to open video file %s: %w", h.Source, err)
	}

	if !capture.IsOpened() {
		capture.Close()
		return fmt.Errorf("video file %s cannot be opened", h.Source)
	}

	h.Capture = capture
	h.IsActive = true

	h.Logger.Info("Video file opened successfully", zap.String("file", h.Source))
	return nil
}

// openNetwork opens a network stream (RTMP, HTTP)
func (h *StreamHandler) openNetwork() error {
	h.Logger.Info("Opening network stream", 
		zap.String("source", h.Source),
		zap.String("type", string(h.Type)))

	// Open network stream
	capture, err := gocv.OpenVideoCapture(h.Source)
	if err != nil {
		return fmt.Errorf("failed to open network stream %s: %w", h.Source, err)
	}

	if !capture.IsOpened() {
		capture.Close()
		return fmt.Errorf("network stream %s cannot be opened", h.Source)
	}

	h.Capture = capture
	h.IsActive = true

	h.Logger.Info("Network stream opened successfully", zap.String("url", h.Source))
	return nil
}

// getStreamProperties retrieves stream properties
func (h *StreamHandler) getStreamProperties() error {
	if h.Capture == nil {
		return fmt.Errorf("capture is not initialized")
	}

	// Get frame rate
	fps := h.Capture.Get(gocv.VideoCaptureFPS)
	if fps <= 0 {
		// Default to 30 FPS if unable to detect
		fps = 30.0
		h.Logger.Warn("Unable to detect frame rate, using default", zap.Float64("fps", fps))
	}
	h.FrameRate = fps

	// Get resolution
	width := int(h.Capture.Get(gocv.VideoCaptureFrameWidth))
	height := int(h.Capture.Get(gocv.VideoCaptureFrameHeight))

	if width <= 0 || height <= 0 {
		// Try to read a frame to get dimensions
		img := gocv.NewMat()
		defer img.Close()

		if h.Capture.Read(&img) {
			width = img.Cols()
			height = img.Rows()
		}

		if width <= 0 || height <= 0 {
			return fmt.Errorf("unable to determine stream resolution")
		}
	}

	h.Width = width
	h.Height = height

	h.Logger.Info("Stream properties detected",
		zap.Float64("fps", h.FrameRate),
		zap.Int("width", h.Width),
		zap.Int("height", h.Height))

	return nil
}

// readFrame reads the next frame from the stream
func (h *StreamHandler) readFrame(ctx context.Context) ([]byte, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if !h.IsActive || h.Capture == nil {
		return nil, fmt.Errorf("stream is not active")
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

	if !h.Capture.Read(&img) {
		// End of stream or error
		if h.Type == domain.StreamTypeFile {
			h.Logger.Info("End of video file reached")
			h.Status = domain.StreamStatusStopped
			return nil, fmt.Errorf("end of stream")
		}
		return nil, fmt.Errorf("failed to read frame")
	}

	if img.Empty() {
		return nil, fmt.Errorf("empty frame received")
	}

	// Convert frame to JPEG bytes
	buf, err := gocv.IMEncode(gocv.JPEGFileExt, img)
	if err != nil {
		return nil, fmt.Errorf("failed to encode frame: %w", err)
	}
	defer buf.Close()

	return buf.GetBytes(), nil
}

// readFrameMat reads the next frame as a gocv.Mat
func (h *StreamHandler) readFrameMat(ctx context.Context) (*gocv.Mat, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if !h.IsActive || h.Capture == nil {
		return nil, fmt.Errorf("stream is not active")
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Read frame
	img := gocv.NewMat()

	if !h.Capture.Read(&img) {
		img.Close()
		// End of stream or error
		if h.Type == domain.StreamTypeFile {
			h.Logger.Info("End of video file reached")
			h.Status = domain.StreamStatusStopped
			return nil, fmt.Errorf("end of stream")
		}
		return nil, fmt.Errorf("failed to read frame")
	}

	if img.Empty() {
		img.Close()
		return nil, fmt.Errorf("empty frame received")
	}

	return &img, nil
}

// stop stops the stream handler
func (h *StreamHandler) stop() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.Logger.Info("Stopping stream handler")

	h.IsActive = false
	h.Status = domain.StreamStatusStopped

	// Signal stop
	select {
	case h.StopChannel <- true:
	default:
	}

	h.close()
}

// close closes the video capture
func (h *StreamHandler) close() {
	if h.Capture != nil {
		h.Capture.Close()
		h.Capture = nil
	}
	close(h.StopChannel)
}

// GetStats returns stream statistics
func (h *StreamHandler) GetStats() map[string]interface{} {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return map[string]interface{}{
		"id":         h.ID,
		"source":     h.Source,
		"type":       string(h.Type),
		"status":     string(h.Status),
		"is_active":  h.IsActive,
		"frame_rate": h.FrameRate,
		"width":      h.Width,
		"height":     h.Height,
	}
}

// SetFrameRate sets a custom frame rate (for throttling)
func (h *StreamHandler) SetFrameRate(fps float64) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if fps > 0 {
		h.FrameRate = fps
		h.Logger.Info("Frame rate updated", zap.Float64("fps", fps))
	}
}

// GetFrameInterval returns the interval between frames in milliseconds
func (h *StreamHandler) GetFrameInterval() time.Duration {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if h.FrameRate <= 0 {
		return time.Millisecond * 33 // Default to ~30 FPS
	}

	return time.Duration(1000.0/h.FrameRate) * time.Millisecond
}
