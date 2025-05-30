resource "kubernetes_namespace" "kafka" {
  metadata {
    name = "kafka-${var.environment}"
  }
}

resource "helm_release" "kafka" {
  name       = var.kafka_instance_name
  repository = "https://charts.bitnami.com/bitnami"
  chart      = "kafka"
  version    = var.kafka_version
  namespace  = kubernetes_namespace.kafka.metadata[0].name

  # Kafka configuration
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
    name  = "externalAccess.enabled"
    value = true
  }

  set {
    name  = "externalAccess.service.type"
    value = "LoadBalancer"
  }
}

# Create Kafka topics
resource "kubernetes_job" "create_topics" {
  metadata {
    name      = "kafka-create-topics"
    namespace = kubernetes_namespace.kafka.metadata[0].name
  }

  spec {
    template {
      metadata {
        name = "kafka-create-topics"
      }

      spec {
        container {
          name    = "kafka-client"
          image   = "bitnami/kafka:${var.kafka_version}"
          command = ["/bin/bash", "-c"]
          args = [
            <<-EOT
            # Wait for Kafka to be ready
            sleep 30

            # Create supply topic
            kafka-topics.sh --create --if-not-exists \
              --bootstrap-server ${var.kafka_instance_name}.${kubernetes_namespace.kafka.metadata[0].name}.svc.cluster.local:9092 \
              --topic supply-events \
              --partitions 3 \
              --replication-factor 3

            # Create order topic
            kafka-topics.sh --create --if-not-exists \
              --bootstrap-server ${var.kafka_instance_name}.${kubernetes_namespace.kafka.metadata[0].name}.svc.cluster.local:9092 \
              --topic order-events \
              --partitions 3 \
              --replication-factor 3

            # Create claim topic
            kafka-topics.sh --create --if-not-exists \
              --bootstrap-server ${var.kafka_instance_name}.${kubernetes_namespace.kafka.metadata[0].name}.svc.cluster.local:9092 \
              --topic claim-events \
              --partitions 3 \
              --replication-factor 3
            EOT
          ]
        }

        restart_policy = "OnFailure"
      }
    }

    backoff_limit = 4
  }

  depends_on = [helm_release.kafka]
}

# Output the Kafka broker addresses
output "kafka_brokers" {
  value = "${var.kafka_instance_name}.${kubernetes_namespace.kafka.metadata[0].name}.svc.cluster.local:9092"
}
