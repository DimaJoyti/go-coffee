# Event-Driven Serverless Orchestration Module
# Manages cross-cloud event routing and workflow orchestration

terraform {
  required_version = ">= 1.6.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }
}

# Local variables
locals {
  name_prefix = "${var.project_name}-${var.environment}"
  
  # Common tags/labels
  common_tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "terraform"
    Component   = "event-orchestration"
    Team        = var.team
  }
  
  # Event routing configuration
  event_routes = {
    # Coffee order events
    "coffee.order.created" = {
      aws_target    = "coffee-order-processor"
      gcp_target    = "coffee-order-processor"
      azure_target  = "coffee-order-processor"
      priority      = 1
      retry_policy  = "exponential_backoff"
    }
    "coffee.order.updated" = {
      aws_target    = "coffee-order-processor"
      gcp_target    = "coffee-order-processor"
      azure_target  = "coffee-order-processor"
      priority      = 1
      retry_policy  = "exponential_backoff"
    }
    
    # AI agent events
    "ai.task.created" = {
      aws_target    = "ai-agent-coordinator"
      gcp_target    = "ai-agent-coordinator"
      azure_target  = "ai-agent-coordinator"
      priority      = 2
      retry_policy  = "linear_backoff"
    }
    "ai.response.needed" = {
      aws_target    = "ai-agent-coordinator"
      gcp_target    = "ai-agent-coordinator"
      azure_target  = "ai-agent-coordinator"
      priority      = 2
      retry_policy  = "linear_backoff"
    }
    
    # DeFi events
    "defi.price.update" = {
      aws_target    = "defi-arbitrage-scanner"
      gcp_target    = "defi-arbitrage-scanner"
      azure_target  = "defi-arbitrage-scanner"
      priority      = 1
      retry_policy  = "immediate"
    }
    "defi.arbitrage.opportunity" = {
      aws_target    = "defi-arbitrage-scanner"
      gcp_target    = "defi-arbitrage-scanner"
      azure_target  = "defi-arbitrage-scanner"
      priority      = 1
      retry_policy  = "immediate"
    }
    
    # Inventory events
    "inventory.low" = {
      aws_target    = "inventory-optimizer"
      gcp_target    = "inventory-optimizer"
      azure_target  = "inventory-optimizer"
      priority      = 2
      retry_policy  = "exponential_backoff"
    }
    "inventory.forecast" = {
      aws_target    = "inventory-optimizer"
      gcp_target    = "inventory-optimizer"
      azure_target  = "inventory-optimizer"
      priority      = 3
      retry_policy  = "linear_backoff"
    }
    
    # Notification events
    "notification.send" = {
      aws_target    = "notification-dispatcher"
      gcp_target    = "notification-dispatcher"
      azure_target  = "notification-dispatcher"
      priority      = 2
      retry_policy  = "exponential_backoff"
    }
    "alert.trigger" = {
      aws_target    = "notification-dispatcher"
      gcp_target    = "notification-dispatcher"
      azure_target  = "notification-dispatcher"
      priority      = 1
      retry_policy  = "immediate"
    }
  }
}

# =============================================================================
# AWS EVENTBRIDGE CONFIGURATION
# =============================================================================

# Custom EventBridge Bus
resource "aws_cloudwatch_event_bus" "main" {
  count = var.enable_aws ? 1 : 0
  
  name = "${local.name_prefix}-event-bus"
  
  tags = local.common_tags
}

# EventBridge Rules for each event type
resource "aws_cloudwatch_event_rule" "event_rules" {
  for_each = var.enable_aws ? local.event_routes : {}
  
  name           = "${local.name_prefix}-${replace(each.key, ".", "-")}"
  description    = "Rule for ${each.key} events"
  event_bus_name = aws_cloudwatch_event_bus.main[0].name
  
  event_pattern = jsonencode({
    source      = ["go-coffee.platform"]
    detail-type = [each.key]
    detail = {
      priority = [each.value.priority]
    }
  })
  
  tags = local.common_tags
}

