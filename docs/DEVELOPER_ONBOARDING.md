# 👨‍💻 Go Coffee Platform - Developer Onboarding Guide

## 🎯 Welcome to the Go Coffee Development Team

Welcome to one of the most innovative coffee technology platforms in the world! This guide will help you get up and running quickly and effectively contribute to our mission of revolutionizing the coffee industry through technology.

## 📋 Prerequisites

### **Required Knowledge**
- **Go Programming**: Intermediate to advanced level
- **Microservices Architecture**: Understanding of distributed systems
- **Docker & Kubernetes**: Container orchestration experience
- **Database Systems**: PostgreSQL, Redis experience
- **API Development**: REST, GraphQL, gRPC
- **Git Workflow**: GitFlow or similar branching strategies

### **Recommended Knowledge**
- **AI/ML Frameworks**: Python, TensorFlow, PyTorch
- **Blockchain/DeFi**: Solidity, Web3 development
- **Cloud Platforms**: GCP, AWS, Azure
- **Infrastructure as Code**: Terraform, Ansible
- **Monitoring**: Prometheus, Grafana, Jaeger

## 🛠️ Development Environment Setup

### **1. System Requirements**
```bash
# Minimum Requirements
- OS: macOS 10.15+, Ubuntu 20.04+, Windows 10 with WSL2
- RAM: 16GB (32GB recommended)
- Storage: 100GB free space
- CPU: 4 cores (8 cores recommended)
- Docker: 20.10+
- Go: 1.21+
- Node.js: 18+
- Python: 3.9+
```

### **2. Required Tools Installation**

#### **Core Development Tools**
```bash
# Install Go
curl -LO https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Install Node.js
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# Install Python
sudo apt-get install python3.9 python3.9-pip python3.9-venv

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

#### **Kubernetes Tools**
```bash
# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Install Helm
curl https://baltocdn.com/helm/signing.asc | gpg --dearmor | sudo tee /usr/share/keyrings/helm.gpg > /dev/null
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/helm.gpg] https://baltocdn.com/helm/stable/debian/ all main" | sudo tee /etc/apt/sources.list.d/helm-stable-debian.list
sudo apt-get update
sudo apt-get install helm

# Install k9s (optional but recommended)
curl -sS https://webinstall.dev/k9s | bash
```

#### **Development Utilities**
```bash
# Install useful CLI tools
go install github.com/air-verse/air@latest           # Hot reload for Go
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest  # Linting
go install github.com/swaggo/swag/cmd/swag@latest   # API documentation
go install github.com/pressly/goose/v3/cmd/goose@latest  # Database migrations

# Install database tools
sudo apt-get install postgresql-client redis-tools

# Install monitoring tools
curl -LO https://github.com/prometheus/prometheus/releases/download/v2.40.0/prometheus-2.40.0.linux-amd64.tar.gz
```

### **3. IDE Setup**

#### **VS Code (Recommended)**
```bash
# Install VS Code
wget -qO- https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor > packages.microsoft.gpg
sudo install -o root -g root -m 644 packages.microsoft.gpg /etc/apt/trusted.gpg.d/
sudo sh -c 'echo "deb [arch=amd64,arm64,armhf signed-by=/etc/apt/trusted.gpg.d/packages.microsoft.gpg] https://packages.microsoft.com/repos/code stable main" > /etc/apt/sources.list.d/vscode.list'
sudo apt update
sudo apt install code

# Essential Extensions
code --install-extension golang.go
code --install-extension ms-vscode.vscode-docker
code --install-extension ms-kubernetes-tools.vscode-kubernetes-tools
code --install-extension bradlc.vscode-tailwindcss
code --install-extension esbenp.prettier-vscode
code --install-extension ms-python.python
```

#### **GoLand (Alternative)**
- Download from JetBrains website
- Configure Go SDK and project structure
- Install Docker and Kubernetes plugins

## 🏗️ Project Structure Overview

```
go-coffee/
├── api-gateway/              # Central API gateway
├── auth-service/             # Authentication service
├── order-service/            # Order management
├── payment-service/          # Payment processing
├── kitchen-service/          # Kitchen operations
├── user-gateway/             # User management
├── security-gateway/         # Security enforcement
├── enterprise-service/       # Enterprise features
├── ai-agents/               # AI agent platform
├── ai-search/               # AI search service
├── ai-arbitrage/            # AI trading service
├── bright-data-hub/         # Data collection service
├── communication-hub/       # Communication service
├── object-detection/        # Computer vision service
├── crypto-wallet/           # Cryptocurrency wallet
├── crypto-terminal/         # Crypto payment terminal
├── dao/                     # DAO platform
├── web-ui/                  # Frontend applications
├── terraform/               # Infrastructure as Code
├── k8s/                     # Kubernetes manifests
├── monitoring/              # Monitoring configuration
├── security/                # Security policies
├── docs/                    # Documentation
├── tests/                   # Test suites
└── scripts/                 # Utility scripts
```

## 🚀 Getting Started

### **1. Repository Setup**
```bash
# Clone the repository
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee

# Set up Git hooks
cp scripts/git-hooks/* .git/hooks/
chmod +x .git/hooks/*

# Install dependencies
make install-deps

# Set up environment
cp .env.example .env
# Edit .env with your local configuration
```

### **2. Local Development Environment**
```bash
# Start local infrastructure
docker-compose -f docker-compose.dev.yml up -d

# Verify services are running
docker-compose ps

# Run database migrations
make migrate-up

# Seed test data
make seed-data

# Start development servers
make dev-start
```

### **3. Verify Setup**
```bash
# Check API Gateway
curl http://localhost:8080/health

# Check individual services
curl http://localhost:8081/health  # Auth Service
curl http://localhost:8082/health  # Order Service
curl http://localhost:8083/health  # Payment Service

# Check frontend
open http://localhost:3000
```

## 🔧 Development Workflow

### **1. Feature Development Process**

#### **Branch Strategy**
```bash
# Create feature branch
git checkout develop
git pull origin develop
git checkout -b feature/COFFEE-123-new-payment-method

# Work on feature
# ... make changes ...

# Commit changes
git add .
git commit -m "feat(payment): add cryptocurrency payment support

- Add Bitcoin and Ethereum payment processing
- Implement wallet integration
- Add transaction validation
- Update payment UI components

Closes COFFEE-123"

# Push and create PR
git push origin feature/COFFEE-123-new-payment-method
# Create PR through GitHub UI
```

#### **Code Quality Standards**
```bash
# Run linting
make lint

# Run tests
make test

# Run security checks
make security-check

# Check code coverage
make coverage

# Format code
make format
```

### **2. Testing Strategy**

#### **Unit Tests**
```bash
# Run unit tests for specific service
cd auth-service
go test ./... -v

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

#### **Integration Tests**
```bash
# Run integration tests
make test-integration

# Run specific integration test
cd tests/integration
go test -v -run TestAuthServiceIntegration
```

#### **End-to-End Tests**
```bash
# Start test environment
make test-env-up

# Run E2E tests
make test-e2e

# Clean up
make test-env-down
```

### **3. Debugging**

#### **Local Debugging**
```bash
# Debug specific service
cd auth-service
dlv debug ./cmd/auth-service

# Debug with VS Code
# Set breakpoints and use F5 to start debugging
```

#### **Container Debugging**
```bash
# Debug running container
docker exec -it go-coffee-auth-service /bin/sh

# View logs
docker logs -f go-coffee-auth-service

# Debug Kubernetes pod
kubectl exec -it auth-service-pod -- /bin/sh
kubectl logs -f auth-service-pod
```

## 📚 Key Concepts and Patterns

### **1. Service Architecture**

#### **Service Structure**
```go
// Standard service structure
service/
├── cmd/
│   └── service/
│       └── main.go          # Entry point
├── internal/
│   ├── config/              # Configuration
│   ├── handlers/            # HTTP handlers
│   ├── services/            # Business logic
│   ├── repository/          # Data access
│   ├── models/              # Data models
│   └── middleware/          # HTTP middleware
├── pkg/                     # Shared packages
├── api/                     # API definitions
├── migrations/              # Database migrations
└── tests/                   # Service tests
```

#### **Dependency Injection Pattern**
```go
// Example service initialization
func NewAuthService(
    userRepo repository.UserRepository,
    sessionRepo repository.SessionRepository,
    emailService services.EmailService,
    config *config.Config,
) *AuthService {
    return &AuthService{
        userRepo:     userRepo,
        sessionRepo:  sessionRepo,
        emailService: emailService,
        config:       config,
    }
}
```

### **2. Error Handling**

#### **Standard Error Types**
```go
// Define custom error types
type AuthError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

func (e *AuthError) Error() string {
    return e.Message
}

// Predefined errors
var (
    ErrInvalidCredentials = &AuthError{
        Code:    "INVALID_CREDENTIALS",
        Message: "Invalid email or password",
    }
    ErrAccountLocked = &AuthError{
        Code:    "ACCOUNT_LOCKED",
        Message: "Account is temporarily locked",
    }
)
```

#### **Error Response Format**
```go
// Standard error response
type ErrorResponse struct {
    Error struct {
        Code      string `json:"code"`
        Message   string `json:"message"`
        Details   string `json:"details,omitempty"`
        RequestID string `json:"request_id"`
        Timestamp string `json:"timestamp"`
    } `json:"error"`
}
```

### **3. Configuration Management**

#### **Environment-based Configuration**
```go
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Redis    RedisConfig    `mapstructure:"redis"`
    JWT      JWTConfig      `mapstructure:"jwt"`
    Email    EmailConfig    `mapstructure:"email"`
}

// Load configuration
func LoadConfig() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("./configs")
    viper.AutomaticEnv()
    
    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }
    
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, err
    }
    
    return &config, nil
}
```

### **4. Database Patterns**

#### **Repository Pattern**
```go
type UserRepository interface {
    CreateUser(ctx context.Context, user *models.User) error
    GetUserByEmail(ctx context.Context, email string) (*models.User, error)
    UpdateUser(ctx context.Context, user *models.User) error
    DeleteUser(ctx context.Context, userID string) error
}

type userRepository struct {
    db *sql.DB
}

func (r *userRepository) CreateUser(ctx context.Context, user *models.User) error {
    query := `
        INSERT INTO users (id, email, password_hash, first_name, last_name, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)`
    
    _, err := r.db.ExecContext(ctx, query,
        user.ID, user.Email, user.PasswordHash,
        user.FirstName, user.LastName, user.CreatedAt)
    
    return err
}
```

#### **Migration Management**
```bash
# Create new migration
goose -dir migrations create add_user_preferences sql

# Run migrations
goose -dir migrations postgres "user=postgres dbname=gocoffee sslmode=disable" up

# Rollback migration
goose -dir migrations postgres "user=postgres dbname=gocoffee sslmode=disable" down
```

## 🔍 Monitoring and Observability

### **1. Logging**
```go
// Structured logging with context
logger.Info("User login attempt",
    "user_id", userID,
    "ip_address", clientIP,
    "user_agent", userAgent,
    "timestamp", time.Now(),
)

// Error logging with stack trace
logger.Error("Database connection failed",
    "error", err,
    "database", "postgres",
    "host", dbHost,
)
```

### **2. Metrics**
```go
// Prometheus metrics
var (
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
        },
        []string{"method", "endpoint", "status"},
    )
    
    requestCount = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
)

// Record metrics
func recordMetrics(method, endpoint string, status int, duration time.Duration) {
    requestDuration.WithLabelValues(method, endpoint, strconv.Itoa(status)).Observe(duration.Seconds())
    requestCount.WithLabelValues(method, endpoint, strconv.Itoa(status)).Inc()
}
```

### **3. Distributed Tracing**
```go
// OpenTelemetry tracing
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
    ctx, span := otel.Tracer("auth-service").Start(ctx, "AuthService.Login")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("user.email", req.Email),
        attribute.String("request.id", getRequestID(ctx)),
    )
    
    // Business logic...
    
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }
    
    span.SetStatus(codes.Ok, "Login successful")
    return response, nil
}
```

## 🚀 Deployment and Operations

### **1. Local Deployment**
```bash
# Build all services
make build

# Build Docker images
make docker-build

# Deploy to local Kubernetes
make k8s-deploy-local

# Check deployment status
kubectl get pods -n go-coffee
```

### **2. Staging Deployment**
```bash
# Deploy to staging
make deploy-staging

# Run smoke tests
make test-staging

# Check service health
make health-check-staging
```

### **3. Production Deployment**
```bash
# Production deployment (requires approval)
make deploy-production

# Monitor deployment
make monitor-production

# Rollback if needed
make rollback-production
```

## 📖 Learning Resources

### **Internal Documentation**
- [Platform Architecture](./PLATFORM_ARCHITECTURE.md)
- [API Documentation](./api/)
- [Testing Guide](./COMPREHENSIVE_TESTING_PLAN.md)
- [Security Guidelines](./security/)

### **External Resources**
- [Go Best Practices](https://golang.org/doc/effective_go.html)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Microservices Patterns](https://microservices.io/patterns/)

### **Training Materials**
- Go Coffee Architecture Deep Dive (Internal)
- Microservices Design Patterns Workshop
- Kubernetes Operations Training
- Security Best Practices Course

## 🤝 Getting Help

### **Team Contacts**
- **Tech Lead**: @tech-lead (Slack)
- **DevOps Team**: @devops-team (Slack)
- **Security Team**: @security-team (Slack)
- **AI/ML Team**: @ai-team (Slack)

### **Communication Channels**
- **#go-coffee-dev**: General development discussions
- **#go-coffee-alerts**: Production alerts and incidents
- **#go-coffee-releases**: Release announcements
- **#go-coffee-random**: Casual team chat

### **Support Process**
1. **Check Documentation**: Start with this guide and related docs
2. **Search Slack**: Look for similar questions in team channels
3. **Ask Team**: Post in appropriate Slack channel
4. **Create Issue**: For bugs or feature requests, create GitHub issue
5. **Escalate**: Contact team leads for urgent issues

---

**Welcome to the team! We're excited to have you contribute to the future of coffee technology.** ☕🚀

For questions about this guide, contact the Tech Lead or post in #go-coffee-dev.
