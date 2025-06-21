package detection

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ModelManager manages detection models
type ModelManager struct {
	logger      *zap.Logger
	storagePath string
	models      map[string]*domain.DetectionModel
	activeModel *domain.DetectionModel
	detector    *Detector
	mutex       sync.RWMutex
}

// ModelManagerConfig configures the model manager
type ModelManagerConfig struct {
	StoragePath     string
	MaxModelSize    int64 // Maximum model file size in bytes
	AllowedFormats  []string
	ValidateOnLoad  bool
}

// DefaultModelManagerConfig returns default model manager configuration
func DefaultModelManagerConfig() ModelManagerConfig {
	return ModelManagerConfig{
		StoragePath:    "./data/models",
		MaxModelSize:   500 * 1024 * 1024, // 500MB
		AllowedFormats: []string{".onnx"},
		ValidateOnLoad: true,
	}
}

// NewModelManager creates a new model manager
func NewModelManager(logger *zap.Logger, config ModelManagerConfig) (*ModelManager, error) {
	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(config.StoragePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	mm := &ModelManager{
		logger:      logger.With(zap.String("component", "model_manager")),
		storagePath: config.StoragePath,
		models:      make(map[string]*domain.DetectionModel),
	}

	// Load existing models
	if err := mm.loadExistingModels(); err != nil {
		logger.Warn("Failed to load existing models", zap.Error(err))
	}

	return mm, nil
}

// UploadModel uploads and validates a new model
func (mm *ModelManager) UploadModel(ctx context.Context, name, version, modelType string, classes []string, modelData io.Reader) (*domain.DetectionModel, error) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	mm.logger.Info("Uploading model",
		zap.String("name", name),
		zap.String("version", version),
		zap.String("type", modelType))

	// Validate input parameters
	if err := mm.validateModelParams(name, version, modelType, classes); err != nil {
		return nil, fmt.Errorf("invalid model parameters: %w", err)
	}

	// Generate model ID
	modelID := uuid.New().String()

	// Create model file path
	fileName := fmt.Sprintf("%s_%s_%s.onnx", name, version, modelID)
	filePath := filepath.Join(mm.storagePath, fileName)

	// Save model file
	if err := mm.saveModelFile(filePath, modelData); err != nil {
		return nil, fmt.Errorf("failed to save model file: %w", err)
	}

	// Calculate file hash
	fileHash, err := mm.calculateFileHash(filePath)
	if err != nil {
		os.Remove(filePath) // Clean up on error
		return nil, fmt.Errorf("failed to calculate file hash: %w", err)
	}

	// Validate model file
	if err := mm.validateModelFile(filePath); err != nil {
		os.Remove(filePath) // Clean up on error
		return nil, fmt.Errorf("model validation failed: %w", err)
	}

	// Create model metadata
	model := &domain.DetectionModel{
		ID:        modelID,
		Name:      name,
		Version:   version,
		Type:      modelType,
		FilePath:  filePath,
		Classes:   classes,
		IsActive:  false,
		FileHash:  fileHash,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store model
	mm.models[modelID] = model

	mm.logger.Info("Model uploaded successfully",
		zap.String("model_id", modelID),
		zap.String("file_path", filePath),
		zap.String("file_hash", fileHash))

	return model, nil
}

// GetModel retrieves a model by ID
func (mm *ModelManager) GetModel(modelID string) (*domain.DetectionModel, error) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	model, exists := mm.models[modelID]
	if !exists {
		return nil, fmt.Errorf("model not found: %s", modelID)
	}

	// Return a copy to avoid race conditions
	modelCopy := *model
	return &modelCopy, nil
}

// GetAllModels retrieves all models
func (mm *ModelManager) GetAllModels() []*domain.DetectionModel {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	models := make([]*domain.DetectionModel, 0, len(mm.models))
	for _, model := range mm.models {
		modelCopy := *model
		models = append(models, &modelCopy)
	}

	return models
}

// GetActiveModel returns the currently active model
func (mm *ModelManager) GetActiveModel() *domain.DetectionModel {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	if mm.activeModel == nil {
		return nil
	}

	// Return a copy
	modelCopy := *mm.activeModel
	return &modelCopy
}

// ActivateModel activates a model for detection
func (mm *ModelManager) ActivateModel(ctx context.Context, modelID string) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	model, exists := mm.models[modelID]
	if !exists {
		return fmt.Errorf("model not found: %s", modelID)
	}

	mm.logger.Info("Activating model",
		zap.String("model_id", modelID),
		zap.String("name", model.Name),
		zap.String("version", model.Version))

	// Validate model file still exists
	if _, err := os.Stat(model.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("model file not found: %s", model.FilePath)
	}

	// Load model into detector if available
	if mm.detector != nil {
		if err := mm.detector.LoadModel(ctx, model.FilePath); err != nil {
			return fmt.Errorf("failed to load model into detector: %w", err)
		}
	}

	// Deactivate previous model
	if mm.activeModel != nil {
		mm.activeModel.IsActive = false
		mm.activeModel.UpdatedAt = time.Now()
	}

	// Activate new model
	model.IsActive = true
	model.UpdatedAt = time.Now()
	mm.activeModel = model

	mm.logger.Info("Model activated successfully", zap.String("model_id", modelID))
	return nil
}

