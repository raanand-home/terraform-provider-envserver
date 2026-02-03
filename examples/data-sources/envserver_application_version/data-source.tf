# Fetch a specific version of an application
data "envserver_application_version" "example" {
  app_id     = "my-application"
  ref_id     = "main"
  version_id = "v1.0.0"
}

# Output the version data
output "version_data" {
  value = data.envserver_application_version.example.version_data
}

output "created_at" {
  value = data.envserver_application_version.example.created_at
}
