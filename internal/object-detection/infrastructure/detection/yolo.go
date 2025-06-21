//go:build opencv
// +build opencv

package detection

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sort"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"github.com/google/uuid"
	"github.com/yalue/onnxruntime_go"
	"go.uber.org/zap"
	"gocv.io/x/gocv"
)

// YOLODetector implements YOLO-specific object detection
type YOLODetector struct {
	*Detector
	inputSize    int
	anchors      [][]float32
	strides      []int
	outputFormat YOLOOutputFormat
}

// YOLOOutputFormat defines the output format of YOLO models
type YOLOOutputFormat int

const (
	YOLOv5Format YOLOOutputFormat = iota // [batch, 25200, 85] format
	YOLOv8Format                         // [batch, 84, 8400] format
)

// YOLOConfig extends DetectorConfig with YOLO-specific settings
type YOLOConfig struct {
	DetectorConfig
	InputSize    int
	OutputFormat YOLOOutputFormat
	Anchors      [][]float32
	Strides      []int
}

// DefaultYOLOv5Config returns default YOLOv5 configuration
func DefaultYOLOv5Config() YOLOConfig {
	return YOLOConfig{
		DetectorConfig: DefaultDetectorConfig(),
		InputSize:      640,
		OutputFormat:   YOLOv5Format,
		Strides:        []int{8, 16, 32},
		Anchors: [][]float32{
			{10, 13, 16, 30, 33, 23},     // P3/8
			{30, 61, 62, 45, 59, 119},    // P4/16
			{116, 90, 156, 198, 373, 326}, // P5/32
		},
	}
}

// DefaultYOLOv8Config returns default YOLOv8 configuration
func DefaultYOLOv8Config() YOLOConfig {
	return YOLOConfig{
		DetectorConfig: DefaultDetectorConfig(),
		InputSize:      640,
		OutputFormat:   YOLOv8Format,
		Strides:        []int{8, 16, 32},
	}
}

// NewYOLODetector creates a new YOLO detector
func NewYOLODetector(logger *zap.Logger, config YOLOConfig) *YOLODetector {
	detector := NewDetector(logger, config.DetectorConfig)
	
	return &YOLODetector{
		Detector:     detector,
		inputSize:    config.InputSize,
		anchors:      config.Anchors,
		strides:      config.Strides,
		outputFormat: config.OutputFormat,
	}
}

// preprocessImage preprocesses an image for YOLO inference
func (d *Detector) preprocessImage(mat *gocv.Mat) (*onnxruntime_go.Tensor[float32], []int, error) {
	if mat.Empty() {
		return nil, nil, fmt.Errorf("input image is empty")
	}

	originalSize := []int{mat.Rows(), mat.Cols()}

	// Convert BGR to RGB
	rgbMat := gocv.NewMat()
	defer rgbMat.Close()
	gocv.CvtColor(*mat, &rgbMat, gocv.ColorBGRToRGB)

	// Resize with padding to maintain aspect ratio
	inputSize := int(d.inputShape[2]) // Assuming NCHW format
	resizedMat, scale, padX, padY := d.resizeWithPadding(&rgbMat, inputSize)
	defer resizedMat.Close()

	// Convert to float32 and normalize to [0, 1]
	floatMat := gocv.NewMat()
	defer floatMat.Close()
	resizedMat.ConvertTo(&floatMat, gocv.MatTypeCV32F)
	floatMat.DivideFloat(255.0)

	// Convert to tensor format (NCHW)
	tensorData, err := d.matToTensorData(&floatMat)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert mat to tensor: %w", err)
	}

	// Create input tensor
	inputTensor, err := onnxruntime_go.NewTensor(d.inputShape, tensorData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create input tensor: %w", err)
	}

	// Store preprocessing info for post-processing
	preprocessInfo := []int{originalSize[0], originalSize[1], int(scale * 1000), padX, padY}

	return inputTensor, preprocessInfo, nil
}

