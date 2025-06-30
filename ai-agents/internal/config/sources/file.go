package sources

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FileSource represents a file-based configuration source
type FileSource struct {
	path     string
	format   string
	required bool
}

// NewFileSource creates a new file-based configuration source
func NewFileSource(path string, required bool) *FileSource {
	format := detectFormat(path)
	return &FileSource{
		path:     path,
		format:   format,
		required: required,
	}
}

// Load loads configuration from the file
func (fs *FileSource) Load(target interface{}) error {
	// Check if file exists
	if _, err := os.Stat(fs.path); os.IsNotExist(err) {
		if fs.required {
			return fmt.Errorf("required configuration file not found: %s", fs.path)
		}
		return nil // Optional file, skip loading
	}
	
	// Read file content
	data, err := os.ReadFile(fs.path)
	if err != nil {
		return fmt.Errorf("failed to read configuration file %s: %w", fs.path, err)
	}
	
	// Parse based on format
	switch fs.format {
	case "yaml", "yml":
		return fs.parseYAML(data, target)
	case "json":
		return fs.parseJSON(data, target)
	default:
		return fmt.Errorf("unsupported configuration file format: %s", fs.format)
	}
}

// parseYAML parses YAML configuration data
func (fs *FileSource) parseYAML(data []byte, target interface{}) error {
	if err := yaml.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to parse YAML configuration from %s: %w", fs.path, err)
	}
	return nil
}

// parseJSON parses JSON configuration data
func (fs *FileSource) parseJSON(data []byte, target interface{}) error {
	// For now, we'll use YAML parser which can handle JSON
	// In a full implementation, you'd use encoding/json
	if err := yaml.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to parse JSON configuration from %s: %w", fs.path, err)
	}
	return nil
}

// GetPath returns the file path
func (fs *FileSource) GetPath() string {
	return fs.path
}

// GetFormat returns the file format
func (fs *FileSource) GetFormat() string {
	return fs.format
}

// IsRequired returns whether the file is required
func (fs *FileSource) IsRequired() bool {
	return fs.required
}

// Exists checks if the configuration file exists
func (fs *FileSource) Exists() bool {
	_, err := os.Stat(fs.path)
	return err == nil
}

// Watch watches the configuration file for changes
func (fs *FileSource) Watch(callback func()) error {
	// This would implement file watching using fsnotify
	// For now, return not implemented
	return fmt.Errorf("file watching not implemented")
}

// detectFormat detects the configuration file format from extension
func detectFormat(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return "yaml"
	case ".json":
		return "json"
	default:
		return "yaml" // Default to YAML
	}
}

// MultiFileSource represents multiple file-based configuration sources
type MultiFileSource struct {
	sources []*FileSource
}

// NewMultiFileSource creates a new multi-file configuration source
func NewMultiFileSource(paths []string, required bool) *MultiFileSource {
	var sources []*FileSource
	for _, path := range paths {
		sources = append(sources, NewFileSource(path, required))
	}
	
	return &MultiFileSource{
		sources: sources,
	}
}

// Load loads configuration from all files in order
func (mfs *MultiFileSource) Load(target interface{}) error {
	for _, source := range mfs.sources {
		if err := source.Load(target); err != nil {
			return err
		}
	}
	return nil
}

// GetSources returns all file sources
func (mfs *MultiFileSource) GetSources() []*FileSource {
	return mfs.sources
}

// EnvironmentFileSource represents environment-specific file configuration
type EnvironmentFileSource struct {
	basePath    string
	environment string
	required    bool
}

// NewEnvironmentFileSource creates a new environment-specific file source
func NewEnvironmentFileSource(basePath, environment string, required bool) *EnvironmentFileSource {
	return &EnvironmentFileSource{
		basePath:    basePath,
		environment: environment,
		required:    required,
	}
}

// Load loads base configuration and environment-specific overrides
func (efs *EnvironmentFileSource) Load(target interface{}) error {
	// Load base configuration
	baseSource := NewFileSource(efs.basePath, efs.required)
	if err := baseSource.Load(target); err != nil {
		return fmt.Errorf("failed to load base configuration: %w", err)
	}
	
	// Load environment-specific configuration
	envPath := efs.getEnvironmentPath()
	envSource := NewFileSource(envPath, false) // Environment files are optional
	if envSource.Exists() {
		if err := envSource.Load(target); err != nil {
			return fmt.Errorf("failed to load environment configuration: %w", err)
		}
	}
	
	return nil
}

// getEnvironmentPath returns the environment-specific configuration file path
func (efs *EnvironmentFileSource) getEnvironmentPath() string {
	dir := filepath.Dir(efs.basePath)
	ext := filepath.Ext(efs.basePath)
	name := strings.TrimSuffix(filepath.Base(efs.basePath), ext)
	
	return filepath.Join(dir, fmt.Sprintf("%s.%s%s", name, efs.environment, ext))
}

// GetBasePath returns the base configuration file path
func (efs *EnvironmentFileSource) GetBasePath() string {
	return efs.basePath
}

// GetEnvironmentPath returns the environment-specific configuration file path
func (efs *EnvironmentFileSource) GetEnvironmentPath() string {
	return efs.getEnvironmentPath()
}

