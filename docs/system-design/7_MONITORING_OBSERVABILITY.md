# 📊 7: Monitoring & Observability

## 📋 Overview

Master observability patterns through Go Coffee's comprehensive monitoring stack. This covers logging strategies, metrics collection, distributed tracing, alerting systems, and performance monitoring using modern observability tools.

## 🎯 Learning Objectives

By the end of this phase, you will:
- Design comprehensive logging strategies
- Implement metrics collection and monitoring systems
- Master distributed tracing for microservices
- Build effective alerting and incident response systems
- Create business metrics and analytics dashboards
- Analyze Go Coffee's observability implementation

---

## 📖 7.1 Logging Strategies & Implementation

### Core Concepts

#### Logging Levels
- **TRACE**: Very detailed information for debugging
- **DEBUG**: Detailed information for diagnosing problems
- **INFO**: General information about application flow
- **WARN**: Potentially harmful situations
- **ERROR**: Error events that allow application to continue
- **FATAL**: Very severe errors that cause application termination

#### Structured Logging
- **JSON Format**: Machine-readable log format
- **Consistent Fields**: Standardized log structure
- **Contextual Information**: Request IDs, user IDs, trace IDs
- **Searchable Logs**: Easy filtering and querying

#### Log Aggregation
- **Centralized Collection**: Single point for all logs
- **Real-time Processing**: Stream processing for immediate insights
- **Long-term Storage**: Archival and compliance requirements
- **Search and Analytics**: Fast querying and visualization

### 🔍 Go Coffee Analysis

#### Study Structured Logging Implementation

<augment_code_snippet path="pkg/logging/structured_logger.go" mode="EXCERPT">
````go
type StructuredLogger struct {
    logger     *slog.Logger
    service    string
    version    string
    deployment string
}

func NewStructuredLogger(service, version, deployment string) *StructuredLogger {
    // Create JSON handler with custom options
    opts := &slog.HandlerOptions{
        Level: slog.LevelDebug,
        ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
            // Customize timestamp format
            if a.Key == slog.TimeKey {
                return slog.Attr{
                    Key:   "timestamp",
                    Value: slog.StringValue(time.Now().UTC().Format(time.RFC3339Nano)),
                }
            }
            return a
        },
    }
    
    handler := slog.NewJSONHandler(os.Stdout, opts)
    logger := slog.New(handler)
    
    return &StructuredLogger{
        logger:     logger,
        service:    service,
        version:    version,
        deployment: deployment,
    }
}

func (sl *StructuredLogger) WithContext(ctx context.Context) *slog.Logger {
    // Extract contextual information
    attrs := []slog.Attr{
        slog.String("service", sl.service),
        slog.String("version", sl.version),
        slog.String("deployment", sl.deployment),
    }
    
    // Add request ID if available
    if requestID := GetRequestID(ctx); requestID != "" {
        attrs = append(attrs, slog.String("request_id", requestID))
    }
    
    // Add trace ID if available
    if traceID := GetTraceID(ctx); traceID != "" {
        attrs = append(attrs, slog.String("trace_id", traceID))
    }
    
    // Add user ID if available
    if userID := GetUserID(ctx); userID != "" {
        attrs = append(attrs, slog.String("user_id", userID))
    }
    
    return sl.logger.With(attrs...)
}

func (sl *StructuredLogger) LogHTTPRequest(ctx context.Context, req *http.Request, statusCode int, duration time.Duration) {
    logger := sl.WithContext(ctx)
    
    logger.Info("HTTP request processed",
        slog.String("method", req.Method),
        slog.String("path", req.URL.Path),
        slog.String("query", req.URL.RawQuery),
        slog.Int("status_code", statusCode),
        slog.Duration("duration", duration),
        slog.String("user_agent", req.UserAgent()),
        slog.String("remote_addr", req.RemoteAddr),
        slog.Int64("content_length", req.ContentLength),
    )
}

