package tracking

import (
	"math"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
)

// KalmanFilter implements a Kalman filter for object tracking
type KalmanFilter struct {
	// State vector: [x, y, vx, vy, w, h, vw, vh]
	// x, y: center coordinates
	// vx, vy: velocities
	// w, h: width, height
	// vw, vh: width/height change rates
	state      []float64 // State vector (8x1)
	covariance [][]float64 // Covariance matrix (8x8)
	
	// Process noise
	processNoise [][]float64 // Process noise covariance (8x8)
	
	// Measurement noise
	measurementNoise [][]float64 // Measurement noise covariance (4x4)
	
	// Transition matrix
	transitionMatrix [][]float64 // State transition matrix (8x8)
	
	// Observation matrix
	observationMatrix [][]float64 // Observation matrix (4x8)
	
	// Time step
	dt float64
	
	// Initialization flag
	initialized bool
}

// KalmanConfig configures the Kalman filter
type KalmanConfig struct {
	ProcessNoiseStd     float64 // Standard deviation of process noise
	MeasurementNoiseStd float64 // Standard deviation of measurement noise
	InitialCovariance   float64 // Initial covariance value
	TimeStep            float64 // Time step between predictions
}

// DefaultKalmanConfig returns default Kalman filter configuration
func DefaultKalmanConfig() KalmanConfig {
	return KalmanConfig{
		ProcessNoiseStd:     1.0,
		MeasurementNoiseStd: 10.0,
		InitialCovariance:   1000.0,
		TimeStep:            1.0, // 1 frame
	}
}

// NewKalmanFilter creates a new Kalman filter
func NewKalmanFilter(config KalmanConfig) *KalmanFilter {
	kf := &KalmanFilter{
		state:      make([]float64, 8),
		covariance: make([][]float64, 8),
		dt:         config.TimeStep,
	}
	
	// Initialize matrices
	for i := 0; i < 8; i++ {
		kf.covariance[i] = make([]float64, 8)
	}
	
	kf.initializeMatrices(config)
	
	return kf
}

// Initialize initializes the Kalman filter with the first detection
func (kf *KalmanFilter) Initialize(detection domain.DetectedObject) {
	// Convert bounding box to center coordinates
	centerX := float64(detection.BoundingBox.X) + float64(detection.BoundingBox.Width)/2
	centerY := float64(detection.BoundingBox.Y) + float64(detection.BoundingBox.Height)/2
	width := float64(detection.BoundingBox.Width)
	height := float64(detection.BoundingBox.Height)
	
	// Initialize state: [x, y, vx, vy, w, h, vw, vh]
	kf.state[0] = centerX
	kf.state[1] = centerY
	kf.state[2] = 0.0 // Initial velocity x
	kf.state[3] = 0.0 // Initial velocity y
	kf.state[4] = width
	kf.state[5] = height
	kf.state[6] = 0.0 // Initial width change rate
	kf.state[7] = 0.0 // Initial height change rate
	
	kf.initialized = true
}

// Predict predicts the next state
func (kf *KalmanFilter) Predict() domain.Rectangle {
	if !kf.initialized {
		return domain.Rectangle{}
	}
	
	// Predict state: x' = F * x
	newState := kf.matrixVectorMultiply(kf.transitionMatrix, kf.state)
	
	// Predict covariance: P' = F * P * F^T + Q
	// P' = F * P * F^T
	FP := kf.matrixMultiply(kf.transitionMatrix, kf.covariance)
	FPFt := kf.matrixMultiplyTranspose(FP, kf.transitionMatrix)
	
	// P' = F * P * F^T + Q
	newCovariance := kf.matrixAdd(FPFt, kf.processNoise)
	
	// Update state and covariance
	kf.state = newState
	kf.covariance = newCovariance
	
	// Convert predicted state back to bounding box
	return kf.stateToBoundingBox()
}