// GetEnvironment returns the environment name
func (efs *EnvironmentFileSource) GetEnvironment() string {
	return efs.environment
}

// ConfigFileLocator helps locate configuration files
type ConfigFileLocator struct {
	searchPaths []string
	fileName    string
}

// NewConfigFileLocator creates a new configuration file locator
func NewConfigFileLocator(fileName string, searchPaths []string) *ConfigFileLocator {
	if len(searchPaths) == 0 {
		searchPaths = getDefaultSearchPaths()
	}
	
	return &ConfigFileLocator{
		searchPaths: searchPaths,
		fileName:    fileName,
	}
}

// Locate finds the configuration file in search paths
func (cfl *ConfigFileLocator) Locate() (string, error) {
	for _, searchPath := range cfl.searchPaths {
		configPath := filepath.Join(searchPath, cfl.fileName)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
	}
	
	return "", fmt.Errorf("configuration file '%s' not found in search paths: %v", 
		cfl.fileName, cfl.searchPaths)
}

// LocateAll finds all configuration files in search paths
func (cfl *ConfigFileLocator) LocateAll() []string {
	var found []string
	
	for _, searchPath := range cfl.searchPaths {
		configPath := filepath.Join(searchPath, cfl.fileName)
		if _, err := os.Stat(configPath); err == nil {
			found = append(found, configPath)
		}
	}
	
	return found
}

// GetSearchPaths returns the search paths
func (cfl *ConfigFileLocator) GetSearchPaths() []string {
	return cfl.searchPaths
}

// AddSearchPath adds a search path
func (cfl *ConfigFileLocator) AddSearchPath(path string) {
	cfl.searchPaths = append(cfl.searchPaths, path)
}

// getDefaultSearchPaths returns default configuration file search paths
func getDefaultSearchPaths() []string {
	homeDir, _ := os.UserHomeDir()
	
	return []string{
		".",                                    // Current directory
		"./config",                            // Config subdirectory
		"./configs",                           // Configs subdirectory
		"/etc/go-coffee",                      // System config directory
		filepath.Join(homeDir, ".go-coffee"), // User config directory
		"/usr/local/etc/go-coffee",           // Local system config
	}
}

// ConfigTemplate represents a configuration template
type ConfigTemplate struct {
	template string
	values   map[string]string
}

// NewConfigTemplate creates a new configuration template
func NewConfigTemplate(template string, values map[string]string) *ConfigTemplate {
	return &ConfigTemplate{
		template: template,
		values:   values,
	}
}

// Render renders the configuration template with values
func (ct *ConfigTemplate) Render() (string, error) {
	result := ct.template
	
	for key, value := range ct.values {
		placeholder := fmt.Sprintf("${%s}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	
	// Check for unresolved placeholders
	if strings.Contains(result, "${") {
		return "", fmt.Errorf("unresolved placeholders found in configuration template")
	}
	
	return result, nil
}

// AddValue adds a template value
func (ct *ConfigTemplate) AddValue(key, value string) {
	if ct.values == nil {
		ct.values = make(map[string]string)
	}
	ct.values[key] = value
}

// GetValues returns all template values
func (ct *ConfigTemplate) GetValues() map[string]string {
	return ct.values
}

// TemplateFileSource represents a template-based file configuration source
type TemplateFileSource struct {
	templatePath string
	values       map[string]string
	required     bool
}

// NewTemplateFileSource creates a new template-based file source
func NewTemplateFileSource(templatePath string, values map[string]string, required bool) *TemplateFileSource {
	return &TemplateFileSource{
		templatePath: templatePath,
		values:       values,
		required:     required,
	}
}

// Load loads and renders the template configuration
func (tfs *TemplateFileSource) Load(target interface{}) error {
	// Check if template file exists
	if _, err := os.Stat(tfs.templatePath); os.IsNotExist(err) {
		if tfs.required {
			return fmt.Errorf("required configuration template not found: %s", tfs.templatePath)
		}
		return nil
	}
	
	// Read template content
	templateData, err := os.ReadFile(tfs.templatePath)
	if err != nil {
		return fmt.Errorf("failed to read configuration template %s: %w", tfs.templatePath, err)
	}
	
	// Create and render template
	template := NewConfigTemplate(string(templateData), tfs.values)
	rendered, err := template.Render()
	if err != nil {
		return fmt.Errorf("failed to render configuration template: %w", err)
	}
	
	// Parse rendered configuration
	format := detectFormat(tfs.templatePath)
	switch format {
	case "yaml", "yml":
		return yaml.Unmarshal([]byte(rendered), target)
	case "json":
		return yaml.Unmarshal([]byte(rendered), target) // YAML parser handles JSON
	default:
		return fmt.Errorf("unsupported template format: %s", format)
	}
}

// GetTemplatePath returns the template file path
func (tfs *TemplateFileSource) GetTemplatePath() string {
	return tfs.templatePath
}

// GetValues returns the template values
func (tfs *TemplateFileSource) GetValues() map[string]string {
	return tfs.values
}

// AddValue adds a template value
func (tfs *TemplateFileSource) AddValue(key, value string) {
	if tfs.values == nil {
		tfs.values = make(map[string]string)
	}
	tfs.values[key] = value
}
