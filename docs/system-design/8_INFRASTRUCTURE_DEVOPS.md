# ☁️ 8: Infrastructure & DevOps

## 📋 Overview

Master infrastructure patterns and DevOps practices through Go Coffee's production-ready deployment architecture. This covers Kubernetes orchestration, CI/CD pipelines, Infrastructure as Code, container optimization, and operational excellence.

## 🎯 Learning Objectives

By the end of this phase, you will:
- Design production Kubernetes deployments
- Implement comprehensive CI/CD pipelines
- Master Infrastructure as Code with Terraform
- Optimize container security and performance
- Build operational excellence practices
- Analyze Go Coffee's infrastructure implementation

---

## 📖 8.1 Kubernetes Orchestration & Deployment

### Core Concepts

#### Kubernetes Architecture
- **Control Plane**: API Server, etcd, Scheduler, Controller Manager
- **Worker Nodes**: kubelet, kube-proxy, Container Runtime
- **Networking**: CNI, Service Mesh, Ingress Controllers
- **Storage**: Persistent Volumes, Storage Classes, CSI Drivers

#### Deployment Patterns
- **Blue-Green Deployment**: Zero-downtime deployments
- **Canary Deployment**: Gradual rollout with traffic splitting
- **Rolling Updates**: Sequential pod replacement
- **A/B Testing**: Feature flag-based deployments

#### Resource Management
- **Resource Requests/Limits**: CPU and memory allocation
- **Quality of Service**: Guaranteed, Burstable, BestEffort
- **Horizontal Pod Autoscaler**: Automatic scaling based on metrics
- **Vertical Pod Autoscaler**: Resource optimization

### 🔍 Go Coffee Analysis

#### Study Kubernetes Deployment Configuration

<augment_code_snippet path="k8s/base/deployment.yaml" mode="EXCERPT">
````yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-coffee-api-gateway
  labels:
    app: api-gateway
    version: v1.0.0
    component: gateway
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: api-gateway
  template:
    metadata:
      labels:
        app: api-gateway
        version: v1.0.0
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: api-gateway-sa
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 2000
      containers:
      - name: api-gateway
        image: go-coffee/api-gateway:v1.0.0
        imagePullPolicy: IfNotPresent
        ports:
        - name: http
          containerPort: 8080
          protocol: TCP
        - name: grpc
          containerPort: 9090
          protocol: TCP
        env:
        - name: ENVIRONMENT
          value: "production"
        - name: LOG_LEVEL
          value: "info"
        - name: POSTGRES_HOST
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: host
        - name: REDIS_URL
          valueFrom:
            configMapKeyRef:
              name: redis-config
              key: url
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: http
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
        volumeMounts:
        - name: config-volume
          mountPath: /app/config
          readOnly: true
        - name: tls-certs
          mountPath: /app/certs
          readOnly: true
      volumes:
      - name: config-volume
        configMap:
          name: api-gateway-config
      - name: tls-certs
        secret:
          secretName: tls-secret
      nodeSelector:
        node-type: application
      tolerations:
      - key: "application"
        operator: "Equal"
        value: "true"
        effect: "NoSchedule"
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - api-gateway
              topologyKey: kubernetes.io/hostname
````
</augment_code_snippet>

### 🛠️ Hands-on Exercise 8.1: Advanced Kubernetes Deployment

#### Step 1: Create Production-Ready Helm Chart
```yaml
# helm/go-coffee/Chart.yaml
apiVersion: v2
name: go-coffee
description: A Helm chart for Go Coffee microservices platform
type: application
version: 1.0.0
appVersion: "1.0.0"
keywords:
  - coffee
  - microservices
  - kubernetes
home: https://github.com/DimaJoyti/go-coffee
sources:
  - https://github.com/DimaJoyti/go-coffee
maintainers:
  - name: Go Coffee Team
    email: team@gocoffee.io

dependencies:
  - name: postgresql
    version: 12.1.9
    repository: https://charts.bitnami.com/bitnami
    condition: postgresql.enabled
  - name: redis
    version: 17.3.7
    repository: https://charts.bitnami.com/bitnami
    condition: redis.enabled
  - name: kafka
    version: 20.0.6
    repository: https://charts.bitnami.com/bitnami
    condition: kafka.enabled
```