// resizeWithPadding resizes image while maintaining aspect ratio with padding
func (d *Detector) resizeWithPadding(mat *gocv.Mat, targetSize int) (*gocv.Mat, float64, int, int) {
	height := mat.Rows()
	width := mat.Cols()

	// Calculate scale factor
	scale := math.Min(float64(targetSize)/float64(width), float64(targetSize)/float64(height))
	
	newWidth := int(float64(width) * scale)
	newHeight := int(float64(height) * scale)

	// Resize image
	resized := gocv.NewMat()
	gocv.Resize(*mat, &resized, image.Point{X: newWidth, Y: newHeight}, 0, 0, gocv.InterpolationLinear)

	// Calculate padding
	padX := (targetSize - newWidth) / 2
	padY := (targetSize - newHeight) / 2

	// Add padding
	padded := gocv.NewMat()
	gocv.CopyMakeBorder(resized, &padded, padY, targetSize-newHeight-padY, padX, targetSize-newWidth-padX, 
		gocv.BorderConstant, color.RGBA{R: 114, G: 114, B: 114, A: 0})
	
	resized.Close()
	return &padded, scale, padX, padY
}

// matToTensorData converts OpenCV Mat to tensor data in NCHW format
func (d *Detector) matToTensorData(mat *gocv.Mat) ([]float32, error) {
	if mat.Channels() != 3 {
		return nil, fmt.Errorf("expected 3-channel image, got %d", mat.Channels())
	}

	height := mat.Rows()
	width := mat.Cols()
	channels := mat.Channels()

	// Split channels
	channelMats := gocv.Split(*mat)
	defer func() {
		for _, ch := range channelMats {
			ch.Close()
		}
	}()

	// Convert to NCHW format: [batch, channels, height, width]
	tensorData := make([]float32, 1*channels*height*width)
	
	for c := 0; c < channels; c++ {
		channelData, err := channelMats[c].DataPtrFloat32()
		if err != nil {
			return nil, fmt.Errorf("failed to get channel data: %w", err)
		}

		offset := c * height * width
		copy(tensorData[offset:offset+height*width], channelData)
	}

	return tensorData, nil
}

// runInference runs model inference
func (d *Detector) runInference(inputTensor *onnxruntime_go.Tensor[float32]) ([]*onnxruntime_go.Tensor[float32], error) {
	// For now, create a mock output since the API is not stable
	// TODO: Implement proper inference when ONNX Runtime Go API is stabilized
	
	// Create a mock output tensor with typical YOLO output shape
	outputShape := []int64{1, 25200, 85} // Typical YOLOv5 output
	outputData := make([]float32, 1*25200*85)
	
	// Fill with random low confidence values to simulate no detections
	for i := range outputData {
		if i%85 == 4 { // confidence score position
			outputData[i] = 0.1 // Low confidence
		}
	}
	
	outputTensor, err := onnxruntime_go.NewTensor(outputShape, outputData)
	if err != nil {
		return nil, fmt.Errorf("failed to create mock output tensor: %w", err)
	}
	
	return []*onnxruntime_go.Tensor[float32]{outputTensor}, nil
}

// postprocessResults processes YOLO model outputs to extract detections
func (d *Detector) postprocessResults(outputs []*onnxruntime_go.Tensor[float32], preprocessInfo []int) ([]domain.DetectedObject, error) {
	if len(outputs) == 0 {
		return nil, fmt.Errorf("no model outputs")
	}

	// Get output data
	outputData := outputs[0].GetData()
	outputShape := outputs[0].GetShape()

	// Parse detections based on YOLO format
	var detections []Detection
	var err error

	// Check if this detector has YOLO-specific methods by checking if it has the outputFormat field
	// We'll default to YOLOv5 format for now
	detections, err = d.parseYOLOv5Output(outputData, outputShape)

	if err != nil {
		return nil, fmt.Errorf("failed to parse YOLO output: %w", err)
	}

	// Apply confidence filtering
	filteredDetections := d.filterByConfidence(detections)

	// Apply Non-Maximum Suppression
	nmsDetections := d.applyNMS(filteredDetections)

	// Convert to domain objects and scale back to original image coordinates
	domainDetections := d.convertToDetectedObjects(nmsDetections, preprocessInfo)

	return domainDetections, nil
}

