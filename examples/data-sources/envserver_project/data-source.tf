# Fetch an existing project
data "envserver_project" "example" {
  id = "my-project"
}

# Use the project data
output "project_description" {
  description = "Description of the existing project"
  value       = data.envserver_project.example.description
}

# Reference in another resource
resource "envserver_application" "api" {
  # This would be defined when we implement the application resource
  # project_id = data.envserver_project.example.id
}