func (sl *StructuredLogger) LogBusinessEvent(ctx context.Context, event string, data map[string]interface{}) {
    logger := sl.WithContext(ctx)
    
    attrs := []slog.Attr{
        slog.String("event_type", "business"),
        slog.String("event_name", event),
        slog.Time("event_time", time.Now()),
    }
    
    // Add business data
    for key, value := range data {
        attrs = append(attrs, slog.Any(key, value))
    }
    
    logger.Info("Business event", attrs...)
}
````
</augment_code_snippet>

### 🛠️ Hands-on Exercise 7.1: Implement Advanced Logging System

#### Step 1: Create Log Aggregation Pipeline
```go
// internal/logging/aggregator.go
package logging

type LogAggregator struct {
    collectors []LogCollector
    processors []LogProcessor
    outputs    []LogOutput
    buffer     *LogBuffer
    config     *AggregatorConfig
    logger     *slog.Logger
}

type LogEntry struct {
    Timestamp   time.Time              `json:"timestamp"`
    Level       string                 `json:"level"`
    Message     string                 `json:"message"`
    Service     string                 `json:"service"`
    Version     string                 `json:"version"`
    TraceID     string                 `json:"trace_id,omitempty"`
    SpanID      string                 `json:"span_id,omitempty"`
    RequestID   string                 `json:"request_id,omitempty"`
    UserID      string                 `json:"user_id,omitempty"`
    Fields      map[string]interface{} `json:"fields,omitempty"`
    Tags        []string               `json:"tags,omitempty"`
}

type LogCollector interface {
    Collect(ctx context.Context) (<-chan *LogEntry, error)
    Stop() error
}

type LogProcessor interface {
    Process(entry *LogEntry) (*LogEntry, error)
}

type LogOutput interface {
    Write(entries []*LogEntry) error
}

func NewLogAggregator(config *AggregatorConfig) *LogAggregator {
    return &LogAggregator{
        collectors: make([]LogCollector, 0),
        processors: make([]LogProcessor, 0),
        outputs:    make([]LogOutput, 0),
        buffer:     NewLogBuffer(config.BufferSize),
        config:     config,
        logger:     slog.Default(),
    }
}

func (la *LogAggregator) Start(ctx context.Context) error {
    // Start all collectors
    logChannels := make([]<-chan *LogEntry, len(la.collectors))
    for i, collector := range la.collectors {
        ch, err := collector.Collect(ctx)
        if err != nil {
            return fmt.Errorf("failed to start collector %d: %w", i, err)
        }
        logChannels[i] = ch
    }
    
    // Merge all log channels
    mergedLogs := la.mergeLogChannels(ctx, logChannels)
    
    // Process logs
    go la.processLogs(ctx, mergedLogs)
    
    // Start batch writer
    go la.batchWriter(ctx)
    
    la.logger.Info("Log aggregator started", 
        "collectors", len(la.collectors),
        "processors", len(la.processors),
        "outputs", len(la.outputs))
    
    return nil
}

func (la *LogAggregator) processLogs(ctx context.Context, logs <-chan *LogEntry) {
    for {
        select {
        case <-ctx.Done():
            return
        case entry := <-logs:
            if entry == nil {
                continue
            }
            
            // Apply processors
            processedEntry := entry
            for _, processor := range la.processors {
                var err error
                processedEntry, err = processor.Process(processedEntry)
                if err != nil {
                    la.logger.Error("Failed to process log entry", "error", err)
                    continue
                }
                if processedEntry == nil {
                    // Entry was filtered out
                    break
                }
            }
            
            if processedEntry != nil {
                la.buffer.Add(processedEntry)
            }
        }
    }
}

func (la *LogAggregator) batchWriter(ctx context.Context) {
    ticker := time.NewTicker(la.config.FlushInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            // Flush remaining logs
            la.flushLogs()
            return
        case <-ticker.C:
            la.flushLogs()
        }
    }
}

func (la *LogAggregator) flushLogs() {
    entries := la.buffer.Flush()
    if len(entries) == 0 {
        return
    }
    
    // Write to all outputs
    for _, output := range la.outputs {
        if err := output.Write(entries); err != nil {
            la.logger.Error("Failed to write logs to output", "error", err)
        }
    }
    
    la.logger.Debug("Flushed log entries", "count", len(entries))
}

// Elasticsearch Log Output
type ElasticsearchOutput struct {
    client *elasticsearch.Client
    index  string
    logger *slog.Logger
}

func (eo *ElasticsearchOutput) Write(entries []*LogEntry) error {
    if len(entries) == 0 {
        return nil
    }
    
    // Prepare bulk request
    var buf bytes.Buffer
    for _, entry := range entries {
        // Index metadata
        meta := map[string]interface{}{
            "index": map[string]interface{}{
                "_index": fmt.Sprintf("%s-%s", eo.index, entry.Timestamp.Format("2006.01.02")),
            },
        }
        metaJSON, _ := json.Marshal(meta)
        buf.Write(metaJSON)
        buf.WriteByte('\n')
        
        // Document
        entryJSON, _ := json.Marshal(entry)
        buf.Write(entryJSON)
        buf.WriteByte('\n')
    }
    
    // Execute bulk request
    res, err := eo.client.Bulk(
        bytes.NewReader(buf.Bytes()),
        eo.client.Bulk.WithIndex(eo.index),
    )
    if err != nil {
        return fmt.Errorf("failed to execute bulk request: %w", err)
    }
    defer res.Body.Close()
    
    if res.IsError() {
        return fmt.Errorf("bulk request failed with status: %s", res.Status())
    }
    
    eo.logger.Debug("Wrote log entries to Elasticsearch", "count", len(entries))
    return nil
}
```

