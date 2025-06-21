//go:build !opencv
// +build !opencv

package video

import "time"

// Frame represents a video frame (stub version for non-opencv builds)
type Frame struct {
	ID        string
	StreamID  string
	Data      []byte
	Width     int
	Height    int
	Timestamp time.Time
	FrameNum  int
}

// NewFrame creates a new frame (stub version for non-opencv builds)
func NewFrame(streamID string, frameNum int, data []byte) *Frame {
	return &Frame{
		ID:        streamID + "_frame_" + string(rune(frameNum)),
		StreamID:  streamID,
		Data:      data,
		Timestamp: time.Now(),
		FrameNum:  frameNum,
	}
}