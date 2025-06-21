package tracking

import (
	"math"
	"sort"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"go.uber.org/zap"
)

// AssociationAlgorithm defines the algorithm used for data association
type AssociationAlgorithm int

const (
	AssociationGreedy AssociationAlgorithm = iota // Greedy nearest neighbor
	AssociationHungarian                          // Hungarian algorithm (optimal)
	AssociationGNN                                // Global Nearest Neighbor
)

// AssociationConfig configures the association algorithm
type AssociationConfig struct {
	Algorithm        AssociationAlgorithm
	IOUThreshold     float64 // IOU threshold for valid associations
	DistanceThreshold float64 // Distance threshold for valid associations
	UseIOU           bool    // Use IOU for distance calculation
	UseEuclidean     bool    // Use Euclidean distance
	UseMahalanobis   bool    // Use Mahalanobis distance (requires covariance)
	WeightIOU        float64 // Weight for IOU in combined distance
	WeightDistance   float64 // Weight for Euclidean distance in combined distance
	WeightClass      float64 // Weight for class similarity
}

// DefaultAssociationConfig returns default association configuration
func DefaultAssociationConfig() AssociationConfig {
	return AssociationConfig{
		Algorithm:         AssociationGreedy,
		IOUThreshold:      0.3,
		DistanceThreshold: 100.0,
		UseIOU:            true,
		UseEuclidean:      true,
		UseMahalanobis:    false,
		WeightIOU:         0.7,
		WeightDistance:    0.3,
		WeightClass:       0.1,
	}
}

// Associator handles detection-track association
type Associator struct {
	logger *zap.Logger
	config AssociationConfig
}

// NewAssociator creates a new associator
func NewAssociator(logger *zap.Logger, config AssociationConfig) *Associator {
	return &Associator{
		logger: logger.With(zap.String("component", "associator")),
		config: config,
	}
}

// Associate associates detections with tracks
func (a *Associator) Associate(detections []domain.DetectedObject, tracks []*Track) ([]Association, []int, []string) {
	if len(detections) == 0 || len(tracks) == 0 {
		// No associations possible
		unassociatedDetections := make([]int, len(detections))
		for i := range detections {
			unassociatedDetections[i] = i
		}
		
		unassociatedTracks := make([]string, len(tracks))
		for i, track := range tracks {
			unassociatedTracks[i] = track.ID
		}
		
		return []Association{}, unassociatedDetections, unassociatedTracks
	}

	// Calculate cost matrix
	costMatrix := a.calculateCostMatrix(detections, tracks)

	// Perform association based on algorithm
	var associations []Association
	switch a.config.Algorithm {
	case AssociationGreedy:
		associations = a.greedyAssociation(costMatrix, detections, tracks)
	case AssociationHungarian:
		associations = a.hungarianAssociation(costMatrix, detections, tracks)
	case AssociationGNN:
		associations = a.gnnAssociation(costMatrix, detections, tracks)
	default:
		associations = a.greedyAssociation(costMatrix, detections, tracks)
	}

	// Filter associations by thresholds
	validAssociations := a.filterAssociations(associations, detections, tracks)

	// Find unassociated detections and tracks
	unassociatedDetections, unassociatedTracks := a.findUnassociated(validAssociations, detections, tracks)

	a.logger.Debug("Association completed",
		zap.Int("detections", len(detections)),
		zap.Int("tracks", len(tracks)),
		zap.Int("associations", len(validAssociations)),
		zap.Int("unassociated_detections", len(unassociatedDetections)),
		zap.Int("unassociated_tracks", len(unassociatedTracks)))

	return validAssociations, unassociatedDetections, unassociatedTracks
}

// calculateCostMatrix calculates the cost matrix for association
func (a *Associator) calculateCostMatrix(detections []domain.DetectedObject, tracks []*Track) [][]float64 {
	rows := len(detections)
	cols := len(tracks)
	costMatrix := make([][]float64, rows)

	for i := 0; i < rows; i++ {
		costMatrix[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			costMatrix[i][j] = a.calculateDistance(detections[i], tracks[j])
		}
	}

	return costMatrix
}

