package tracking

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"go.uber.org/zap"
)

// TrajectoryRecorder records and analyzes object trajectories
type TrajectoryRecorder struct {
	logger      *zap.Logger
	trajectories map[string]*TrajectoryData
	config      TrajectoryConfig
	mutex       sync.RWMutex
	stats       *TrajectoryStats
}

// TrajectoryConfig configures trajectory recording
type TrajectoryConfig struct {
	MaxTrajectoryLength int           // Maximum number of points to keep
	SamplingInterval    time.Duration // Minimum time between recorded points
	SmoothingWindow     int           // Window size for trajectory smoothing
	EnablePrediction    bool          // Enable trajectory prediction
	PredictionHorizon   time.Duration // How far ahead to predict
	MinTrajectoryLength int           // Minimum points before analysis
}

// TrajectoryData stores trajectory information for a track
type TrajectoryData struct {
	TrackID           string
	Points            []TrajectoryPoint
	SmoothedPoints    []TrajectoryPoint
	Predictions       []TrajectoryPoint
	TotalDistance     float64
	AverageVelocity   Velocity
	MaxVelocity       Velocity
	Direction         float64 // Direction in radians
	IsStationary      bool
	LastUpdated       time.Time
	CreatedAt         time.Time
	mutex             sync.RWMutex
}

// TrajectoryStats tracks trajectory recording statistics
type TrajectoryStats struct {
	ActiveTrajectories int
	TotalTrajectories  int64
	TotalPoints        int64
	AverageLength      float64
	LongestTrajectory  int
	StartTime          time.Time
	LastUpdate         time.Time
	mutex              sync.RWMutex
}

// TrajectoryAnalysis contains analysis results for a trajectory
type TrajectoryAnalysis struct {
	TrackID           string
	TotalDistance     float64
	AverageSpeed      float64
	MaxSpeed          float64
	Direction         float64
	IsStationary      bool
	StationaryTime    time.Duration
	MovingTime        time.Duration
	TurningPoints     []TrajectoryPoint
	Predictions       []TrajectoryPoint
	BoundingBox       domain.Rectangle
	StartTime         time.Time
	EndTime           time.Time
}

// DefaultTrajectoryConfig returns default trajectory configuration
func DefaultTrajectoryConfig() TrajectoryConfig {
	return TrajectoryConfig{
		MaxTrajectoryLength: 1000,
		SamplingInterval:    100 * time.Millisecond,
		SmoothingWindow:     5,
		EnablePrediction:    true,
		PredictionHorizon:   2 * time.Second,
		MinTrajectoryLength: 3,
	}
}

// NewTrajectoryRecorder creates a new trajectory recorder
func NewTrajectoryRecorder(logger *zap.Logger, config TrajectoryConfig) *TrajectoryRecorder {
	return &TrajectoryRecorder{
		logger:       logger.With(zap.String("component", "trajectory_recorder")),
		trajectories: make(map[string]*TrajectoryData),
		config:       config,
		stats: &TrajectoryStats{
			StartTime: time.Now(),
		},
	}
}

// RecordPoint records a new trajectory point for a track
func (tr *TrajectoryRecorder) RecordPoint(trackID string, point TrajectoryPoint) {
	tr.mutex.Lock()
	defer tr.mutex.Unlock()

	// Get or create trajectory data
	trajectory, exists := tr.trajectories[trackID]
	if !exists {
		trajectory = &TrajectoryData{
			TrackID:   trackID,
			Points:    make([]TrajectoryPoint, 0),
			CreatedAt: time.Now(),
		}
		tr.trajectories[trackID] = trajectory
		tr.stats.TotalTrajectories++
	}

	trajectory.mutex.Lock()
	defer trajectory.mutex.Unlock()

	// Check sampling interval
	if len(trajectory.Points) > 0 {
		lastPoint := trajectory.Points[len(trajectory.Points)-1]
		if point.Timestamp.Sub(lastPoint.Timestamp) < tr.config.SamplingInterval {
			return // Skip this point due to sampling interval
		}
	}

	// Add point to trajectory
	trajectory.Points = append(trajectory.Points, point)
	trajectory.LastUpdated = time.Now()
	tr.stats.TotalPoints++

	// Limit trajectory length
	if len(trajectory.Points) > tr.config.MaxTrajectoryLength {
		trajectory.Points = trajectory.Points[1:]
	}

	// Update trajectory analysis
	tr.updateTrajectoryAnalysis(trajectory)

	// Smooth trajectory if enough points
	if len(trajectory.Points) >= tr.config.SmoothingWindow {
		tr.smoothTrajectory(trajectory)
	}

	// Generate predictions if enabled
	if tr.config.EnablePrediction && len(trajectory.Points) >= tr.config.MinTrajectoryLength {
		tr.predictTrajectory(trajectory)
	}

	tr.logger.Debug("Trajectory point recorded",
		zap.String("track_id", trackID),
		zap.Int("total_points", len(trajectory.Points)),
		zap.Float64("x", point.X),
		zap.Float64("y", point.Y))
}

