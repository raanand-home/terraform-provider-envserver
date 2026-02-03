# Example using provider defaults for project_id and env_id
# Configure the provider with defaults:
# provider "envserver" {
#   endpoint           = "https://api.envserver.example.com"
#   default_project_id = "my-project"
#   default_env_id     = "production"
# }

# Register an application instance using provider defaults
resource "envserver_application_instance" "with_defaults" {
  app_id      = "my-application"
  instance_id = "instance-001"

  tags = {
    version     = "1.0.0"
    region      = "us-east-1"
    environment = "production"
  }
}

# Register an application instance with explicit project_id and env_id
resource "envserver_application_instance" "explicit" {
  project_id  = "my-project"
  env_id      = "production"
  app_id      = "my-application"
  instance_id = "instance-002"

  tags = {
    version     = "1.0.0"
    region      = "us-east-1"
    environment = "production"
  }
}

# Example with minimal configuration (no tags, using provider defaults)
resource "envserver_application_instance" "minimal" {
  app_id      = "my-application"
  instance_id = "instance-003"
}

# Example referencing an existing project (overriding provider defaults)
resource "envserver_project" "example" {
  id          = "my-project"
  description = "Example project for application instances"
}

resource "envserver_application_instance" "with_project" {
  project_id  = envserver_project.example.id
  env_id      = "development"
  app_id      = "backend-service"
  instance_id = "dev-instance-001"

  tags = {
    managed_by = "terraform"
    team       = "platform"
  }
}
