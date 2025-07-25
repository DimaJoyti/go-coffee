# ☕ Go Coffee - Advanced Monitoring and Observability

## 🎯 Overview

This directory contains a comprehensive monitoring and observability stack for the Go Coffee platform, implementing industry best practices for cloud-native applications with OpenTelemetry, Prometheus, Grafana, Jaeger, and Loki.

## 🏗️ Architecture

### Observability Stack Components

```
┌─────────────────────────────────────────────────────────────┐
│                    Observability Stack                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Grafana   │  │ Prometheus  │  │ Alertmanager│         │
│  │ Dashboards  │  │   Metrics   │  │   Alerts    │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Jaeger    │  │    Loki     │  │ OpenTelemetry│        │
│  │   Tracing   │  │    Logs     │  │  Collector   │        │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │ Fluent Bit  │  │ Node Exporter│  │Kube State   │        │
│  │Log Collection│  │Infrastructure│  │ Metrics     │        │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────────────┘
```

### Data Flow

```
Go Coffee Services → OpenTelemetry Collector → Storage & Processing
                                            ↓
                  ┌─────────────────────────────────────────┐
                  │                                         │
                  ▼                 ▼                 ▼     │
            Prometheus         Jaeger           Loki        │
            (Metrics)         (Traces)         (Logs)      │
                  │                 │                 │     │
                  └─────────────────┼─────────────────┘     │
                                    ▼                       │
                                Grafana                     │
                              (Visualization)               │
                                    │                       │
                                    ▼                       │
                              Alertmanager ←────────────────┘
                              (Notifications)
```

## 📊 Key Features

### **Comprehensive Metrics Collection**
- **Application Metrics**: Request rates, latency, errors, business KPIs
- **Infrastructure Metrics**: CPU, memory, disk, network, Kubernetes resources
- **Custom Metrics**: AI agent performance, Web3 transactions, DeFi trading
- **Business Metrics**: Orders/minute, revenue, customer satisfaction

### **Distributed Tracing**
- **End-to-End Tracing**: Complete request journey across all services
- **Performance Analysis**: Identify bottlenecks and optimization opportunities
- **Error Correlation**: Link errors to specific traces and spans
- **Service Dependencies**: Visualize service interactions and dependencies

### **Centralized Logging**
- **Structured Logs**: JSON-formatted logs with consistent schema
- **Log Aggregation**: Centralized collection from all services and infrastructure
- **Log Correlation**: Link logs to traces and metrics
- **Real-Time Analysis**: Live log streaming and search capabilities

### **Intelligent Alerting**
- **Multi-Channel Notifications**: Slack, email, PagerDuty integration
- **Smart Routing**: Team-specific alert routing based on service ownership
- **Alert Correlation**: Suppress redundant alerts and group related issues
- **Escalation Policies**: Automatic escalation for critical issues

### **Rich Visualizations**
- **Business Dashboards**: Executive-level KPIs and business metrics
- **Technical Dashboards**: Service performance and infrastructure health
- **Custom Views**: Role-specific dashboards for different teams
- **Real-Time Updates**: Live data with configurable refresh intervals

## 🚀 Quick Start

### **1. Deploy Monitoring Stack**

```bash
# Make deployment script executable
chmod +x monitoring/deploy-monitoring-stack.sh

# Deploy complete monitoring stack
./monitoring/deploy-monitoring-stack.sh deploy

# Verify deployment
./monitoring/deploy-monitoring-stack.sh verify
```

### **2. Access Dashboards**

```bash
# Grafana (Primary Dashboard)
kubectl port-forward svc/grafana 3000:80 -n go-coffee-monitoring
# Open: http://localhost:3000 (admin/admin123)

# Prometheus (Metrics)
kubectl port-forward svc/prometheus-kube-prometheus-prometheus 9090:9090 -n go-coffee-monitoring
# Open: http://localhost:9090

# Jaeger (Tracing)
kubectl port-forward svc/jaeger-query 16686:16686 -n go-coffee-monitoring
# Open: http://localhost:16686

# Alertmanager (Alerts)
kubectl port-forward svc/alertmanager 9093:9093 -n go-coffee-monitoring
# Open: http://localhost:9093
```

