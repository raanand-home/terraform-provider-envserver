# Fetch an existing application by ID
data "envserver_application" "example" {
  id = "my-application"
}

# Output the application tags
output "application_tags" {
  value = data.envserver_application.example.tags
}
    