// Update updates the filter with a new measurement
func (kf *KalmanFilter) Update(detection domain.DetectedObject) {
	if !kf.initialized {
		kf.Initialize(detection)
		return
	}
	
	// Convert detection to measurement vector [x, y, w, h]
	measurement := kf.detectionToMeasurement(detection)
	
	// Innovation: y = z - H * x
	Hx := kf.matrixVectorMultiply(kf.observationMatrix, kf.state)
	innovation := kf.vectorSubtract(measurement, Hx)
	
	// Innovation covariance: S = H * P * H^T + R
	HP := kf.matrixMultiply(kf.observationMatrix, kf.covariance)
	HPHt := kf.matrixMultiplyTranspose(HP, kf.observationMatrix)
	S := kf.matrixAdd(HPHt, kf.measurementNoise)
	
	// Kalman gain: K = P * H^T * S^-1
	PHt := kf.matrixMultiplyTranspose(kf.covariance, kf.observationMatrix)
	SInv := kf.matrixInverse(S)
	K := kf.matrixMultiply(PHt, SInv)
	
	// Update state: x = x + K * y
	Ky := kf.matrixVectorMultiply(K, innovation)
	kf.state = kf.vectorAdd(kf.state, Ky)
	
	// Update covariance: P = (I - K * H) * P
	KH := kf.matrixMultiply(K, kf.observationMatrix)
	I := kf.identityMatrix(8)
	IKH := kf.matrixSubtract(I, KH)
	kf.covariance = kf.matrixMultiply(IKH, kf.covariance)
}

// GetState returns the current state as a bounding box
func (kf *KalmanFilter) GetState() domain.Rectangle {
	if !kf.initialized {
		return domain.Rectangle{}
	}
	return kf.stateToBoundingBox()
}

// GetVelocity returns the current velocity
func (kf *KalmanFilter) GetVelocity() Velocity {
	if !kf.initialized {
		return Velocity{}
	}
	return Velocity{
		VX: kf.state[2],
		VY: kf.state[3],
	}
}

// IsInitialized returns whether the filter is initialized
func (kf *KalmanFilter) IsInitialized() bool {
	return kf.initialized
}

// initializeMatrices initializes the filter matrices
func (kf *KalmanFilter) initializeMatrices(config KalmanConfig) {
	// Initialize transition matrix F (8x8)
	// State: [x, y, vx, vy, w, h, vw, vh]
	kf.transitionMatrix = make([][]float64, 8)
	for i := 0; i < 8; i++ {
		kf.transitionMatrix[i] = make([]float64, 8)
	}
	
	// Identity matrix with velocity integration
	for i := 0; i < 8; i++ {
		kf.transitionMatrix[i][i] = 1.0
	}
	kf.transitionMatrix[0][2] = kf.dt // x += vx * dt
	kf.transitionMatrix[1][3] = kf.dt // y += vy * dt
	kf.transitionMatrix[4][6] = kf.dt // w += vw * dt
	kf.transitionMatrix[5][7] = kf.dt // h += vh * dt
	
	// Initialize observation matrix H (4x8)
	// Measurement: [x, y, w, h]
	kf.observationMatrix = make([][]float64, 4)
	for i := 0; i < 4; i++ {
		kf.observationMatrix[i] = make([]float64, 8)
	}
	kf.observationMatrix[0][0] = 1.0 // x
	kf.observationMatrix[1][1] = 1.0 // y
	kf.observationMatrix[2][4] = 1.0 // w
	kf.observationMatrix[3][5] = 1.0 // h
	
	// Initialize process noise Q (8x8)
	kf.processNoise = make([][]float64, 8)
	for i := 0; i < 8; i++ {
		kf.processNoise[i] = make([]float64, 8)
		kf.processNoise[i][i] = config.ProcessNoiseStd * config.ProcessNoiseStd
	}
	
	// Initialize measurement noise R (4x4)
	kf.measurementNoise = make([][]float64, 4)
	for i := 0; i < 4; i++ {
		kf.measurementNoise[i] = make([]float64, 4)
		kf.measurementNoise[i][i] = config.MeasurementNoiseStd * config.MeasurementNoiseStd
	}
	
	// Initialize covariance matrix P (8x8)
	for i := 0; i < 8; i++ {
		kf.covariance[i][i] = config.InitialCovariance
	}
}

// stateToBoundingBox converts state vector to bounding box
func (kf *KalmanFilter) stateToBoundingBox() domain.Rectangle {
	centerX := kf.state[0]
	centerY := kf.state[1]
	width := kf.state[4]
	height := kf.state[5]
	
	// Convert center coordinates to top-left coordinates
	x := centerX - width/2
	y := centerY - height/2
	
	return domain.Rectangle{
		X:      int(math.Round(x)),
		Y:      int(math.Round(y)),
		Width:  int(math.Round(width)),
		Height: int(math.Round(height)),
	}
}

