# 404 Error Handling for API Keys

## Problem
When Terraform tried to read an API key that no longer exists (404 Not Found), the provider would fail with an error instead of recognizing that the resource needs to be recreated. This caused Terraform operations to fail with messages like:

```
Could not read API key 73015828: failed to get API key: API error (status 404): {"detail":"API key not found"}
```

## Solution
The provider now properly handles 404 errors by:

1. **Detecting 404 errors**: Added an `APIError` type in [`client.go`](../internal/client/client.go) that captures the HTTP status code
2. **Removing from state**: When a 404 is detected during a Read operation, the resource is removed from Terraform state
3. **Triggering recreation**: Terraform automatically detects the missing resource and recreates it on the next apply

## Implementation Details

### Changes to `internal/client/client.go`

Added structured error handling:

```go
// APIError represents an error returned by the API
type APIError struct {
    StatusCode int
    Message    string
}

// IsNotFound returns true if the error is a 404 Not Found error
func (e *APIError) IsNotFound() bool {
    return e.StatusCode == http.StatusNotFound
}

// IsNotFoundError checks if an error is a 404 Not Found error
func IsNotFoundError(err error) bool {
    if apiErr, ok := err.(*APIError); ok {
        return apiErr.IsNotFound()
    }
    return false
}
```

### Changes to `internal/provider/service_account_api_key_resource.go`

Modified the `Read` function to handle 404 errors gracefully:

```go
func (r *ServiceAccountAPIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    // ... get state ...
    
    apiKey, err := r.client.GetServiceAccountAPIKey(ctx, state.KeyID.ValueString())
    if err != nil {
        // If the API key is not found (404), remove it from state so Terraform will recreate it
        if client.IsNotFoundError(err) {
            resp.State.RemoveResource(ctx)
            return
        }
        
        // Other errors still fail as before
        resp.Diagnostics.AddError(...)
        return
    }
    
    // ... update state ...
}
```

## Behavior

### Before
- Terraform plan/apply would fail when encountering a deleted API key
- User had to manually remove the resource from state using `terraform state rm`
- Required manual intervention to recover

### After
- Terraform automatically detects the missing API key
- Removes it from state during the refresh phase
- Plans to recreate the resource on the next apply
- No manual intervention required

## Example Workflow

1. API key exists in Terraform state but was deleted externally
2. User runs `terraform plan` or `terraform apply`
3. Provider attempts to read the API key
4. API returns 404 Not Found
5. Provider removes the resource from state
6. Terraform detects the missing resource and plans to recreate it
7. On apply, a new API key is created

## Testing

To test this behavior:

1. Create an API key using Terraform
2. Manually delete it from the Environment Server (or let it expire)
3. Run `terraform plan`
4. Verify that Terraform plans to recreate the resource instead of failing

## Related Resources

This pattern should be applied to other resources that may be deleted externally:
- Service accounts
- Policies
- Projects
- Applications
- Environments

## References

- [Terraform Plugin Framework - Resource Read](https://developer.hashicorp.com/terraform/plugin/framework/resources/read)
- [Handling Missing Resources](https://developer.hashicorp.com/terraform/plugin/framework/resources/read#recommendations)