// GetTrajectory returns trajectory data for a track
func (tr *TrajectoryRecorder) GetTrajectory(trackID string) (*TrajectoryData, bool) {
	tr.mutex.RLock()
	defer tr.mutex.RUnlock()

	trajectory, exists := tr.trajectories[trackID]
	if !exists {
		return nil, false
	}

	// Return a copy to avoid race conditions
	trajectory.mutex.RLock()
	defer trajectory.mutex.RUnlock()

	trajectoryCopy := &TrajectoryData{
		TrackID:         trajectory.TrackID,
		Points:          make([]TrajectoryPoint, len(trajectory.Points)),
		SmoothedPoints:  make([]TrajectoryPoint, len(trajectory.SmoothedPoints)),
		Predictions:     make([]TrajectoryPoint, len(trajectory.Predictions)),
		TotalDistance:   trajectory.TotalDistance,
		AverageVelocity: trajectory.AverageVelocity,
		MaxVelocity:     trajectory.MaxVelocity,
		Direction:       trajectory.Direction,
		IsStationary:    trajectory.IsStationary,
		LastUpdated:     trajectory.LastUpdated,
		CreatedAt:       trajectory.CreatedAt,
	}

	copy(trajectoryCopy.Points, trajectory.Points)
	copy(trajectoryCopy.SmoothedPoints, trajectory.SmoothedPoints)
	copy(trajectoryCopy.Predictions, trajectory.Predictions)

	return trajectoryCopy, true
}

// AnalyzeTrajectory performs comprehensive trajectory analysis
func (tr *TrajectoryRecorder) AnalyzeTrajectory(trackID string) (*TrajectoryAnalysis, error) {
	trajectory, exists := tr.GetTrajectory(trackID)
	if !exists {
		return nil, fmt.Errorf("trajectory not found for track: %s", trackID)
	}

	if len(trajectory.Points) < tr.config.MinTrajectoryLength {
		return nil, fmt.Errorf("insufficient trajectory points for analysis: %d", len(trajectory.Points))
	}

	analysis := &TrajectoryAnalysis{
		TrackID:   trackID,
		StartTime: trajectory.Points[0].Timestamp,
		EndTime:   trajectory.Points[len(trajectory.Points)-1].Timestamp,
	}

	// Calculate total distance and speeds
	tr.calculateDistanceAndSpeed(trajectory, analysis)

	// Calculate direction
	tr.calculateDirection(trajectory, analysis)

	// Detect stationary periods
	tr.detectStationaryPeriods(trajectory, analysis)

	// Find turning points
	tr.findTurningPoints(trajectory, analysis)

	// Calculate bounding box
	tr.calculateBoundingBox(trajectory, analysis)

	// Add predictions
	analysis.Predictions = trajectory.Predictions

	return analysis, nil
}

// RemoveTrajectory removes trajectory data for a track
func (tr *TrajectoryRecorder) RemoveTrajectory(trackID string) {
	tr.mutex.Lock()
	defer tr.mutex.Unlock()

	if trajectory, exists := tr.trajectories[trackID]; exists {
		delete(tr.trajectories, trackID)
		tr.logger.Debug("Trajectory removed",
			zap.String("track_id", trackID),
			zap.Int("points", len(trajectory.Points)))
	}
}

// GetActiveTrajectories returns all active trajectory IDs
func (tr *TrajectoryRecorder) GetActiveTrajectories() []string {
	tr.mutex.RLock()
	defer tr.mutex.RUnlock()

	trajectories := make([]string, 0, len(tr.trajectories))
	for trackID := range tr.trajectories {
		trajectories = append(trajectories, trackID)
	}
	return trajectories
}

