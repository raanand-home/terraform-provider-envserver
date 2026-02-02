# Create a project
resource "envserver_project" "example" {
  id          = "my-project"
  description = "Example project managed by Terraform"
}

# Create another project
resource "envserver_project" "production" {
  id          = "production"
  description = "Production environment project"
}

# Output the project ID
output "project_id" {
  description = "The ID of the created project"
  value       = envserver_project.example.id
}
