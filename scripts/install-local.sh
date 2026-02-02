#!/bin/bash
set -e

# Script to install the provider locally for testing

echo "Building terraform-provider-envserver..."
go build -o terraform-provider-envserver

# Detect OS and architecture
OS=$(go env GOOS)
ARCH=$(go env GOARCH)

# Create plugin directory
PLUGIN_DIR="${HOME}/.terraform.d/plugins/registry.terraform.io/raanand-home/envserver/0.1.0/${OS}_${ARCH}"
echo "Creating plugin directory: ${PLUGIN_DIR}"
mkdir -p "${PLUGIN_DIR}"

# Copy binary
echo "Installing provider to ${PLUGIN_DIR}"
cp terraform-provider-envserver "${PLUGIN_DIR}/"

echo "✅ Provider installed successfully!"
echo ""
echo "You can now use it in your Terraform configurations:"
echo ""
echo "terraform {"
echo "  required_providers {"
echo "    envserver = {"
echo "      source  = \"raanand-home/envserver\""
echo "      version = \"0.1.0\""
echo "    }"
echo "  }"
echo "}"