//go:build opencv
// +build opencv

package video

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"gocv.io/x/gocv"
)

// Frame represents a video frame with metadata
type Frame struct {
	ID        string
	StreamID  string
	Data      []byte
	Mat       *gocv.Mat
	Width     int
	Height    int
	Timestamp time.Time
	FrameNum  int
}

// FrameProcessor defines the interface for frame processing
type FrameProcessor interface {
	ProcessFrame(ctx context.Context, frame *Frame) (*Frame, error)
}

// Pipeline manages the video processing pipeline
type Pipeline struct {
	logger      *zap.Logger
	streamID    string
	input       chan *Frame
	output      chan *Frame
	processors  []FrameProcessor
	workers     int
	isRunning   bool
	stopChannel chan bool
	wg          sync.WaitGroup
	mutex       sync.RWMutex
	stats       *PipelineStats
}

// PipelineStats tracks pipeline performance
type PipelineStats struct {
	FramesProcessed   int64
	FramesDropped     int64
	AverageProcessTime time.Duration
	TotalProcessTime  time.Duration
	StartTime         time.Time
	LastFrameTime     time.Time
	FPS               float64
	mutex             sync.RWMutex
}

// PipelineConfig configures the processing pipeline
type PipelineConfig struct {
	StreamID       string
	Workers        int
	BufferSize     int
	MaxProcessTime time.Duration
	DropFrames     bool
}

// NewPipeline creates a new video processing pipeline
func NewPipeline(logger *zap.Logger, config PipelineConfig) *Pipeline {
	if config.Workers <= 0 {
		config.Workers = 1
	}
	if config.BufferSize <= 0 {
		config.BufferSize = 10
	}

	return &Pipeline{
		logger:      logger.With(zap.String("component", "pipeline"), zap.String("stream_id", config.StreamID)),
		streamID:    config.StreamID,
		input:       make(chan *Frame, config.BufferSize),
		output:      make(chan *Frame, config.BufferSize),
		processors:  make([]FrameProcessor, 0),
		workers:     config.Workers,
		isRunning:   false,
		stopChannel: make(chan bool, 1),
		stats: &PipelineStats{
			StartTime: time.Now(),
		},
	}
}

// AddProcessor adds a frame processor to the pipeline
func (p *Pipeline) AddProcessor(processor FrameProcessor) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.processors = append(p.processors, processor)
	p.logger.Info("Added processor to pipeline", zap.Int("total_processors", len(p.processors)))
}

// Start starts the processing pipeline
func (p *Pipeline) Start(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.isRunning {
		return fmt.Errorf("pipeline is already running")
	}

	p.logger.Info("Starting processing pipeline", zap.Int("workers", p.workers))

	p.isRunning = true
	p.stats.StartTime = time.Now()

	// Start worker goroutines
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i)
	}

	p.logger.Info("Processing pipeline started")
	return nil
}

// Stop stops the processing pipeline
func (p *Pipeline) Stop() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.isRunning {
		return fmt.Errorf("pipeline is not running")
	}

	p.logger.Info("Stopping processing pipeline")

	p.isRunning = false

	// Signal stop
	select {
	case p.stopChannel <- true:
	default:
	}

	// Close input channel
	close(p.input)

	// Wait for workers to finish
	p.wg.Wait()

	// Close output channel
	close(p.output)

	p.logger.Info("Processing pipeline stopped")
	return nil
}

// ProcessFrame sends a frame to the pipeline for processing
func (p *Pipeline) ProcessFrame(frame *Frame) error {
	if !p.isRunning {
		return fmt.Errorf("pipeline is not running")
	}

	select {
	case p.input <- frame:
		return nil
	default:
		// Buffer is full, drop frame if configured
		p.stats.mutex.Lock()
		p.stats.FramesDropped++
		p.stats.mutex.Unlock()
		return fmt.Errorf("pipeline buffer full, frame dropped")
	}
}

// GetOutputChannel returns the output channel for processed frames
func (p *Pipeline) GetOutputChannel() <-chan *Frame {
	return p.output
}

// worker processes frames in the pipeline
func (p *Pipeline) worker(ctx context.Context, workerID int) {
	defer p.wg.Done()

	logger := p.logger.With(zap.Int("worker_id", workerID))
	logger.Info("Pipeline worker started")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Pipeline worker stopped by context")
			return

		case <-p.stopChannel:
			logger.Info("Pipeline worker stopped by signal")
			return

		case frame, ok := <-p.input:
			if !ok {
				logger.Info("Pipeline worker stopped - input channel closed")
				return
			}

			if frame == nil {
				continue
			}

			startTime := time.Now()

			// Process frame through all processors
			processedFrame, err := p.processFrameThroughPipeline(ctx, frame)
			if err != nil {
				logger.Error("Frame processing failed", zap.Error(err))
				continue
			}

			processTime := time.Since(startTime)

			// Update stats
			p.updateStats(processTime)

			// Send processed frame to output
			select {
			case p.output <- processedFrame:
			case <-ctx.Done():
				return
			default:
				// Output buffer full, drop frame
				p.stats.mutex.Lock()
				p.stats.FramesDropped++
				p.stats.mutex.Unlock()
				logger.Warn("Output buffer full, dropping processed frame")
			}
		}
	}
}

