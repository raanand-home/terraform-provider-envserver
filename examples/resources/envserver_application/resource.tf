# Create an application
resource "envserver_application" "example" {
  id = "my-application"

  tags = {
    environment = "production"
    team        = "platform"
  }
}

# Create an application with minimal configuration
resource "envserver_application" "minimal" {
  id = "minimal-app"
}