// DeactivateModel deactivates the currently active model
func (mm *ModelManager) DeactivateModel() error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if mm.activeModel == nil {
		return fmt.Errorf("no active model to deactivate")
	}

	mm.logger.Info("Deactivating model", zap.String("model_id", mm.activeModel.ID))

	mm.activeModel.IsActive = false
	mm.activeModel.UpdatedAt = time.Now()
	mm.activeModel = nil

	mm.logger.Info("Model deactivated successfully")
	return nil
}

// DeleteModel deletes a model
func (mm *ModelManager) DeleteModel(modelID string) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	model, exists := mm.models[modelID]
	if !exists {
		return fmt.Errorf("model not found: %s", modelID)
	}

	// Cannot delete active model
	if model.IsActive {
		return fmt.Errorf("cannot delete active model: %s", modelID)
	}

	mm.logger.Info("Deleting model",
		zap.String("model_id", modelID),
		zap.String("name", model.Name),
		zap.String("file_path", model.FilePath))

	// Delete model file
	if err := os.Remove(model.FilePath); err != nil && !os.IsNotExist(err) {
		mm.logger.Warn("Failed to delete model file", zap.Error(err))
	}

	// Remove from models map
	delete(mm.models, modelID)

	mm.logger.Info("Model deleted successfully", zap.String("model_id", modelID))
	return nil
}

// SetDetector sets the detector for model loading
func (mm *ModelManager) SetDetector(detector *Detector) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	mm.detector = detector
	mm.logger.Info("Detector set for model manager")
}

// validateModelParams validates model parameters
func (mm *ModelManager) validateModelParams(name, version, modelType string, classes []string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("model name cannot be empty")
	}

	if strings.TrimSpace(version) == "" {
		return fmt.Errorf("model version cannot be empty")
	}

	if strings.TrimSpace(modelType) == "" {
		return fmt.Errorf("model type cannot be empty")
	}

	if len(classes) == 0 {
		return fmt.Errorf("model must have at least one class")
	}

	// Check for duplicate model name/version combination
	for _, model := range mm.models {
		if model.Name == name && model.Version == version {
			return fmt.Errorf("model with name '%s' and version '%s' already exists", name, version)
		}
	}

	return nil
}

// saveModelFile saves model data to file
func (mm *ModelManager) saveModelFile(filePath string, modelData io.Reader) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create model file: %w", err)
	}
	defer file.Close()

	// Copy data with size limit
	limitedReader := io.LimitReader(modelData, 500*1024*1024) // 500MB limit
	_, err = io.Copy(file, limitedReader)
	if err != nil {
		return fmt.Errorf("failed to write model data: %w", err)
	}

	return nil
}

// calculateFileHash calculates SHA256 hash of the model file
func (mm *ModelManager) calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file for hashing: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// validateModelFile validates the model file format and structure
func (mm *ModelManager) validateModelFile(filePath string) error {
	// Check file extension
	ext := filepath.Ext(filePath)
	if ext != ".onnx" {
		return fmt.Errorf("unsupported model format: %s", ext)
	}

	// Check file size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	if fileInfo.Size() == 0 {
		return fmt.Errorf("model file is empty")
	}

	if fileInfo.Size() > 500*1024*1024 { // 500MB limit
		return fmt.Errorf("model file too large: %d bytes", fileInfo.Size())
	}

	// TODO: Add more sophisticated ONNX model validation
	// For now, we just check basic file properties

	return nil
}

// loadExistingModels loads models from the storage directory
func (mm *ModelManager) loadExistingModels() error {
	files, err := os.ReadDir(mm.storagePath)
	if err != nil {
		return fmt.Errorf("failed to read storage directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".onnx") {
			continue
		}

		filePath := filepath.Join(mm.storagePath, file.Name())
		
		// Try to parse model info from filename
		// Format: name_version_id.onnx
		parts := strings.Split(strings.TrimSuffix(file.Name(), ".onnx"), "_")
		if len(parts) < 3 {
			mm.logger.Warn("Skipping model file with invalid name format", zap.String("file", file.Name()))
			continue
		}

		name := parts[0]
		version := parts[1]
		modelID := parts[2]

		// Calculate file hash
		fileHash, err := mm.calculateFileHash(filePath)
		if err != nil {
			mm.logger.Warn("Failed to calculate hash for existing model", 
				zap.String("file", file.Name()), zap.Error(err))
			continue
		}

		// Create model metadata (with default values for missing info)
		model := &domain.DetectionModel{
			ID:        modelID,
			Name:      name,
			Version:   version,
			Type:      "yolo", // Default type
			FilePath:  filePath,
			Classes:   GetCOCOClasses(), // Default classes
			IsActive:  false,
			FileHash:  fileHash,
			CreatedAt: time.Now(), // We don't have the original creation time
			UpdatedAt: time.Now(),
		}

		mm.models[modelID] = model
		mm.logger.Info("Loaded existing model", 
			zap.String("model_id", modelID),
			zap.String("name", name),
			zap.String("version", version))
	}

	return nil
}

// GetStats returns model manager statistics
func (mm *ModelManager) GetStats() map[string]interface{} {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	activeModelID := ""
	if mm.activeModel != nil {
		activeModelID = mm.activeModel.ID
	}

	return map[string]interface{}{
		"total_models":     len(mm.models),
		"active_model_id":  activeModelID,
		"storage_path":     mm.storagePath,
	}
}
