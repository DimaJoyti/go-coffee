# Docker and Kubernetes for Coffee Order System

This document describes the use of Docker and Kubernetes for deploying the Coffee Order System.

## Docker

### Docker Images

The Coffee Order System consists of three main components, each with its own Docker image:

1. **Producer** (`coffee-producer`): HTTP server that receives coffee orders and sends them to Kafka.
2. **Streams Processor** (`coffee-streams`): Service that processes coffee orders using Kafka Streams.
3. **Consumer** (`coffee-consumer`): Service that consumes processed coffee orders from Kafka and executes them.

### Building Images

To build the images, use the following commands:

```bash
# Producer
cd producer
docker build -t coffee-producer:latest .

# Streams Processor
cd streams
docker build -t coffee-streams:latest .

# Consumer
cd consumer
docker build -t coffee-consumer:latest .
```

### Running with Docker Compose

For local development and testing, you can use Docker Compose:

```bash
docker-compose up -d
```

This will start all system components, including Kafka and Zookeeper.

## Kubernetes

### Deploying to Kubernetes

The Coffee Order System can be deployed to Kubernetes using Helm or standard manifests.

#### Deploying with Helm

```bash
# Add Helm repository (if needed)
helm repo add coffee-system https://example.com/helm-charts
helm repo update

# Deploy
helm install coffee-system coffee-system/coffee-system \
  --namespace coffee-system \
  --create-namespace \
  --set global.registry=your-registry/ \
  --set producer.tag=latest \
  --set consumer.tag=latest \
  --set streams.tag=latest
```

#### Deploying with kubectl

```bash
# Create namespace
kubectl apply -f kubernetes/manifests/00-namespace.yaml

# Deploy ConfigMap
kubectl apply -f kubernetes/manifests/01-configmap.yaml

# Deploy Kafka and Zookeeper
kubectl apply -f kubernetes/manifests/02-kafka.yaml

# Deploy Producer
kubectl apply -f kubernetes/manifests/03-producer.yaml

# Deploy Streams Processor
kubectl apply -f kubernetes/manifests/04-streams.yaml

# Deploy Consumer
kubectl apply -f kubernetes/manifests/05-consumer.yaml

# Deploy HorizontalPodAutoscaler
kubectl apply -f kubernetes/manifests/06-hpa.yaml
```

### Configuration

#### ConfigMap

The system configuration is stored in the `coffee-config` ConfigMap. You can modify the configuration by editing this ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: coffee-config
  namespace: coffee-system
data:
  KAFKA_BROKERS: '["kafka-service:9092"]'
  KAFKA_TOPIC: coffee_orders
  KAFKA_PROCESSED_TOPIC: processed_orders
  KAFKA_RETRY_MAX: "5"
  KAFKA_REQUIRED_ACKS: "all"
  KAFKA_APPLICATION_ID: coffee-streams-app
  KAFKA_AUTO_OFFSET_RESET: earliest
  KAFKA_PROCESSING_GUARANTEE: at_least_once
  KAFKA_CONSUMER_GROUP: coffee-consumer-group
  KAFKA_WORKER_POOL_SIZE: "3"
```

#### Ingress

To access the Producer API, an Ingress is used:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: producer-ingress
  namespace: coffee-system
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: coffee-api.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: producer-service
            port:
              number: 80
```

Change the `host` to your domain.

### Scaling

The system supports automatic scaling using HorizontalPodAutoscaler:

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: producer-hpa
  namespace: coffee-system
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: producer
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

You can configure scaling parameters by changing the values of `minReplicas`, `maxReplicas`, and resource utilization thresholds.

## CI/CD

The system uses GitHub Actions for CI/CD. The workflow is located in the `.github/workflows/ci-cd.yaml` file.

The workflow performs the following steps:

1. Build and test code
2. Build and publish Docker images
3. Deploy to Kubernetes using Helm

To configure CI/CD, you need to add the following secrets to your GitHub repository:

- `KUBECONFIG`: Kubernetes configuration for cluster access

## Monitoring

For system monitoring, it is recommended to use Prometheus and Grafana. You can install them using Helm:

```bash
# Add Helm repository
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Install Prometheus Stack (includes Prometheus, Alertmanager, and Grafana)
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace
```

## Logging

For logging, it is recommended to use Elasticsearch, Fluentd, and Kibana (EFK stack). You can install them using Helm:

```bash
# Add Helm repository
helm repo add elastic https://helm.elastic.co
helm repo update

# Install Elasticsearch
helm install elasticsearch elastic/elasticsearch \
  --namespace logging \
  --create-namespace

# Install Kibana
helm install kibana elastic/kibana \
  --namespace logging \
  --set service.type=ClusterIP

# Install Fluentd
helm install fluentd stable/fluentd \
  --namespace logging \
  --set elasticsearch.host=elasticsearch-client
```

## Conclusion

Docker and Kubernetes provide powerful tools for deploying, scaling, and managing the Coffee Order System. Using these technologies allows you to easily deploy the system in different environments and ensure its reliability and scalability.
