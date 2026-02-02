# Complete Authentication Setup Example
# This example demonstrates how to set up service accounts, policies, and access control

terraform {
  required_providers {
    envserver = {
      source  = "raanand-home/envserver"
      version = "0.1.0"
    }
  }
}

provider "envserver" {
  endpoint = "https://envserver.example.com"
  api_key  = var.admin_api_key
}

# Create a project
resource "envserver_project" "main" {
  id          = "my-application"
  description = "Main application project"
}

# ============================================================================
# Service Accounts
# ============================================================================

# CI/CD Service Account
resource "envserver_service_account" "ci_cd" {
  id          = "ci-cd-automation"
  description = "Service account for CI/CD pipelines"
}

# Create API key for CI/CD
resource "envserver_service_account_api_key" "ci_cd_key" {
  service_account_id = envserver_service_account.ci_cd.id
}

# Monitoring Service Account
resource "envserver_service_account" "monitoring" {
  id          = "monitoring-system"
  description = "Service account for monitoring and observability"
}

# Create API key for monitoring
resource "envserver_service_account_api_key" "monitoring_key" {
  service_account_id = envserver_service_account.monitoring.id
}

# ============================================================================
# Policies
# ============================================================================

# Read-only policy for monitoring
resource "envserver_policy" "readonly" {
  id          = "readonly-policy"
  description = "Read-only access to all resources"

  policy = jsonencode({
    Statements = [
      {
        actions  = ["projects:get", "application:Get", "application:query", "environments:get", "operations:get"]
        effect   = "Allow"
        resource = ".*"
      }
    ]
  })
}

# Deployment policy for CI/CD
resource "envserver_policy" "deployer" {
  id          = "deployment-policy"
  description = "Full deployment access"

  policy = jsonencode({
    Statements = [
      {
        actions  = ["application:*", "environments:*", "operations:*", "versions:*"]
        effect   = "Allow"
        resource = "arn:project:${envserver_project.main.id}"
      }
    ]
  })
}

# Project admin policy
resource "envserver_policy" "project_admin" {
  id          = "project-admin-policy"
  description = "Full project administration access"

  policy = jsonencode({
    Statements = [
      {
        actions  = ["projects:*", "application:*", "environments:*", "operations:*", "versions:*"]
        effect   = "Allow"
        resource = "arn:project:${envserver_project.main.id}"
      }
    ]
  })
}

# ============================================================================
# Policy Attachments
# ============================================================================

# Attach deployment policy to CI/CD service account
resource "envserver_service_account_policy_attachment" "ci_cd_deployer" {
  service_account_id = envserver_service_account.ci_cd.id
  policy_id          = envserver_policy.deployer.id
}

# Attach readonly policy to monitoring service account
resource "envserver_service_account_policy_attachment" "monitoring_readonly" {
  service_account_id = envserver_service_account.monitoring.id
  policy_id          = envserver_policy.readonly.id
}

# ============================================================================
# Outputs
# ============================================================================

output "ci_cd_api_key" {
  description = "API key for CI/CD service account - STORE SECURELY!"
  value       = envserver_service_account_api_key.ci_cd_key.full_key
  sensitive   = true
}

output "ci_cd_key_id" {
  description = "CI/CD API key ID"
  value       = envserver_service_account_api_key.ci_cd_key.key_id
}

output "monitoring_api_key" {
  description = "API key for monitoring service account - STORE SECURELY!"
  value       = envserver_service_account_api_key.monitoring_key.full_key
  sensitive   = true
}

output "monitoring_key_id" {
  description = "Monitoring API key ID"
  value       = envserver_service_account_api_key.monitoring_key.key_id
}

output "setup_summary" {
  description = "Summary of the authentication setup"
  value = {
    project_id = envserver_project.main.id
    service_accounts = {
      ci_cd      = envserver_service_account.ci_cd.id
      monitoring = envserver_service_account.monitoring.id
    }
    policies = {
      readonly      = envserver_policy.readonly.id
      deployer      = envserver_policy.deployer.id
      project_admin = envserver_policy.project_admin.id
    }
  }
}

# ============================================================================
# Usage Instructions
# ============================================================================

# To retrieve the sensitive API keys after apply:
# terraform output -raw ci_cd_api_key
# terraform output -raw monitoring_api_key

# To use the CI/CD API key in your pipeline:
# export ENVSERVER_API_KEY=$(terraform output -raw ci_cd_api_key)

# To import existing resources:
# terraform import envserver_service_account.ci_cd ci-cd-automation
# terraform import envserver_policy.readonly readonly-policy
# terraform import envserver_service_account_policy_attachment.ci_cd_deployer ci-cd-automation/deployment-policy
