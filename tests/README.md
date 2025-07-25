# 🧪 Go Coffee Platform - Testing Framework

## 📋 Overview

This directory contains the comprehensive testing framework for the Go Coffee platform, including unit tests, integration tests, performance tests, and end-to-end tests.

## 🏗️ Testing Structure

```
tests/
├── unit/                    # Unit tests for individual services
│   ├── api-gateway/        # API Gateway unit tests
│   ├── auth-service/       # Authentication service tests
│   ├── order-service/      # Order service tests
│   ├── payment-service/    # Payment service tests
│   └── ...                 # Other service tests
├── integration/            # Integration tests
│   ├── services_test.go    # Cross-service integration tests
│   ├── database_test.go    # Database integration tests
│   └── api_test.go         # API contract tests
├── performance/            # Performance and load tests
│   ├── load_test.go        # Load testing framework
│   ├── stress_test.go      # Stress testing scenarios
│   └── benchmark_test.go   # Performance benchmarks
├── e2e/                    # End-to-end tests
│   ├── user_journey_test.go # Complete user workflows
│   ├── admin_test.go       # Admin functionality tests
│   └── api_flow_test.go    # API workflow tests
├── security/               # Security tests
│   ├── auth_test.go        # Authentication security tests
│   ├── injection_test.go   # SQL injection tests
│   └── xss_test.go         # XSS protection tests
├── testutils/              # Test utilities and helpers
│   ├── fixtures.go         # Test data fixtures
│   ├── mocks.go           # Mock implementations
│   └── helpers.go         # Test helper functions
└── README.md              # This file
```

## 🚀 Quick Start

### **Prerequisites**
```bash
# Install Go testing tools
go install github.com/onsi/ginkgo/v2/ginkgo@latest
go install github.com/onsi/gomega/...@latest
go install github.com/golang/mock/mockgen@latest

# Install test dependencies
go mod download
```

### **Running Tests**

#### **Unit Tests**
```bash
# Run all unit tests
make test-unit

# Run specific service tests
cd tests/unit/auth-service
go test -v ./...

# Run with coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### **Integration Tests**
```bash
# Start test environment
make test-env-up

# Run integration tests
make test-integration

# Clean up
make test-env-down
```

#### **Performance Tests**
```bash
# Run load tests
cd tests/performance
go test -v -run TestAPIGatewayLoad

# Run benchmarks
go test -bench=. -benchmem
```

#### **End-to-End Tests**
```bash
# Run E2E tests
make test-e2e

# Run specific E2E scenario
go test -v -run TestCompleteOrderFlow
```

## 🔧 Test Configuration

### **Environment Variables**
```bash
# Test Configuration
export TEST_ENV=local
export TEST_BASE_URL=http://localhost:8080
export TEST_DATABASE_URL=postgres://test:test@localhost:5432/gocoffee_test
export TEST_REDIS_URL=redis://localhost:6379/1

# Test Credentials
export TEST_USER_EMAIL=test@gocoffee.com
export TEST_USER_PASSWORD=TestPassword123!
export TEST_ADMIN_EMAIL=admin@gocoffee.com
export TEST_ADMIN_PASSWORD=AdminPassword123!
```

### **Test Database Setup**
```bash
# Create test database
createdb gocoffee_test

# Run test migrations
migrate -path migrations -database "postgres://test:test@localhost:5432/gocoffee_test?sslmode=disable" up

# Seed test data
go run scripts/seed-test-data.go
```

## 📊 Testing Standards

### **Unit Test Guidelines**
- **Coverage Target**: Minimum 90% code coverage
- **Test Structure**: Use table-driven tests where appropriate
- **Mocking**: Mock all external dependencies
- **Assertions**: Use testify for assertions
- **Naming**: Test functions should be descriptive

#### **Example Unit Test**
```go
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name        string
        input       *CreateUserRequest
        mockSetup   func(*MockUserRepository)
        expected    *User
        expectError bool
    }{
        {
            name: "successful user creation",
            input: &CreateUserRequest{
                Email:     "test@example.com",
                Password:  "SecurePassword123!",
                FirstName: "John",
                LastName:  "Doe",
            },
            mockSetup: func(repo *MockUserRepository) {
                repo.On("CreateUser", mock.Anything, mock.AnythingOfType("*User")).Return(nil)
            },
            expected: &User{
                Email:     "test@example.com",
                FirstName: "John",
                LastName:  "Doe",
            },
            expectError: false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := new(MockUserRepository)
            tt.mockSetup(mockRepo)
            
            service := NewUserService(mockRepo)
            result, err := service.CreateUser(context.Background(), tt.input)
            
            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected.Email, result.Email)
            }
            
            mockRepo.AssertExpectations(t)
        })
    }
}
```

### **Integration Test Guidelines**
- **Real Dependencies**: Use real databases and services
- **Test Containers**: Use testcontainers for isolated testing
- **Data Cleanup**: Clean up test data after each test
- **Parallel Execution**: Tests should be safe to run in parallel

#### **Example Integration Test**
```go
func TestOrderServiceIntegration(t *testing.T) {
    // Setup test container
    ctx := context.Background()
    postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "postgres:13",
            ExposedPorts: []string{"5432/tcp"},
            Env: map[string]string{
                "POSTGRES_DB":       "testdb",
                "POSTGRES_USER":     "test",
                "POSTGRES_PASSWORD": "test",
            },
        },
        Started: true,
    })
    require.NoError(t, err)
    defer postgres.Terminate(ctx)

    // Get connection details
    host, _ := postgres.Host(ctx)
    port, _ := postgres.MappedPort(ctx, "5432")
    
    // Setup service
    db := setupTestDatabase(host, port.Port())
    orderService := NewOrderService(db)
    
    // Test order creation
    order, err := orderService.CreateOrder(ctx, &CreateOrderRequest{
        UserID: "test-user",
        Items: []OrderItem{
            {MenuItemID: "coffee-1", Quantity: 1},
        },
    })
    
    assert.NoError(t, err)
    assert.NotEmpty(t, order.ID)
    assert.Equal(t, "pending", order.Status)
}
```

### **Performance Test Guidelines**
- **Realistic Load**: Test with realistic user loads
- **Gradual Ramp-up**: Gradually increase load
- **Multiple Scenarios**: Test different usage patterns
- **Performance Metrics**: Track response times, throughput, error rates

### **Security Test Guidelines**
- **Authentication Tests**: Test all auth mechanisms
- **Authorization Tests**: Verify access controls
- **Input Validation**: Test for injection attacks
- **Data Protection**: Verify encryption and data handling

## 📈 Test Metrics and Reporting

### **Coverage Reports**
```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Coverage by service
go test -coverprofile=auth-coverage.out ./auth-service/...
go test -coverprofile=order-coverage.out ./order-service/...
```

### **Performance Reports**
```bash
# Generate performance report
go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof

# Analyze profiles
go tool pprof cpu.prof
go tool pprof mem.prof
```

### **Test Results**
- **JUnit XML**: For CI/CD integration
- **HTML Reports**: For human-readable results
- **Metrics Export**: For monitoring dashboards

## 🔄 Continuous Integration

### **GitHub Actions Integration**
```yaml
name: Test Suite
on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: 1.21
      - run: make test-unit
      
  integration-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:13
        env:
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: 1.21
      - run: make test-integration
```

## 🛠️ Test Utilities

### **Mock Generation**
```bash
# Generate mocks for interfaces
mockgen -source=internal/repository/user.go -destination=tests/mocks/user_repository.go

# Generate mocks for all interfaces
make generate-mocks
```

### **Test Data Fixtures**
```go
// User fixtures
func CreateTestUser() *User {
    return &User{
        ID:        uuid.New().String(),
        Email:     "test@example.com",
        FirstName: "Test",
        LastName:  "User",
        CreatedAt: time.Now(),
    }
}

// Order fixtures
func CreateTestOrder(userID string) *Order {
    return &Order{
        ID:     uuid.New().String(),
        UserID: userID,
        Status: "pending",
        Items: []OrderItem{
            {MenuItemID: "coffee-1", Quantity: 1, Price: 4.50},
        },
        Total:     4.50,
        CreatedAt: time.Now(),
    }
}
```

### **Test Helpers**
```go
// HTTP test helpers
func MakeAuthenticatedRequest(t *testing.T, method, url string, body interface{}) *http.Response {
    token := GetTestAuthToken(t)
    return MakeRequestWithAuth(t, method, url, body, token)
}

// Database test helpers
func SetupTestDB(t *testing.T) *sql.DB {
    db := ConnectTestDB()
    CleanupTestData(db)
    SeedTestData(db)
    return db
}
```

## 📚 Best Practices

### **Test Organization**
- **One test file per source file**: `user.go` → `user_test.go`
- **Descriptive test names**: `TestUserService_CreateUser_WithValidData_ReturnsUser`
- **Group related tests**: Use subtests for related scenarios
- **Setup and teardown**: Use `TestMain` for test suite setup

### **Test Data Management**
- **Isolated test data**: Each test should have its own data
- **Cleanup after tests**: Remove test data to avoid interference
- **Realistic test data**: Use data that resembles production
- **Data builders**: Use builder pattern for complex test objects

### **Error Testing**
- **Test error paths**: Verify error handling works correctly
- **Specific error types**: Test for specific error conditions
- **Error messages**: Verify error messages are helpful
- **Recovery testing**: Test system recovery from errors

### **Performance Testing**
- **Baseline measurements**: Establish performance baselines
- **Regression testing**: Detect performance regressions
- **Load patterns**: Test with realistic load patterns
- **Resource monitoring**: Monitor CPU, memory, and I/O during tests

## 🎯 Quality Gates

### **Required Checks**
- ✅ All unit tests pass
- ✅ Code coverage > 90%
- ✅ Integration tests pass
- ✅ Performance tests within SLA
- ✅ Security tests pass
- ✅ No critical vulnerabilities

### **Performance SLAs**
- **API Response Time**: P95 < 500ms
- **Database Queries**: P95 < 100ms
- **Order Processing**: < 2 seconds
- **Authentication**: < 200ms
- **Search Queries**: < 300ms

---

**This testing framework ensures the Go Coffee platform maintains the highest quality standards while enabling rapid, confident development and deployment.** 🧪✅