# EventBridge Targets
resource "aws_cloudwatch_event_target" "lambda_targets" {
  for_each = var.enable_aws ? local.event_routes : {}
  
  rule           = aws_cloudwatch_event_rule.event_rules[each.key].name
  event_bus_name = aws_cloudwatch_event_bus.main[0].name
  target_id      = "${each.value.aws_target}-target"
  arn            = "arn:aws:lambda:${var.aws_region}:${data.aws_caller_identity.current[0].account_id}:function:${local.name_prefix}-${each.value.aws_target}"
  
  # Retry policy configuration
  retry_policy {
    maximum_retry_attempts       = var.max_retry_attempts
    maximum_event_age_in_seconds = var.max_event_age_seconds
  }
  
  # Dead letter queue configuration
  dead_letter_config {
    arn = aws_sqs_queue.dlq[0].arn
  }
  
  # Input transformation
  input_transformer {
    input_paths = {
      timestamp = "$.time"
      source    = "$.source"
      detail    = "$.detail"
    }
    
    input_template = jsonencode({
      event_type = each.key
      timestamp  = "<timestamp>"
      source     = "<source>"
      payload    = "<detail>"
      metadata = {
        retry_policy = each.value.retry_policy
        priority     = each.value.priority
        target       = each.value.aws_target
      }
    })
  }
}

# Dead Letter Queue
resource "aws_sqs_queue" "dlq" {
  count = var.enable_aws ? 1 : 0
  
  name                       = "${local.name_prefix}-event-dlq"
  message_retention_seconds  = var.dlq_retention_seconds
  visibility_timeout_seconds = 300
  
  redrive_allow_policy = jsonencode({
    redrivePermission = "byQueue"
    sourceQueueArns   = ["arn:aws:sqs:${var.aws_region}:${data.aws_caller_identity.current[0].account_id}:*"]
  })
  
  tags = local.common_tags
}

# EventBridge Archive for event replay
resource "aws_cloudwatch_event_archive" "main" {
  count = var.enable_aws && var.enable_event_replay ? 1 : 0
  
  name             = "${local.name_prefix}-event-archive"
  event_source_arn = aws_cloudwatch_event_bus.main[0].arn
  description      = "Archive for Go Coffee platform events"
  retention_days   = var.event_archive_retention_days
  
  event_pattern = jsonencode({
    source = ["go-coffee.platform"]
  })
}

# Cross-region replication
resource "aws_cloudwatch_event_rule" "cross_region_replication" {
  count = var.enable_aws && var.enable_cross_region_replication ? 1 : 0
  
  name           = "${local.name_prefix}-cross-region-replication"
  description    = "Replicate events to other regions"
  event_bus_name = aws_cloudwatch_event_bus.main[0].name
  
  event_pattern = jsonencode({
    source = ["go-coffee.platform"]
    detail = {
      replicate = ["true"]
    }
  })
  
  tags = local.common_tags
}

# =============================================================================
# GOOGLE CLOUD PUB/SUB CONFIGURATION
# =============================================================================

# Pub/Sub Topics for each event type
resource "google_pubsub_topic" "event_topics" {
  for_each = var.enable_gcp ? local.event_routes : {}
  
  name = "${local.name_prefix}-${replace(each.key, ".", "-")}"
  
  # Message retention
  message_retention_duration = "${var.pubsub_message_retention_days * 24}h"
  
  # Schema settings
  schema_settings {
    schema   = google_pubsub_schema.event_schema[0].id
    encoding = "JSON"
  }
  
  labels = local.common_tags
}

# Pub/Sub Schema for event validation
resource "google_pubsub_schema" "event_schema" {
  count = var.enable_gcp ? 1 : 0
  
  name = "${local.name_prefix}-event-schema"
  type = "AVRO"
  
  definition = jsonencode({
    type = "record"
    name = "GoCoffeeEvent"
    fields = [
      {
        name = "event_type"
        type = "string"
      },
      {
        name = "timestamp"
        type = "string"
      },
      {
        name = "source"
        type = "string"
      },
      {
        name = "payload"
        type = "string"
      },
      {
        name = "metadata"
        type = {
          type = "record"
          name = "EventMetadata"
          fields = [
            { name = "priority", type = "int" },
            { name = "retry_policy", type = "string" },
            { name = "target", type = "string" }
          ]
        }
      }
    ]
  })
}

