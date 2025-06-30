package sources

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// SecretSource represents a source for retrieving secrets
type SecretSource interface {
	GetSecret(ctx context.Context, path string) (map[string]interface{}, error)
	GetSecretValue(ctx context.Context, path, key string) (string, error)
	PutSecret(ctx context.Context, path string, data map[string]interface{}) error
	DeleteSecret(ctx context.Context, path string) error
	ListSecrets(ctx context.Context, path string) ([]string, error)
}

// VaultConfig holds Vault configuration
type VaultConfig struct {
	Address   string `yaml:"address" json:"address"`
	Token     string `yaml:"token" json:"token"`
	Namespace string `yaml:"namespace" json:"namespace"`
	
	// Authentication
	AuthMethod string `yaml:"auth_method" json:"auth_method"` // token, userpass, aws, kubernetes
	Username   string `yaml:"username" json:"username"`
	Password   string `yaml:"password" json:"password"`
	RoleID     string `yaml:"role_id" json:"role_id"`
	SecretID   string `yaml:"secret_id" json:"secret_id"`
	
	// TLS configuration
	TLSSkipVerify bool   `yaml:"tls_skip_verify" json:"tls_skip_verify"`
	CACert        string `yaml:"ca_cert" json:"ca_cert"`
	ClientCert    string `yaml:"client_cert" json:"client_cert"`
	ClientKey     string `yaml:"client_key" json:"client_key"`
	
	// Connection settings
	Timeout       time.Duration `yaml:"timeout" json:"timeout"`
	RetryAttempts int           `yaml:"retry_attempts" json:"retry_attempts"`
	
	// Secret engine settings
	SecretEngine string `yaml:"secret_engine" json:"secret_engine"` // kv, kv-v2
	MountPath    string `yaml:"mount_path" json:"mount_path"`
}

// VaultSource implements SecretSource for HashiCorp Vault
type VaultSource struct {
	config VaultConfig
	client VaultClient
	mutex  sync.RWMutex
	cache  map[string]cachedSecret
}

// VaultClient interface for Vault operations (allows mocking)
type VaultClient interface {
	Read(path string) (*VaultSecret, error)
	Write(path string, data map[string]interface{}) (*VaultSecret, error)
	Delete(path string) (*VaultSecret, error)
	List(path string) (*VaultSecret, error)
	SetToken(token string)
	Token() string
}

// VaultSecret represents a Vault secret response
type VaultSecret struct {
	Data     map[string]interface{} `json:"data"`
	Metadata map[string]interface{} `json:"metadata"`
	LeaseID  string                 `json:"lease_id"`
	LeaseDuration int                `json:"lease_duration"`
	Renewable bool                   `json:"renewable"`
}

// cachedSecret represents a cached secret with TTL
type cachedSecret struct {
	data      map[string]interface{}
	expiresAt time.Time
}

// NewVaultSource creates a new Vault secret source
func NewVaultSource(config VaultConfig) (*VaultSource, error) {
	// Set defaults
	if config.Address == "" {
		config.Address = os.Getenv("VAULT_ADDR")
		if config.Address == "" {
			config.Address = "http://localhost:8200"
		}
	}
	
	if config.Token == "" {
		config.Token = os.Getenv("VAULT_TOKEN")
	}
	
	if config.Namespace == "" {
		config.Namespace = os.Getenv("VAULT_NAMESPACE")
	}
	
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	
	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}
	
	if config.SecretEngine == "" {
		config.SecretEngine = "kv-v2"
	}
	
	if config.MountPath == "" {
		config.MountPath = "secret"
	}
	
	// Create Vault client (this would be the actual Vault client in production)
	client := &mockVaultClient{
		token:   config.Token,
		secrets: make(map[string]*VaultSecret),
	}
	
	return &VaultSource{
		config: config,
		client: client,
		cache:  make(map[string]cachedSecret),
	}, nil
}