// GetStats returns trajectory recording statistics
func (tr *TrajectoryRecorder) GetStats() *TrajectoryStats {
	tr.stats.mutex.RLock()
	defer tr.stats.mutex.RUnlock()

	tr.mutex.RLock()
	activeTrajectories := len(tr.trajectories)
	longestTrajectory := 0
	totalPoints := int64(0)

	for _, trajectory := range tr.trajectories {
		trajectory.mutex.RLock()
		pointCount := len(trajectory.Points)
		totalPoints += int64(pointCount)
		if pointCount > longestTrajectory {
			longestTrajectory = pointCount
		}
		trajectory.mutex.RUnlock()
	}
	tr.mutex.RUnlock()

	averageLength := 0.0
	if activeTrajectories > 0 {
		averageLength = float64(totalPoints) / float64(activeTrajectories)
	}

	return &TrajectoryStats{
		ActiveTrajectories: activeTrajectories,
		TotalTrajectories:  tr.stats.TotalTrajectories,
		TotalPoints:        tr.stats.TotalPoints,
		AverageLength:      averageLength,
		LongestTrajectory:  longestTrajectory,
		StartTime:          tr.stats.StartTime,
		LastUpdate:         time.Now(),
	}
}

// updateTrajectoryAnalysis updates basic trajectory analysis
func (tr *TrajectoryRecorder) updateTrajectoryAnalysis(trajectory *TrajectoryData) {
	if len(trajectory.Points) < 2 {
		return
	}

	// Calculate total distance
	totalDistance := 0.0
	velocities := make([]Velocity, 0)

	for i := 1; i < len(trajectory.Points); i++ {
		prev := trajectory.Points[i-1]
		curr := trajectory.Points[i]

		// Calculate distance
		dx := curr.X - prev.X
		dy := curr.Y - prev.Y
		distance := math.Sqrt(dx*dx + dy*dy)
		totalDistance += distance

		// Calculate velocity
		dt := curr.Timestamp.Sub(prev.Timestamp).Seconds()
		if dt > 0 {
			velocity := Velocity{
				VX: dx / dt,
				VY: dy / dt,
			}
			velocities = append(velocities, velocity)
		}
	}

	trajectory.TotalDistance = totalDistance

	// Calculate average and max velocity
	if len(velocities) > 0 {
		var sumVX, sumVY, maxSpeed float64
		for _, v := range velocities {
			sumVX += v.VX
			sumVY += v.VY
			speed := math.Sqrt(v.VX*v.VX + v.VY*v.VY)
			if speed > maxSpeed {
				maxSpeed = speed
				trajectory.MaxVelocity = v
			}
		}

		trajectory.AverageVelocity = Velocity{
			VX: sumVX / float64(len(velocities)),
			VY: sumVY / float64(len(velocities)),
		}

		// Check if stationary
		avgSpeed := math.Sqrt(trajectory.AverageVelocity.VX*trajectory.AverageVelocity.VX +
			trajectory.AverageVelocity.VY*trajectory.AverageVelocity.VY)
		trajectory.IsStationary = avgSpeed < 1.0 // Less than 1 pixel/second
	}
}

// smoothTrajectory applies smoothing to trajectory points
func (tr *TrajectoryRecorder) smoothTrajectory(trajectory *TrajectoryData) {
	if len(trajectory.Points) < tr.config.SmoothingWindow {
		return
	}

	smoothed := make([]TrajectoryPoint, 0, len(trajectory.Points))
	halfWindow := tr.config.SmoothingWindow / 2

	for i := 0; i < len(trajectory.Points); i++ {
		start := max(0, i-halfWindow)
		end := min(len(trajectory.Points), i+halfWindow+1)

		var sumX, sumY float64
		count := 0

		for j := start; j < end; j++ {
			sumX += trajectory.Points[j].X
			sumY += trajectory.Points[j].Y
			count++
		}

		smoothedPoint := TrajectoryPoint{
			X:         sumX / float64(count),
			Y:         sumY / float64(count),
			Timestamp: trajectory.Points[i].Timestamp,
			FrameID:   trajectory.Points[i].FrameID,
			Velocity:  trajectory.Points[i].Velocity,
		}

		smoothed = append(smoothed, smoothedPoint)
	}

	trajectory.SmoothedPoints = smoothed
}

