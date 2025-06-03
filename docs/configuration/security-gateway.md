# Security Gateway Configuration Reference

## Overview

This document provides a comprehensive reference for configuring the Security Gateway service. The configuration supports multiple formats (YAML, JSON, environment variables) and includes detailed explanations for all available options.

## Configuration Sources

The Security Gateway loads configuration from multiple sources in the following order of precedence:

1. **Command-line flags** (highest priority)
2. **Environment variables**
3. **Configuration files** (YAML/JSON)
4. **Default values** (lowest priority)

## Configuration File Structure

### Main Configuration File (`config/config.yaml`)

```yaml
# =============================================================================
# SERVER CONFIGURATION
# =============================================================================
server:
  # Server listening configuration
  port: 8080                    # Server port (env: SERVER_PORT)
  host: "0.0.0.0"              # Server host (env: SERVER_HOST)
  
  # Timeout settings
  read_timeout: "30s"           # Request read timeout
  write_timeout: "30s"          # Response write timeout
  idle_timeout: "120s"          # Connection idle timeout
  
  # TLS configuration
  tls:
    enabled: false              # Enable TLS/HTTPS
    cert_file: ""               # Path to TLS certificate
    key_file: ""                # Path to TLS private key
    min_version: "1.2"          # Minimum TLS version (1.0, 1.1, 1.2, 1.3)

# =============================================================================
# LOGGING CONFIGURATION
# =============================================================================
logging:
  level: "info"                 # Log level: debug, info, warn, error
  format: "json"                # Log format: json, text
  output: "stdout"              # Log output: stdout, stderr, file
  file_path: "/var/log/security-gateway.log"  # Log file path (if output=file)
  max_size: 100                 # Max log file size in MB
  max_backups: 5                # Number of log file backups
  max_age: 30                   # Max age of log files in days
  compress: true                # Compress old log files

# =============================================================================
# REDIS CONFIGURATION
# =============================================================================
redis:
  # Connection settings
  url: "redis://localhost:6379" # Redis connection URL (env: REDIS_URL)
  db: 0                         # Redis database number
  password: ""                  # Redis password (env: REDIS_PASSWORD)
  
  # Connection pool settings
  pool_size: 10                 # Maximum number of connections
  min_idle_conns: 5             # Minimum idle connections
  max_retries: 3                # Maximum retry attempts
  
  # Timeout settings
  dial_timeout: "5s"            # Connection timeout
  read_timeout: "3s"            # Read timeout
  write_timeout: "3s"           # Write timeout
  pool_timeout: "4s"            # Pool timeout
  idle_timeout: "5m"            # Idle connection timeout
  idle_check_frequency: "1m"    # Idle connection check frequency

# =============================================================================
# SECURITY CONFIGURATION
# =============================================================================
security:
  # Encryption settings
  encryption:
    aes_key: ""                 # AES-256 encryption key (env: AES_KEY)
    rsa_private_key: ""         # RSA private key (env: RSA_PRIVATE_KEY)
    rsa_public_key: ""          # RSA public key (env: RSA_PUBLIC_KEY)
  
  # JWT settings
  jwt:
    secret: ""                  # JWT signing secret (env: JWT_SECRET)
    expiry: "24h"               # JWT token expiry
    refresh_expiry: "720h"      # Refresh token expiry
    issuer: "go-coffee-security" # JWT issuer
    audience: "go-coffee-api"   # JWT audience

# =============================================================================
# RATE LIMITING CONFIGURATION
# =============================================================================
rate_limit:
  enabled: true                 # Enable rate limiting (env: RATE_LIMIT_ENABLED)
  
  # Global settings
  requests_per_minute: 100      # Default requests per minute
  burst_size: 20                # Burst size for token bucket
  cleanup_interval: "1m"        # Cleanup interval for expired keys
  window_size: "1m"             # Time window for rate limiting
  
  # Specific rate limits
  limits:
    # IP-based rate limiting
    ip:
      requests_per_minute: 100
      burst_size: 20
      window_size: "1m"
    
    # User-based rate limiting
    user:
      requests_per_minute: 200
      burst_size: 40
      window_size: "1m"
    
    # Endpoint-specific rate limiting
    endpoint:
      requests_per_minute: 50
      burst_size: 10
      window_size: "1m"
    
    # Global rate limiting
    global:
      requests_per_minute: 10000
      burst_size: 1000
      window_size: "1m"

# =============================================================================
# WEB APPLICATION FIREWALL (WAF) CONFIGURATION
# =============================================================================
waf:
  enabled: true                 # Enable WAF protection (env: WAF_ENABLED)
  
  # Request filtering
  max_request_size: 10485760    # Maximum request size (10MB)
  max_header_size: 8192         # Maximum header size (8KB)
  max_url_length: 2048          # Maximum URL length
  
  # Geographic filtering
  geo_blocking:
    enabled: true               # Enable geo-blocking
    allowed_countries: ["US", "CA", "GB", "DE", "FR", "UA"]
    blocked_countries: ["CN", "RU", "KP", "IR"]
    default_action: "allow"     # Default action for unlisted countries
  
  # Bot detection
  bot_detection:
    enabled: true               # Enable bot detection
    block_malicious_bots: true  # Block known malicious bots
    challenge_suspicious: true  # Challenge suspicious bots
    whitelist_good_bots: true   # Whitelist known good bots
  
  # Rule configuration
  rules:
    # SQL Injection protection
    sql_injection:
      enabled: true
      action: "block"           # block, log, challenge
      sensitivity: "medium"     # low, medium, high
    
    # XSS protection
    xss:
      enabled: true
      action: "block"
      sensitivity: "medium"
    
    # Path traversal protection
    path_traversal:
      enabled: true
      action: "block"
      sensitivity: "high"
    
    # Command injection protection
    command_injection:
      enabled: true
      action: "block"
      sensitivity: "high"
    
    # LDAP injection protection
    ldap_injection:
      enabled: true
      action: "block"
      sensitivity: "medium"

# =============================================================================
# INPUT VALIDATION CONFIGURATION
# =============================================================================
validation:
  # General validation settings
  max_input_length: 10000       # Maximum input length
  strict_mode: true             # Enable strict validation mode
  enable_sanitization: true     # Enable input sanitization
  
  # File upload validation
  file_upload:
    max_file_size: 10485760     # Maximum file size (10MB)
    allowed_types: ["jpg", "jpeg", "png", "gif", "pdf", "doc", "docx"]
    scan_for_malware: true      # Enable malware scanning
  
  # URL validation
  url_validation:
    allow_private_ips: false    # Allow private IP addresses
    allow_localhost: false      # Allow localhost URLs
    max_redirects: 3            # Maximum redirect follows
    timeout: "10s"              # URL validation timeout

# =============================================================================
# FRAUD DETECTION CONFIGURATION
# =============================================================================
fraud_detection:
  enabled: true                 # Enable fraud detection (env: FRAUD_DETECTION_ENABLED)
  
  # Analysis settings
  velocity_checks: true         # Enable velocity analysis
  amount_analysis: true         # Enable amount analysis
  location_analysis: true       # Enable location analysis
  device_analysis: true         # Enable device analysis
  behavior_analysis: true       # Enable behavior analysis
  
  # Thresholds
  velocity_window: "1h"         # Time window for velocity checks
  max_transactions_per_hour: 10 # Maximum transactions per hour
  suspicious_amount_multiplier: 3.0  # Multiplier for suspicious amounts
  max_distance_km: 1000         # Maximum distance for location analysis
  min_time_between_locations: "30m"  # Minimum time between locations
  
  # Risk scoring
  risk_thresholds:
    low: 0.4                    # Low risk threshold
    medium: 0.6                 # Medium risk threshold
    high: 0.8                   # High risk threshold
  
  # Machine learning model
  model:
    version: "v2.1.0"           # Model version
    update_interval: "24h"      # Model update interval
    confidence_threshold: 0.8   # Minimum confidence threshold

# =============================================================================
# MONITORING CONFIGURATION
# =============================================================================
monitoring:
  # General monitoring settings
  enabled: true                 # Enable monitoring
  real_time_monitoring: true    # Enable real-time monitoring
  threat_intelligence: true     # Enable threat intelligence
  
  # Event retention
  event_retention: "720h"       # Event retention period (30 days)
  max_events_per_minute: 1000   # Maximum events per minute
  
  # Metrics collection
  metrics:
    enabled: true               # Enable metrics collection
    interval: "15s"             # Metrics collection interval
    retention: "168h"           # Metrics retention period (7 days)
  
  # Alerting
  alerting:
    enabled: true               # Enable alerting
    
    # Alert thresholds
    thresholds:
      failed_login_attempts: 5  # Failed login attempts threshold
      suspicious_ip_requests: 100  # Suspicious IP requests threshold
      high_risk_events: 10      # High risk events threshold
      time_window: "5m"         # Time window for thresholds
      critical_event_threshold: 1  # Critical event threshold
    
    # Notification channels
    channels:
      email:
        enabled: true
        smtp_host: "smtp.gmail.com"
        smtp_port: 587
        username: ""            # SMTP username (env: SMTP_USERNAME)
        password: ""            # SMTP password (env: SMTP_PASSWORD)
        from_email: "security@go-coffee.com"
        to_emails: ["admin@go-coffee.com"]
      
      slack:
        enabled: false
        webhook_url: ""         # Slack webhook URL (env: SLACK_WEBHOOK_URL)
        channel: "#security-alerts"
      
      webhook:
        enabled: false
        url: ""                 # Webhook URL (env: WEBHOOK_URL)
        secret: ""              # Webhook secret (env: WEBHOOK_SECRET)

# =============================================================================
# CORS CONFIGURATION
# =============================================================================
cors:
  enabled: true                 # Enable CORS
  allowed_origins: ["*"]        # Allowed origins
  allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"]
  allowed_headers: ["Origin", "Content-Type", "Authorization", "X-Requested-With"]
  exposed_headers: ["Content-Length", "X-Request-ID", "X-Rate-Limit-Remaining"]
  allow_credentials: true       # Allow credentials
  max_age: 86400               # Preflight cache duration (24 hours)

# =============================================================================
# SERVICE DISCOVERY CONFIGURATION
# =============================================================================
services:
  # Backend service URLs
  auth_service:
    url: "http://localhost:8081"  # Auth service URL (env: AUTH_SERVICE_URL)
    timeout: "30s"              # Request timeout
    retries: 3                  # Retry attempts
    circuit_breaker:
      enabled: true             # Enable circuit breaker
      failure_threshold: 5      # Failure threshold
      recovery_timeout: "30s"   # Recovery timeout
      max_requests: 3           # Max requests in half-open state
  
  order_service:
    url: "http://localhost:8082"  # Order service URL (env: ORDER_SERVICE_URL)
    timeout: "30s"
    retries: 3
    circuit_breaker:
      enabled: true
      failure_threshold: 5
      recovery_timeout: "30s"
      max_requests: 3
  
  payment_service:
    url: "http://localhost:8083"  # Payment service URL (env: PAYMENT_SERVICE_URL)
    timeout: "30s"
    retries: 3
    circuit_breaker:
      enabled: true
      failure_threshold: 5
      recovery_timeout: "30s"
      max_requests: 3
  
  user_service:
    url: "http://localhost:8084"  # User service URL (env: USER_SERVICE_URL)
    timeout: "30s"
    retries: 3
    circuit_breaker:
      enabled: true
      failure_threshold: 5
      recovery_timeout: "30s"
      max_requests: 3

# =============================================================================
# HEALTH CHECK CONFIGURATION
# =============================================================================
health:
  enabled: true                 # Enable health checks
  interval: "30s"               # Health check interval
  timeout: "5s"                 # Health check timeout
  
  # External dependencies
  dependencies:
    redis:
      enabled: true             # Check Redis health
      timeout: "3s"
    
    services:
      enabled: true             # Check backend services health
      timeout: "5s"

# =============================================================================
# PERFORMANCE CONFIGURATION
# =============================================================================
performance:
  # Connection settings
  max_concurrent_requests: 1000 # Maximum concurrent requests
  request_timeout: "30s"        # Request timeout
  idle_connection_timeout: "90s" # Idle connection timeout
  max_idle_connections: 100     # Maximum idle connections
  max_connections_per_host: 10  # Maximum connections per host
  
  # Caching
  cache:
    enabled: true               # Enable caching
    ttl: "5m"                   # Cache TTL
    max_size: 1000              # Maximum cache entries
  
  # Compression
  compression:
    enabled: true               # Enable response compression
    level: 6                    # Compression level (1-9)
    min_size: 1024              # Minimum size to compress

# =============================================================================
# DEVELOPMENT CONFIGURATION
# =============================================================================
development:
  # Debug settings
  debug: false                  # Enable debug mode (env: DEBUG)
  enable_debug_endpoints: false # Enable debug endpoints
  enable_swagger: true          # Enable Swagger documentation
  enable_pprof: false           # Enable pprof profiling
  
  # Testing
  enable_test_endpoints: false  # Enable test endpoints
  test_mode: false              # Enable test mode
```