# Pub/Sub Subscriptions
resource "google_pubsub_subscription" "event_subscriptions" {
  for_each = var.enable_gcp ? local.event_routes : {}
  
  name  = "${local.name_prefix}-${replace(each.key, ".", "-")}-subscription"
  topic = google_pubsub_topic.event_topics[each.key].name
  
  # Acknowledgment deadline
  ack_deadline_seconds = var.pubsub_ack_deadline_seconds
  
  # Message retention
  message_retention_duration = "${var.pubsub_message_retention_days * 24}h"
  
  # Retry policy
  retry_policy {
    minimum_backoff = "${var.pubsub_min_backoff_seconds}s"
    maximum_backoff = "${var.pubsub_max_backoff_seconds}s"
  }
  
  # Dead letter policy
  dead_letter_policy {
    dead_letter_topic     = google_pubsub_topic.dlq[0].id
    max_delivery_attempts = var.max_retry_attempts
  }
  
  # Push configuration for Cloud Functions
  push_config {
    push_endpoint = "https://${var.gcp_region}-${var.gcp_project_id}.cloudfunctions.net/${local.name_prefix}-${each.value.gcp_target}"
    
    attributes = {
      x-goog-version = "v1"
    }
    
    # OIDC token for authentication
    oidc_token {
      service_account_email = google_service_account.pubsub_invoker[0].email
    }
  }
  
  # Enable message ordering
  enable_message_ordering = var.enable_message_ordering
  
  labels = local.common_tags
}

# Dead Letter Topic
resource "google_pubsub_topic" "dlq" {
  count = var.enable_gcp ? 1 : 0
  
  name = "${local.name_prefix}-event-dlq"
  
  message_retention_duration = "${var.dlq_retention_seconds}s"
  
  labels = local.common_tags
}

# Service Account for Pub/Sub to invoke Cloud Functions
resource "google_service_account" "pubsub_invoker" {
  count = var.enable_gcp ? 1 : 0
  
  account_id   = "${local.name_prefix}-pubsub-invoker"
  display_name = "Pub/Sub Cloud Functions Invoker"
  description  = "Service account for Pub/Sub to invoke Cloud Functions"
}

# IAM binding for Cloud Functions invoker
resource "google_project_iam_member" "pubsub_invoker" {
  count = var.enable_gcp ? 1 : 0
  
  project = var.gcp_project_id
  role    = "roles/cloudfunctions.invoker"
  member  = "serviceAccount:${google_service_account.pubsub_invoker[0].email}"
}

# =============================================================================
# AZURE EVENT GRID CONFIGURATION
# =============================================================================

# Event Grid Topics
resource "azurerm_eventgrid_topic" "event_topics" {
  for_each = var.enable_azure ? local.event_routes : {}
  
  name                = "${local.name_prefix}-${replace(each.key, ".", "-")}"
  location            = var.azure_location
  resource_group_name = var.azure_resource_group_name
  
  # Input schema
  input_schema = "EventGridSchema"
  
  tags = local.common_tags
}

# Event Grid Subscriptions
resource "azurerm_eventgrid_event_subscription" "event_subscriptions" {
  for_each = var.enable_azure ? local.event_routes : {}
  
  name  = "${local.name_prefix}-${replace(each.key, ".", "-")}-subscription"
  scope = azurerm_eventgrid_topic.event_topics[each.key].id
  
  # Azure Function endpoint
  azure_function_endpoint {
    function_id                       = "/subscriptions/${data.azurerm_client_config.current[0].subscription_id}/resourceGroups/${var.azure_resource_group_name}/providers/Microsoft.Web/sites/${local.name_prefix}-${each.value.azure_target}/functions/${each.value.azure_target}"
    max_events_per_batch             = var.azure_max_events_per_batch
    preferred_batch_size_in_kilobytes = var.azure_preferred_batch_size_kb
  }
  
  # Retry policy
  retry_policy {
    max_delivery_attempts = var.max_retry_attempts
    event_time_to_live    = var.max_event_age_seconds
  }
  
  # Dead letter destination
  storage_blob_dead_letter_destination {
    storage_account_id          = azurerm_storage_account.event_storage[0].id
    storage_blob_container_name = azurerm_storage_container.dlq[0].name
  }
  
  # Event filtering
  subject_filter {
    subject_begins_with = each.key
    case_sensitive      = false
  }
  
  # Advanced filters
  advanced_filter {
    string_in {
      key    = "data.priority"
      values = [tostring(each.value.priority)]
    }
  }
  
  labels = local.common_tags
}

