# Multi-Cloud Serverless Orchestrator Module
# Manages serverless functions across AWS, GCP, and Azure

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
    archive = {
      source  = "hashicorp/archive"
      version = "~> 2.0"
    }
  }
}

# Local variables
locals {
  name_prefix = "${var.project_name}-${var.environment}"
  
  # Common tags
  common_tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "terraform"
    Component   = "serverless-orchestrator"
    Team        = "platform"
  }
  
  # Serverless function configurations
  functions = {
    coffee_order_processor = {
      runtime = "go1.x"
      handler = "main"
      timeout = 30
      memory  = 512
      triggers = ["order.created", "order.updated"]
    }
    ai_agent_coordinator = {
      runtime = "go1.x"
      handler = "main"
      timeout = 60
      memory  = 1024
      triggers = ["agent.task.created", "agent.response.needed"]
    }
    defi_arbitrage_scanner = {
      runtime = "go1.x"
      handler = "main"
      timeout = 300
      memory  = 2048
      triggers = ["price.update", "arbitrage.opportunity"]
    }
    inventory_optimizer = {
      runtime = "go1.x"
      handler = "main"
      timeout = 120
      memory  = 1024
      triggers = ["inventory.low", "demand.forecast"]
    }
    notification_dispatcher = {
      runtime = "go1.x"
      handler = "main"
      timeout = 30
      memory  = 256
      triggers = ["notification.send", "alert.trigger"]
    }
  }
}

# =============================================================================
# AWS SERVERLESS COMPONENTS
# =============================================================================

# AWS Lambda Functions
resource "aws_lambda_function" "functions" {
  for_each = var.enable_aws ? local.functions : {}
  
  filename         = "${path.module}/functions/${each.key}/deployment.zip"
  function_name    = "${local.name_prefix}-${each.key}"
  role            = aws_iam_role.lambda_role[0].arn
  handler         = each.value.handler
  runtime         = each.value.runtime
  timeout         = each.value.timeout
  memory_size     = each.value.memory
  
  environment {
    variables = {
      ENVIRONMENT     = var.environment
      PROJECT_NAME    = var.project_name
      KAFKA_BROKERS   = var.kafka_brokers
      REDIS_URL       = var.redis_url
      DATABASE_URL    = var.database_url
      LOG_LEVEL       = var.log_level
    }
  }
  
  vpc_config {
    subnet_ids         = var.aws_subnet_ids
    security_group_ids = [aws_security_group.lambda_sg[0].id]
  }
  
  tags = local.common_tags
  
  depends_on = [
    aws_iam_role_policy_attachment.lambda_basic,
    aws_iam_role_policy_attachment.lambda_vpc,
    aws_cloudwatch_log_group.lambda_logs,
  ]
}

# AWS EventBridge Rules for Function Triggers
resource "aws_cloudwatch_event_rule" "function_triggers" {
  for_each = var.enable_aws ? {
    for func_name, func_config in local.functions : func_name => func_config
    if length(func_config.triggers) > 0
  } : {}
  
  name        = "${local.name_prefix}-${each.key}-trigger"
  description = "Trigger for ${each.key} function"
  
  event_pattern = jsonencode({
    source      = ["go-coffee.platform"]
    detail-type = each.value.triggers
  })
  
  tags = local.common_tags
}

# EventBridge Targets
resource "aws_cloudwatch_event_target" "lambda_targets" {
  for_each = var.enable_aws ? {
    for func_name, func_config in local.functions : func_name => func_config
    if length(func_config.triggers) > 0
  } : {}
  
  rule      = aws_cloudwatch_event_rule.function_triggers[each.key].name
  target_id = "${each.key}-target"
  arn       = aws_lambda_function.functions[each.key].arn
}

# Lambda permissions for EventBridge
resource "aws_lambda_permission" "eventbridge_invoke" {
  for_each = var.enable_aws ? {
    for func_name, func_config in local.functions : func_name => func_config
    if length(func_config.triggers) > 0
  } : {}
  
  statement_id  = "AllowExecutionFromEventBridge"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.functions[each.key].function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.function_triggers[each.key].arn
}

# =============================================================================
# GOOGLE CLOUD SERVERLESS COMPONENTS
# =============================================================================