```yaml
# helm/go-coffee/values.yaml
global:
  imageRegistry: ""
  imagePullSecrets: []
  storageClass: ""

replicaCount: 3

image:
  registry: docker.io
  repository: gocoffee/api-gateway
  tag: "1.0.0"
  pullPolicy: IfNotPresent

nameOverride: ""
fullnameOverride: ""

serviceAccount:
  create: true
  annotations: {}
  name: ""

podAnnotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8080"
  prometheus.io/path: "/metrics"

podSecurityContext:
  runAsNonRoot: true
  runAsUser: 1000
  fsGroup: 2000

securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
    - ALL
  readOnlyRootFilesystem: true
  runAsNonRoot: true
  runAsUser: 1000

service:
  type: ClusterIP
  port: 80
  targetPort: 8080
  annotations: {}

ingress:
  enabled: true
  className: "nginx"
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
  hosts:
    - host: api.gocoffee.io
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: gocoffee-tls
      hosts:
        - api.gocoffee.io

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 250m
    memory: 256Mi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 50
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

nodeSelector:
  node-type: application

tolerations:
  - key: "application"
    operator: "Equal"
    value: "true"
    effect: "NoSchedule"

affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchExpressions:
          - key: app.kubernetes.io/name
            operator: In
            values:
            - go-coffee
        topologyKey: kubernetes.io/hostname

# Database configuration
postgresql:
  enabled: true
  auth:
    postgresPassword: "secure-password"
    database: "go_coffee"
  primary:
    persistence:
      enabled: true
      size: 100Gi
      storageClass: "fast-ssd"
  readReplicas:
    replicaCount: 2
    persistence:
      enabled: true
      size: 100Gi

redis:
  enabled: true
  auth:
    enabled: true
    password: "redis-password"
  master:
    persistence:
      enabled: true
      size: 20Gi
  replica:
    replicaCount: 2
    persistence:
      enabled: true
      size: 20Gi

kafka:
  enabled: true
  replicaCount: 3
  persistence:
    enabled: true
    size: 50Gi
  zookeeper:
    replicaCount: 3
    persistence:
      enabled: true
      size: 10Gi

# Monitoring
monitoring:
  enabled: true
  serviceMonitor:
    enabled: true
    interval: 30s
    scrapeTimeout: 10s

# Backup configuration
backup:
  enabled: true
  schedule: "0 2 * * *"
  retention: "30d"
  storage:
    type: "s3"
    bucket: "gocoffee-backups"
    region: "us-west-2"
```

#### Step 2: Implement Advanced Deployment Strategies
```yaml
# k8s/deployments/canary-deployment.yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: go-coffee-api-gateway
spec:
  replicas: 10
  strategy:
    canary:
      steps:
      - setWeight: 10
      - pause: {duration: 2m}
      - setWeight: 20
      - pause: {duration: 2m}
      - setWeight: 50
      - pause: {duration: 5m}
      - setWeight: 80
      - pause: {duration: 2m}
      canaryService: api-gateway-canary
      stableService: api-gateway-stable
      trafficRouting:
        nginx:
          stableIngress: api-gateway-stable
          annotationPrefix: nginx.ingress.kubernetes.io
          additionalIngressAnnotations:
            canary-by-header: X-Canary
      analysis:
        templates:
        - templateName: success-rate
        startingStep: 2
        args:
        - name: service-name
          value: api-gateway-canary
  selector:
    matchLabels:
      app: api-gateway
  template:
    metadata:
      labels:
        app: api-gateway
    spec:
      containers:
      - name: api-gateway
        image: go-coffee/api-gateway:v1.1.0
        ports:
        - containerPort: 8080
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

#### Step 3: Service Mesh Integration with Istio
```yaml
# k8s/istio/virtual-service.yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: go-coffee-api-gateway
spec:
  hosts:
  - api.gocoffee.io
  gateways:
  - go-coffee-gateway
  http:
  - match:
    - headers:
        canary:
          exact: "true"
    route:
    - destination:
        host: api-gateway-canary
        port:
          number: 80
      weight: 100
  - route:
    - destination:
        host: api-gateway-stable
        port:
          number: 80
      weight: 90
    - destination:
        host: api-gateway-canary
        port:
          number: 80
      weight: 10
    fault:
      delay:
        percentage:
          value: 0.1
        fixedDelay: 5s
    retries:
      attempts: 3
      perTryTimeout: 2s
      retryOn: 5xx,reset,connect-failure,refused-stream
    timeout: 10s