#### Step 2: Implement Log Processing Pipeline
```go
// internal/logging/processors.go
package logging

// Enrichment Processor - adds contextual information
type EnrichmentProcessor struct {
    geoIP     GeoIPService
    userCache UserCacheService
}

func (ep *EnrichmentProcessor) Process(entry *LogEntry) (*LogEntry, error) {
    // Add geographic information if IP address is present
    if ip, exists := entry.Fields["remote_addr"]; exists {
        if ipStr, ok := ip.(string); ok {
            if location, err := ep.geoIP.Lookup(ipStr); err == nil {
                entry.Fields["geo_country"] = location.Country
                entry.Fields["geo_city"] = location.City
                entry.Fields["geo_lat"] = location.Latitude
                entry.Fields["geo_lon"] = location.Longitude
            }
        }
    }
    
    // Enrich user information
    if userID, exists := entry.Fields["user_id"]; exists {
        if userIDStr, ok := userID.(string); ok {
            if user, err := ep.userCache.GetUser(userIDStr); err == nil {
                entry.Fields["user_tier"] = user.Tier
                entry.Fields["user_country"] = user.Country
                entry.Fields["user_signup_date"] = user.SignupDate
            }
        }
    }
    
    return entry, nil
}

// Filtering Processor - filters out unwanted logs
type FilteringProcessor struct {
    rules []FilterRule
}

type FilterRule struct {
    Field     string      `json:"field"`
    Operator  string      `json:"operator"`
    Value     interface{} `json:"value"`
    Action    string      `json:"action"` // "include" or "exclude"
}

func (fp *FilteringProcessor) Process(entry *LogEntry) (*LogEntry, error) {
    for _, rule := range fp.rules {
        if fp.matchesRule(entry, rule) {
            if rule.Action == "exclude" {
                return nil, nil // Filter out this entry
            }
        }
    }
    return entry, nil
}

func (fp *FilteringProcessor) matchesRule(entry *LogEntry, rule FilterRule) bool {
    var fieldValue interface{}
    
    switch rule.Field {
    case "level":
        fieldValue = entry.Level
    case "service":
        fieldValue = entry.Service
    case "message":
        fieldValue = entry.Message
    default:
        fieldValue = entry.Fields[rule.Field]
    }
    
    switch rule.Operator {
    case "equals":
        return fieldValue == rule.Value
    case "contains":
        if str, ok := fieldValue.(string); ok {
            if substr, ok := rule.Value.(string); ok {
                return strings.Contains(str, substr)
            }
        }
    case "regex":
        if str, ok := fieldValue.(string); ok {
            if pattern, ok := rule.Value.(string); ok {
                matched, _ := regexp.MatchString(pattern, str)
                return matched
            }
        }
    }
    
    return false
}

// Sampling Processor - reduces log volume
type SamplingProcessor struct {
    sampleRate float64
    random     *rand.Rand
}

func NewSamplingProcessor(sampleRate float64) *SamplingProcessor {
    return &SamplingProcessor{
        sampleRate: sampleRate,
        random:     rand.New(rand.NewSource(time.Now().UnixNano())),
    }
}

func (sp *SamplingProcessor) Process(entry *LogEntry) (*LogEntry, error) {
    // Always include ERROR and FATAL logs
    if entry.Level == "ERROR" || entry.Level == "FATAL" {
        return entry, nil
    }
    
    // Sample other logs based on rate
    if sp.random.Float64() < sp.sampleRate {
        return entry, nil
    }
    
    return nil, nil // Filter out
}
```