// Detection represents a raw detection from YOLO
type Detection struct {
	X          float32
	Y          float32
	Width      float32
	Height     float32
	Confidence float32
	ClassID    int
	ClassProb  float32
}

// parseYOLOv5Output parses YOLOv5 output format [batch, 25200, 85]
func (d *Detector) parseYOLOv5Output(data []float32, shape []int64) ([]Detection, error) {
	if len(shape) != 3 {
		return nil, fmt.Errorf("expected 3D output shape, got %dD", len(shape))
	}

	batch := int(shape[0])
	numDetections := int(shape[1])
	numClasses := int(shape[2]) - 5 // 5 = x, y, w, h, confidence

	var detections []Detection

	for b := 0; b < batch; b++ {
		for i := 0; i < numDetections; i++ {
			offset := b*numDetections*(numClasses+5) + i*(numClasses+5)

			x := data[offset+0]
			y := data[offset+1]
			w := data[offset+2]
			h := data[offset+3]
			confidence := data[offset+4]

			if confidence < float32(d.confidenceThreshold) {
				continue
			}

			// Find best class
			maxClassProb := float32(0)
			bestClassID := 0

			for c := 0; c < numClasses; c++ {
				classProb := data[offset+5+c]
				if classProb > maxClassProb {
					maxClassProb = classProb
					bestClassID = c
				}
			}

			finalConfidence := confidence * maxClassProb
			if finalConfidence < float32(d.confidenceThreshold) {
				continue
			}

			detections = append(detections, Detection{
				X:          x,
				Y:          y,
				Width:      w,
				Height:     h,
				Confidence: finalConfidence,
				ClassID:    bestClassID,
				ClassProb:  maxClassProb,
			})
		}
	}

	return detections, nil
}

// parseYOLOv8Output parses YOLOv8 output format [batch, 84, 8400]
func (d *Detector) parseYOLOv8Output(data []float32, shape []int64) ([]Detection, error) {
	if len(shape) != 3 {
		return nil, fmt.Errorf("expected 3D output shape, got %dD", len(shape))
	}

	batch := int(shape[0])
	numFeatures := int(shape[1]) // 84 = 4 (bbox) + 80 (classes)
	numDetections := int(shape[2])
	numClasses := numFeatures - 4

	var detections []Detection

	for b := 0; b < batch; b++ {
		for i := 0; i < numDetections; i++ {
			// Get bbox coordinates
			x := data[b*numFeatures*numDetections + 0*numDetections + i]
			y := data[b*numFeatures*numDetections + 1*numDetections + i]
			w := data[b*numFeatures*numDetections + 2*numDetections + i]
			h := data[b*numFeatures*numDetections + 3*numDetections + i]

			// Find best class
			maxClassProb := float32(0)
			bestClassID := 0

			for c := 0; c < numClasses; c++ {
				classProb := data[b*numFeatures*numDetections + (4+c)*numDetections + i]
				if classProb > maxClassProb {
					maxClassProb = classProb
					bestClassID = c
				}
			}

			if maxClassProb < float32(d.confidenceThreshold) {
				continue
			}

			detections = append(detections, Detection{
				X:          x,
				Y:          y,
				Width:      w,
				Height:     h,
				Confidence: maxClassProb,
				ClassID:    bestClassID,
				ClassProb:  maxClassProb,
			})
		}
	}

	return detections, nil
}

// filterByConfidence filters detections by confidence threshold
func (d *Detector) filterByConfidence(detections []Detection) []Detection {
	var filtered []Detection
	for _, det := range detections {
		if det.Confidence >= float32(d.confidenceThreshold) {
			filtered = append(filtered, det)
		}
	}
	return filtered
}