---
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: go-coffee-api-gateway
spec:
  host: api-gateway
  trafficPolicy:
    connectionPool:
      tcp:
        maxConnections: 100
      http:
        http1MaxPendingRequests: 50
        http2MaxRequests: 100
        maxRequestsPerConnection: 10
        maxRetries: 3
        consecutiveGatewayErrors: 5
        interval: 30s
        baseEjectionTime: 30s
        maxEjectionPercent: 50
    loadBalancer:
      simple: LEAST_CONN
    circuitBreaker:
      consecutiveGatewayErrors: 5
      interval: 30s
      baseEjectionTime: 30s
      maxEjectionPercent: 50
  subsets:
  - name: v1
    labels:
      version: v1
  - name: v2
    labels:
      version: v2
```

### 💡 Practice Question 8.1
**"Design a Kubernetes deployment strategy for Go Coffee that supports zero-downtime deployments, auto-scaling, and multi-region disaster recovery."**

**Solution Framework:**
1. **Deployment Strategy**
   - Blue-green deployments for critical services
   - Canary deployments for gradual rollouts
   - Rolling updates for non-critical services
   - Feature flags for A/B testing

2. **Auto-Scaling Implementation**
   - HPA based on CPU, memory, and custom metrics
   - VPA for resource optimization
   - Cluster autoscaler for node scaling
   - Predictive scaling based on traffic patterns

3. **Multi-Region Setup**
   - Active-active deployment across regions
   - Cross-region database replication
   - Global load balancing with health checks
   - Automated failover mechanisms

---

## 📖 8.2 CI/CD Pipeline Design & Implementation

### Core Concepts

#### CI/CD Pipeline Stages
- **Source Control**: Git workflows and branching strategies
- **Build**: Compilation, testing, and artifact creation
- **Test**: Unit, integration, and end-to-end testing
- **Security**: Vulnerability scanning and compliance checks
- **Deploy**: Automated deployment to environments
- **Monitor**: Post-deployment monitoring and rollback

#### Pipeline Patterns
- **GitOps**: Git as single source of truth for deployments
- **Trunk-based Development**: Short-lived feature branches
- **Feature Flags**: Runtime feature toggling
- **Progressive Delivery**: Gradual feature rollout

### 🔍 Go Coffee Analysis

#### Study GitHub Actions CI/CD Pipeline

<augment_code_snippet path=".github/workflows/ci-cd.yml" mode="EXCERPT">
````yaml
name: Go Coffee CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: [1.21]
        service: [producer, consumer, api-gateway, auth-service]
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}
    
    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-
    
    - name: Run tests
      run: |
        cd ${{ matrix.service }}
        go mod download
        go test -v -race -coverprofile=coverage.out ./...
        go tool cover -html=coverage.out -o coverage.html
    
    - name: Upload coverage reports
      uses: codecov/codecov-action@v3
      with:
        file: ./${{ matrix.service }}/coverage.out
        flags: ${{ matrix.service }}

  security-scan:
    runs-on: ubuntu-latest
    needs: test
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Run Gosec Security Scanner
      uses: securecodewarrior/github-action-gosec@master
      with:
        args: '-fmt sarif -out gosec.sarif ./...'
    
    - name: Upload SARIF file
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: gosec.sarif
    
    - name: Run Trivy vulnerability scanner
      uses: aquasecurity/trivy-action@master
      with:
        scan-type: 'fs'
        scan-ref: '.'
        format: 'sarif'
        output: 'trivy-results.sarif'

  build-and-push:
    runs-on: ubuntu-latest
    needs: [test, security-scan]
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    
    strategy:
      matrix:
        service: [producer, consumer, api-gateway, auth-service]
    
    steps:
    - name: Checkout repository
      uses: actions/checkout@v4
    
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3
    
    - name: Log in to Container Registry
      uses: docker/login-action@v3
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}
    
    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v5
      with:
        images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}/${{ matrix.service }}
        tags: |
          type=ref,event=branch
          type=ref,event=pr
          type=sha,prefix={{branch}}-
          type=raw,value=latest,enable={{is_default_branch}}
    
    - name: Build and push Docker image
      uses: docker/build-push-action@v5
      with:
        context: ./${{ matrix.service }}
        platforms: linux/amd64,linux/arm64
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max
        build-args: |
          VERSION=${{ github.sha }}
          BUILD_DATE=${{ github.event.head_commit.timestamp }}

  deploy-staging:
    runs-on: ubuntu-latest
    needs: build-and-push
    environment: staging
    
    steps:
    - name: Checkout repository
      uses: actions/checkout@v4
    
    - name: Configure AWS credentials
      uses: aws-actions/configure-aws-credentials@v4
      with:
        aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        aws-region: us-west-2
    
    - name: Update kubeconfig
      run: |
        aws eks update-kubeconfig --name go-coffee-staging --region us-west-2
    
    - name: Deploy to staging
      run: |
        helm upgrade --install go-coffee ./helm/go-coffee \
          --namespace staging \
          --create-namespace \
          --values ./helm/go-coffee/values-staging.yaml \
          --set image.tag=${{ github.sha }} \
          --wait --timeout=10m
    
    - name: Run smoke tests
      run: |
        kubectl wait --for=condition=ready pod -l app=go-coffee -n staging --timeout=300s
        ./scripts/smoke-tests.sh staging

  deploy-production:
    runs-on: ubuntu-latest
    needs: deploy-staging
    environment: production
    if: github.ref == 'refs/heads/main'
    
    steps:
    - name: Checkout repository
      uses: actions/checkout@v4
    
    - name: Configure AWS credentials
      uses: aws-actions/configure-aws-credentials@v4
      with:
        aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        aws-region: us-west-2
    
    - name: Update kubeconfig
      run: |
        aws eks update-kubeconfig --name go-coffee-production --region us-west-2
    
    - name: Deploy to production with canary
      run: |
        # Deploy canary version
        helm upgrade --install go-coffee-canary ./helm/go-coffee \
          --namespace production \
          --values ./helm/go-coffee/values-production.yaml \
          --set image.tag=${{ github.sha }} \
          --set replicaCount=1 \
          --set service.name=go-coffee-canary \
          --wait --timeout=10m
        
        # Wait and monitor canary
        sleep 300
        
        # Check canary metrics
        if ./scripts/check-canary-health.sh; then
          # Promote canary to full deployment
          helm upgrade --install go-coffee ./helm/go-coffee \
            --namespace production \
            --values ./helm/go-coffee/values-production.yaml \
            --set image.tag=${{ github.sha }} \
            --wait --timeout=10m
          
          # Clean up canary
          helm uninstall go-coffee-canary --namespace production
        else
          echo "Canary deployment failed, rolling back"
          helm uninstall go-coffee-canary --namespace production
          exit 1
        fi
````
</augment_code_snippet>

### 🛠️ Hands-on Exercise 8.2: Advanced CI/CD Implementation

#### Step 1: GitOps with ArgoCD
```yaml
# argocd/applications/go-coffee-staging.yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: go-coffee-staging
  namespace: argocd
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: default
  source:
    repoURL: https://github.com/DimaJoyti/go-coffee
    targetRevision: develop
    path: helm/go-coffee
    helm:
      valueFiles:
        - values-staging.yaml
      parameters:
        - name: image.tag
          value: "develop-latest"
        - name: replicaCount
          value: "2"
  destination:
    server: https://kubernetes.default.svc
    namespace: staging
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
      allowEmpty: false
    syncOptions:
      - CreateNamespace=true
      - PrunePropagationPolicy=foreground
      - PruneLast=true
    retry:
      limit: 5
      backoff:
        duration: 5s
        factor: 2
        maxDuration: 3m
  revisionHistoryLimit: 10
