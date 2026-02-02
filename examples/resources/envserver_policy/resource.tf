# Example 1: Create a policy using JSON format (traditional approach)
resource "envserver_policy" "developer" {
  id          = "developer-policy"
  description = "Developer access policy"

  policy = jsonencode({
    Statements = [
      {
        actions  = ["projects:get", "application:Get", "application:query"]
        effect   = "Allow"
        resource = "arn:project:my-project"
      },
      {
        actions  = ["environments:get"]
        effect   = "Allow"
        resource = "arn:environment:my-project:*"
      }
    ]
  })
}

# Example 2: Create a policy using statement blocks (new approach)
resource "envserver_policy" "deployer" {
  id          = "deployer-policy"
  description = "Deployment automation policy"

  statement {
    actions  = ["application:*", "environments:*", "operations:*"]
    effect   = "Allow"
    resource = "arn:project:my-project"
  }
}

# Example 3: Multiple statement blocks
resource "envserver_policy" "admin" {
  id          = "admin-policy"
  description = "Administrator policy with multiple statements"

  statement {
    actions  = ["projects:*"]
    effect   = "Allow"
    resource = ".*"
  }

  statement {
    actions  = ["application:*", "environments:*"]
    effect   = "Allow"
    resource = "arn:project:.*"
  }

  statement {
    actions  = ["operations:execute"]
    effect   = "Allow"
    resource = "arn:operation:.*"
  }
}

# Example 4: Read-only policy using statement blocks
resource "envserver_policy" "readonly" {
  id          = "readonly-policy"
  description = "Read-only access policy"

  statement {
    actions = [
      "projects:get",
      "application:Get",
      "application:query",
      "environments:get",
      "operations:get"
    ]
    effect   = "Allow"
    resource = ".*"
  }
}

# Attach policy to service account
resource "envserver_service_account_policy_attachment" "ci_cd_deployer" {
  service_account_id = envserver_service_account.ci_cd.id
  policy_id          = envserver_policy.deployer.id
}

# Note: You can use either 'policy' (JSON/YAML string) OR 'statement' blocks, but not both.
# The following would be INVALID:
#
# resource "envserver_policy" "invalid" {
#   id          = "invalid-policy"
#   description = "This will fail validation"
#
#   policy = jsonencode({
#     Statements = [...]
#   })
#
#   statement {
#     actions  = ["..."]
#     effect   = "Allow"
#     resource = "..."
#   }
# }