### 💡 Practice Question 7.1
**"Design a logging strategy for Go Coffee that can handle 1TB of logs per day while providing real-time search capabilities and maintaining cost efficiency."**

**Solution Framework:**
1. **Log Tiering Strategy**
   - Hot tier: Recent logs (7 days) in Elasticsearch
   - Warm tier: Medium-term logs (30 days) in compressed storage
   - Cold tier: Long-term logs (1+ years) in object storage
   - Archive tier: Compliance logs in glacier storage

2. **Processing Pipeline**
   - Real-time streaming with Kafka
   - Parallel processing with multiple workers
   - Intelligent sampling and filtering
   - Batch processing for analytics

3. **Cost Optimization**
   - Log level-based retention policies
   - Compression and deduplication
   - Intelligent indexing strategies
   - Auto-scaling based on volume

---

## 📖 7.2 Metrics Collection & Monitoring

### Core Concepts

#### Metric Types
- **Counters**: Monotonically increasing values (requests, errors)
- **Gauges**: Point-in-time values (CPU usage, memory)
- **Histograms**: Distribution of values (response times)
- **Summaries**: Quantiles and totals (P95, P99 latencies)

#### Monitoring Patterns
- **RED Method**: Rate, Errors, Duration
- **USE Method**: Utilization, Saturation, Errors
- **Four Golden Signals**: Latency, Traffic, Errors, Saturation

#### Time Series Data
- **High Cardinality**: Many unique label combinations
- **Retention Policies**: Different resolution for different time ranges
- **Aggregation**: Downsampling for long-term storage
- **Alerting**: Threshold-based and anomaly detection

### 🔍 Go Coffee Analysis

#### Study Prometheus Metrics Implementation

<augment_code_snippet path="pkg/metrics/prometheus_metrics.go" mode="EXCERPT">
````go
type PrometheusMetrics struct {
    // HTTP metrics
    httpRequestsTotal    *prometheus.CounterVec
    httpRequestDuration  *prometheus.HistogramVec
    httpRequestsInFlight *prometheus.GaugeVec
    
    // Business metrics
    ordersTotal          *prometheus.CounterVec
    orderValue           *prometheus.HistogramVec
    inventoryLevel       *prometheus.GaugeVec
    
    // System metrics
    dbConnections        *prometheus.GaugeVec
    cacheHitRate         *prometheus.GaugeVec
    queueDepth           *prometheus.GaugeVec
    
    registry *prometheus.Registry
}

func NewPrometheusMetrics() *PrometheusMetrics {
    pm := &PrometheusMetrics{
        registry: prometheus.NewRegistry(),
    }
    
    // HTTP metrics
    pm.httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status_code", "service"},
    )
    
    pm.httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
        },
        []string{"method", "endpoint", "service"},
    )
    
    pm.httpRequestsInFlight = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "http_requests_in_flight",
            Help: "Current number of HTTP requests being processed",
        },
        []string{"service"},
    )
    
    // Business metrics
    pm.ordersTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "coffee_orders_total",
            Help: "Total number of coffee orders",
        },
        []string{"shop_id", "coffee_type", "status"},
    )
    
    pm.orderValue = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "coffee_order_value_usd",
            Help:    "Coffee order value in USD",
            Buckets: []float64{1, 5, 10, 20, 50, 100, 200},
        },
        []string{"shop_id", "payment_method"},
    )
    
    // Register all metrics
    pm.registry.MustRegister(
        pm.httpRequestsTotal,
        pm.httpRequestDuration,
        pm.httpRequestsInFlight,
        pm.ordersTotal,
        pm.orderValue,
        pm.inventoryLevel,
        pm.dbConnections,
        pm.cacheHitRate,
        pm.queueDepth,
    )
    
    return pm
}

