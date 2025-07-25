package testutils

import (
	"os"
	"sync"
)

var (
	baseURL string
	once    sync.Once
)

// GetTestBaseURL returns the base URL for testing
func GetTestBaseURL() string {
	once.Do(func() {
		if url := os.Getenv("TEST_BASE_URL"); url != "" {
			baseURL = url
		} else {
			baseURL = "http://localhost:8080"
		}
	})
	return baseURL
}

// SetTestBaseURL sets a custom base URL for testing
func SetTestBaseURL(url string) {
	baseURL = url
}