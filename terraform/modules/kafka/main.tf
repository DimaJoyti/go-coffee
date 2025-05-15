# Модуль для розгортання Kafka в Kubernetes за допомогою Helm

# Створення простору імен для Kafka
resource "kubernetes_namespace" "kafka" {
  metadata {
    name = "kafka-${var.environment}"
  }
}

# Розгортання Kafka за допомогою Helm
resource "helm_release" "kafka" {
  name       = var.kafka_instance_name
  repository = "https://charts.bitnami.com/bitnami"
  chart      = "kafka"
  version    = var.kafka_version
  namespace  = kubernetes_namespace.kafka.metadata[0].name

  # Налаштування Kafka
  set {
    name  = "replicaCount"
    value = 3
  }

  set {
    name  = "persistence.enabled"
    value = true
  }

  set {
    name  = "persistence.size"
    value = "10Gi"
  }

  set {
    name  = "zookeeper.enabled"
    value = true
  }

  set {
    name  = "zookeeper.replicaCount"
    value = 3
  }

  set {
    name  = "zookeeper.persistence.enabled"
    value = true
  }

  set {
    name  = "zookeeper.persistence.size"
    value = "5Gi"
  }

  set {
    name  = "metrics.kafka.enabled"
    value = true
  }

  set {
    name  = "metrics.jmx.enabled"
    value = true
  }

  set {
    name  = "metrics.serviceMonitor.enabled"
    value = true
  }

  # Налаштування ресурсів
  set {
    name  = "resources.requests.memory"
    value = "1Gi"
  }

  set {
    name  = "resources.requests.cpu"
    value = "500m"
  }

  set {
    name  = "resources.limits.memory"
    value = "2Gi"
  }

  set {
    name  = "resources.limits.cpu"
    value = "1000m"
  }
}

# Створення топіків Kafka
resource "kubernetes_job" "create_topics" {
  depends_on = [helm_release.kafka]

  metadata {
    name      = "create-kafka-topics-${var.environment}"
    namespace = kubernetes_namespace.kafka.metadata[0].name
  }

  spec {
    template {
      metadata {
        name = "create-kafka-topics"
      }

      spec {
        container {
          name    = "kafka-client"
          image   = "bitnami/kafka:${var.kafka_version}"
          command = ["/bin/bash", "-c"]
          args = [
            <<-EOT
            # Чекаємо, поки Kafka буде готова
            sleep 30

            # Створюємо топік для замовлень
            kafka-topics.sh --create --if-not-exists \
              --bootstrap-server ${var.kafka_instance_name}.${kubernetes_namespace.kafka.metadata[0].name}.svc.cluster.local:9092 \
              --topic ${var.kafka_topic_name} \
              --partitions 3 \
              --replication-factor 3

            # Створюємо топік для оброблених замовлень
            kafka-topics.sh --create --if-not-exists \
              --bootstrap-server ${var.kafka_instance_name}.${kubernetes_namespace.kafka.metadata[0].name}.svc.cluster.local:9092 \
              --topic ${var.kafka_processed_topic_name} \
              --partitions 3 \
              --replication-factor 3
            EOT
          ]
        }

        restart_policy = "OnFailure"
      }
    }

    backoff_limit = 5
  }
}
