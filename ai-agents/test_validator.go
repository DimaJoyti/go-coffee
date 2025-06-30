package main

import (
	"fmt"
	"go-coffee-ai-agents/internal/config"
	"log"
)

func testValidator() {
	validator := config.NewConfigValidator()

	// Create a config instance to validate
	cfg := &config.Config{}
	
	// Test the validator with some configuration
	if err := validator.Validate(cfg); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	fmt.Println("Configuration validation passed successfully")
}