## Environment Variables

All configuration options can be overridden using environment variables. The naming convention follows the pattern: `SECTION_SUBSECTION_OPTION` in uppercase.

### Core Environment Variables

```bash
# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
READ_TIMEOUT=30s
WRITE_TIMEOUT=30s
IDLE_TIMEOUT=120s

# TLS Configuration
TLS_ENABLED=false
TLS_CERT_FILE=/path/to/cert.pem
TLS_KEY_FILE=/path/to/key.pem
TLS_MIN_VERSION=1.2

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=stdout
LOG_FILE_PATH=/var/log/security-gateway.log

# Redis Configuration
REDIS_URL=redis://localhost:6379
REDIS_DB=0
REDIS_PASSWORD=
REDIS_POOL_SIZE=10
REDIS_MAX_RETRIES=3

# Security Configuration
AES_KEY=your-base64-encoded-aes-key
RSA_PRIVATE_KEY=your-rsa-private-key
RSA_PUBLIC_KEY=your-rsa-public-key
JWT_SECRET=your-jwt-secret
JWT_EXPIRY=24h

# Rate Limiting Configuration
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MINUTE=100
RATE_LIMIT_BURST_SIZE=20
RATE_LIMIT_CLEANUP_INTERVAL=1m

# WAF Configuration
WAF_ENABLED=true
WAF_MAX_REQUEST_SIZE=10485760
WAF_GEO_BLOCKING_ENABLED=true
WAF_BOT_DETECTION_ENABLED=true

# Fraud Detection Configuration
FRAUD_DETECTION_ENABLED=true
FRAUD_VELOCITY_CHECKS=true
FRAUD_AMOUNT_ANALYSIS=true
FRAUD_LOCATION_ANALYSIS=true

# Monitoring Configuration
MONITORING_ENABLED=true
MONITORING_REAL_TIME_MONITORING=true
MONITORING_EVENT_RETENTION=720h

# Service URLs
AUTH_SERVICE_URL=http://localhost:8081
ORDER_SERVICE_URL=http://localhost:8082
PAYMENT_SERVICE_URL=http://localhost:8083
USER_SERVICE_URL=http://localhost:8084

# Notification Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/your/webhook
WEBHOOK_URL=https://your-webhook-endpoint.com
WEBHOOK_SECRET=your-webhook-secret
```