### **3. Configure Alerts**

```bash
# Update Slack webhook URL
kubectl patch secret alertmanager-secrets -n go-coffee-monitoring \
  --type='json' -p='[{"op": "replace", "path": "/data/slack-api-url", "value": "'$(echo -n "YOUR_SLACK_WEBHOOK" | base64)'"}]'

# Update email settings
kubectl patch secret alertmanager-secrets -n go-coffee-monitoring \
  --type='json' -p='[{"op": "replace", "path": "/data/smtp-password", "value": "'$(echo -n "YOUR_SMTP_PASSWORD" | base64)'"}]'
```

## 📁 Directory Structure

```
monitoring/
├── opentelemetry/
│   └── otel-collector.yaml          # OpenTelemetry Collector configuration
├── prometheus/
│   └── prometheus-config.yaml       # Prometheus configuration and rules
├── grafana/
│   └── dashboards/
│       └── go-coffee-overview.json  # Business overview dashboard
├── jaeger/
│   └── jaeger-deployment.yaml       # Jaeger tracing stack
├── loki/
│   └── loki-config.yaml            # Loki log aggregation
├── alertmanager/
│   └── alertmanager-config.yaml     # Alert routing and notifications
├── deploy-monitoring-stack.sh       # Complete deployment script
└── README.md                        # This file
```

## 🔧 Configuration

### **OpenTelemetry Collector**

The OpenTelemetry Collector serves as the central hub for all observability data:

- **Receivers**: OTLP, Prometheus, Jaeger, Zipkin, Kubernetes
- **Processors**: Batch, memory limiter, resource attribution, sampling
- **Exporters**: Prometheus, Jaeger, external OTLP endpoints
- **Extensions**: Health check, pprof, zpages for debugging

### **Prometheus Configuration**

Comprehensive metrics collection with:

- **Service Discovery**: Automatic discovery of Go Coffee services
- **Recording Rules**: Pre-computed business and SLI metrics
- **Alerting Rules**: Critical, warning, and business alerts
- **Retention**: 30-day metric retention with efficient storage

### **Grafana Dashboards**

Pre-built dashboards include:

- **☕ Go Coffee Overview**: Business KPIs, orders, revenue, satisfaction
- **🏗️ Service Performance**: Latency, throughput, error rates
- **🤖 AI Agent Performance**: Model metrics, inference times
- **🌐 Web3 Metrics**: Transaction success rates, gas optimization
- **📊 Infrastructure Health**: Cluster, node, and pod metrics

### **Jaeger Tracing**

Distributed tracing configuration:

- **Sampling**: Configurable per-service sampling rates
- **Storage**: Elasticsearch backend with 7-day retention
- **UI Integration**: Links to logs and metrics from traces
- **Performance**: Optimized for high-throughput environments

### **Loki Logging**

Centralized log management:

- **Log Parsing**: Structured JSON log parsing
- **Retention**: 31-day log retention with compression
- **Alerting**: Log-based alerts for errors and security events
- **Integration**: Seamless integration with Grafana and Jaeger

### **Alertmanager**

Intelligent alert routing:

- **Team Routing**: Alerts routed to appropriate teams
- **Escalation**: Critical alerts escalated to PagerDuty
- **Suppression**: Smart alert grouping and suppression
- **Templates**: Rich notification templates for different channels

## 📈 Metrics and KPIs

### **Business Metrics**

| Metric | Description | Target |
|--------|-------------|---------|
| Orders per Minute | Coffee orders processed | > 10/min |
| Revenue per Minute | Revenue generated | > $100/min |
| Average Order Value | Average transaction value | > $8 |
| Customer Satisfaction | 5-star rating average | > 4.0 |
| AI Agent Efficiency | Task completion rate | > 90% |
| Web3 Success Rate | Blockchain transaction success | > 95% |

### **Technical Metrics**

| Metric | Description | SLI Target |
|--------|-------------|------------|
| Availability | Service uptime | 99.9% |
| Latency P99 | 99th percentile response time | < 500ms |
| Error Rate | 5xx error percentage | < 1% |
| Throughput | Requests per second | Variable |