# Storage Account for dead letter events
resource "azurerm_storage_account" "event_storage" {
  count = var.enable_azure ? 1 : 0
  
  name                     = replace("${local.name_prefix}eventstorage", "-", "")
  resource_group_name      = var.azure_resource_group_name
  location                 = var.azure_location
  account_tier             = "Standard"
  account_replication_type = "LRS"
  
  tags = local.common_tags
}

# Storage Container for dead letter events
resource "azurerm_storage_container" "dlq" {
  count = var.enable_azure ? 1 : 0
  
  name                  = "event-dlq"
  storage_account_name  = azurerm_storage_account.event_storage[0].name
  container_access_type = "private"
}

# =============================================================================
# CROSS-CLOUD EVENT ROUTING
# =============================================================================

# Event Router Function (deployed to primary cloud)
resource "aws_lambda_function" "cross_cloud_event_router" {
  count = var.enable_aws && var.enable_cross_cloud_routing ? 1 : 0
  
  filename         = "${path.module}/functions/cross-cloud-event-router.zip"
  function_name    = "${local.name_prefix}-cross-cloud-event-router"
  role            = aws_iam_role.event_router_role[0].arn
  handler         = "main"
  runtime         = "go1.x"
  timeout         = 60
  memory_size     = 512
  
  environment {
    variables = {
      ENVIRONMENT           = var.environment
      PROJECT_NAME          = var.project_name
      ENABLE_GCP_ROUTING    = var.enable_gcp
      ENABLE_AZURE_ROUTING  = var.enable_azure
      GCP_PROJECT_ID        = var.gcp_project_id
      AZURE_RESOURCE_GROUP  = var.azure_resource_group_name
      EVENT_ROUTING_TABLE   = jsonencode(local.event_routes)
    }
  }
  
  tags = local.common_tags
}

# IAM Role for Cross-Cloud Event Router
resource "aws_iam_role" "event_router_role" {
  count = var.enable_aws && var.enable_cross_cloud_routing ? 1 : 0
  
  name = "${local.name_prefix}-event-router-role"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
  
  tags = local.common_tags
}

# IAM Policy for Cross-Cloud Event Router
resource "aws_iam_role_policy" "event_router_policy" {
  count = var.enable_aws && var.enable_cross_cloud_routing ? 1 : 0
  
  name = "${local.name_prefix}-event-router-policy"
  role = aws_iam_role.event_router_role[0].id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:aws:logs:*:*:*"
      },
      {
        Effect = "Allow"
        Action = [
          "events:PutEvents"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue"
        ]
        Resource = "arn:aws:secretsmanager:*:*:secret:${local.name_prefix}/*"
      }
    ]
  })
}

# =============================================================================
# MONITORING AND OBSERVABILITY
# =============================================================================

# CloudWatch Dashboard for Event Metrics
resource "aws_cloudwatch_dashboard" "event_metrics" {
  count = var.enable_aws && var.monitoring_enabled ? 1 : 0
  
  dashboard_name = "${local.name_prefix}-event-metrics"
  
  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "metric"
        x      = 0
        y      = 0
        width  = 12
        height = 6
        
        properties = {
          metrics = [
            ["AWS/Events", "SuccessfulInvocations", "RuleName", "${local.name_prefix}-coffee-order-created"],
            [".", "FailedInvocations", ".", "."],
            [".", "SuccessfulInvocations", "RuleName", "${local.name_prefix}-ai-task-created"],
            [".", "FailedInvocations", ".", "."]
          ]
          view    = "timeSeries"
          stacked = false
          region  = var.aws_region
          title   = "Event Processing Metrics"
          period  = 300
        }
      }
    ]
  })
}

# Data sources
data "aws_caller_identity" "current" {
  count = var.enable_aws ? 1 : 0
}

data "azurerm_client_config" "current" {
  count = var.enable_azure ? 1 : 0
}