// GetSecret retrieves a secret from Vault
func (vs *VaultSource) GetSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	// Check cache first
	vs.mutex.RLock()
	if cached, exists := vs.cache[path]; exists && time.Now().Before(cached.expiresAt) {
		vs.mutex.RUnlock()
		return cached.data, nil
	}
	vs.mutex.RUnlock()
	
	// Build full path
	fullPath := vs.buildSecretPath(path)
	
	// Retrieve from Vault with retry
	var secret *VaultSecret
	var err error
	
	for attempt := 0; attempt < vs.config.RetryAttempts; attempt++ {
		secret, err = vs.client.Read(fullPath)
		if err == nil {
			break
		}
		
		if attempt < vs.config.RetryAttempts-1 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to read secret from Vault path %s: %w", fullPath, err)
	}
	
	if secret == nil {
		return nil, fmt.Errorf("secret not found at path %s", fullPath)
	}
	
	// Extract data based on secret engine version
	var data map[string]interface{}
	if vs.config.SecretEngine == "kv-v2" {
		if dataField, exists := secret.Data["data"]; exists {
			if dataMap, ok := dataField.(map[string]interface{}); ok {
				data = dataMap
			}
		}
	} else {
		data = secret.Data
	}
	
	if data == nil {
		return nil, fmt.Errorf("no data found in secret at path %s", fullPath)
	}
	
	// Cache the secret
	vs.mutex.Lock()
	vs.cache[path] = cachedSecret{
		data:      data,
		expiresAt: time.Now().Add(5 * time.Minute), // 5-minute cache TTL
	}
	vs.mutex.Unlock()
	
	return data, nil
}

// GetSecretValue retrieves a specific value from a secret
func (vs *VaultSource) GetSecretValue(ctx context.Context, path, key string) (string, error) {
	data, err := vs.GetSecret(ctx, path)
	if err != nil {
		return "", err
	}
	
	value, exists := data[key]
	if !exists {
		return "", fmt.Errorf("key %s not found in secret at path %s", key, path)
	}
	
	stringValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("value for key %s is not a string", key)
	}
	
	return stringValue, nil
}

// PutSecret stores a secret in Vault
func (vs *VaultSource) PutSecret(ctx context.Context, path string, data map[string]interface{}) error {
	fullPath := vs.buildSecretPath(path)
	
	// Prepare data based on secret engine version
	var writeData map[string]interface{}
	if vs.config.SecretEngine == "kv-v2" {
		writeData = map[string]interface{}{
			"data": data,
		}
	} else {
		writeData = data
	}
	
	// Write to Vault with retry
	var err error
	for attempt := 0; attempt < vs.config.RetryAttempts; attempt++ {
		_, err = vs.client.Write(fullPath, writeData)
		if err == nil {
			break
		}
		
		if attempt < vs.config.RetryAttempts-1 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}
	
	if err != nil {
		return fmt.Errorf("failed to write secret to Vault path %s: %w", fullPath, err)
	}
	
	// Invalidate cache
	vs.mutex.Lock()
	delete(vs.cache, path)
	vs.mutex.Unlock()
	
	return nil
}

// DeleteSecret deletes a secret from Vault
func (vs *VaultSource) DeleteSecret(ctx context.Context, path string) error {
	fullPath := vs.buildSecretPath(path)
	
	// Delete from Vault with retry
	var err error
	for attempt := 0; attempt < vs.config.RetryAttempts; attempt++ {
		_, err = vs.client.Delete(fullPath)
		if err == nil {
			break
		}
		
		if attempt < vs.config.RetryAttempts-1 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}
	
	if err != nil {
		return fmt.Errorf("failed to delete secret from Vault path %s: %w", fullPath, err)
	}
	
	// Invalidate cache
	vs.mutex.Lock()
	delete(vs.cache, path)
	vs.mutex.Unlock()
	
	return nil
}

// ListSecrets lists secrets at a given path
func (vs *VaultSource) ListSecrets(ctx context.Context, path string) ([]string, error) {
	fullPath := vs.buildSecretPath(path)
	
	// List from Vault with retry
	var secret *VaultSecret
	var err error
	
	for attempt := 0; attempt < vs.config.RetryAttempts; attempt++ {
		secret, err = vs.client.List(fullPath)
		if err == nil {
			break
		}
		
		if attempt < vs.config.RetryAttempts-1 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets from Vault path %s: %w", fullPath, err)
	}
	
	if secret == nil || secret.Data == nil {
		return []string{}, nil
	}
	
	keys, exists := secret.Data["keys"]
	if !exists {
		return []string{}, nil
	}
	
	keySlice, ok := keys.([]interface{})
	if !ok {
		return []string{}, nil
	}
	
	var result []string
	for _, key := range keySlice {
		if keyStr, ok := key.(string); ok {
			result = append(result, keyStr)
		}
	}
	
	return result, nil
}

