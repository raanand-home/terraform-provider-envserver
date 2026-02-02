terraform {
  required_providers {
    envserver = {
      source = "raanand-home/envserver"
    }
  }
}

# Configure the Environment Server Provider
provider "envserver" {
  # API endpoint (required)
  endpoint = "https://envserver.example.com"

  # Authentication - choose one method:

  # Option 1: Service Account API Key (recommended for automation)
  api_key = var.envserver_api_key

  # Option 2: Username and Password
  # username = var.envserver_username
  # password = var.envserver_password

  # Option 3: Okta Token
  # okta_token = var.envserver_okta_token

  # Optional settings
  timeout = 30 # API timeout in seconds
}

# Example variable definitions
variable "envserver_api_key" {
  description = "Environment Server API key"
  type        = string
  sensitive   = true
}
