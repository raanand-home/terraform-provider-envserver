# How to Rebuild and Install the Provider

After making code changes, you need to rebuild and reinstall the provider for Terraform to use the updated version.

## Quick Steps

1. **Navigate to the provider directory:**
   ```bash
   cd terraform-provider-envserver
   ```

2. **Build the provider:**
   ```bash
   go build -o terraform-provider-envserver
   ```

3. **Install it locally:**
   ```bash
   bash scripts/install-local.sh
   ```

4. **Clear Terraform cache (important!):**
   ```bash
   cd /path/to/your/terraform/project
   rm -rf .terraform
   rm -rf .terraform.lock.hcl
   ```

5. **Reinitialize Terraform:**
   ```bash
   terraform init
   ```

6. **Run your Terraform command:**
   ```bash
   terraform plan
   # or
   terraform apply
   ```

## Alternative: Manual Installation

If the script doesn't work, you can manually install:

```bash
# Build
cd terraform-provider-envserver
go build -o terraform-provider-envserver

# Detect OS and architecture
OS=$(go env GOOS)
ARCH=$(go env GOARCH)

# Create plugin directory
PLUGIN_DIR="${HOME}/.terraform.d/plugins/registry.terraform.io/raanand-home/envserver/0.1.0/${OS}_${ARCH}"
mkdir -p "${PLUGIN_DIR}"

# Copy binary
cp terraform-provider-envserver "${PLUGIN_DIR}/"

# Clear Terraform cache in your project
cd /path/to/your/terraform/project
rm -rf .terraform .terraform.lock.hcl

# Reinitialize
terraform init
```

## Troubleshooting

### Still getting old errors?

1. **Check if the binary was updated:**
   ```bash
   ls -la ~/.terraform.d/plugins/registry.terraform.io/raanand-home/envserver/0.1.0/*/terraform-provider-envserver
   ```

2. **Verify the build succeeded:**
   ```bash
   cd terraform-provider-envserver
   go build -o terraform-provider-envserver
   echo $?  # Should output 0
   ```

3. **Make sure Terraform is using the local provider:**
   Check your `terraform` block in your `.tf` files - it should reference the local provider.

4. **Clear ALL Terraform caches:**
   ```bash
   # In your Terraform project directory
   rm -rf .terraform
   rm -rf .terraform.lock.hcl
   rm -rf terraform.tfstate.d  # if using workspaces
   ```

5. **Rebuild with verbose output:**
   ```bash
   cd terraform-provider-envserver
   go build -v -o terraform-provider-envserver
   ```

## Verifying the Fix

After rebuilding and reinstalling, when you run `terraform plan` or `terraform apply`:

- **Before fix:** You would see an error like:
  ```
  Error Reading Service Account API Key
  Could not read API key 73015828: failed to get API key: API error (status 404): {"detail":"API key not found"}
  ```

- **After fix:** Terraform should:
  1. Detect the missing API key (404)
  2. Remove it from state
  3. Plan to recreate it
  4. Show in the plan: `# envserver_service_account_api_key.name will be created`