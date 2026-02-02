# Quick Reference Guide

## Build Commands

```bash
# Initialize dependencies
make init

# Build the provider
make build

# Install locally for testing
make install
# OR
./scripts/install-local.sh

# Run tests
make test

# Run acceptance tests
make testacc

# Format code
make fmt

# Clean build artifacts
make clean
```

## Provider Configuration

### Using API Key (Recommended)
```hcl
provider "envserver" {
  endpoint = "https://envserver.example.com"
  api_key  = var.api_key
}
```

### Using Username/Password
```hcl
provider "envserver" {
  endpoint = "https://envserver.example.com"
  username = var.username
  password = var.password
}
```

### Using Okta Token
```hcl
provider "envserver" {
  endpoint   = "https://envserver.example.com"
  okta_token = var.okta_token
}
```

### Environment Variables
```bash
export ENVSERVER_ENDPOINT="https://envserver.example.com"
export ENVSERVER_API_KEY="your-api-key"
```

## Resources

### Project Resource

```hcl
resource "envserver_project" "example" {
  id          = "my-project"
  description = "My project description"
}
```

**Attributes:**
- `id` (String, Required, ForceNew) - Project identifier
- `description` (String, Required) - Project description

**Import:**
```bash
terraform import envserver_project.example my-project
```

## Data Sources

### Project Data Source

```hcl
data "envserver_project" "example" {
  id = "my-project"
}

output "description" {
  value = data.envserver_project.example.description
}
```

## File Structure

```
terraform-provider-envserver/
├── main.go                          # Entry point
├── go.mod                           # Dependencies
├── Makefile                         # Build automation
├── internal/
│   ├── client/                      # API client
│   │   ├── client.go               # Base client
│   │   └── projects.go             # Project API
│   └── provider/                    # Provider implementation
│       ├── provider.go             # Provider config
│       ├── project_resource.go     # Project resource
│       └── project_data_source.go  # Project data source
├── examples/                        # Usage examples
└── scripts/                         # Helper scripts
```

## Development Workflow

### 1. Make Changes
Edit files in `internal/` directory

### 2. Build
```bash
make build
```

### 3. Install Locally
```bash
make install
```

### 4. Test
Create a test Terraform configuration and run:
```bash
terraform init
terraform plan
terraform apply
```

### 5. Debug
Enable debug logging:
```bash
export TF_LOG=DEBUG
export TF_LOG_PATH=terraform.log
terraform plan
```

## Adding a New Resource

### 1. Create API Client Methods
File: `internal/client/your_resource.go`
```go
package client

type YourResource struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

func (c *Client) CreateYourResource(ctx context.Context, r *YourResource) (*YourResource, error) {
    var result YourResource
    err := c.Post(ctx, "/api/your-resources/create", r, &result)
    return &result, err
}

// Add Get, Update, Delete methods...
```

### 2. Create Resource Implementation
File: `internal/provider/your_resource_resource.go`
```go
package provider

import (
    "context"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/your-org/terraform-provider-envserver/internal/client"
)

type YourResourceResource struct {
    client *client.Client
}

type YourResourceModel struct {
    ID   types.String `tfsdk:"id"`
    Name types.String `tfsdk:"name"`
}

func NewYourResourceResource() resource.Resource {
    return &YourResourceResource{}
}

// Implement Metadata, Schema, Configure, Create, Read, Update, Delete, ImportState
```

### 3. Register Resource
File: `internal/provider/provider.go`
```go
func (p *EnvServerProvider) Resources(ctx context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        NewProjectResource,
        NewYourResourceResource,  // Add this line
    }
}
```

### 4. Create Examples
File: `examples/resources/envserver_your_resource/resource.tf`
```hcl
resource "envserver_your_resource" "example" {
  id   = "example"
  name = "Example Resource"
}
```

### 5. Build and Test
```bash
make build
make install
cd examples/resources/envserver_your_resource
terraform init
terraform plan
```

## Common Patterns

### Reading Configuration
```go
var data YourResourceModel
resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
if resp.Diagnostics.HasError() {
    return
}
```

### Calling API
```go
result, err := r.client.CreateYourResource(ctx, &client.YourResource{
    ID:   data.ID.ValueString(),
    Name: data.Name.ValueString(),
})
if err != nil {
    resp.Diagnostics.AddError("Error Creating Resource", err.Error())
    return
}
```

### Updating State
```go
data.ID = types.StringValue(result.ID)
data.Name = types.StringValue(result.Name)
resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
```

## Troubleshooting

### Provider Not Found
```bash
# Reinstall
make install

# Verify
ls -la ~/.terraform.d/plugins/registry.terraform.io/your-org/envserver/

# Re-init Terraform
rm -rf .terraform .terraform.lock.hcl
terraform init
```

### Build Errors
```bash
# Update dependencies
go mod tidy

# Rebuild
make clean
make build
```

### Import Errors
```bash
# Check module path in go.mod
# Ensure all imports use: github.com/your-org/terraform-provider-envserver
```

## Useful Links

- [STATUS.md](STATUS.md) - Current implementation status
- [GETTING_STARTED.md](GETTING_STARTED.md) - Detailed setup guide
- [Architecture](../plans/terraform-provider-architecture.md) - Complete design
- [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)