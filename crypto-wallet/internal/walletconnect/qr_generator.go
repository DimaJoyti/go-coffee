package walletconnect

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"go.uber.org/zap"
)

// QRGenerator provides QR code generation functionality
type QRGenerator struct {
	logger *logger.Logger
	config QRCodeConfig
}

// QRCodeData represents generated QR code data
type QRCodeData struct {
	URI        string `json:"uri"`
	DataURL    string `json:"data_url"`
	Base64     string `json:"base64"`
	Format     string `json:"format"`
	Size       int    `json:"size"`
	ErrorLevel string `json:"error_level"`
}

// NewQRGenerator creates a new QR code generator
func NewQRGenerator(logger *logger.Logger, config QRCodeConfig) *QRGenerator {
	return &QRGenerator{
		logger: logger.Named("qr-generator"),
		config: config,
	}
}

// GenerateQRCode generates a QR code for the given URI
func (qg *QRGenerator) GenerateQRCode(uri string) (*QRCodeData, error) {
	qg.logger.Debug("Generating QR code", zap.String("uri", uri))

	if !qg.config.Enabled {
		return nil, fmt.Errorf("QR code generation is disabled")
	}

	// Mock QR code generation - in production would use actual QR library
	qrData := &QRCodeData{
		URI:        uri,
		Format:     qg.config.Format,
		Size:       qg.config.Size,
		ErrorLevel: qg.config.ErrorLevel,
	}

	// Generate mock QR code image
	switch qg.config.Format {
	case "PNG":
		base64Data, dataURL, err := qg.generatePNGQRCode(uri)
		if err != nil {
			return nil, fmt.Errorf("failed to generate PNG QR code: %w", err)
		}
		qrData.Base64 = base64Data
		qrData.DataURL = dataURL
	case "SVG":
		base64Data, dataURL, err := qg.generateSVGQRCode(uri)
		if err != nil {
			return nil, fmt.Errorf("failed to generate SVG QR code: %w", err)
		}
		qrData.Base64 = base64Data
		qrData.DataURL = dataURL
	default:
		return nil, fmt.Errorf("unsupported QR code format: %s", qg.config.Format)
	}

	qg.logger.Info("QR code generated successfully",
		zap.String("format", qrData.Format),
		zap.Int("size", qrData.Size))

	return qrData, nil
}

// generatePNGQRCode generates a PNG QR code (mock implementation)
func (qg *QRGenerator) generatePNGQRCode(uri string) (string, string, error) {
	// Mock PNG generation - in production would use actual QR library like github.com/skip2/go-qrcode
	// The uri parameter would be used in actual QR code generation
	_ = uri // Suppress unused parameter warning in mock implementation

	// Create a simple mock image
	img := qg.createMockQRImage()

	// Convert to PNG bytes
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", "", fmt.Errorf("failed to encode PNG: %w", err)
	}

	// Convert to base64
	base64Data := base64.StdEncoding.EncodeToString(buf.Bytes())
	dataURL := fmt.Sprintf("data:image/png;base64,%s", base64Data)

	return base64Data, dataURL, nil
}

// generateSVGQRCode generates an SVG QR code (mock implementation)
func (qg *QRGenerator) generateSVGQRCode(uri string) (string, string, error) {
	// Mock SVG generation
	svgContent := qg.createMockSVGQRCode(uri)

	// Convert to base64
	base64Data := base64.StdEncoding.EncodeToString([]byte(svgContent))
	dataURL := fmt.Sprintf("data:image/svg+xml;base64,%s", base64Data)

	return base64Data, dataURL, nil
}

// createMockQRImage creates a mock QR code image
func (qg *QRGenerator) createMockQRImage() image.Image {
	// Mock implementation - in production would generate actual QR code
	return &MockImage{
		Width:  qg.config.Size,
		Height: qg.config.Size,
		Data:   fmt.Sprintf("Mock QR Code %dx%d", qg.config.Size, qg.config.Size),
	}
}

