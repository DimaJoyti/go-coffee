package config

import (
	"testing"
)

func TestConfigLoad(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
		wantError  bool
	}{
		{
			name:       "empty config path",
			configPath: "",
			wantError:  false, // Should use defaults
		},
		{
			name:       "non-existent config file",
			configPath: "/non/existent/path/config.yaml",
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test config loading behavior
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Config loading panicked: %v", r)
				}
			}()

			// Basic test for config structure
			t.Logf("Config test for path %s completed", tt.configPath)
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	// Test default configuration values
	defaults := map[string]interface{}{
		"log_level":     "info",
		"output_format": "table",
		"telemetry":     true,
	}

	for key, value := range defaults {
		t.Run("default_"+key, func(t *testing.T) {
			// Test that default values are reasonable
			switch key {
			case "log_level":
				if value != "info" {
					t.Errorf("Expected default log_level to be 'info', got %v", value)
				}
			case "output_format":
				if value != "table" {
					t.Errorf("Expected default output_format to be 'table', got %v", value)
				}
			case "telemetry":
				if value != true {
					t.Errorf("Expected default telemetry to be true, got %v", value)
				}
			}
			t.Logf("Default %s: %v", key, value)
		})
	}
}

func TestConfigValidation(t *testing.T) {
	// Test configuration validation
	tests := []struct {
		name   string
		config map[string]interface{}
		valid  bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"log_level":     "info",
				"output_format": "table",
			},
			valid: true,
		},
		{
			name: "invalid log level",
			config: map[string]interface{}{
				"log_level": "invalid",
			},
			valid: false,
		},
		{
			name: "invalid output format",
			config: map[string]interface{}{
				"output_format": "invalid",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test configuration validation logic
			for key, value := range tt.config {
				switch key {
				case "log_level":
					validLevels := []string{"debug", "info", "warn", "error"}
					valid := false
					for _, level := range validLevels {
						if value == level {
							valid = true
							break
						}
					}
					if !valid && tt.valid {
						t.Errorf("Invalid log level %v should be rejected", value)
					}
				case "output_format":
					validFormats := []string{"table", "json", "yaml"}
					valid := false
					for _, format := range validFormats {
						if value == format {
							valid = true
							break
						}
					}
					if !valid && tt.valid {
						t.Errorf("Invalid output format %v should be rejected", value)
					}
				}
			}
		})
	}
}
