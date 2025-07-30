# 🛠️ Go Coffee Scripts - Quick Reference Guide

## 📋 **Available Scripts**

### **🔨 Build & Compilation**
```bash
./build_all.sh [OPTIONS]           # Build all microservices
```
**Options:**
- `--core-only` - Build only production services (20 services)
- `--test-only` - Build only test services (7 services)  
- `--ai-only` - Build only AI services (8 services)
- `--help` - Show detailed help

### **🧪 Testing**
```bash
./scripts/test-all-services.sh [OPTIONS]    # Test all services
```
**Options:**
- `--core-only` - Test only core services
- `--fast` - Skip integration tests
- `--parallel` - Run tests in parallel
- `--verbose` - Detailed test output
- `--help` - Show detailed help

### **🚀 Service Management**
```bash
./scripts/start-all-services.sh [OPTIONS]   # Start all services
```
**Options:**
- `--core-only` - Start only core services
- `--dev-mode` - Development mode with hot reload
- `--production` - Production mode with optimizations
- `--monitor` - Enable continuous monitoring
- `--help` - Show detailed help

### **🏥 Health Monitoring**
```bash
./scripts/health-check.sh [OPTIONS]         # Check service health
```
**Options:**
- `--comprehensive` - Deep system analysis
- `--monitoring` - Continuous monitoring mode
- `--report` - Generate detailed JSON report
- `--environment ENV` - Specify environment
- `--help` - Show detailed help

## 🎯 **Common Workflows**

### **🔄 Development Workflow**
```bash
# 1. Build all services
./build_all.sh

# 2. Run tests
./scripts/test-all-services.sh --fast

# 3. Start services in dev mode
./scripts/start-all-services.sh --dev-mode

# 4. Check health
./scripts/health-check.sh
```

### **🚀 Production Deployment**
```bash
# 1. Build core services only
./build_all.sh --core-only

# 2. Run comprehensive tests
./scripts/test-all-services.sh --core-only --verbose

# 3. Start in production mode with monitoring
./scripts/start-all-services.sh --core-only --production --monitor

# 4. Comprehensive health check with report
./scripts/health-check.sh --comprehensive --report
```

### **🤖 AI Services Focus**
```bash
# 1. Build AI services
./build_all.sh --ai-only

# 2. Test AI services
./scripts/test-all-services.sh --ai-only --verbose

# 3. Start AI services
./scripts/start-all-services.sh --ai-only --dev-mode

# 4. Monitor AI services
./scripts/health-check.sh --monitoring
```

## 📊 **Service Categories**

### **Core Production Services (20)**
- **AI Services**: ai-search, ai-service, ai-arbitrage-service, ai-order-service
- **Infrastructure**: auth-service, communication-hub, user-gateway, security-gateway  
- **Business Logic**: kitchen-service, order-service, payment-service, api-gateway
- **External Integration**: market-data-service, defi-service, bright-data-hub-service
- **AI Orchestration**: llm-orchestrator, llm-orchestrator-simple, mcp-ai-integration
- **Supporting**: redis-mcp-server, task-cli

### **Test Services (7)**
- ai-arbitrage-demo, auth-test, config-test, test-server
- simple-auth, redis-mcp-demo, llm-orchestrator-minimal

## 🔧 **Troubleshooting**

### **Build Issues**
```bash
# Check Go version
go version

# Update dependencies
go mod tidy

# Debug single service
go build -v cmd/SERVICE/main.go

# Verbose build output
DEBUG=true ./build_all.sh
```

### **Service Issues**
```bash
# Check service logs
tail -f logs/SERVICE.log

# Check port usage
netstat -tulpn | grep PORT

# Restart specific service
./scripts/start-all-services.sh --core-only

# Health check with details
./scripts/health-check.sh --comprehensive
```

### **Test Issues**
```bash
# Run single service tests
go test -v ./cmd/SERVICE/...

# Check test coverage
go tool cover -html=coverage/SERVICE.out

# Verbose test output
./scripts/test-all-services.sh --verbose
```