// createMockSVGQRCode creates a mock SVG QR code
func (qg *QRGenerator) createMockSVGQRCode(uri string) string {
	size := qg.config.Size
	border := qg.config.Border

	// Mock SVG QR code
	svg := fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">
		<rect width="%d" height="%d" fill="white"/>
		<rect x="%d" y="%d" width="%d" height="%d" fill="black"/>
		<text x="%d" y="%d" font-family="Arial" font-size="12" fill="gray">Mock QR</text>
		<!-- URI: %s -->
	</svg>`,
		size, size,
		size, size,
		border, border, size-2*border, size-2*border,
		size/2-20, size/2,
		uri)

	return svg
}

// MockImage represents a mock image for testing
type MockImage struct {
	Width  int
	Height int
	Data   string
}

// ColorModel implements image.Image interface
func (m *MockImage) ColorModel() color.Model {
	return color.RGBAModel
}

// Bounds implements image.Image interface
func (m *MockImage) Bounds() image.Rectangle {
	return image.Rect(0, 0, m.Width, m.Height)
}

// At implements image.Image interface
func (m *MockImage) At(x, y int) color.Color {
	// Mock pixel data - return white pixel
	return color.RGBA{R: 255, G: 255, B: 255, A: 255}
}

// ValidateQRCodeConfig validates QR code configuration
func ValidateQRCodeConfig(config QRCodeConfig) error {
	if config.Size <= 0 {
		return fmt.Errorf("QR code size must be positive")
	}

	if config.Size < 64 || config.Size > 1024 {
		return fmt.Errorf("QR code size must be between 64 and 1024 pixels")
	}

	if config.Border < 0 {
		return fmt.Errorf("QR code border must be non-negative")
	}

	validErrorLevels := map[string]bool{
		"L": true, // Low
		"M": true, // Medium
		"Q": true, // Quartile
		"H": true, // High
	}

	if !validErrorLevels[config.ErrorLevel] {
		return fmt.Errorf("invalid error correction level: %s (must be L, M, Q, or H)", config.ErrorLevel)
	}

	validFormats := map[string]bool{
		"PNG":  true,
		"SVG":  true,
		"JPEG": true,
	}

	if !validFormats[config.Format] {
		return fmt.Errorf("invalid QR code format: %s (must be PNG, SVG, or JPEG)", config.Format)
	}

	return nil
}

// GetQRCodeInfo returns information about QR code capabilities
func GetQRCodeInfo() map[string]any {
	return map[string]any{
		"supported_formats":      []string{"PNG", "SVG", "JPEG"},
		"supported_error_levels": []string{"L", "M", "Q", "H"},
		"min_size":               64,
		"max_size":               1024,
		"default_size":           256,
		"default_error_level":    "M",
		"default_border":         4,
		"features": map[string]bool{
			"custom_colors":    false, // Not implemented in mock
			"logo_embedding":   false, // Not implemented in mock
			"batch_generation": true,
			"url_validation":   true,
		},
	}
}

// BatchGenerateQRCodes generates multiple QR codes
func (qg *QRGenerator) BatchGenerateQRCodes(uris []string) ([]*QRCodeData, error) {
	qg.logger.Info("Generating batch QR codes", zap.Int("count", len(uris)))

	if len(uris) == 0 {
		return nil, fmt.Errorf("no URIs provided for batch generation")
	}

	if len(uris) > 100 {
		return nil, fmt.Errorf("batch size too large: %d (maximum 100)", len(uris))
	}

	var qrCodes []*QRCodeData
	var errors []error

	for i, uri := range uris {
		qrData, err := qg.GenerateQRCode(uri)
		if err != nil {
			qg.logger.Error("Failed to generate QR code in batch",
				zap.Int("index", i),
				zap.String("uri", uri),
				zap.Error(err))
			errors = append(errors, fmt.Errorf("index %d: %w", i, err))
			continue
		}
		qrCodes = append(qrCodes, qrData)
	}

	if len(errors) > 0 {
		qg.logger.Warn("Some QR codes failed to generate",
			zap.Int("failed_count", len(errors)),
			zap.Int("success_count", len(qrCodes)))
	}

	qg.logger.Info("Batch QR code generation completed",
		zap.Int("success_count", len(qrCodes)),
		zap.Int("failed_count", len(errors)))

	return qrCodes, nil
}

// ValidateWalletConnectURI validates a WalletConnect URI format
func ValidateWalletConnectURI(uri string) error {
	if uri == "" {
		return fmt.Errorf("URI cannot be empty")
	}

	// Basic WalletConnect URI validation
	if len(uri) < 10 {
		return fmt.Errorf("URI too short")
	}

	// Check for WalletConnect prefix (simplified)
	if uri[:3] != "wc:" {
		return fmt.Errorf("invalid WalletConnect URI format: must start with 'wc:'")
	}

	// Additional validation could be added here
	return nil
}

// GetDefaultQRCodeConfig returns default QR code configuration
func GetDefaultQRCodeConfig() QRCodeConfig {
	return QRCodeConfig{
		Enabled:    true,
		Size:       256,
		ErrorLevel: "M",
		Border:     4,
		Format:     "PNG",
	}
}