### **Infrastructure Metrics**

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| CPU Usage | Node CPU utilization | > 80% |
| Memory Usage | Node memory utilization | > 85% |
| Disk Usage | Storage utilization | > 90% |
| Pod Restarts | Container restart count | > 5/hour |

## 🚨 Alerting Rules

### **Critical Alerts**
- Service down (1 minute)
- High error rate (> 5% for 5 minutes)
- Payment service failures
- Web3 transaction failures
- Database connection errors
- Application panics

### **Warning Alerts**
- High latency (> 1s for 5 minutes)
- High resource usage (> 80% for 10 minutes)
- AI agent efficiency drop (< 80% for 5 minutes)
- Business metric anomalies

### **Business Alerts**
- Order rate drop (< 1/min for 10 minutes)
- Revenue drop (< $10/min for 15 minutes)
- Customer satisfaction drop (< 3.0 for 30 minutes)

## 🔒 Security and Compliance

### **Data Protection**
- **Encryption**: All data encrypted in transit and at rest
- **Access Control**: RBAC for dashboard and metric access
- **Audit Logging**: Complete audit trail of all access
- **Data Retention**: Configurable retention policies

### **Privacy**
- **PII Scrubbing**: Automatic removal of sensitive data
- **Anonymization**: Customer data anonymization in metrics
- **Compliance**: GDPR and SOC2 compliance ready

## 🛠️ Maintenance

### **Regular Tasks**

```bash
# Check monitoring stack health
kubectl get pods -n go-coffee-monitoring

# View recent alerts
kubectl logs deployment/alertmanager -n go-coffee-monitoring

# Check Prometheus targets
kubectl port-forward svc/prometheus-kube-prometheus-prometheus 9090:9090 -n go-coffee-monitoring
# Visit: http://localhost:9090/targets

# Backup Grafana dashboards
kubectl get configmap grafana-dashboards -n go-coffee-monitoring -o yaml > grafana-backup.yaml
```

### **Scaling**

```bash
# Scale Prometheus for higher load
kubectl patch prometheus prometheus-kube-prometheus-prometheus -n go-coffee-monitoring \
  --type='merge' -p='{"spec":{"replicas":2}}'

# Scale Jaeger collector
kubectl scale deployment jaeger-collector --replicas=3 -n go-coffee-monitoring

# Increase storage for metrics
kubectl patch prometheus prometheus-kube-prometheus-prometheus -n go-coffee-monitoring \
  --type='merge' -p='{"spec":{"storage":{"volumeClaimTemplate":{"spec":{"resources":{"requests":{"storage":"200Gi"}}}}}}}'
```

### **Troubleshooting**

```bash
# Check OpenTelemetry Collector health
kubectl port-forward svc/otel-collector 13133:13133 -n go-coffee-monitoring
curl http://localhost:13133

# View Prometheus configuration
kubectl get prometheus prometheus-kube-prometheus-prometheus -n go-coffee-monitoring -o yaml

# Check Jaeger storage
kubectl logs statefulset/elasticsearch-master -n go-coffee-monitoring

# Debug Loki ingestion
kubectl logs statefulset/loki -n go-coffee-monitoring
```

## 🎯 Best Practices

### **Metric Design**
- Use consistent naming conventions
- Include relevant labels for filtering
- Avoid high cardinality metrics
- Implement proper sampling for traces

### **Dashboard Design**
- Focus on user journey and business impact
- Use appropriate visualization types
- Include context and drill-down capabilities
- Optimize for different screen sizes

### **Alert Design**
- Alert on symptoms, not causes
- Include runbook links in alerts
- Use appropriate severity levels
- Test alert routing regularly

### **Performance Optimization**
- Configure appropriate retention periods
- Use recording rules for expensive queries
- Implement proper sampling strategies
- Monitor monitoring system resource usage

## 📚 Additional Resources

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/)
- [Grafana Dashboard Design](https://grafana.com/docs/grafana/latest/best-practices/)
- [Jaeger Performance Tuning](https://www.jaegertracing.io/docs/latest/performance-tuning/)
- [Go Coffee Monitoring Runbooks](https://runbooks.gocoffee.dev)

---

**The Go Coffee monitoring stack provides enterprise-grade observability for modern cloud-native applications, ensuring optimal performance and reliability.** ☕📊
