# Terraform Provider for Environment Server

This is a Terraform provider for managing [Environment Server](../README.md) resources.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21

## Building The Provider

```bash
go build -o terraform-provider-envserver
```

## Using the Provider

```hcl
terraform {
  required_providers {
    envserver = {
      source = "raanand-home/envserver"
      version = "~> 1.0"
    }
  }
}

provider "envserver" {
  endpoint = "https://envserver.example.com"
  api_key  = var.envserver_api_key
}

resource "envserver_project" "example" {
  id          = "my-project"
  description = "My project"
}
```

## Developing the Provider

### Building

```bash
go build -o terraform-provider-envserver
```

### Testing

```bash
go test ./...
```

### Installing Locally

```bash
# Build
go build -o terraform-provider-envserver

# Install to local Terraform plugins directory
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/raanand-home/envserver/0.1.0/$(go env GOOS)_$(go env GOARCH)
cp terraform-provider-envserver ~/.terraform.d/plugins/registry.terraform.io/raanand-home/envserver/0.1.0/$(go env GOOS)_$(go env GOARCH)/
```

### Running Acceptance Tests

```bash
TF_ACC=1 go test ./... -v -timeout 120m
```

## Documentation

See the [plans](../plans/) directory for detailed architecture and implementation guides.

## License

[Your License Here]