```

#### Step 2: Progressive Delivery with Flagger
```yaml
# k8s/flagger/canary.yaml
apiVersion: flagger.app/v1beta1
kind: Canary
metadata:
  name: go-coffee-api-gateway
  namespace: production
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-gateway
  progressDeadlineSeconds: 60
  service:
    port: 80
    targetPort: 8080
    gateways:
    - go-coffee-gateway
    hosts:
    - api.gocoffee.io
  analysis:
    interval: 1m
    threshold: 5
    maxWeight: 50
    stepWeight: 10
    metrics:
    - name: request-success-rate
      thresholdRange:
        min: 99
      interval: 1m
    - name: request-duration
      thresholdRange:
        max: 500
      interval: 1m
    - name: cpu-usage
      thresholdRange:
        max: 80
      interval: 1m
    webhooks:
    - name: load-test
      url: http://flagger-loadtester.test/
      timeout: 5s
      metadata:
        cmd: "hey -z 1m -q 10 -c 2 http://api-gateway-canary.production:80/health"
    - name: slack-notification
      url: https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK
      metadata:
        channel: "#deployments"
```

### 💡 Practice Question 8.2
**"Design a CI/CD pipeline for Go Coffee that supports multiple environments, automated testing, security scanning, and progressive delivery with automatic rollback."**

**Solution Framework:**
1. **Pipeline Stages**
   - Source control with Git workflows
   - Automated testing (unit, integration, e2e)
   - Security scanning and compliance checks
   - Multi-stage deployments with approvals
   - Progressive delivery with monitoring

2. **Environment Strategy**
   - Development: Feature branch deployments
   - Staging: Integration testing environment
   - Production: Blue-green with canary analysis
   - Disaster recovery: Cross-region failover

3. **Quality Gates**
   - Code coverage thresholds
   - Security vulnerability scanning
   - Performance regression testing
   - Manual approval for production
   - Automated rollback on failure

---

## 📖 8.3 Infrastructure as Code (IaC)

### Core Concepts

#### Terraform Fundamentals
- **Providers**: Cloud platform integrations
- **Resources**: Infrastructure components
- **Modules**: Reusable infrastructure patterns
- **State Management**: Infrastructure state tracking
- **Workspaces**: Environment separation

#### IaC Best Practices
- **Version Control**: All infrastructure code in Git
- **Modular Design**: Reusable and composable modules
- **Environment Separation**: Isolated infrastructure per environment
- **State Management**: Remote state with locking
- **Security**: Secrets management and least privilege

### 🔍 Go Coffee Analysis

#### Study Terraform Infrastructure

<augment_code_snippet path="terraform/main.tf" mode="EXCERPT">
````hcl
terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.23"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.11"
    }
  }
  
  backend "s3" {
    bucket         = "go-coffee-terraform-state"
    key            = "infrastructure/terraform.tfstate"
    region         = "us-west-2"
    encrypt        = true
    dynamodb_table = "terraform-state-lock"
  }
}

provider "aws" {
  region = var.aws_region
  
  default_tags {
    tags = {
      Project     = "go-coffee"
      Environment = var.environment
      ManagedBy   = "terraform"
      Owner       = "platform-team"
    }
  }
}

# VPC and Networking
module "vpc" {
  source = "./modules/vpc"
  
  name               = "${var.project_name}-${var.environment}"
  cidr               = var.vpc_cidr
  availability_zones = var.availability_zones
  
  enable_nat_gateway   = true
  enable_vpn_gateway   = false
  enable_dns_hostnames = true
  enable_dns_support   = true
  
  tags = local.common_tags
}

# EKS Cluster
module "eks" {
  source = "./modules/eks"
  
  cluster_name    = "${var.project_name}-${var.environment}"
  cluster_version = var.kubernetes_version
  
  vpc_id          = module.vpc.vpc_id
  subnet_ids      = module.vpc.private_subnets
  
  node_groups = {
    application = {
      instance_types = ["t3.large", "t3.xlarge"]
      capacity_type  = "ON_DEMAND"
      min_size       = 3
      max_size       = 20
      desired_size   = 6
      
      k8s_labels = {
        node-type = "application"
      }
      
      taints = [
        {
          key    = "application"
          value  = "true"
          effect = "NO_SCHEDULE"
        }
      ]
    }
    
    ai_workloads = {
      instance_types = ["g4dn.xlarge", "g4dn.2xlarge"]
      capacity_type  = "SPOT"
      min_size       = 0
      max_size       = 10
      desired_size   = 2
      
      k8s_labels = {
        node-type = "ai-workloads"
        gpu       = "nvidia-t4"
      }
      
      taints = [
        {
          key    = "ai-workloads"
          value  = "true"
          effect = "NO_SCHEDULE"
        }
      ]
    }
  }
  
  tags = local.common_tags
}

# RDS PostgreSQL
module "rds" {
  source = "./modules/rds"
  
  identifier = "${var.project_name}-${var.environment}"
  
  engine         = "postgres"
  engine_version = "15.4"
  instance_class = var.rds_instance_class
  
  allocated_storage     = 100
  max_allocated_storage = 1000
  storage_encrypted     = true
  
  db_name  = "go_coffee"
  username = "go_coffee_user"
  
  vpc_security_group_ids = [module.security_groups.rds_sg_id]
  db_subnet_group_name   = module.vpc.database_subnet_group
  
  backup_retention_period = 7
  backup_window          = "03:00-04:00"
  maintenance_window     = "sun:04:00-sun:05:00"
  
  performance_insights_enabled = true
  monitoring_interval         = 60
  
  tags = local.common_tags
}

# ElastiCache Redis
module "redis" {
  source = "./modules/redis"
  
  cluster_id = "${var.project_name}-${var.environment}"
  
  node_type               = var.redis_node_type
  num_cache_nodes         = 3
  parameter_group_name    = "default.redis7"
  port                    = 6379
  
  subnet_group_name       = module.vpc.elasticache_subnet_group
  security_group_ids      = [module.security_groups.redis_sg_id]
  
  at_rest_encryption_enabled = true
  transit_encryption_enabled = true
  
  tags = local.common_tags
}

# MSK Kafka
module "kafka" {
  source = "./modules/kafka"
  
  cluster_name = "${var.project_name}-${var.environment}"
  
  kafka_version   = "2.8.1"
  number_of_nodes = 3
  instance_type   = var.kafka_instance_type
  
  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets
  
  encryption_in_transit_client_broker = "TLS"
  encryption_in_transit_in_cluster    = true
  encryption_at_rest_kms_key_id       = module.kms.kafka_key_id
  
  tags = local.common_tags
}

# Application Load Balancer
module "alb" {
  source = "./modules/alb"
  
  name = "${var.project_name}-${var.environment}"
  
  vpc_id          = module.vpc.vpc_id
  subnets         = module.vpc.public_subnets
  security_groups = [module.security_groups.alb_sg_id]
  
  certificate_arn = module.acm.certificate_arn
  
  tags = local.common_tags
}

# Monitoring and Observability
module "monitoring" {
  source = "./modules/monitoring"
  
  cluster_name = module.eks.cluster_name
  vpc_id       = module.vpc.vpc_id
  
  prometheus_storage_size = "100Gi"
  grafana_admin_password  = var.grafana_admin_password
  
  tags = local.common_tags
}

locals {
  common_tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "terraform"
    Owner       = "platform-team"
  }
}
````
</augment_code_snippet>

### 💡 Practice Question 8.3
**"Design an Infrastructure as Code strategy for Go Coffee that supports multi-environment deployments, disaster recovery, and cost optimization across multiple cloud providers."**

**Solution Framework:**
1. **Multi-Cloud Strategy**
   - Terraform modules for AWS, GCP, Azure
   - Provider-agnostic resource abstractions
   - Cross-cloud networking and data replication
   - Unified monitoring and management

2. **Environment Management**
   - Workspace-based environment separation
   - Environment-specific variable files
   - Automated environment provisioning
   - Resource tagging and cost allocation

3. **Disaster Recovery**
   - Cross-region infrastructure replication
   - Automated backup and restore procedures
   - RTO/RPO requirements implementation
   - Failover testing and validation

---

## 🎯 8 Completion Checklist

### Knowledge Mastery
- [ ] Understand Kubernetes orchestration and deployment patterns
- [ ] Can design comprehensive CI/CD pipelines
- [ ] Know Infrastructure as Code best practices
- [ ] Understand container security and optimization
- [ ] Can implement operational excellence practices

### Practical Skills
- [ ] Can deploy production-ready Kubernetes applications
- [ ] Can build automated CI/CD pipelines with quality gates
- [ ] Can manage infrastructure with Terraform
- [ ] Can implement progressive delivery strategies
- [ ] Can design disaster recovery procedures

### Go Coffee Analysis
- [ ] Analyzed Kubernetes deployment configurations
- [ ] Studied CI/CD pipeline implementations
- [ ] Examined Infrastructure as Code patterns
- [ ] Understood operational monitoring and alerting
- [ ] Identified optimization opportunities

###  Readiness
- [ ] Can design production deployment strategies
- [ ] Can explain CI/CD pipeline trade-offs
- [ ] Can implement infrastructure automation
- [ ] Can handle operational excellence scenarios
- [ ] Can discuss disaster recovery planning

---

## 🚀 Next Steps

Ready for **9: Advanced Distributed Systems**:
- Consensus algorithms and distributed coordination
- Blockchain integration and smart contracts
- Edge computing and IoT integration
- Advanced AI/ML system architecture
- Next-generation distributed patterns

**Excellent progress on mastering infrastructure and DevOps! 🎉**
