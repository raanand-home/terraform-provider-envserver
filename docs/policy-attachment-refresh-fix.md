# Policy Attachment Refresh Fix

## Problem

The `envserver_service_account_policy_attachment` resource was not detecting when a policy was deleted from the backend during a Terraform refresh operation. This meant that if a policy was manually deleted or removed through the API, Terraform would not detect the drift and would continue to show the attachment as existing in the state.

## Root Cause

The issue was in the `Read` function of the service account policy attachment resource. While it was fetching the service account to verify it still exists, it was not checking whether the specific policy was still attached to that service account.

The backend API's `/api/auth/service_account/get/{id}` endpoint returns a `PublicServiceAccount` object that includes a `policies` array, but the Go client's `ServiceAccount` struct was missing this field, and the Read function had a TODO comment indicating this check was not implemented.

## Solution

The fix involved two changes:

### 1. Updated the ServiceAccount struct

**File:** [`terraform-provider-envserver/internal/client/service_accounts.go`](terraform-provider-envserver/internal/client/service_accounts.go:41-47)

Added the `Policies` field to the `ServiceAccount` struct to capture the policies returned by the API:

```go
type ServiceAccount struct {
	ID          string       `json:"id"`
	Description string       `json:"description"`
	CreatedTime FlexibleTime `json:"created_time"`
	CreatedBy   string       `json:"created_by"`
	Policies    []Policy     `json:"policies,omitempty"`
}
```

### 2. Implemented Policy Attachment Verification

**File:** [`terraform-provider-envserver/internal/provider/service_account_policy_attachment_resource.go`](terraform-provider-envserver/internal/provider/service_account_policy_attachment_resource.go:117-157)

Updated the `Read` function to check if the policy is still attached:

```go
// Check if the policy is still attached to the service account
policyAttached := false
for _, policy := range sa.Policies {
	if policy.ID == state.PolicyID.ValueString() {
		policyAttached = true
		break
	}
}

// If the policy is not attached, remove the resource from state
if !policyAttached {
	resp.State.RemoveResource(ctx)
	return
}
```

## Behavior After Fix

Now when running `terraform refresh` or `terraform plan`:

1. The provider fetches the service account from the backend
2. It checks if the service account exists (404 handling remains unchanged)
3. **NEW:** It verifies that the specific policy is still in the service account's policies list
4. If the policy is not found in the list, the resource is removed from Terraform state
5. On the next `terraform apply`, Terraform will detect the drift and recreate the attachment if it's still defined in the configuration

## Testing

To test this fix:

1. Create a service account policy attachment using Terraform
2. Manually delete the policy attachment using the API or UI
3. Run `terraform refresh`
4. Verify that Terraform detects the attachment has been removed and updates the state accordingly
5. Run `terraform plan` to see that Terraform wants to recreate the attachment

## Related Files

- [`terraform-provider-envserver/internal/client/service_accounts.go`](terraform-provider-envserver/internal/client/service_accounts.go) - Client struct definition
- [`terraform-provider-envserver/internal/provider/service_account_policy_attachment_resource.go`](terraform-provider-envserver/internal/provider/service_account_policy_attachment_resource.go) - Resource implementation
- [`src/enviorment_server/auth/routers/service_account.py`](src/enviorment_server/auth/routers/service_account.py) - Backend API endpoint