package sources

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// EnvSource represents an environment variable configuration source
type EnvSource struct {
	prefix    string
	separator string
	required  map[string]bool
}

// NewEnvSource creates a new environment variable configuration source
func NewEnvSource(prefix string) *EnvSource {
	return &EnvSource{
		prefix:    prefix,
		separator: "_",
		required:  make(map[string]bool),
	}
}

// WithSeparator sets the separator for nested fields
func (es *EnvSource) WithSeparator(separator string) *EnvSource {
	es.separator = separator
	return es
}

// WithRequired marks environment variables as required
func (es *EnvSource) WithRequired(vars ...string) *EnvSource {
	for _, v := range vars {
		es.required[v] = true
	}
	return es
}

// Load loads configuration from environment variables
func (es *EnvSource) Load(target interface{}) error {
	return es.loadStruct(reflect.ValueOf(target), "")
}

// loadStruct recursively loads struct fields from environment variables
func (es *EnvSource) loadStruct(v reflect.Value, prefix string) error {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %s", v.Kind())
	}
	
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		
		// Skip unexported fields
		if !field.CanSet() {
			continue
		}
		
		// Get field name from tag or use field name
		fieldName := es.getFieldName(fieldType)
		if fieldName == "-" {
			continue // Skip this field
		}
		
		// Build environment variable name
		envName := es.buildEnvName(prefix, fieldName)
		
		// Handle nested structs
		if field.Kind() == reflect.Struct || (field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct) {
			if err := es.loadStruct(field, envName); err != nil {
				return err
			}
			continue
		}
		
		// Load value from environment variable
		if err := es.loadField(field, envName, fieldType); err != nil {
			return err
		}
	}
	
	return nil
}

// loadField loads a single field from environment variable
func (es *EnvSource) loadField(field reflect.Value, envName string, fieldType reflect.StructField) error {
	envValue := os.Getenv(envName)
	
	// Check if required
	if es.required[envName] && envValue == "" {
		return fmt.Errorf("required environment variable %s is not set", envName)
	}
	
	// Skip if not set and not required
	if envValue == "" {
		return nil
	}
	
	// Convert and set value based on field type
	return es.setFieldValue(field, envValue, fieldType)
}

// setFieldValue sets the field value based on its type
func (es *EnvSource) setFieldValue(field reflect.Value, value string, fieldType reflect.StructField) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
		
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value for %s: %s", fieldType.Name, value)
		}
		field.SetBool(boolVal)
		
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			// Handle time.Duration
			duration, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid duration value for %s: %s", fieldType.Name, value)
			}
			field.SetInt(int64(duration))
		} else {
			intVal, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid integer value for %s: %s", fieldType.Name, value)
			}
			field.SetInt(intVal)
		}
		
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid unsigned integer value for %s: %s", fieldType.Name, value)
		}
		field.SetUint(uintVal)
		
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value for %s: %s", fieldType.Name, value)
		}
		field.SetFloat(floatVal)
		
	case reflect.Slice:
		return es.setSliceValue(field, value, fieldType)
		
	case reflect.Map:
		return es.setMapValue(field, value, fieldType)
		
	case reflect.Ptr:
		// Handle pointer types
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return es.setFieldValue(field.Elem(), value, fieldType)
		
	default:
		return fmt.Errorf("unsupported field type %s for field %s", field.Kind(), fieldType.Name)
	}
	
	return nil
}

// setSliceValue sets slice values from comma-separated string
func (es *EnvSource) setSliceValue(field reflect.Value, value string, fieldType reflect.StructField) error {
	if value == "" {
		return nil
	}
	
	parts := strings.Split(value, ",")
	sliceType := field.Type().Elem()
	slice := reflect.MakeSlice(field.Type(), len(parts), len(parts))
	
	for i, part := range parts {
		part = strings.TrimSpace(part)
		elem := slice.Index(i)
		
		switch sliceType.Kind() {
		case reflect.String:
			elem.SetString(part)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intVal, err := strconv.ParseInt(part, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid integer value in slice for %s: %s", fieldType.Name, part)
			}
			elem.SetInt(intVal)
		default:
			return fmt.Errorf("unsupported slice element type %s for field %s", sliceType.Kind(), fieldType.Name)
		}
	}
	
	field.Set(slice)
	return nil
}

