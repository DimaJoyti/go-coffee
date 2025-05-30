package main

import (
	"fmt"
	"log"

	aiorder "github.com/DimaJoyti/go-coffee/internal/ai-order"
)

func main() {
	fmt.Println("🧪 Testing AI Order Service build...")

	// Test creating simple service
	service := aiorder.NewSimpleService()
	if service != nil {
		log.Println("✅ AI Order Service created successfully")
	} else {
		log.Println("❌ Failed to create AI Order Service")
	}

	fmt.Println("🎉 Build test completed!")
}