func (pm *PrometheusMetrics) RecordHTTPRequest(method, endpoint, statusCode, service string, duration time.Duration) {
    pm.httpRequestsTotal.WithLabelValues(method, endpoint, statusCode, service).Inc()
    pm.httpRequestDuration.WithLabelValues(method, endpoint, service).Observe(duration.Seconds())
}

func (pm *PrometheusMetrics) RecordOrder(shopID, coffeeType, status string, value float64, paymentMethod string) {
    pm.ordersTotal.WithLabelValues(shopID, coffeeType, status).Inc()
    pm.orderValue.WithLabelValues(shopID, paymentMethod).Observe(value)
}

func (pm *PrometheusMetrics) UpdateSystemMetrics(dbConns, cacheHitRate, queueDepth float64, service string) {
    pm.dbConnections.WithLabelValues(service).Set(dbConns)
    pm.cacheHitRate.WithLabelValues(service).Set(cacheHitRate)
    pm.queueDepth.WithLabelValues(service).Set(queueDepth)
}
````
</augment_code_snippet>

### 🛠️ Hands-on Exercise 7.2: Implement Advanced Metrics System

#### Step 1: Create Custom Metrics Collector
```go
// internal/metrics/collector.go
package metrics

type MetricsCollector struct {
    collectors map[string]Collector
    registry   *prometheus.Registry
    interval   time.Duration
    logger     *slog.Logger
}

type Collector interface {
    Name() string
    Collect() ([]Metric, error)
    Describe() []*prometheus.Desc
}

type Metric struct {
    Name      string
    Value     float64
    Labels    map[string]string
    Timestamp time.Time
    Type      MetricType
}

type MetricType string

const (
    MetricTypeCounter   MetricType = "counter"
    MetricTypeGauge     MetricType = "gauge"
    MetricTypeHistogram MetricType = "histogram"
)

// Business Metrics Collector
type BusinessMetricsCollector struct {
    orderRepo     OrderRepository
    inventoryRepo InventoryRepository
    userRepo      UserRepository
    logger        *slog.Logger
}

func (bmc *BusinessMetricsCollector) Name() string {
    return "business_metrics"
}

func (bmc *BusinessMetricsCollector) Collect() ([]Metric, error) {
    var metrics []Metric
    now := time.Now()
    
    // Collect order metrics
    orderMetrics, err := bmc.collectOrderMetrics(now)
    if err != nil {
        bmc.logger.Error("Failed to collect order metrics", "error", err)
    } else {
        metrics = append(metrics, orderMetrics...)
    }
    
    // Collect inventory metrics
    inventoryMetrics, err := bmc.collectInventoryMetrics(now)
    if err != nil {
        bmc.logger.Error("Failed to collect inventory metrics", "error", err)
    } else {
        metrics = append(metrics, inventoryMetrics...)
    }
    
    // Collect user metrics
    userMetrics, err := bmc.collectUserMetrics(now)
    if err != nil {
        bmc.logger.Error("Failed to collect user metrics", "error", err)
    } else {
        metrics = append(metrics, userMetrics...)
    }
    
    return metrics, nil
}

func (bmc *BusinessMetricsCollector) collectOrderMetrics(timestamp time.Time) ([]Metric, error) {
    var metrics []Metric
    
    // Orders by status in last hour
    orderStats, err := bmc.orderRepo.GetOrderStatsByStatus(time.Now().Add(-time.Hour))
    if err != nil {
        return nil, fmt.Errorf("failed to get order stats: %w", err)
    }
    
    for status, count := range orderStats {
        metrics = append(metrics, Metric{
            Name:      "coffee_orders_by_status",
            Value:     float64(count),
            Labels:    map[string]string{"status": status},
            Timestamp: timestamp,
            Type:      MetricTypeGauge,
        })
    }
    
    // Average order value by shop
    avgOrderValues, err := bmc.orderRepo.GetAverageOrderValueByShop(time.Now().Add(-24*time.Hour))
    if err != nil {
        return nil, fmt.Errorf("failed to get average order values: %w", err)
    }
    
    for shopID, avgValue := range avgOrderValues {
        metrics = append(metrics, Metric{
            Name:      "coffee_average_order_value",
            Value:     avgValue,
            Labels:    map[string]string{"shop_id": shopID},
            Timestamp: timestamp,
            Type:      MetricTypeGauge,
        })
    }
    
    // Order processing time percentiles
    processingTimes, err := bmc.orderRepo.GetProcessingTimePercentiles(time.Now().Add(-time.Hour))
    if err != nil {
        return nil, fmt.Errorf("failed to get processing times: %w", err)
    }
    
    for percentile, duration := range processingTimes {
        metrics = append(metrics, Metric{
            Name:      "coffee_order_processing_time_seconds",
            Value:     duration.Seconds(),
            Labels:    map[string]string{"percentile": percentile},
            Timestamp: timestamp,
            Type:      MetricTypeGauge,
        })
    }
    
    return metrics, nil
}

func (bmc *BusinessMetricsCollector) collectInventoryMetrics(timestamp time.Time) ([]Metric, error) {
    var metrics []Metric
    
    // Low stock alerts
    lowStockItems, err := bmc.inventoryRepo.GetLowStockItems()
    if err != nil {
        return nil, fmt.Errorf("failed to get low stock items: %w", err)
    }
    
    for _, item := range lowStockItems {
        metrics = append(metrics, Metric{
            Name:      "coffee_inventory_level",
            Value:     float64(item.Quantity),
            Labels: map[string]string{
                "shop_id":     item.ShopID,
                "product_id":  item.ProductID,
                "product_name": item.ProductName,
            },
            Timestamp: timestamp,
            Type:      MetricTypeGauge,
        })
    }
    
    // Inventory turnover rate
    turnoverRates, err := bmc.inventoryRepo.GetInventoryTurnoverRates()
    if err != nil {
        return nil, fmt.Errorf("failed to get turnover rates: %w", err)
    }
    
    for shopID, rate := range turnoverRates {
        metrics = append(metrics, Metric{
            Name:      "coffee_inventory_turnover_rate",
            Value:     rate,
            Labels:    map[string]string{"shop_id": shopID},
            Timestamp: timestamp,
            Type:      MetricTypeGauge,
        })
    }
    
    return metrics, nil
}
```

