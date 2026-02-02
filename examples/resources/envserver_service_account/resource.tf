# Create a service account
resource "envserver_service_account" "ci_cd" {
  id          = "ci-cd-account"
  description = "Service account for CI/CD automation"
}

# Create an API key for the service account
resource "envserver_service_account_api_key" "ci_cd_key" {
  service_account_id = envserver_service_account.ci_cd.id
}

# Output the API key (sensitive)
output "ci_cd_api_key" {
  description = "API key for CI/CD service account - store this securely!"
  value       = envserver_service_account_api_key.ci_cd_key.full_key
  sensitive   = true
}

output "ci_cd_key_id" {
  description = "API key ID"
  value       = envserver_service_account_api_key.ci_cd_key.key_id
}
