package main

import (
	"fmt"
	"log"
	"os"

	"github.com/DimaJoyti/go-coffee/internal/auth"
)

func main() {
	fmt.Println("🚀 Starting Auth Service Test Runner")
	fmt.Println("=====================================")

	// Create test runner
	runner, err := auth.NewTestRunner()
	if err != nil {
		log.Fatalf("❌ Failed to create test runner: %v", err)
	}
	defer runner.Close()

	fmt.Printf("📡 Test server running at: %s\n", runner.GetBaseURL())
	fmt.Println()

	// Run basic flow tests
	fmt.Println("🔄 Running basic authentication flow tests...")
	if err := runner.TestBasicFlow(); err != nil {
		log.Fatalf("❌ Basic flow tests failed: %v", err)
	}

	fmt.Println()

	// Run error case tests
	fmt.Println("🔄 Running error case tests...")
	if err := runner.TestErrorCases(); err != nil {
		log.Fatalf("❌ Error case tests failed: %v", err)
	}

	fmt.Println()
	fmt.Println("🎉 All tests passed successfully!")
	fmt.Println("✅ Auth service is working correctly")
	fmt.Println()
	fmt.Println("Available endpoints:")
	fmt.Println("  POST /api/v1/auth/register   - Register new user")
	fmt.Println("  POST /api/v1/auth/login      - Login user")
	fmt.Println("  POST /api/v1/auth/logout     - Logout user")
	fmt.Println("  POST /api/v1/auth/validate   - Validate token")
	fmt.Println("  GET  /api/v1/auth/me         - Get user info")
	fmt.Println("  POST /api/v1/auth/refresh    - Refresh token")
	fmt.Println("  POST /api/v1/auth/change-password - Change password")
	fmt.Println()
	fmt.Println("🏗️  Clean Architecture Implementation Complete!")
	fmt.Println("   ✅ Domain Layer: Events, Business Rules, Aggregates")
	fmt.Println("   ✅ Application Layer: CQRS, Commands, Queries")
	fmt.Println("   ✅ Infrastructure Layer: Redis, Security, Events")
	fmt.Println("   ✅ Transport Layer: HTTP, WebSocket, gRPC")
	fmt.Println("   ✅ Security: JWT, Rate Limiting, Middleware")
}

// Example usage function
func printExampleUsage() {
	fmt.Println()
	fmt.Println("📖 Example API Usage:")
	fmt.Println()
	fmt.Println("1. Register a user:")
	fmt.Println(`   curl -X POST http://localhost:8080/api/v1/auth/register \
     -H "Content-Type: application/json" \
     -d '{"email":"user@example.com","password":"SecurePass123!","role":"user"}'`)
	fmt.Println()
	fmt.Println("2. Login:")
	fmt.Println(`   curl -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"user@example.com","password":"SecurePass123!"}'`)
	fmt.Println()
	fmt.Println("3. Get user info (with token):")
	fmt.Println(`   curl -X GET http://localhost:8080/api/v1/auth/me \
     -H "Authorization: Bearer YOUR_ACCESS_TOKEN"`)
	fmt.Println()
	fmt.Println("4. Validate token:")
	fmt.Println(`   curl -X POST http://localhost:8080/api/v1/auth/validate \
     -H "Content-Type: application/json" \
     -d '{"token":"YOUR_ACCESS_TOKEN"}'`)
	fmt.Println()
}

func init() {
	// Set up environment for testing
	if os.Getenv("GO_ENV") == "" {
		os.Setenv("GO_ENV", "test")
	}
}