#### Step 2: Implement Alerting System
```go
// internal/alerting/alert_manager.go
package alerting

type AlertManager struct {
    rules       []AlertRule
    evaluator   RuleEvaluator
    notifiers   []Notifier
    storage     AlertStorage
    logger      *slog.Logger
}

type AlertRule struct {
    ID          string        `json:"id"`
    Name        string        `json:"name"`
    Query       string        `json:"query"`
    Condition   string        `json:"condition"`
    Threshold   float64       `json:"threshold"`
    Duration    time.Duration `json:"duration"`
    Severity    Severity      `json:"severity"`
    Labels      map[string]string `json:"labels"`
    Annotations map[string]string `json:"annotations"`
    Enabled     bool          `json:"enabled"`
}

type Severity string

const (
    SeverityCritical Severity = "critical"
    SeverityWarning  Severity = "warning"
    SeverityInfo     Severity = "info"
)

type Alert struct {
    ID          string            `json:"id"`
    RuleID      string            `json:"rule_id"`
    Name        string            `json:"name"`
    Severity    Severity          `json:"severity"`
    Status      AlertStatus       `json:"status"`
    Value       float64           `json:"value"`
    Threshold   float64           `json:"threshold"`
    Labels      map[string]string `json:"labels"`
    Annotations map[string]string `json:"annotations"`
    StartsAt    time.Time         `json:"starts_at"`
    EndsAt      *time.Time        `json:"ends_at,omitempty"`
    UpdatedAt   time.Time         `json:"updated_at"`
}

type AlertStatus string

const (
    AlertStatusFiring   AlertStatus = "firing"
    AlertStatusResolved AlertStatus = "resolved"
)

func (am *AlertManager) EvaluateRules(ctx context.Context) error {
    for _, rule := range am.rules {
        if !rule.Enabled {
            continue
        }
        
        // Evaluate rule
        result, err := am.evaluator.Evaluate(ctx, rule.Query)
        if err != nil {
            am.logger.Error("Failed to evaluate rule", 
                "rule_id", rule.ID, 
                "rule_name", rule.Name, 
                "error", err)
            continue
        }
        
        // Check condition
        shouldFire := am.checkCondition(rule.Condition, result.Value, rule.Threshold)
        
        // Get existing alert
        existingAlert, err := am.storage.GetActiveAlert(rule.ID)
        if err != nil && !errors.Is(err, ErrAlertNotFound) {
            am.logger.Error("Failed to get existing alert", "rule_id", rule.ID, "error", err)
            continue
        }
        
        if shouldFire {
            if existingAlert == nil {
                // Create new alert
                alert := &Alert{
                    ID:          uuid.New().String(),
                    RuleID:      rule.ID,
                    Name:        rule.Name,
                    Severity:    rule.Severity,
                    Status:      AlertStatusFiring,
                    Value:       result.Value,
                    Threshold:   rule.Threshold,
                    Labels:      rule.Labels,
                    Annotations: rule.Annotations,
                    StartsAt:    time.Now(),
                    UpdatedAt:   time.Now(),
                }
                
                if err := am.storage.CreateAlert(alert); err != nil {
                    am.logger.Error("Failed to create alert", "error", err)
                    continue
                }
                
                // Send notifications
                am.sendNotifications(alert)
                
                am.logger.Warn("Alert fired", 
                    "alert_id", alert.ID,
                    "rule_name", rule.Name,
                    "value", result.Value,
                    "threshold", rule.Threshold)
                
            } else {
                // Update existing alert
                existingAlert.Value = result.Value
                existingAlert.UpdatedAt = time.Now()
                
                if err := am.storage.UpdateAlert(existingAlert); err != nil {
                    am.logger.Error("Failed to update alert", "error", err)
                }
            }
        } else {
            if existingAlert != nil {
                // Resolve alert
                now := time.Now()
                existingAlert.Status = AlertStatusResolved
                existingAlert.EndsAt = &now
                existingAlert.UpdatedAt = now
                
                if err := am.storage.UpdateAlert(existingAlert); err != nil {
                    am.logger.Error("Failed to resolve alert", "error", err)
                    continue
                }
                
                // Send resolution notification
                am.sendResolutionNotification(existingAlert)
                
                am.logger.Info("Alert resolved", 
                    "alert_id", existingAlert.ID,
                    "rule_name", rule.Name)
            }
        }
    }
    
    return nil
}

func (am *AlertManager) checkCondition(condition string, value, threshold float64) bool {
    switch condition {
    case "greater_than":
        return value > threshold
    case "less_than":
        return value < threshold
    case "equals":
        return value == threshold
    case "not_equals":
        return value != threshold
    default:
        return false
    }
}

func (am *AlertManager) sendNotifications(alert *Alert) {
    for _, notifier := range am.notifiers {
        go func(n Notifier) {
            if err := n.SendAlert(alert); err != nil {
                am.logger.Error("Failed to send alert notification", 
                    "notifier", n.Name(), 
                    "alert_id", alert.ID, 
                    "error", err)
            }
        }(notifier)
    }
}
```

