# Retrieve a single environment configuration by key
data "envserver_environment_config" "database_url" {
  project_id = "my-project"
  env_id     = "production"
  key        = "DATABASE_URL"
}

# Use the configuration value
output "database_url" {
  value     = data.envserver_environment_config.database_url.value
  sensitive = true
}

# Using provider defaults for project_id and env_id
data "envserver_environment_config" "api_key" {
  key = "API_KEY"
}