// calculateDistance calculates the distance between a detection and a track
func (a *Associator) calculateDistance(detection domain.DetectedObject, track *Track) float64 {
	var totalDistance float64
	var weightSum float64

	// Get predicted position from track
	var trackBox domain.Rectangle
	if track.KalmanFilter != nil && track.KalmanFilter.IsInitialized() {
		trackBox = track.KalmanFilter.GetState()
	} else if track.LastDetection != nil {
		trackBox = track.LastDetection.BoundingBox
	} else {
		return math.Inf(1) // Invalid track
	}

	// Calculate IOU distance
	if a.config.UseIOU && a.config.WeightIOU > 0 {
		iou := a.calculateIOU(detection.BoundingBox, trackBox)
		iouDistance := 1.0 - iou // Convert IOU to distance (higher IOU = lower distance)
		totalDistance += iouDistance * a.config.WeightIOU
		weightSum += a.config.WeightIOU
	}

	// Calculate Euclidean distance
	if a.config.UseEuclidean && a.config.WeightDistance > 0 {
		euclideanDistance := a.calculateEuclideanDistance(detection.BoundingBox, trackBox)
		// Normalize distance by image size (assuming max distance of 1000 pixels)
		normalizedDistance := euclideanDistance / 1000.0
		totalDistance += normalizedDistance * a.config.WeightDistance
		weightSum += a.config.WeightDistance
	}

	// Calculate class similarity
	if a.config.WeightClass > 0 {
		classSimilarity := a.calculateClassSimilarity(detection.Class, track.Class)
		classDistance := 1.0 - classSimilarity
		totalDistance += classDistance * a.config.WeightClass
		weightSum += a.config.WeightClass
	}

	// Calculate Mahalanobis distance if enabled and available
	if a.config.UseMahalanobis && track.KalmanFilter != nil && track.KalmanFilter.IsInitialized() {
		// TODO: Implement Mahalanobis distance using covariance matrix
		// For now, skip this component
	}

	// Normalize by total weight
	if weightSum > 0 {
		totalDistance /= weightSum
	}

	return totalDistance
}

// calculateIOU calculates Intersection over Union between two bounding boxes
func (a *Associator) calculateIOU(box1, box2 domain.Rectangle) float64 {
	// Calculate intersection coordinates
	x1 := math.Max(float64(box1.X), float64(box2.X))
	y1 := math.Max(float64(box1.Y), float64(box2.Y))
	x2 := math.Min(float64(box1.X+box1.Width), float64(box2.X+box2.Width))
	y2 := math.Min(float64(box1.Y+box1.Height), float64(box2.Y+box2.Height))

	// Check if there's no intersection
	if x2 <= x1 || y2 <= y1 {
		return 0.0
	}

	// Calculate intersection area
	intersection := (x2 - x1) * (y2 - y1)

	// Calculate union area
	area1 := float64(box1.Width * box1.Height)
	area2 := float64(box2.Width * box2.Height)
	union := area1 + area2 - intersection

	if union <= 0 {
		return 0.0
	}

	return intersection / union
}

// calculateEuclideanDistance calculates Euclidean distance between box centers
func (a *Associator) calculateEuclideanDistance(box1, box2 domain.Rectangle) float64 {
	center1X := float64(box1.X) + float64(box1.Width)/2
	center1Y := float64(box1.Y) + float64(box1.Height)/2
	center2X := float64(box2.X) + float64(box2.Width)/2
	center2Y := float64(box2.Y) + float64(box2.Height)/2

	dx := center1X - center2X
	dy := center1Y - center2Y

	return math.Sqrt(dx*dx + dy*dy)
}

// calculateClassSimilarity calculates similarity between two classes
func (a *Associator) calculateClassSimilarity(class1, class2 string) float64 {
	if class1 == class2 {
		return 1.0
	}
	return 0.0 // Could be extended to handle class hierarchies
}