### 💡 Practice Question 7.2
**"Design a metrics and alerting system for Go Coffee that can monitor 100+ services with sub-second alert response times while maintaining high availability."**

**Solution Framework:**
1. **Metrics Architecture**
   - Pull-based metrics collection (Prometheus)
   - Push-based for ephemeral jobs (Pushgateway)
   - Multi-dimensional time series data
   - Efficient storage and compression

2. **Alerting Strategy**
   - Hierarchical alert rules (service → team → organization)
   - Smart alert grouping and deduplication
   - Escalation policies and on-call rotation
   - Alert fatigue prevention

3. **High Availability**
   - Federated Prometheus setup
   - Alert manager clustering
   - Cross-region replication
   - Automated failover mechanisms

---

## 📖 7.3 Distributed Tracing

### Core Concepts

#### Tracing Fundamentals
- **Trace**: Complete request journey across services
- **Span**: Individual operation within a trace
- **Context Propagation**: Passing trace context between services
- **Sampling**: Reducing trace volume while maintaining visibility

#### OpenTelemetry
- **Unified Standard**: Single API for metrics, logs, and traces
- **Auto-instrumentation**: Automatic span creation
- **Manual Instrumentation**: Custom spans for business logic
- **Exporters**: Send data to various backends (Jaeger, Zipkin)

