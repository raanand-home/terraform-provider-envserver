.PHONY: build install test testacc fmt lint docs clean

# Build the provider
build:
	go build -o terraform-provider-envserver

# Install the provider locally for testing
install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/your-org/envserver/0.1.0/$$(go env GOOS)_$$(go env GOARCH)
	cp terraform-provider-envserver ~/.terraform.d/plugins/registry.terraform.io/your-org/envserver/0.1.0/$$(go env GOOS)_$$(go env GOARCH)/

# Run unit tests
test:
	go test -v ./...

# Run acceptance tests
testacc:
	TF_ACC=1 go test -v ./... -timeout 120m

# Format code
fmt:
	go fmt ./...
	terraform fmt -recursive ./examples/

# Lint code
lint:
	golangci-lint run

# Generate documentation
docs:
	tfplugindocs generate

# Clean build artifacts
clean:
	rm -f terraform-provider-envserver
	rm -rf dist/

# Download dependencies
deps:
	go mod download
	go mod tidy

# Initialize the project (first time setup)
init: deps
	@echo "Project initialized. Run 'make build' to build the provider."