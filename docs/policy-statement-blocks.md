# Policy Statement Blocks

The `envserver_policy` resource now supports two mutually exclusive ways to define policy documents:

1. **JSON/YAML String Format** (using the `policy` attribute)
2. **Statement Blocks** (using `statement` blocks)

## Overview

You can choose either approach based on your preference, but you **cannot use both** in the same resource. The provider will validate this and return an error if both are specified.

## Method 1: JSON/YAML String Format

This is the traditional approach where you define the entire policy as a JSON or YAML string.

### Example

```hcl
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
```

### Advantages
- Familiar to users of AWS IAM policies
- Can use `jsonencode()` or `yamlencode()` functions
- Easy to copy/paste from existing policy documents
- Supports complex policy structures

## Method 2: Statement Blocks

This is a more Terraform-native approach using HCL blocks.

### Example

```hcl
resource "envserver_policy" "deployer" {
  id          = "deployer-policy"
  description = "Deployment automation policy"

  statement {
    actions  = ["application:*", "environments:*", "operations:*"]
    effect   = "Allow"
    resource = "arn:project:my-project"
  }
}
```

### Multiple Statements

You can define multiple statement blocks:

```hcl
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
```

### Advantages
- More readable and maintainable
- Better IDE support (autocomplete, validation)
- Type-safe (Terraform validates the structure)
- No need for `jsonencode()` or `yamlencode()`
- Easier to add/remove individual statements

## Statement Block Schema

Each `statement` block supports the following attributes:

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `actions` | list(string) | Yes | List of actions (e.g., `["projects:get", "application:*"]`) |
| `effect` | string | Yes | Effect of the statement. Must be `"Allow"` or `"Deny"` |
| `resource` | string | Yes | Resource ARN pattern (e.g., `"arn:project:my-project"` or `".*"`) |

## Validation

The provider enforces that you use **exactly one** of the two methods:

### ✅ Valid Configurations

```hcl
# Using policy string only
resource "envserver_policy" "example1" {
  id          = "policy1"
  description = "Using JSON format"
  policy      = jsonencode({ Statements = [...] })
}

# Using statement blocks only
resource "envserver_policy" "example2" {
  id          = "policy2"
  description = "Using statement blocks"
  
  statement {
    actions  = ["..."]
    effect   = "Allow"
    resource = "..."
  }
}
```

### ❌ Invalid Configuration

```hcl
# This will fail validation
resource "envserver_policy" "invalid" {
  id          = "invalid-policy"
  description = "This will fail"
  
  # Cannot use both!
  policy = jsonencode({ Statements = [...] })
  
  statement {
    actions  = ["..."]
    effect   = "Allow"
    resource = "..."
  }
}
```

**Error message:**
```
Error: Invalid Attribute Combination

Exactly one of these attributes must be configured:
[policy,statement]
```

## Migration Guide

### From JSON String to Statement Blocks

If you have an existing policy using the JSON string format:

```hcl
resource "envserver_policy" "old" {
  id          = "my-policy"
  description = "My policy"
  
  policy = jsonencode({
    Statements = [
      {
        actions  = ["projects:get"]
        effect   = "Allow"
        resource = "arn:project:my-project"
      }
    ]
  })
}
```

You can convert it to statement blocks:

```hcl
resource "envserver_policy" "new" {
  id          = "my-policy"
  description = "My policy"
  
  statement {
    actions  = ["projects:get"]
    effect   = "Allow"
    resource = "arn:project:my-project"
  }
}
```

**Note:** This will trigger a resource update, not a replacement, since the policy ID remains the same.

## Best Practices

1. **Choose one format and stick with it** across your Terraform configuration for consistency
2. **Use statement blocks** for new policies - they're more maintainable
3. **Use JSON format** when:
   - Importing existing policies from other systems
   - Working with complex policy structures
   - Generating policies programmatically
4. **Use statement blocks** when:
   - Writing policies from scratch
   - Policies are relatively simple
   - You want better IDE support and validation

## Implementation Details

### Internal Conversion

When you use statement blocks, the provider automatically converts them to the JSON format expected by the Environment Server API:

```hcl
statement {
  actions  = ["application:*"]
  effect   = "Allow"
  resource = "arn:project:my-project"
}
```

Becomes:

```json
{
  "Statements": [
    {
      "actions": ["application:*"],
      "effect": "Allow",
      "resource": "arn:project:my-project"
    }
  ]
}
```

### State Management

The provider preserves your chosen format in the Terraform state:
- If you use `policy`, the state stores the policy as a JSON string
- If you use `statement` blocks, the state stores individual statement blocks

This means you can safely run `terraform plan` and `terraform apply` without unexpected changes.

## Examples

See the [examples/resources/envserver_policy/resource.tf](../examples/resources/envserver_policy/resource.tf) file for complete working examples of both formats.