### 🔍 Go Coffee Analysis

#### Study Distributed Tracing Implementation

<augment_code_snippet path="pkg/tracing/opentelemetry.go" mode="EXCERPT">
````go
func (ot *OpenTelemetryTracer) TraceHTTPHandler(handler http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract trace context from headers
        ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
        
        // Start new span
        spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
        ctx, span := ot.tracer.Start(ctx, spanName,
            trace.WithSpanKind(trace.SpanKindServer),
            trace.WithAttributes(
                semconv.HTTPMethodKey.String(r.Method),
                semconv.HTTPURLKey.String(r.URL.String()),
                semconv.HTTPUserAgentKey.String(r.UserAgent()),
                semconv.HTTPClientIPKey.String(r.RemoteAddr),
            ),
        )
        defer span.End()
        
        // Wrap response writer to capture status code
        wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}
        
        // Process request
        handler.ServeHTTP(wrapped, r.WithContext(ctx))
        
        // Add response attributes
        span.SetAttributes(
            semconv.HTTPStatusCodeKey.Int(wrapped.statusCode),
            semconv.HTTPResponseSizeKey.Int64(wrapped.bytesWritten),
        )
        
        // Set span status based on HTTP status code
        if wrapped.statusCode >= 400 {
            span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", wrapped.statusCode))
        }
    })
}
````
</augment_code_snippet>

### 💡 Practice Question 7.3
**"Design a distributed tracing strategy for Go Coffee that provides end-to-end visibility across 50+ microservices while maintaining performance and cost efficiency."**

**Solution Framework:**
1. **Sampling Strategy**
   - Head-based sampling for predictable costs
   - Tail-based sampling for error traces
   - Adaptive sampling based on service load
   - Priority sampling for critical paths

2. **Context Propagation**
   - W3C Trace Context standard
   - Baggage for cross-cutting concerns
   - Service mesh integration
   - Database query correlation

3. **Performance Optimization**
   - Async span export
   - Batch processing
   - Resource attribution
   - Intelligent span filtering

---

## 🎯 7 Completion Checklist

### Knowledge Mastery
- [ ] Understand logging strategies and structured logging
- [ ] Can design metrics collection and monitoring systems
- [ ] Know distributed tracing concepts and implementation
- [ ] Understand alerting and incident response
- [ ] Can create business metrics and analytics

### Practical Skills
- [ ] Can implement comprehensive logging pipelines
- [ ] Can build custom metrics collectors and dashboards
- [ ] Can set up distributed tracing across microservices
- [ ] Can design effective alerting strategies
- [ ] Can create observability for business metrics

### Go Coffee Analysis
- [ ] Analyzed logging and metrics implementations
- [ ] Studied distributed tracing setup and configuration
- [ ] Examined alerting rules and notification systems
- [ ] Understood observability best practices
- [ ] Identified monitoring optimization opportunities

###  Readiness
- [ ] Can design observability for large-scale systems
- [ ] Can explain monitoring trade-offs and strategies
- [ ] Can implement distributed tracing solutions
- [ ] Can handle incident response scenarios
- [ ] Can discuss observability cost optimization

---

## 🚀 Next Steps

Ready for **8: Infrastructure & DevOps**:
- Kubernetes orchestration and deployment
- CI/CD pipeline design and implementation
- Infrastructure as Code with Terraform
- Container security and optimization
- Production deployment strategies

**Excellent progress on mastering monitoring and observability! 🎉**
