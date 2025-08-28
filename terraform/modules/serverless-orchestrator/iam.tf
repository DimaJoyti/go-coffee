# IAM Resources for Multi-Cloud Serverless Orchestrator

# =============================================================================
# AWS IAM RESOURCES
# =============================================================================

# Lambda Execution Role
resource "aws_iam_role" "lambda_role" {
  count = var.enable_aws ? 1 : 0
  
  name = "${local.name_prefix}-lambda-role"
  
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

# Basic Lambda execution policy
resource "aws_iam_role_policy_attachment" "lambda_basic" {
  count = var.enable_aws ? 1 : 0
  
  role       = aws_iam_role.lambda_role[0].name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# VPC access policy for Lambda
resource "aws_iam_role_policy_attachment" "lambda_vpc" {
  count = var.enable_aws && length(var.aws_subnet_ids) > 0 ? 1 : 0
  
  role       = aws_iam_role.lambda_role[0].name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

# Custom policy for Lambda functions
resource "aws_iam_role_policy" "lambda_custom_policy" {
  count = var.enable_aws ? 1 : 0
  
  name = "${local.name_prefix}-lambda-custom-policy"
  role = aws_iam_role.lambda_role[0].id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "events:PutEvents",
          "events:List*",
          "events:Describe*"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "sqs:SendMessage",
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes"
        ]
        Resource = "arn:aws:sqs:*:*:${local.name_prefix}-*"
      },
      {
        Effect = "Allow"
        Action = [
          "sns:Publish",
          "sns:Subscribe",
          "sns:Unsubscribe"
        ]
        Resource = "arn:aws:sns:*:*:${local.name_prefix}-*"
      },
      {
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
          "dynamodb:DeleteItem",
          "dynamodb:Query",
          "dynamodb:Scan"
        ]
        Resource = "arn:aws:dynamodb:*:*:table/${local.name_prefix}-*"
      },
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret"
        ]
        Resource = "arn:aws:secretsmanager:*:*:secret:${local.name_prefix}/*"
      },
      {
        Effect = "Allow"
        Action = [
          "kms:Decrypt",
          "kms:DescribeKey"
        ]
        Resource = "*"
        Condition = {
          StringEquals = {
            "kms:ViaService" = "lambda.${var.aws_region}.amazonaws.com"
          }
        }
      }
    ]
  })
}

# Cross-cloud access policy for event router
resource "aws_iam_role_policy" "cross_cloud_policy" {
  count = var.enable_aws && var.enable_cross_cloud_routing ? 1 : 0
  
  name = "${local.name_prefix}-cross-cloud-policy"
  role = aws_iam_role.lambda_role[0].id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sts:AssumeRole"
        ]
        Resource = [
          "arn:aws:iam::*:role/${local.name_prefix}-cross-cloud-*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "events:PutEvents"
        ]
        Resource = "*"
      }
    ]
  })
}

# Security Group for Lambda functions
resource "aws_security_group" "lambda_sg" {
  count = var.enable_aws && length(var.aws_subnet_ids) > 0 ? 1 : 0
  
  name_prefix = "${local.name_prefix}-lambda-"
  vpc_id      = var.aws_vpc_id
  description = "Security group for Lambda functions"
  
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "All outbound traffic"
  }
  
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTPS inbound"
  }
  
  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-lambda-sg"
  })
}

# =============================================================================
# GOOGLE CLOUD IAM RESOURCES
# =============================================================================

# Service Account for Cloud Functions
resource "google_service_account" "function_sa" {
  count = var.enable_gcp ? 1 : 0
  
  account_id   = "${local.name_prefix}-functions"
  display_name = "Go Coffee Serverless Functions Service Account"
  description  = "Service account for serverless functions"
}

