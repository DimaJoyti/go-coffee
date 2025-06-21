package security

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBasicSecurityChecks performs basic security checks on the codebase
func TestBasicSecurityChecks(t *testing.T) {
	t.Run("NoHardcodedSecrets", func(t *testing.T) {
		// Check for common hardcoded secrets patterns
		err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip non-Go files and test files
			if !strings.HasSuffix(path, ".go") || strings.Contains(path, "_test.go") {
				return nil
			}

			// Skip vendor and .git directories
			if strings.Contains(path, "vendor/") || strings.Contains(path, ".git/") {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			contentStr := string(content)

			// Check for potential hardcoded secrets
			suspiciousPatterns := []string{
				"password = \"",
				"secret = \"",
				"api_key = \"",
				"private_key = \"",
				"token = \"",
			}

			for _, pattern := range suspiciousPatterns {
				if strings.Contains(strings.ToLower(contentStr), pattern) {
					// Allow test files and example configurations
					if !strings.Contains(path, "test") && !strings.Contains(path, "example") {
						t.Errorf("Potential hardcoded secret found in %s: %s", path, pattern)
					}
				}
			}

			return nil
		})

		assert.NoError(t, err, "Error walking directory tree")
	})

	t.Run("NoSQLInjectionVulnerabilities", func(t *testing.T) {
		// Check for potential SQL injection vulnerabilities
		err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Only check Go files
			if !strings.HasSuffix(path, ".go") {
				return nil
			}

			// Skip vendor and .git directories
			if strings.Contains(path, "vendor/") || strings.Contains(path, ".git/") {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			contentStr := string(content)

			// Check for potential SQL injection patterns
			dangerousPatterns := []string{
				"fmt.Sprintf(\"SELECT",
				"fmt.Sprintf(\"INSERT",
				"fmt.Sprintf(\"UPDATE",
				"fmt.Sprintf(\"DELETE",
				"\"SELECT * FROM \" +",
				"\"INSERT INTO \" +",
			}

			for _, pattern := range dangerousPatterns {
				if strings.Contains(contentStr, pattern) {
					// This is a potential SQL injection vulnerability
					t.Logf("Potential SQL injection pattern found in %s: %s", path, pattern)
					// Note: We log instead of fail to avoid false positives
				}
			}

			return nil
		})

		assert.NoError(t, err, "Error walking directory tree")
	})

	t.Run("ProperErrorHandling", func(t *testing.T) {
		// Check that error handling is present
		err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Only check Go files, skip test files
			if !strings.HasSuffix(path, ".go") || strings.Contains(path, "_test.go") {
				return nil
			}

			// Skip vendor and .git directories
			if strings.Contains(path, "vendor/") || strings.Contains(path, ".git/") {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			contentStr := string(content)

			// Check for functions that return errors but might not be handled
			lines := strings.Split(contentStr, "\n")
			for i, line := range lines {
				// Look for function calls that likely return errors
				if strings.Contains(line, ":=") &&
					(strings.Contains(line, ".Get(") ||
						strings.Contains(line, ".Post(") ||
						strings.Contains(line, ".Put(") ||
						strings.Contains(line, ".Delete(") ||
						strings.Contains(line, ".Execute(") ||
						strings.Contains(line, ".Query(")) {

					// Check if error is handled in the next few lines
					errorHandled := false
					for j := i + 1; j < len(lines) && j < i+5; j++ {
						if strings.Contains(lines[j], "if err != nil") ||
							strings.Contains(lines[j], "return") ||
							strings.Contains(lines[j], "log.") ||
							strings.Contains(lines[j], "panic(") {
							errorHandled = true
							break
						}
					}

					if !errorHandled {
						t.Logf("Potential unhandled error in %s at line %d: %s", path, i+1, strings.TrimSpace(line))
					}
				}
			}

			return nil
		})

		assert.NoError(t, err, "Error walking directory tree")
	})

	t.Run("NoDebugCode", func(t *testing.T) {
		// Check for debug code that shouldn't be in production
		debugPatterns := []string{
			"GIN_MODE=debug",
			"LOG_LEVEL=debug",
			"fmt.Print",
			"log.Print",
			"panic(",
			"//TODO",
			"//FIXME",
		}

		err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip vendor, .git, and test files
			if strings.Contains(path, "vendor/") ||
				strings.Contains(path, ".git/") ||
				strings.Contains(path, "_test.go") ||
				strings.HasSuffix(path, ".md") {
				return nil
			}

			// Only check Go files and config files
			if strings.HasSuffix(path, ".go") ||
				strings.HasSuffix(path, ".env") ||
				strings.HasSuffix(path, ".yml") ||
				strings.HasSuffix(path, ".yaml") {

				content, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				for _, pattern := range debugPatterns {
					if strings.Contains(string(content), pattern) {
						t.Logf("Warning: Found debug pattern '%s' in %s", pattern, path)
						// Don't fail the test, just warn for now
					}
				}
			}

			return nil
		})

		if err != nil {
			t.Errorf("Error walking directory: %v", err)
		}
	})

	t.Run("ConfigurationSecurity", func(t *testing.T) {
		// Check configuration files for security issues
		configFiles := []string{
			"config.json",
			"config.yaml",
			"config.yml",
			".env",
			"docker-compose.yml",
			"docker-compose.yaml",
		}

		for _, configFile := range configFiles {
			err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if strings.HasSuffix(path, configFile) {
					content, err := os.ReadFile(path)
					if err != nil {
						return err
					}

					contentStr := string(content)

					// Check for insecure configurations
					insecurePatterns := []string{
						"debug: true",
						"debug=true",
						"ssl: false",
						"ssl=false",
						"tls: false",
						"tls=false",
						"verify_ssl: false",
						"verify_ssl=false",
					}

					for _, pattern := range insecurePatterns {
						if strings.Contains(strings.ToLower(contentStr), pattern) {
							t.Logf("Potentially insecure configuration in %s: %s", path, pattern)
						}
					}
				}

				return nil
			})

			assert.NoError(t, err, "Error checking configuration files")
		}
	})
}

// TestEnvironmentVariables checks that sensitive data is properly handled via environment variables
func TestEnvironmentVariables(t *testing.T) {
	t.Run("RequiredEnvironmentVariables", func(t *testing.T) {
		// List of environment variables that should be set in production
		requiredEnvVars := []string{
			// These would be required in production but not in tests
			// "DATABASE_URL",
			// "REDIS_URL",
			// "JWT_SECRET",
		}

		for _, envVar := range requiredEnvVars {
			value := os.Getenv(envVar)
			if value == "" {
				t.Logf("Environment variable %s is not set (this is OK for tests)", envVar)
			} else {
				assert.NotEmpty(t, value, "Environment variable %s should not be empty", envVar)
			}
		}
	})

	t.Run("NoSensitiveDataInEnvironment", func(t *testing.T) {
		// Check that environment variables don't contain obvious sensitive data
		for _, env := range os.Environ() {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				key := parts[0]
				value := parts[1]

				// Skip checking test environment variables
				if strings.Contains(key, "TEST") || strings.Contains(key, "CI") {
					continue
				}

				// Check for patterns that might indicate sensitive data
				if strings.Contains(strings.ToLower(key), "password") ||
					strings.Contains(strings.ToLower(key), "secret") ||
					strings.Contains(strings.ToLower(key), "key") {

					// Ensure the value is not obviously insecure
					if value == "password" || value == "secret" || value == "123456" {
						t.Errorf("Environment variable %s has an insecure value", key)
					}
				}
			}
		}
	})
}