## Configuration Validation

The Security Gateway validates all configuration options on startup. Invalid configurations will prevent the service from starting and display detailed error messages.

### Common Configuration Errors

1. **Invalid Redis URL**: Ensure Redis URL format is correct
2. **Missing encryption keys**: AES and RSA keys are required
3. **Invalid time durations**: Use Go duration format (e.g., "30s", "5m", "1h")
4. **Invalid log level**: Must be one of: debug, info, warn, error
5. **Invalid service URLs**: Backend service URLs must be valid HTTP/HTTPS URLs

## Configuration Examples

### Production Configuration

```yaml
server:
  port: 8080
  host: "0.0.0.0"
  tls:
    enabled: true
    cert_file: "/etc/ssl/certs/security-gateway.crt"
    key_file: "/etc/ssl/private/security-gateway.key"

logging:
  level: "info"
  format: "json"
  output: "file"
  file_path: "/var/log/security-gateway.log"

redis:
  url: "redis://redis-cluster:6379"
  password: "production-redis-password"
  pool_size: 20

rate_limit:
  enabled: true
  requests_per_minute: 1000
  burst_size: 100

waf:
  enabled: true
  geo_blocking:
    enabled: true
    blocked_countries: ["CN", "RU", "KP", "IR", "SY"]

monitoring:
  enabled: true
  real_time_monitoring: true
  alerting:
    enabled: true
    channels:
      email:
        enabled: true
        to_emails: ["security@company.com", "ops@company.com"]
      slack:
        enabled: true
        webhook_url: "https://hooks.slack.com/services/..."
```

