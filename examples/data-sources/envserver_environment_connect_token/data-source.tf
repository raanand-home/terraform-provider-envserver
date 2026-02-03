# Example: Get a connect token for AWS OIDC federation
# This token can be used to authenticate with AWS via IAM OIDC Identity Provider

# Using explicit project_id and env_id
data "envserver_environment_connect_token" "aws_token" {
  project_id = "my-project"
  env_id     = "production"
  audience   = "sts.amazonaws.com"
  exp        = 3600 # 1 hour (default)
}

# Using provider defaults (when default_project_id and default_env_id are set)
data "envserver_environment_connect_token" "aws_token_with_defaults" {
  audience = "sts.amazonaws.com"
}

# Use the token with AWS provider for OIDC authentication
# provider "aws" {
#   assume_role_with_web_identity {
#     role_arn                = "arn:aws:iam::123456789012:role/MyOIDCRole"
#     web_identity_token      = data.envserver_environment_connect_token.aws_token.token
#     session_name            = "terraform-session"
#   }
# }

output "token_preview" {
  value     = substr(data.envserver_environment_connect_token.aws_token.token, 0, 20)
  sensitive = true
}
