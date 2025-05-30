# Monitoring and Logging

This guide explains how to monitor and log the Accounts Service in a production environment.

## Monitoring

The Accounts Service exposes Prometheus metrics for monitoring. These metrics can be collected by Prometheus and visualized using Grafana.

### Available Metrics

The service exposes the following metrics:

#### HTTP Metrics

- `http_requests_total`: Total number of HTTP requests (labels: method, path, status)
- `http_request_duration_seconds`: Duration of HTTP requests in seconds (labels: method, path)

#### Database Metrics

- `database_queries_total`: Total number of database queries (labels: operation, table, status)
- `database_query_duration_seconds`: Duration of database queries in seconds (labels: operation, table)

#### Kafka Metrics

- `kafka_messages_total`: Total number of Kafka messages (labels: topic, event_type, status)
- `kafka_message_duration_seconds`: Duration of Kafka message processing in seconds (labels: topic, event_type)

### Setting Up Prometheus

1. **Deploy Prometheus to Kubernetes**

```bash
# Add Prometheus Helm repository
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Install Prometheus
helm install prometheus prometheus-community/prometheus
```

2. **Configure Prometheus to Scrape Metrics**

```yaml
# prometheus-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
    scrape_configs:
      - job_name: 'kubernetes-pods'
        kubernetes_sd_configs:
          - role: pod
        relabel_configs:
          - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
            action: keep
            regex: true
          - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
            action: replace
            target_label: __metrics_path__
            regex: (.+)
          - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
            action: replace
            regex: ([^:]+)(?::\d+)?;(\d+)
            replacement: $1:$2
            target_label: __address__
          - action: labelmap
            regex: __meta_kubernetes_pod_label_(.+)
          - source_labels: [__meta_kubernetes_namespace]
            action: replace
            target_label: kubernetes_namespace
          - source_labels: [__meta_kubernetes_pod_name]
            action: replace
            target_label: kubernetes_pod_name
```

3. **Apply the ConfigMap**

```bash
kubectl apply -f prometheus-config.yaml
```

### Setting Up Grafana

1. **Deploy Grafana to Kubernetes**

```bash
# Add Grafana Helm repository
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

# Install Grafana
helm install grafana grafana/grafana
```

2. **Configure Grafana Data Source**

Add Prometheus as a data source in Grafana:

- Name: Prometheus
- Type: Prometheus
- URL: http://prometheus-server
- Access: Server (default)

3. **Import Dashboards**

Import the following dashboards:

- [JVM Micrometer](https://grafana.com/grafana/dashboards/4701)
- [Kubernetes Cluster](https://grafana.com/grafana/dashboards/6417)
- [Accounts Service Dashboard](https://grafana.com/grafana/dashboards/12345) (custom dashboard)

### Custom Grafana Dashboard

You can create a custom Grafana dashboard for the Accounts Service with the following panels:

1. **HTTP Request Rate**
```
sum(rate(http_requests_total[5m])) by (method, path)
```

2. **HTTP Request Duration (p95)**
```
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, method, path))
```

3. **Database Query Rate**
```
sum(rate(database_queries_total[5m])) by (operation, table)
```

4. **Database Query Duration (p95)**
```
histogram_quantile(0.95, sum(rate(database_query_duration_seconds_bucket[5m])) by (le, operation, table))
```

5. **Kafka Message Rate**
```
sum(rate(kafka_messages_total[5m])) by (topic, event_type)
```

6. **Kafka Message Duration (p95)**
```
histogram_quantile(0.95, sum(rate(kafka_message_duration_seconds_bucket[5m])) by (le, topic, event_type))
```

## Logging

The Accounts Service uses structured logging with zap. Logs are output in JSON format by default, making them easy to parse and analyze with tools like Elasticsearch, Logstash, and Kibana (ELK stack).

### Log Format

The logs are structured in JSON format with the following fields:

- `timestamp`: The time the log was generated
- `level`: The log level (debug, info, warn, error, fatal)
- `logger`: The name of the logger
- `caller`: The file and line number where the log was generated
- `message`: The log message
- `stacktrace`: The stack trace (for error logs)
- Additional fields depending on the context

### Setting Up ELK Stack

1. **Deploy ELK Stack to Kubernetes**

```bash
# Add Elastic Helm repository
helm repo add elastic https://helm.elastic.co
helm repo update

# Install Elasticsearch
helm install elasticsearch elastic/elasticsearch

# Install Logstash
helm install logstash elastic/logstash

# Install Kibana
helm install kibana elastic/kibana

# Install Filebeat
helm install filebeat elastic/filebeat
```

2. **Configure Filebeat**

```yaml
# filebeat-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: filebeat-config
data:
  filebeat.yml: |
    filebeat.inputs:
    - type: container
      paths:
        - /var/log/containers/*.log
      processors:
        - add_kubernetes_metadata:
            host: ${NODE_NAME}
            matchers:
            - logs_path:
                logs_path: "/var/log/containers/"

    output.elasticsearch:
      hosts: ["elasticsearch-master:9200"]
```

3. **Apply the ConfigMap**

```bash
kubectl apply -f filebeat-config.yaml
```

### Log Levels

The service supports the following log levels:

- `debug`: Detailed information for debugging
- `info`: General information about the service
- `warn`: Warning messages
- `error`: Error messages
- `fatal`: Fatal errors that cause the service to exit

You can configure the log level using the `LOG_LEVEL` environment variable or the `logging.level` field in the configuration file.

### Development Mode

In development mode, logs are output in a more human-readable format. You can enable development mode using the `LOG_DEVELOPMENT` environment variable or the `logging.development` field in the configuration file.

### Log Encoding

You can configure the log encoding using the `LOG_ENCODING` environment variable or the `logging.encoding` field in the configuration file. The supported encodings are:

- `json`: JSON format (default)
- `console`: Human-readable format