## 📁 **File Structure**
```
scripts/
├── lib/
│   └── common.sh              # Shared library functions
├── test-all-services.sh       # Comprehensive testing
├── start-all-services.sh      # Service orchestration
├── health-check.sh           # Health monitoring
└── README.md                 # This file

build_all.sh                  # Main build script (root level)
```

## 🎨 **Features**

### **✅ Enhanced Error Handling**
- Timeout protection for all operations
- Detailed error messages and troubleshooting
- Graceful failure handling with cleanup
- Exit codes for CI/CD integration

### **✅ Performance Optimizations**
- Parallel execution for builds and tests
- Optimized service startup sequences
- Resource usage monitoring
- Performance metrics tracking

### **✅ Developer Experience**
- Comprehensive help documentation
- Progress indicators and status updates
- Multiple execution modes
- Consistent output formatting

### **✅ Production Ready**
- Health monitoring and alerting
- Service dependency management
- Graceful shutdown procedures
- Comprehensive logging

## 🚀 **Quick Start**

1. **First Time Setup**:
   ```bash
   chmod +x build_all.sh scripts/*.sh
   ```

2. **Build Everything**:
   ```bash
   ./build_all.sh
   ```

3. **Start Development Environment**:
   ```bash
   ./scripts/start-all-services.sh --dev-mode
   ```

4. **Monitor Health**:
   ```bash
   ./scripts/health-check.sh --monitoring
   ```

## 📞 **Support**

For issues or questions:
1. Check the troubleshooting section above
2. Run scripts with `--help` for detailed usage
3. Use `DEBUG=true` for verbose output
4. Check service logs in `logs/` directory

## 🎯 **3: Advanced Features (NEW!)**

### **🔬 Advanced Monitoring & Observability**
```bash
./scripts/monitoring/setup-observability.sh [OPTIONS]    # Complete monitoring stack
```
**Features:**
- Prometheus, Grafana, Jaeger, Loki, AlertManager
- Custom Go Coffee dashboards and alerts
- Service discovery and auto-configuration
- Multi-environment support

### **⚡ Performance Testing & Load Analysis**
```bash
./scripts/performance/load-test.sh [OPTIONS]             # Advanced load testing
```
**Features:**
- Multi-tool support (wrk, hey, ab)
- Scenario-based testing (API, crypto, web, full)
- Load profiles (light, medium, heavy, stress, spike)
- Comprehensive reporting and analysis

### **🔒 Security Scanning & Vulnerability Assessment**
```bash
./scripts/security/security-scan.sh [OPTIONS]            # Security scanning
```
**Features:**
- Multi-scope scanning (code, deps, containers, APIs)
- Security tool integration (gosec, nancy, trivy, semgrep)
- Vulnerability classification and reporting
- CI/CD security gates

### **🚀 Advanced CI/CD Pipeline**
```bash
./scripts/ci-cd/pipeline.sh [OPTIONS] STAGE              # CI/CD automation
```
**Features:**
- 8-stage pipeline (validate, build, test, security, package, deploy, verify, notify)
- Multi-environment deployment with rollback
- Docker containerization and Kubernetes deployment
- Slack/Teams notifications

### **📚 Automated Documentation Generation**
```bash
./scripts/docs/generate-docs.sh [OPTIONS]                # Documentation automation
```
**Features:**
- Multi-format documentation (HTML, Markdown, PDF)
- API documentation with OpenAPI/Swagger
- Architecture and deployment guides
- Interactive web interface

### **🌟 Unified Ecosystem Management**
```bash
./scripts/ecosystem/manage.sh [OPTIONS] COMMAND          # Unified management
```
**Features:**
- Single interface for entire ecosystem (43+ services)
- Component orchestration and health monitoring
- Lifecycle management (start, stop, build, test, deploy)
- Environment-aware operations

## 📊 **Complete Platform Coverage**

### **All Phases Combined**
- ✅ **27 Core Microservices** (1)
- ✅ **8 Crypto Wallet Services** (2)
- ✅ **3 Web UI Services** (2)
- ✅ **5 Monitoring Services** (3)
- ✅ **Advanced Tooling & Automation** (3)

**Total: 43+ services with enterprise-grade management! 🎯**

---
**Go Coffee Microservices Platform** - Complete enterprise-grade infrastructure 🚀☕