// processFrameThroughPipeline processes a frame through all processors
func (p *Pipeline) processFrameThroughPipeline(ctx context.Context, frame *Frame) (*Frame, error) {
	currentFrame := frame

	for i, processor := range p.processors {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		processedFrame, err := processor.ProcessFrame(ctx, currentFrame)
		if err != nil {
			return nil, fmt.Errorf("processor %d failed: %w", i, err)
		}

		currentFrame = processedFrame
	}

	return currentFrame, nil
}

// updateStats updates pipeline statistics
func (p *Pipeline) updateStats(processTime time.Duration) {
	p.stats.mutex.Lock()
	defer p.stats.mutex.Unlock()

	p.stats.FramesProcessed++
	p.stats.TotalProcessTime += processTime
	p.stats.AverageProcessTime = p.stats.TotalProcessTime / time.Duration(p.stats.FramesProcessed)
	p.stats.LastFrameTime = time.Now()

	// Calculate FPS
	elapsed := time.Since(p.stats.StartTime).Seconds()
	if elapsed > 0 {
		p.stats.FPS = float64(p.stats.FramesProcessed) / elapsed
	}
}

// GetStats returns pipeline statistics
func (p *Pipeline) GetStats() *PipelineStats {
	p.stats.mutex.RLock()
	defer p.stats.mutex.RUnlock()

	// Return a copy to avoid race conditions
	return &PipelineStats{
		FramesProcessed:    p.stats.FramesProcessed,
		FramesDropped:      p.stats.FramesDropped,
		AverageProcessTime: p.stats.AverageProcessTime,
		TotalProcessTime:   p.stats.TotalProcessTime,
		StartTime:          p.stats.StartTime,
		LastFrameTime:      p.stats.LastFrameTime,
		FPS:                p.stats.FPS,
	}
}

// IsRunning returns whether the pipeline is running
func (p *Pipeline) IsRunning() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.isRunning
}

// GetBufferUsage returns the current buffer usage
func (p *Pipeline) GetBufferUsage() (input, output int) {
	return len(p.input), len(p.output)
}

// NewFrame creates a new frame from image data
func NewFrame(streamID string, frameNum int, data []byte) *Frame {
	return &Frame{
		ID:        fmt.Sprintf("%s_frame_%d", streamID, frameNum),
		StreamID:  streamID,
		Data:      data,
		Timestamp: time.Now(),
		FrameNum:  frameNum,
	}
}

// NewFrameFromMat creates a new frame from a gocv.Mat
func NewFrameFromMat(streamID string, frameNum int, mat *gocv.Mat) (*Frame, error) {
	if mat == nil || mat.Empty() {
		return nil, fmt.Errorf("invalid or empty mat")
	}

	// Encode mat to JPEG
	buf, err := gocv.IMEncode(gocv.JPEGFileExt, *mat)
	if err != nil {
		return nil, fmt.Errorf("failed to encode mat: %w", err)
	}
	defer buf.Close()

	frame := &Frame{
		ID:        fmt.Sprintf("%s_frame_%d", streamID, frameNum),
		StreamID:  streamID,
		Data:      buf.GetBytes(),
		Mat:       mat,
		Width:     mat.Cols(),
		Height:    mat.Rows(),
		Timestamp: time.Now(),
		FrameNum:  frameNum,
	}

	return frame, nil
}

// Clone creates a copy of the frame
func (f *Frame) Clone() *Frame {
	clone := &Frame{
		ID:        f.ID,
		StreamID:  f.StreamID,
		Width:     f.Width,
		Height:    f.Height,
		Timestamp: f.Timestamp,
		FrameNum:  f.FrameNum,
	}

	// Copy data
	if f.Data != nil {
		clone.Data = make([]byte, len(f.Data))
		copy(clone.Data, f.Data)
	}

	// Clone mat if present
	if f.Mat != nil && !f.Mat.Empty() {
		cloneMat := f.Mat.Clone()
		clone.Mat = &cloneMat
	}

	return clone
}

// Release releases resources held by the frame
func (f *Frame) Release() {
	if f.Mat != nil {
		f.Mat.Close()
		f.Mat = nil
	}
	f.Data = nil
}

// ToMat converts frame data to gocv.Mat
func (f *Frame) ToMat() (*gocv.Mat, error) {
	if f.Mat != nil && !f.Mat.Empty() {
		return f.Mat, nil
	}

	if f.Data == nil {
		return nil, fmt.Errorf("no frame data available")
	}

	// Decode JPEG data to Mat
	mat, err := gocv.IMDecode(f.Data, gocv.IMReadColor)
	if err != nil {
		return nil, fmt.Errorf("failed to decode frame data: %w", err)
	}

	f.Mat = &mat
	f.Width = mat.Cols()
	f.Height = mat.Rows()

	return &mat, nil
}

// GetSize returns the frame size in bytes
func (f *Frame) GetSize() int {
	if f.Data != nil {
		return len(f.Data)
	}
	return 0
}
