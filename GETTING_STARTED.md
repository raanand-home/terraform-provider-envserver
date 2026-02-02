# Getting Started with terraform-provider-envserver

This guide will help you get started with developing and using the Environment Server Terraform provider.

## Prerequisites

- Go 1.21 or later
- Terraform 1.0 or later
- Access to an Environment Server instance
- Git

## Initial Setup

### 1. Clone and Initialize

```bash
cd terraform-provider-envserver

# Download dependencies and tidy go.mod
make init
```

### 2. Build the Provider

```bash
# Build the provider binary
make build

# This creates: terraform-provider-envserver
```

### 3. Install Locally for Testing

```bash
# Install to your local Terraform plugins directory
make install
```

This installs the provider to:
```
~/.terraform.d/plugins/registry.terraform.io/your-org/envserver/0.1.0/<OS>_<ARCH>/
```

## Using the Provider

### 1. Create a Test Configuration

Create a new directory for testing:

```bash
mkdir -p test-config
cd test-config
```

Create `main.tf`:

```hcl
terraform {
  required_providers {
    envserver = {
      source = "your-org/envserver"
      version = "0.1.0"
    }
  }
}

provider "envserver" {
  endpoint = "https://your-envserver.example.com"
  api_key  = "your-api-key-here"
}

resource "envserver_project" "test" {
  id          = "test-project"
  description = "Test project created by Terraform"
}

output "project_id" {
  value = envserver_project.test.id
}
```

### 2. Initialize Terraform

```bash
terraform init
```

### 3. Plan and Apply

```bash
# See what will be created
terraform plan

# Create the resources
terraform apply
```

### 4. Import Existing Resources

```bash
# Import an existing project
terraform import envserver_project.existing existing-project-id
```

## Development Workflow

### Running Tests

```bash
# Run unit tests
make test

# Run acceptance tests (requires a test Environment Server)
export ENVSERVER_ENDPOINT="https://test.example.com"
export ENVSERVER_API_KEY="test-api-key"
make testacc
```

### Code Formatting

```bash
# Format Go code and Terraform examples
make fmt
```

### Generating Documentation

```bash
# Install terraform-plugin-docs if not already installed
go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest

# Generate documentation
make docs
```

## Project Structure

```
terraform-provider-envserver/
├── main.go                          # Provider entry point
├── go.mod                           # Go module definition
├── Makefile                         # Build automation
├── internal/
│   ├── provider/                    # Provider implementation
│   │   ├── provider.go             # Provider configuration
│   │   ├── project_resource.go     # Project resource
│   │   └── project_data_source.go  # Project data source
│   └── client/                      # API client
│       ├── client.go               # Base client
│       └── projects.go             # Project API methods
├── examples/                        # Example configurations
│   ├── provider/                   # Provider examples
│   ├── resources/                  # Resource examples
│   └── data-sources/               # Data source examples
└── docs/                           # Generated documentation
```

## Authentication Methods

The provider supports three authentication methods:

### 1. Service Account API Key (Recommended)

```hcl
provider "envserver" {
  endpoint = "https://envserver.example.com"
  api_key  = var.api_key
}
```

### 2. Username and Password

```hcl
provider "envserver" {
  endpoint = "https://envserver.example.com"
  username = var.username
  password = var.password
}
```

### 3. Okta Token

```hcl
provider "envserver" {
  endpoint   = "https://envserver.example.com"
  okta_token = var.okta_token
}
```

## Environment Variables

All provider configuration can be set via environment variables:

```bash
export ENVSERVER_ENDPOINT="https://envserver.example.com"
export ENVSERVER_API_KEY="your-api-key"
# OR
export ENVSERVER_USERNAME="user@example.com"
export ENVSERVER_PASSWORD="password"
# OR
export ENVSERVER_OKTA_TOKEN="okta-token"
```

## Next Steps

### Implementing Additional Resources

To add a new resource (e.g., Application):

1. **Create API client methods** in `internal/client/applications.go`
2. **Create resource implementation** in `internal/provider/application_resource.go`
3. **Register resource** in `internal/provider/provider.go`
4. **Add examples** in `examples/resources/envserver_application/`
5. **Write tests** in `internal/provider/application_resource_test.go`
6. **Generate docs** with `make docs`

### Current Implementation Status

✅ **Completed:**
- Project structure
- Base API client with authentication
- Provider configuration
- Project resource (CRUD operations)
- Project data source
- Examples and documentation structure

🚧 **In Progress:**
- Additional resources (Applications, Environments, etc.)
- Comprehensive testing
- CI/CD pipeline

📋 **Planned:**
- Application resource
- Application Version resource
- Environment resource
- Environment Config resource
- Version resource
- Operation Template resource
- User and Service Account resources
- Policy resources

## Troubleshooting

### Provider Not Found

If Terraform can't find the provider:

```bash
# Reinstall the provider
make install

# Verify installation
ls -la ~/.terraform.d/plugins/registry.terraform.io/your-org/envserver/0.1.0/

# Re-initialize Terraform
cd your-config-directory
rm -rf .terraform .terraform.lock.hcl
terraform init
```

### Import Errors

If you see Go import errors:

```bash
# Download and tidy dependencies
go mod download
go mod tidy

# Rebuild
make build
```

### API Connection Issues

Enable debug logging:

```bash
export TF_LOG=DEBUG
export TF_LOG_PATH=terraform.log
terraform plan
```

## Contributing

See the [architecture documentation](../plans/terraform-provider-architecture.md) for detailed information about the provider design and implementation guidelines.

## Resources

- [Terraform Plugin Framework Documentation](https://developer.hashicorp.com/terraform/plugin/framework)
- [Environment Server API Documentation](../Agent.md)
- [Provider Architecture](../plans/terraform-provider-architecture.md)
- [Implementation Guide](../plans/terraform-provider-implementation-guide.md)