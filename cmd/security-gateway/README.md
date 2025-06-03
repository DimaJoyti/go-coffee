# Security Gateway Service

Enterprise-grade security gateway for the Go Coffee microservices platform, providing comprehensive protection against modern threats while maintaining high performance.

## 🛡️ Features

### Core Security Features
- **Web Application Firewall (WAF)** - Protection against OWASP Top 10
- **Rate Limiting** - Distributed rate limiting with Redis backend
- **Input Validation** - Comprehensive request sanitization
- **Threat Detection** - Real-time ML-powered threat analysis
- **API Gateway** - Secure request routing and load balancing
- **Security Headers** - HSTS, CSP, X-Frame-Options enforcement

### Advanced Protection
- **Geo-blocking** - Country-based access control
- **Bot Detection** - Automated bot identification and blocking
- **Device Fingerprinting** - Track and analyze client devices
- **Behavioral Analysis** - Detect unusual usage patterns
- **DDoS Protection** - Distributed denial of service mitigation
- **SSL/TLS Termination** - Secure communication handling

### Monitoring & Analytics
- **Real-time Metrics** - Prometheus-compatible metrics
- **Security Events** - Comprehensive event logging
- **Threat Intelligence** - External threat feed integration
- **Alerting** - Automated incident notifications
- **Dashboards** - Grafana-based visualization

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- Redis 7.0+
- Docker & Docker Compose (optional)

### Local Development

1. **Clone and setup**:
```bash
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee
```

2. **Install dependencies**:
```bash
make -f Makefile.security-gateway deps
```

3. **Start Redis**:
```bash
make -f Makefile.security-gateway redis-start
```

4. **Run the service**:
```bash
make -f Makefile.security-gateway run-dev
```

### Docker Deployment

1. **Start all services**:
```bash
docker-compose -f docker-compose.security-gateway.yml up -d
```

2. **Check health**:
```bash
curl http://localhost:8080/health
```

## 📋 Configuration

### Environment Variables

```bash
# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Redis Configuration
REDIS_URL=redis://localhost:6379
REDIS_DB=0
REDIS_PASSWORD=

# Security Configuration
AES_KEY=your-aes-key-base64
RSA_KEY=your-rsa-private-key-pem
JWT_SECRET=your-jwt-secret

# Service URLs
AUTH_SERVICE_URL=http://localhost:8081
ORDER_SERVICE_URL=http://localhost:8082
PAYMENT_SERVICE_URL=http://localhost:8083
USER_SERVICE_URL=http://localhost:8084

# Logging
LOG_LEVEL=info
ENVIRONMENT=development
```

### Configuration File

The service uses `config/config.yaml` for detailed configuration:

```yaml
# Security Gateway Configuration
security:
  # Rate Limiting
  rate_limit:
    enabled: true
    requests_per_minute: 100
    burst_size: 20
    cleanup_interval: "1m"

  # Web Application Firewall
  waf:
    enabled: true
    block_suspicious_ip: true
    allowed_countries: ["US", "CA", "GB", "DE", "FR", "UA"]
    blocked_countries: ["CN", "RU", "KP"]
    max_request_size: 10485760 # 10MB

  # Input Validation
  validation:
    max_input_length: 10000
    strict_mode: true
    enable_sanitization: true

  # Monitoring
  monitoring:
    enable_real_time_monitoring: true
    retention_period: "720h" # 30 days
    enable_threat_intelligence: true
```

## 🔧 API Endpoints

### Health & Monitoring

```bash
# Health check
GET /health

# Metrics (Prometheus format)
GET /metrics

# Security metrics
GET /api/v1/security/metrics

# Security alerts
GET /api/v1/security/alerts
```

### Security Operations

```bash
# Validate input
POST /api/v1/security/validate
{
  "type": "email|password|url|ip|input",
  "value": "input-to-validate"
}

# Get security metrics
GET /api/v1/security/metrics

# Get security alerts
GET /api/v1/security/alerts?limit=50&status=open&severity=high
```

### Gateway Operations

```bash
# Proxy to auth service
ANY /api/v1/gateway/auth/*

# Proxy to order service
ANY /api/v1/gateway/order/*

# Proxy to payment service
ANY /api/v1/gateway/payment/*

# Proxy to user service
ANY /api/v1/gateway/user/*
```

## 🛠️ Development

### Building

```bash
# Build binary
make -f Makefile.security-gateway build

# Build for multiple platforms
make -f Makefile.security-gateway build-all

# Build Docker image
make -f Makefile.security-gateway docker-build
```

### Testing

```bash
# Run tests
make -f Makefile.security-gateway test

# Run tests with coverage
make -f Makefile.security-gateway test-coverage

# Run integration tests
make -f Makefile.security-gateway test-integration

# Run benchmarks
make -f Makefile.security-gateway benchmark
```