# IAM bindings for Cloud Functions service account
resource "google_project_iam_member" "function_sa_bindings" {
  for_each = var.enable_gcp ? toset([
    "roles/cloudsql.client",
    "roles/pubsub.publisher",
    "roles/pubsub.subscriber",
    "roles/storage.objectViewer",
    "roles/secretmanager.secretAccessor",
    "roles/monitoring.metricWriter",
    "roles/logging.logWriter",
    "roles/cloudtrace.agent"
  ]) : toset([])
  
  project = var.gcp_project_id
  role    = each.value
  member  = "serviceAccount:${google_service_account.function_sa[0].email}"
}

# Storage bucket for function source code
resource "google_storage_bucket" "function_source" {
  count = var.enable_gcp ? 1 : 0
  
  name     = "${local.name_prefix}-function-source"
  location = var.gcp_region
  
  uniform_bucket_level_access = true
  
  versioning {
    enabled = true
  }
  
  lifecycle_rule {
    condition {
      age = 30
    }
    action {
      type = "Delete"
    }
  }
  
  labels = local.common_tags
}

# Function source code objects
resource "google_storage_bucket_object" "function_source" {
  for_each = var.enable_gcp ? local.functions : {}
  
  name   = "${each.key}/source.zip"
  bucket = google_storage_bucket.function_source[0].name
  source = "${path.module}/functions/${each.key}/deployment.zip"
}

# =============================================================================
# AZURE IAM RESOURCES
# =============================================================================

# Service Plan for Azure Functions
resource "azurerm_service_plan" "function_plan" {
  count = var.enable_azure ? 1 : 0
  
  name                = "${local.name_prefix}-function-plan"
  resource_group_name = var.azure_resource_group_name
  location            = var.azure_location
  os_type             = "Linux"
  sku_name            = "Y1"  # Consumption plan
  
  tags = local.common_tags
}

# Storage Account for Azure Functions
resource "azurerm_storage_account" "function_storage" {
  count = var.enable_azure ? 1 : 0
  
  name                     = replace("${local.name_prefix}funcstorage", "-", "")
  resource_group_name      = var.azure_resource_group_name
  location                 = var.azure_location
  account_tier             = "Standard"
  account_replication_type = "LRS"
  
  tags = local.common_tags
}

# User Assigned Identity for Azure Functions
resource "azurerm_user_assigned_identity" "function_identity" {
  count = var.enable_azure ? 1 : 0
  
  name                = "${local.name_prefix}-function-identity"
  resource_group_name = var.azure_resource_group_name
  location            = var.azure_location
  
  tags = local.common_tags
}

# Role assignments for Azure Functions
resource "azurerm_role_assignment" "function_roles" {
  for_each = var.enable_azure ? toset([
    "Storage Blob Data Contributor",
    "EventGrid Data Sender",
    "Key Vault Secrets User"
  ]) : toset([])
  
  scope                = "/subscriptions/${data.azurerm_client_config.current.subscription_id}/resourceGroups/${var.azure_resource_group_name}"
  role_definition_name = each.value
  principal_id         = azurerm_user_assigned_identity.function_identity[0].principal_id
}

# Data source for current Azure configuration
data "azurerm_client_config" "current" {
  count = var.enable_azure ? 1 : 0
}

# =============================================================================
# CROSS-CLOUD SERVICE ACCOUNTS
# =============================================================================

# AWS role for cross-cloud access from GCP
resource "aws_iam_role" "gcp_cross_cloud_role" {
  count = var.enable_aws && var.enable_gcp && var.enable_cross_cloud_routing ? 1 : 0
  
  name = "${local.name_prefix}-gcp-cross-cloud-role"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = "accounts.google.com"
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "accounts.google.com:aud" = var.gcp_project_id
          }
        }
      }
    ]
  })
  
  tags = local.common_tags
}

# Policy for GCP cross-cloud access
resource "aws_iam_role_policy" "gcp_cross_cloud_policy" {
  count = var.enable_aws && var.enable_gcp && var.enable_cross_cloud_routing ? 1 : 0
  
  name = "${local.name_prefix}-gcp-cross-cloud-policy"
  role = aws_iam_role.gcp_cross_cloud_role[0].id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "events:PutEvents",
          "lambda:InvokeFunction"
        ]
        Resource = "*"
      }
    ]
  })
}