// buildSecretPath builds the full secret path for Vault
func (vs *VaultSource) buildSecretPath(path string) string {
	// Remove leading slash
	path = strings.TrimPrefix(path, "/")
	
	if vs.config.SecretEngine == "kv-v2" {
		return fmt.Sprintf("%s/data/%s", vs.config.MountPath, path)
	}
	
	return fmt.Sprintf("%s/%s", vs.config.MountPath, path)
}

// ClearCache clears the secret cache
func (vs *VaultSource) ClearCache() {
	vs.mutex.Lock()
	vs.cache = make(map[string]cachedSecret)
	vs.mutex.Unlock()
}

// GetConfig returns the Vault configuration
func (vs *VaultSource) GetConfig() VaultConfig {
	return vs.config
}

// mockVaultClient is a mock implementation for testing
type mockVaultClient struct {
	token   string
	secrets map[string]*VaultSecret
	mutex   sync.RWMutex
}

// Read implements VaultClient.Read
func (mvc *mockVaultClient) Read(path string) (*VaultSecret, error) {
	mvc.mutex.RLock()
	defer mvc.mutex.RUnlock()
	
	secret, exists := mvc.secrets[path]
	if !exists {
		return nil, fmt.Errorf("secret not found at path %s", path)
	}
	
	return secret, nil
}

// Write implements VaultClient.Write
func (mvc *mockVaultClient) Write(path string, data map[string]interface{}) (*VaultSecret, error) {
	mvc.mutex.Lock()
	defer mvc.mutex.Unlock()
	
	secret := &VaultSecret{
		Data: data,
	}
	
	mvc.secrets[path] = secret
	return secret, nil
}

// Delete implements VaultClient.Delete
func (mvc *mockVaultClient) Delete(path string) (*VaultSecret, error) {
	mvc.mutex.Lock()
	defer mvc.mutex.Unlock()
	
	delete(mvc.secrets, path)
	return nil, nil
}

// List implements VaultClient.List
func (mvc *mockVaultClient) List(path string) (*VaultSecret, error) {
	mvc.mutex.RLock()
	defer mvc.mutex.RUnlock()
	
	var keys []interface{}
	prefix := path + "/"
	
	for secretPath := range mvc.secrets {
		if strings.HasPrefix(secretPath, prefix) {
			relativePath := strings.TrimPrefix(secretPath, prefix)
			if !strings.Contains(relativePath, "/") {
				keys = append(keys, relativePath)
			}
		}
	}
	
	return &VaultSecret{
		Data: map[string]interface{}{
			"keys": keys,
		},
	}, nil
}

// SetToken implements VaultClient.SetToken
func (mvc *mockVaultClient) SetToken(token string) {
	mvc.token = token
}

// Token implements VaultClient.Token
func (mvc *mockVaultClient) Token() string {
	return mvc.token
}

// FileSecretSource implements SecretSource for file-based secrets
type FileSecretSource struct {
	basePath string
}

// NewFileSecretSource creates a new file-based secret source
func NewFileSecretSource(basePath string) *FileSecretSource {
	return &FileSecretSource{
		basePath: basePath,
	}
}

// GetSecret retrieves a secret from file
func (fss *FileSecretSource) GetSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	filePath := fmt.Sprintf("%s/%s", fss.basePath, path)
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret file %s: %w", filePath, err)
	}
	
	return map[string]interface{}{
		"value": string(data),
	}, nil
}

// GetSecretValue retrieves a specific value from a secret
func (fss *FileSecretSource) GetSecretValue(ctx context.Context, path, key string) (string, error) {
	data, err := fss.GetSecret(ctx, path)
	if err != nil {
		return "", err
	}
	
	value, exists := data[key]
	if !exists {
		return "", fmt.Errorf("key %s not found in secret", key)
	}
	
	stringValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("value is not a string")
	}
	
	return stringValue, nil
}

// PutSecret stores a secret to file
func (fss *FileSecretSource) PutSecret(ctx context.Context, path string, data map[string]interface{}) error {
	return fmt.Errorf("file secret source does not support writing secrets")
}

// DeleteSecret deletes a secret file
func (fss *FileSecretSource) DeleteSecret(ctx context.Context, path string) error {
	return fmt.Errorf("file secret source does not support deleting secrets")
}

// ListSecrets lists secret files
func (fss *FileSecretSource) ListSecrets(ctx context.Context, path string) ([]string, error) {
	return nil, fmt.Errorf("file secret source does not support listing secrets")
}