// setMapValue sets map values from key=value pairs
func (es *EnvSource) setMapValue(field reflect.Value, value string, fieldType reflect.StructField) error {
	if value == "" {
		return nil
	}
	
	mapType := field.Type()
	keyType := mapType.Key()
	valueType := mapType.Elem()
	
	// Only support string keys for now
	if keyType.Kind() != reflect.String {
		return fmt.Errorf("unsupported map key type %s for field %s", keyType.Kind(), fieldType.Name)
	}
	
	mapValue := reflect.MakeMap(mapType)
	pairs := strings.Split(value, ",")
	
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			return fmt.Errorf("invalid key=value pair for %s: %s", fieldType.Name, pair)
		}
		
		key := reflect.ValueOf(strings.TrimSpace(kv[0]))
		val := reflect.New(valueType).Elem()
		
		switch valueType.Kind() {
		case reflect.String:
			val.SetString(strings.TrimSpace(kv[1]))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intVal, err := strconv.ParseInt(strings.TrimSpace(kv[1]), 10, 64)
			if err != nil {
				return fmt.Errorf("invalid integer value in map for %s: %s", fieldType.Name, kv[1])
			}
			val.SetInt(intVal)
		default:
			return fmt.Errorf("unsupported map value type %s for field %s", valueType.Kind(), fieldType.Name)
		}
		
		mapValue.SetMapIndex(key, val)
	}
	
	field.Set(mapValue)
	return nil
}

// getFieldName gets the field name from struct tag or field name
func (es *EnvSource) getFieldName(fieldType reflect.StructField) string {
	// Check for env tag first
	if envTag := fieldType.Tag.Get("env"); envTag != "" {
		return envTag
	}
	
	// Check for yaml tag
	if yamlTag := fieldType.Tag.Get("yaml"); yamlTag != "" {
		parts := strings.Split(yamlTag, ",")
		if parts[0] != "" {
			return strings.ToUpper(parts[0])
		}
	}
	
	// Check for json tag
	if jsonTag := fieldType.Tag.Get("json"); jsonTag != "" {
		parts := strings.Split(jsonTag, ",")
		if parts[0] != "" {
			return strings.ToUpper(parts[0])
		}
	}
	
	// Use field name as fallback
	return strings.ToUpper(fieldType.Name)
}

// buildEnvName builds the full environment variable name
func (es *EnvSource) buildEnvName(prefix, fieldName string) string {
	parts := []string{}
	
	if es.prefix != "" {
		parts = append(parts, es.prefix)
	}
	
	if prefix != "" {
		parts = append(parts, prefix)
	}
	
	parts = append(parts, fieldName)
	
	return strings.Join(parts, es.separator)
}

// GetPrefix returns the environment variable prefix
func (es *EnvSource) GetPrefix() string {
	return es.prefix
}

// GetSeparator returns the separator used for nested fields
func (es *EnvSource) GetSeparator() string {
	return es.separator
}

// GetRequired returns the list of required environment variables
func (es *EnvSource) GetRequired() []string {
	var required []string
	for k, v := range es.required {
		if v {
			required = append(required, k)
		}
	}
	return required
}

// EnvVarInfo represents information about an environment variable
type EnvVarInfo struct {
	Name        string `json:"name"`
	Value       string `json:"value,omitempty"`
	Required    bool   `json:"required"`
	Description string `json:"description,omitempty"`
	Example     string `json:"example,omitempty"`
}

// EnvDocGenerator generates documentation for environment variables
type EnvDocGenerator struct {
	prefix    string
	separator string
	vars      []EnvVarInfo
}

// NewEnvDocGenerator creates a new environment variable documentation generator
func NewEnvDocGenerator(prefix string) *EnvDocGenerator {
	return &EnvDocGenerator{
		prefix:    prefix,
		separator: "_",
		vars:      []EnvVarInfo{},
	}
}

// AddVar adds an environment variable to the documentation
func (edg *EnvDocGenerator) AddVar(name, description, example string, required bool) {
	edg.vars = append(edg.vars, EnvVarInfo{
		Name:        name,
		Required:    required,
		Description: description,
		Example:     example,
	})
}

// GenerateMarkdown generates Markdown documentation for environment variables
func (edg *EnvDocGenerator) GenerateMarkdown() string {
	var sb strings.Builder
	
	sb.WriteString("# Environment Variables\n\n")
	sb.WriteString("The following environment variables can be used to configure the application:\n\n")
	
	sb.WriteString("| Variable | Required | Description | Example |\n")
	sb.WriteString("|----------|----------|-------------|----------|\n")
	
	for _, v := range edg.vars {
		required := "No"
		if v.Required {
			required = "Yes"
		}
		
		sb.WriteString(fmt.Sprintf("| `%s` | %s | %s | `%s` |\n",
			v.Name, required, v.Description, v.Example))
	}
	
	return sb.String()
}

// GenerateShell generates shell script with environment variable examples
func (edg *EnvDocGenerator) GenerateShell() string {
	var sb strings.Builder
	
	sb.WriteString("#!/bin/bash\n")
	sb.WriteString("# Environment variables for go-coffee-ai-agents\n\n")
	
	for _, v := range edg.vars {
		if v.Description != "" {
			sb.WriteString(fmt.Sprintf("# %s\n", v.Description))
		}
		if v.Required {
			sb.WriteString("# Required\n")
		}
		sb.WriteString(fmt.Sprintf("export %s=\"%s\"\n\n", v.Name, v.Example))
	}
	
	return sb.String()
}

// GetVars returns all environment variable information
func (edg *EnvDocGenerator) GetVars() []EnvVarInfo {
	return edg.vars
}