// predictTrajectory generates trajectory predictions
func (tr *TrajectoryRecorder) predictTrajectory(trajectory *TrajectoryData) {
	if len(trajectory.Points) < 2 {
		return
	}

	// Use last few points to estimate velocity
	recentPoints := 3
	if len(trajectory.Points) < recentPoints {
		recentPoints = len(trajectory.Points)
	}

	startIdx := len(trajectory.Points) - recentPoints
	lastPoint := trajectory.Points[len(trajectory.Points)-1]

	// Calculate average velocity from recent points
	var avgVX, avgVY float64
	count := 0

	for i := startIdx; i < len(trajectory.Points)-1; i++ {
		curr := trajectory.Points[i]
		next := trajectory.Points[i+1]
		dt := next.Timestamp.Sub(curr.Timestamp).Seconds()

		if dt > 0 {
			avgVX += (next.X - curr.X) / dt
			avgVY += (next.Y - curr.Y) / dt
			count++
		}
	}

	if count > 0 {
		avgVX /= float64(count)
		avgVY /= float64(count)

		// Generate predictions
		predictions := make([]TrajectoryPoint, 0)
		predictionSteps := 10
		stepDuration := tr.config.PredictionHorizon / time.Duration(predictionSteps)

		for i := 1; i <= predictionSteps; i++ {
			dt := float64(i) * stepDuration.Seconds()
			predX := lastPoint.X + avgVX*dt
			predY := lastPoint.Y + avgVY*dt

			prediction := TrajectoryPoint{
				X:         predX,
				Y:         predY,
				Timestamp: lastPoint.Timestamp.Add(time.Duration(i) * stepDuration),
				FrameID:   fmt.Sprintf("pred_%d", i),
				Velocity:  Velocity{VX: avgVX, VY: avgVY},
			}

			predictions = append(predictions, prediction)
		}

		trajectory.Predictions = predictions
	}
}

// Helper functions for trajectory analysis
func (tr *TrajectoryRecorder) calculateDistanceAndSpeed(trajectory *TrajectoryData, analysis *TrajectoryAnalysis) {
	// Implementation would calculate detailed distance and speed metrics
	analysis.TotalDistance = trajectory.TotalDistance
	analysis.AverageSpeed = math.Sqrt(trajectory.AverageVelocity.VX*trajectory.AverageVelocity.VX +
		trajectory.AverageVelocity.VY*trajectory.AverageVelocity.VY)
	analysis.MaxSpeed = math.Sqrt(trajectory.MaxVelocity.VX*trajectory.MaxVelocity.VX +
		trajectory.MaxVelocity.VY*trajectory.MaxVelocity.VY)
}

func (tr *TrajectoryRecorder) calculateDirection(trajectory *TrajectoryData, analysis *TrajectoryAnalysis) {
	if len(trajectory.Points) < 2 {
		return
	}

	first := trajectory.Points[0]
	last := trajectory.Points[len(trajectory.Points)-1]

	dx := last.X - first.X
	dy := last.Y - first.Y

	analysis.Direction = math.Atan2(dy, dx)
}

func (tr *TrajectoryRecorder) detectStationaryPeriods(trajectory *TrajectoryData, analysis *TrajectoryAnalysis) {
	analysis.IsStationary = trajectory.IsStationary
	// Additional stationary period detection logic would go here
}

func (tr *TrajectoryRecorder) findTurningPoints(trajectory *TrajectoryData, analysis *TrajectoryAnalysis) {
	// Implementation would detect significant direction changes
	analysis.TurningPoints = make([]TrajectoryPoint, 0)
}

func (tr *TrajectoryRecorder) calculateBoundingBox(trajectory *TrajectoryData, analysis *TrajectoryAnalysis) {
	if len(trajectory.Points) == 0 {
		return
	}

	minX, maxX := trajectory.Points[0].X, trajectory.Points[0].X
	minY, maxY := trajectory.Points[0].Y, trajectory.Points[0].Y

	for _, point := range trajectory.Points {
		if point.X < minX {
			minX = point.X
		}
		if point.X > maxX {
			maxX = point.X
		}
		if point.Y < minY {
			minY = point.Y
		}
		if point.Y > maxY {
			maxY = point.Y
		}
	}

	analysis.BoundingBox = domain.Rectangle{
		X:      int(minX),
		Y:      int(minY),
		Width:  int(maxX - minX),
		Height: int(maxY - minY),
	}
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Close closes the trajectory recorder and cleans up resources
func (tr *TrajectoryRecorder) Close() error {
	tr.mutex.Lock()
	defer tr.mutex.Unlock()

	tr.logger.Info("Closing trajectory recorder",
		zap.Int("active_trajectories", len(tr.trajectories)))

	// Clear all trajectories
	tr.trajectories = make(map[string]*TrajectoryData)

	tr.logger.Info("Trajectory recorder closed")
	return nil
}
