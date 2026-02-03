# Retrieve all environment configurations
data "envserver_environment_configs" "all" {
  project_id = "my-project"
  env_id     = "production"
}

# Output all configurations
output "all_configs" {
  value = data.envserver_environment_configs.all.configs
}

# Access a specific config from the list
output "first_config_key" {
  value = length(data.envserver_environment_configs.all.configs) > 0 ? data.envserver_environment_configs.all.configs[0].key : ""
}

# Using provider defaults for project_id and env_id
data "envserver_environment_configs" "default_env" {
}

# Create a map from the configs list for easier access
locals {
  config_map = { for config in data.envserver_environment_configs.all.configs : config.key => config.value }
}

output "database_url_from_map" {
  value     = lookup(local.config_map, "DATABASE_URL", "")
  sensitive = true
}
