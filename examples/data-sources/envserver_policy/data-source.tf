# Fetch an existing policy by ID
data "envserver_policy" "admin" {
  id = "admin"
}

# Use the policy data in outputs
output "admin_policy_description" {
  value = data.envserver_policy.admin.description
}

output "admin_policy_document" {
  value = data.envserver_policy.admin.policy
}

output "is_managed_policy" {
  value = data.envserver_policy.admin.managed
}

# Fetch a custom policy
data "envserver_policy" "custom" {
  id = "admin"
}

# Reference the policy in a service account attachment
resource "envserver_service_account_policy_attachment" "example" {
  service_account_id = envserver_service_account.example.id
  policy_id          = data.envserver_policy.custom.id
}