# Google Cloud Functions
resource "google_cloudfunctions2_function" "functions" {
  for_each = var.enable_gcp ? local.functions : {}
  
  name        = "${local.name_prefix}-${each.key}"
  location    = var.gcp_region
  description = "Serverless function for ${each.key}"
  
  build_config {
    runtime     = "go121"
    entry_point = "Handler"
    source {
      storage_source {
        bucket = google_storage_bucket.function_source[0].name
        object = google_storage_bucket_object.function_source[each.key].name
      }
    }
  }
  
  service_config {
    max_instance_count = var.max_instances
    min_instance_count = 0
    available_memory   = "${each.value.memory}Mi"
    timeout_seconds    = each.value.timeout
    
    environment_variables = {
      ENVIRONMENT     = var.environment
      PROJECT_NAME    = var.project_name
      KAFKA_BROKERS   = var.kafka_brokers
      REDIS_URL       = var.redis_url
      DATABASE_URL    = var.database_url
      LOG_LEVEL       = var.log_level
    }
    
    vpc_connector                 = var.gcp_vpc_connector
    vpc_connector_egress_settings = "ALL_TRAFFIC"
    ingress_settings             = "ALLOW_INTERNAL_ONLY"
  }
  
  event_trigger {
    trigger_region = var.gcp_region
    event_type     = "google.cloud.pubsub.topic.v1.messagePublished"
    pubsub_topic   = google_pubsub_topic.function_triggers[each.key].id
  }
  
  labels = local.common_tags
}

# Pub/Sub Topics for Function Triggers
resource "google_pubsub_topic" "function_triggers" {
  for_each = var.enable_gcp ? local.functions : {}
  
  name = "${local.name_prefix}-${each.key}-trigger"
  
  labels = local.common_tags
}

# =============================================================================
# AZURE SERVERLESS COMPONENTS
# =============================================================================

# Azure Function App
resource "azurerm_linux_function_app" "functions" {
  for_each = var.enable_azure ? local.functions : {}
  
  name                = "${local.name_prefix}-${each.key}"
  resource_group_name = var.azure_resource_group_name
  location            = var.azure_location
  service_plan_id     = azurerm_service_plan.function_plan[0].id
  storage_account_name       = azurerm_storage_account.function_storage[0].name
  storage_account_access_key = azurerm_storage_account.function_storage[0].primary_access_key
  
  site_config {
    application_stack {
      go_version = "1.21"
    }
    
    application_insights_key = azurerm_application_insights.function_insights[0].instrumentation_key
  }
  
  app_settings = {
    ENVIRONMENT     = var.environment
    PROJECT_NAME    = var.project_name
    KAFKA_BROKERS   = var.kafka_brokers
    REDIS_URL       = var.redis_url
    DATABASE_URL    = var.database_url
    LOG_LEVEL       = var.log_level
    FUNCTIONS_WORKER_RUNTIME = "custom"
  }
  
  tags = local.common_tags
}

# Event Grid Topics for Function Triggers
resource "azurerm_eventgrid_topic" "function_triggers" {
  for_each = var.enable_azure ? local.functions : {}
  
  name                = "${local.name_prefix}-${each.key}-trigger"
  location            = var.azure_location
  resource_group_name = var.azure_resource_group_name
  
  tags = local.common_tags
}

# =============================================================================
# CROSS-CLOUD EVENT ROUTING
# =============================================================================

# Event Router Lambda (AWS)
resource "aws_lambda_function" "event_router" {
  count = var.enable_aws && var.enable_cross_cloud_routing ? 1 : 0
  
  filename         = "${path.module}/functions/event-router/deployment.zip"
  function_name    = "${local.name_prefix}-event-router"
  role            = aws_iam_role.lambda_role[0].arn
  handler         = "main"
  runtime         = "go1.x"
  timeout         = 60
  memory_size     = 512
  
  environment {
    variables = {
      ENVIRONMENT           = var.environment
      GCP_PROJECT_ID       = var.gcp_project_id
      AZURE_RESOURCE_GROUP = var.azure_resource_group_name
      CROSS_CLOUD_ROUTING  = "enabled"
    }
  }
  
  tags = local.common_tags
}

# =============================================================================
# MONITORING AND OBSERVABILITY
# =============================================================================

# CloudWatch Log Groups (AWS)
resource "aws_cloudwatch_log_group" "lambda_logs" {
  for_each = var.enable_aws ? local.functions : {}
  
  name              = "/aws/lambda/${local.name_prefix}-${each.key}"
  retention_in_days = var.log_retention_days
  
  tags = local.common_tags
}

# Application Insights (Azure)
resource "azurerm_application_insights" "function_insights" {
  count = var.enable_azure ? 1 : 0
  
  name                = "${local.name_prefix}-insights"
  location            = var.azure_location
  resource_group_name = var.azure_resource_group_name
  application_type    = "other"
  
  tags = local.common_tags
}