// detectionToMeasurement converts detection to measurement vector
func (kf *KalmanFilter) detectionToMeasurement(detection domain.DetectedObject) []float64 {
	centerX := float64(detection.BoundingBox.X) + float64(detection.BoundingBox.Width)/2
	centerY := float64(detection.BoundingBox.Y) + float64(detection.BoundingBox.Height)/2
	width := float64(detection.BoundingBox.Width)
	height := float64(detection.BoundingBox.Height)
	
	return []float64{centerX, centerY, width, height}
}

// Matrix operations (simplified implementations)

func (kf *KalmanFilter) matrixVectorMultiply(matrix [][]float64, vector []float64) []float64 {
	rows := len(matrix)
	result := make([]float64, rows)
	
	for i := 0; i < rows; i++ {
		for j := 0; j < len(vector); j++ {
			result[i] += matrix[i][j] * vector[j]
		}
	}
	
	return result
}

func (kf *KalmanFilter) matrixMultiply(a, b [][]float64) [][]float64 {
	rows := len(a)
	cols := len(b[0])
	result := make([][]float64, rows)
	
	for i := 0; i < rows; i++ {
		result[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			for k := 0; k < len(b); k++ {
				result[i][j] += a[i][k] * b[k][j]
			}
		}
	}
	
	return result
}

func (kf *KalmanFilter) matrixMultiplyTranspose(a [][]float64, b [][]float64) [][]float64 {
	rows := len(a)
	cols := len(b)
	result := make([][]float64, rows)
	
	for i := 0; i < rows; i++ {
		result[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			for k := 0; k < len(b[0]); k++ {
				result[i][j] += a[i][k] * b[j][k]
			}
		}
	}
	
	return result
}

func (kf *KalmanFilter) matrixAdd(a, b [][]float64) [][]float64 {
	rows := len(a)
	cols := len(a[0])
	result := make([][]float64, rows)
	
	for i := 0; i < rows; i++ {
		result[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			result[i][j] = a[i][j] + b[i][j]
		}
	}
	
	return result
}

func (kf *KalmanFilter) matrixSubtract(a, b [][]float64) [][]float64 {
	rows := len(a)
	cols := len(a[0])
	result := make([][]float64, rows)
	
	for i := 0; i < rows; i++ {
		result[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			result[i][j] = a[i][j] - b[i][j]
		}
	}
	
	return result
}

func (kf *KalmanFilter) vectorAdd(a, b []float64) []float64 {
	result := make([]float64, len(a))
	for i := 0; i < len(a); i++ {
		result[i] = a[i] + b[i]
	}
	return result
}

func (kf *KalmanFilter) vectorSubtract(a, b []float64) []float64 {
	result := make([]float64, len(a))
	for i := 0; i < len(a); i++ {
		result[i] = a[i] - b[i]
	}
	return result
}

func (kf *KalmanFilter) identityMatrix(size int) [][]float64 {
	result := make([][]float64, size)
	for i := 0; i < size; i++ {
		result[i] = make([]float64, size)
		result[i][i] = 1.0
	}
	return result
}

// Simplified matrix inverse for small matrices (4x4)
func (kf *KalmanFilter) matrixInverse(matrix [][]float64) [][]float64 {
	size := len(matrix)
	
	// Create augmented matrix [A|I]
	augmented := make([][]float64, size)
	for i := 0; i < size; i++ {
		augmented[i] = make([]float64, 2*size)
		for j := 0; j < size; j++ {
			augmented[i][j] = matrix[i][j]
		}
		augmented[i][i+size] = 1.0
	}
	
	// Gaussian elimination
	for i := 0; i < size; i++ {
		// Find pivot
		pivot := augmented[i][i]
		if math.Abs(pivot) < 1e-10 {
			// Add small value to diagonal for numerical stability
			pivot = 1e-6
			augmented[i][i] = pivot
		}
		
		// Scale row
		for j := 0; j < 2*size; j++ {
			augmented[i][j] /= pivot
		}
		
		// Eliminate column
		for k := 0; k < size; k++ {
			if k != i {
				factor := augmented[k][i]
				for j := 0; j < 2*size; j++ {
					augmented[k][j] -= factor * augmented[i][j]
				}
			}
		}
	}
	
	// Extract inverse matrix
	result := make([][]float64, size)
	for i := 0; i < size; i++ {
		result[i] = make([]float64, size)
		for j := 0; j < size; j++ {
			result[i][j] = augmented[i][j+size]
		}
	}
	
	return result
}