// greedyAssociation performs greedy nearest neighbor association
func (a *Associator) greedyAssociation(costMatrix [][]float64, detections []domain.DetectedObject, tracks []*Track) []Association {
	var associations []Association
	usedDetections := make(map[int]bool)
	usedTracks := make(map[int]bool)

	// Create list of all possible associations with their costs
	type costEntry struct {
		detectionIdx int
		trackIdx     int
		cost         float64
	}

	var costs []costEntry
	for i := 0; i < len(costMatrix); i++ {
		for j := 0; j < len(costMatrix[i]); j++ {
			costs = append(costs, costEntry{
				detectionIdx: i,
				trackIdx:     j,
				cost:         costMatrix[i][j],
			})
		}
	}

	// Sort by cost (ascending)
	sort.Slice(costs, func(i, j int) bool {
		return costs[i].cost < costs[j].cost
	})

	// Greedily assign associations
	for _, entry := range costs {
		if !usedDetections[entry.detectionIdx] && !usedTracks[entry.trackIdx] {
			iou := a.calculateIOU(detections[entry.detectionIdx].BoundingBox, 
				tracks[entry.trackIdx].KalmanFilter.GetState())
			
			associations = append(associations, Association{
				DetectionIndex: entry.detectionIdx,
				TrackID:        tracks[entry.trackIdx].ID,
				Distance:       entry.cost,
				IOU:            iou,
				Confidence:     detections[entry.detectionIdx].Confidence,
			})

			usedDetections[entry.detectionIdx] = true
			usedTracks[entry.trackIdx] = true
		}
	}

	return associations
}

// hungarianAssociation performs Hungarian algorithm association (simplified implementation)
func (a *Associator) hungarianAssociation(costMatrix [][]float64, detections []domain.DetectedObject, tracks []*Track) []Association {
	// For now, use greedy as a placeholder
	// A full Hungarian algorithm implementation would be more complex
	return a.greedyAssociation(costMatrix, detections, tracks)
}

// gnnAssociation performs Global Nearest Neighbor association
func (a *Associator) gnnAssociation(costMatrix [][]float64, detections []domain.DetectedObject, tracks []*Track) []Association {
	// For now, use greedy as a placeholder
	// GNN would consider global optimization
	return a.greedyAssociation(costMatrix, detections, tracks)
}

// filterAssociations filters associations by thresholds
func (a *Associator) filterAssociations(associations []Association, detections []domain.DetectedObject, tracks []*Track) []Association {
	var validAssociations []Association

	for _, assoc := range associations {
		valid := true

		// Check IOU threshold
		if a.config.UseIOU && assoc.IOU < a.config.IOUThreshold {
			valid = false
		}

		// Check distance threshold
		if valid && assoc.Distance > a.config.DistanceThreshold {
			valid = false
		}

		if valid {
			validAssociations = append(validAssociations, assoc)
		}
	}

	return validAssociations
}

// findUnassociated finds unassociated detections and tracks
func (a *Associator) findUnassociated(associations []Association, detections []domain.DetectedObject, tracks []*Track) ([]int, []string) {
	// Find unassociated detections
	associatedDetections := make(map[int]bool)
	for _, assoc := range associations {
		associatedDetections[assoc.DetectionIndex] = true
	}

	var unassociatedDetections []int
	for i := range detections {
		if !associatedDetections[i] {
			unassociatedDetections = append(unassociatedDetections, i)
		}
	}

	// Find unassociated tracks
	associatedTracks := make(map[string]bool)
	for _, assoc := range associations {
		associatedTracks[assoc.TrackID] = true
	}

	var unassociatedTracks []string
	for _, track := range tracks {
		if !associatedTracks[track.ID] {
			unassociatedTracks = append(unassociatedTracks, track.ID)
		}
	}

	return unassociatedDetections, unassociatedTracks
}

// UpdateConfig updates the association configuration
func (a *Associator) UpdateConfig(config AssociationConfig) {
	a.config = config
	a.logger.Info("Association configuration updated",
		zap.String("algorithm", config.Algorithm.String()),
		zap.Float64("iou_threshold", config.IOUThreshold),
		zap.Float64("distance_threshold", config.DistanceThreshold))
}

// String returns string representation of association algorithm
func (aa AssociationAlgorithm) String() string {
	switch aa {
	case AssociationGreedy:
		return "greedy"
	case AssociationHungarian:
		return "hungarian"
	case AssociationGNN:
		return "gnn"
	default:
		return "unknown"
	}
}