### Development Configuration

```yaml
server:
  port: 8080
  host: "localhost"

logging:
  level: "debug"
  format: "text"
  output: "stdout"

redis:
  url: "redis://localhost:6379"
  db: 1

rate_limit:
  enabled: true
  requests_per_minute: 1000

waf:
  enabled: true
  geo_blocking:
    enabled: false

development:
  debug: true
  enable_debug_endpoints: true
  enable_swagger: true
  enable_test_endpoints: true
```

## Configuration Management

### Using Configuration Files

1. **Default location**: `config/config.yaml`
2. **Custom location**: Use `--config` flag or `CONFIG_FILE` environment variable
3. **Multiple files**: Configuration can be split across multiple files

### Using Environment Variables

Environment variables take precedence over configuration files:

```bash
export RATE_LIMIT_ENABLED=false
export WAF_ENABLED=true
export LOG_LEVEL=debug
./security-gateway
```

### Using Command-line Flags

Command-line flags have the highest precedence:

```bash
./security-gateway \
  --server.port=9090 \
  --logging.level=debug \
  --rate-limit.enabled=false
```

## Security Considerations

1. **Sensitive Data**: Store sensitive configuration (passwords, keys) in environment variables or secure vaults
2. **File Permissions**: Ensure configuration files have appropriate permissions (600 or 640)
3. **Encryption**: Encrypt configuration files containing sensitive data
4. **Rotation**: Regularly rotate encryption keys and passwords
5. **Validation**: Always validate configuration before deployment

## Troubleshooting

### Common Issues

1. **Service won't start**: Check configuration validation errors in logs
2. **Redis connection failed**: Verify Redis URL and credentials
3. **High memory usage**: Adjust cache settings and connection pools
4. **Rate limiting not working**: Check Redis connectivity and configuration
5. **WAF blocking legitimate requests**: Review and adjust WAF rules

### Debug Configuration

Enable debug mode to troubleshoot configuration issues:

```yaml
development:
  debug: true
  enable_debug_endpoints: true
```

Access debug endpoints:
- `GET /debug/config` - View current configuration
- `GET /debug/health` - Detailed health information
- `GET /debug/metrics` - Internal metrics

---

For more information, see the [API Documentation](../api/security-gateway.md) and [Deployment Guide](../deployment/security-gateway.md).