// applyNMS applies Non-Maximum Suppression
func (d *Detector) applyNMS(detections []Detection) []Detection {
	if len(detections) == 0 {
		return detections
	}

	// Group detections by class
	classBuckets := make(map[int][]Detection)
	for _, det := range detections {
		classBuckets[det.ClassID] = append(classBuckets[det.ClassID], det)
	}

	var result []Detection

	// Apply NMS per class
	for _, classDetections := range classBuckets {
		nmsResult := d.nmsPerClass(classDetections)
		result = append(result, nmsResult...)
	}

	return result
}

// nmsPerClass applies NMS for a single class
func (d *Detector) nmsPerClass(detections []Detection) []Detection {
	if len(detections) <= 1 {
		return detections
	}

	// Sort by confidence (descending)
	sort.Slice(detections, func(i, j int) bool {
		return detections[i].Confidence > detections[j].Confidence
	})

	var result []Detection
	suppressed := make([]bool, len(detections))

	for i := 0; i < len(detections); i++ {
		if suppressed[i] {
			continue
		}

		result = append(result, detections[i])

		// Suppress overlapping detections
		for j := i + 1; j < len(detections); j++ {
			if suppressed[j] {
				continue
			}

			iou := d.calculateIOU(detections[i], detections[j])
			if iou > d.nmsThreshold {
				suppressed[j] = true
			}
		}
	}

	return result
}

// calculateIOU calculates Intersection over Union
func (d *Detector) calculateIOU(det1, det2 Detection) float64 {
	// Convert center coordinates to corner coordinates
	x1_min := det1.X - det1.Width/2
	y1_min := det1.Y - det1.Height/2
	x1_max := det1.X + det1.Width/2
	y1_max := det1.Y + det1.Height/2

	x2_min := det2.X - det2.Width/2
	y2_min := det2.Y - det2.Height/2
	x2_max := det2.X + det2.Width/2
	y2_max := det2.Y + det2.Height/2

	// Calculate intersection
	inter_x_min := math.Max(float64(x1_min), float64(x2_min))
	inter_y_min := math.Max(float64(y1_min), float64(y2_min))
	inter_x_max := math.Min(float64(x1_max), float64(x2_max))
	inter_y_max := math.Min(float64(y1_max), float64(y2_max))

	if inter_x_max <= inter_x_min || inter_y_max <= inter_y_min {
		return 0.0
	}

	intersection := (inter_x_max - inter_x_min) * (inter_y_max - inter_y_min)

	// Calculate union
	area1 := float64(det1.Width * det1.Height)
	area2 := float64(det2.Width * det2.Height)
	union := area1 + area2 - intersection

	if union <= 0 {
		return 0.0
	}

	return intersection / union
}

// convertToDetectedObjects converts detections to domain objects
func (d *Detector) convertToDetectedObjects(detections []Detection, preprocessInfo []int) []domain.DetectedObject {
	originalHeight := preprocessInfo[0]
	originalWidth := preprocessInfo[1]
	scale := float64(preprocessInfo[2]) / 1000.0
	padX := preprocessInfo[3]
	padY := preprocessInfo[4]

	var objects []domain.DetectedObject

	for _, det := range detections {
		// Scale back to original coordinates
		x := (float64(det.X) - float64(padX)) / scale
		y := (float64(det.Y) - float64(padY)) / scale
		width := float64(det.Width) / scale
		height := float64(det.Height) / scale

		// Convert center coordinates to top-left coordinates
		x1 := x - width/2
		y1 := y - height/2

		// Clamp to image boundaries
		x1 = math.Max(0, math.Min(x1, float64(originalWidth)))
		y1 = math.Max(0, math.Min(y1, float64(originalHeight)))
		width = math.Max(0, math.Min(width, float64(originalWidth)-x1))
		height = math.Max(0, math.Min(height, float64(originalHeight)-y1))

		className := ""
		if det.ClassID < len(d.classes) {
			className = d.classes[det.ClassID]
		}

		objects = append(objects, domain.DetectedObject{
			ID:         uuid.New().String(),
			Class:      className,
			Confidence: float64(det.Confidence),
			BoundingBox: domain.Rectangle{
				X:      int(x1),
				Y:      int(y1),
				Width:  int(width),
				Height: int(height),
			},
			Timestamp: time.Now(),
		})
	}

	return objects
}
