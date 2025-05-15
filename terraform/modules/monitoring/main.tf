# Модуль для розгортання моніторингу (Prometheus + Grafana) в Kubernetes

# Створення простору імен для моніторингу
resource "kubernetes_namespace" "monitoring" {
  metadata {
    name = "monitoring-${var.environment}"
  }
}

# Розгортання Prometheus за допомогою Helm
resource "helm_release" "prometheus" {
  count = var.enable_monitoring ? 1 : 0

  name       = "prometheus"
  repository = "https://prometheus-community.github.io/helm-charts"
  chart      = "prometheus"
  version    = "15.0.0"
  namespace  = kubernetes_namespace.monitoring.metadata[0].name

  # Налаштування Prometheus
  set {
    name  = "server.persistentVolume.enabled"
    value = true
  }

  set {
    name  = "server.persistentVolume.size"
    value = "10Gi"
  }

  set {
    name  = "alertmanager.persistentVolume.enabled"
    value = true
  }

  set {
    name  = "alertmanager.persistentVolume.size"
    value = "2Gi"
  }

  set {
    name  = "server.service.type"
    value = "ClusterIP"
  }

  # Налаштування ресурсів
  set {
    name  = "server.resources.requests.memory"
    value = "1Gi"
  }

  set {
    name  = "server.resources.requests.cpu"
    value = "500m"
  }

  set {
    name  = "server.resources.limits.memory"
    value = "2Gi"
  }

  set {
    name  = "server.resources.limits.cpu"
    value = "1000m"
  }
}

# Розгортання Grafana за допомогою Helm
resource "helm_release" "grafana" {
  count = var.enable_monitoring ? 1 : 0

  name       = "grafana"
  repository = "https://grafana.github.io/helm-charts"
  chart      = "grafana"
  version    = "6.50.0"
  namespace  = kubernetes_namespace.monitoring.metadata[0].name

  # Налаштування Grafana
  set {
    name  = "persistence.enabled"
    value = true
  }

  set {
    name  = "persistence.size"
    value = "5Gi"
  }

  set {
    name  = "service.type"
    value = "ClusterIP"
  }

  set {
    name  = "adminPassword"
    value = var.grafana_admin_password
  }

  # Налаштування ресурсів
  set {
    name  = "resources.requests.memory"
    value = "512Mi"
  }

  set {
    name  = "resources.requests.cpu"
    value = "250m"
  }

  set {
    name  = "resources.limits.memory"
    value = "1Gi"
  }

  set {
    name  = "resources.limits.cpu"
    value = "500m"
  }

  # Налаштування datasource для Prometheus
  set {
    name  = "datasources.datasources\\.yaml.apiVersion"
    value = "1"
  }

  set {
    name  = "datasources.datasources\\.yaml.datasources[0].name"
    value = "Prometheus"
  }

  set {
    name  = "datasources.datasources\\.yaml.datasources[0].type"
    value = "prometheus"
  }

  set {
    name  = "datasources.datasources\\.yaml.datasources[0].url"
    value = "http://prometheus-server.${kubernetes_namespace.monitoring.metadata[0].name}.svc.cluster.local:80"
  }

  set {
    name  = "datasources.datasources\\.yaml.datasources[0].access"
    value = "proxy"
  }

  set {
    name  = "datasources.datasources\\.yaml.datasources[0].isDefault"
    value = "true"
  }
}