### Code Quality

```bash
# Format code
make -f Makefile.security-gateway fmt

# Run linter
make -f Makefile.security-gateway lint

# Security scan
make -f Makefile.security-gateway security-scan

# Vulnerability check
make -f Makefile.security-gateway vulnerability-check
```

## 📊 Monitoring

### Metrics

The service exposes Prometheus metrics at `/metrics`:

```
# Request metrics
security_gateway_total_requests
security_gateway_blocked_requests
security_gateway_allowed_requests

# Security metrics
security_gateway_threat_detections
security_gateway_waf_blocks
security_gateway_rate_limit_violations

# Performance metrics
security_gateway_request_duration_seconds
security_gateway_response_size_bytes
```

### Dashboards

Access monitoring dashboards:
- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Jaeger**: http://localhost:16686

### Alerting

Configure alerts in `monitoring/alertmanager/alertmanager.yml`:

```yaml
route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'

receivers:
- name: 'web.hook'
  webhook_configs:
  - url: 'http://localhost:5001/'
```

## 🔒 Security Features

### WAF Rules

The WAF includes protection against:

- **SQL Injection** - Database manipulation attempts
- **Cross-Site Scripting (XSS)** - Client-side code injection
- **Path Traversal** - File system access attempts
- **Command Injection** - System command execution
- **Malicious User Agents** - Known attack tools

### Rate Limiting

Multiple rate limiting strategies:

- **IP-based** - Per IP address limits
- **User-based** - Per authenticated user limits
- **Endpoint-based** - Per API endpoint limits
- **Global** - Overall system limits

### Threat Detection

Real-time threat detection includes:

- **Velocity Analysis** - Rapid request detection
- **Pattern Recognition** - Attack pattern identification
- **Anomaly Detection** - Unusual behavior detection
- **Reputation Checking** - IP/domain reputation validation

## 🚨 Incident Response

### Automated Response

The gateway can automatically:

- **Block malicious IPs** - Temporary or permanent blocking
- **Rate limit abusers** - Dynamic rate limit adjustment
- **Alert security team** - Real-time notifications
- **Log security events** - Comprehensive audit trail

### Manual Response

Security team can:

- **Review alerts** - Investigate security incidents
- **Adjust rules** - Modify WAF and rate limit rules
- **Block/unblock IPs** - Manual IP management
- **Generate reports** - Security analysis reports

## 📈 Performance

### Benchmarks

Typical performance metrics:

- **Latency**: < 5ms additional latency
- **Throughput**: 10,000+ requests/second
- **Memory**: < 100MB base memory usage
- **CPU**: < 5% additional CPU overhead

### Optimization

Performance optimization features:

- **Connection pooling** - Efficient backend connections
- **Request caching** - Cache validation results
- **Async processing** - Non-blocking security checks
- **Load balancing** - Distribute traffic efficiently

## 🔧 Troubleshooting

### Common Issues

1. **High latency**:
   - Check Redis connectivity
   - Review WAF rule complexity
   - Monitor resource usage

2. **False positives**:
   - Review WAF rules
   - Adjust sensitivity settings
   - Whitelist legitimate traffic

3. **Rate limit issues**:
   - Check rate limit configuration
   - Review Redis performance
   - Monitor request patterns

### Debug Mode

Enable debug logging:

```bash
LOG_LEVEL=debug ./security-gateway
```

### Health Checks

Monitor service health:

```bash
# Basic health
curl http://localhost:8080/health

# Detailed metrics
curl http://localhost:8080/metrics

# Security status
curl http://localhost:8080/api/v1/security/metrics
```

## 📚 Documentation

- [Security Architecture](../../docs/SECURITY-ARCHITECTURE.md)
- [API Documentation](../../docs/api/security-gateway.md)
- [Configuration Reference](../../docs/configuration/security-gateway.md)
- [Deployment Guide](../../docs/deployment/security-gateway.md)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run security scans
6. Submit a pull request

### Development Guidelines

- Follow Go best practices
- Write comprehensive tests
- Document security implications
- Update configuration examples
- Maintain backward compatibility

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](../../LICENSE) file for details.

## 🆘 Support

- **Issues**: [GitHub Issues](https://github.com/DimaJoyti/go-coffee/issues)
- **Discussions**: [GitHub Discussions](https://github.com/DimaJoyti/go-coffee/discussions)
- **Security**: security@go-coffee.com
- **Documentation**: [Wiki](https://github.com/DimaJoyti/go-coffee/wiki)

---

**Security Gateway Service** - Protecting your microservices with enterprise-grade security 🛡️
