terraform {
  required_providers {
    envserver = {
      source = "raanand-home/envserver"
    }
  }
}

# Configure the Environment Server Provider
#
# Configuration can be provided via:
# 1. Provider block attributes (highest priority)
# 2. Environment variables (ENVSERVER_ENDPOINT, ENVSERVER_USERNAME, ENVSERVER_PASSWORD, ENV_SERVER_TOKEN, etc.)
# 3. Config file at ~/.env_server.toml (lowest priority)
#
# The ~/.env_server.toml file format:
#   username = "admin@local.com"
#   url = "https://envserver.example.com"
#   password = "your-password"
#
provider "envserver" {
  # API endpoint (required - can also be set via ENVSERVER_ENDPOINT or 'url' in ~/.env_server.toml)
  endpoint = "https://envserver.example.com"

  # Authentication - choose one method:

  # Option 1: Pre-authenticated Token (highest priority - can also be set via ENV_SERVER_TOKEN)
  # token = var.envserver_token

  # Option 2: Service Account API Key (recommended for automation)
  api_key = var.envserver_api_key

  # Option 3: Username and Password (can also be set in ~/.env_server.toml)
  # username = var.envserver_username
  # password = var.envserver_password

  # Option 4: Okta Token
  # okta_token = var.envserver_okta_token

  # Optional settings
  timeout = 30 # API timeout in seconds

  # Default project and environment IDs
  # These are used as defaults for data sources like envserver_connect_token
  default_project_id = var.default_project_id
  default_env_id     = var.default_env_id
}

# Example variable definitions
variable "envserver_api_key" {
  description = "Environment Server API key"
  type        = string
  sensitive   = true
}

variable "default_project_id" {
  description = "Default project ID for resources and data sources"
  type        = string
  default     = ""
}

variable "default_env_id" {
  description = "Default environment ID for resources and data sources"
  type        = string
  default     = ""
}
