package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/raanand-home/terraform-provider-envserver/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ServiceAccountPolicyAttachmentResource{}
var _ resource.ResourceWithImportState = &ServiceAccountPolicyAttachmentResource{}

// NewServiceAccountPolicyAttachmentResource creates a new service account policy attachment resource
func NewServiceAccountPolicyAttachmentResource() resource.Resource {
	return &ServiceAccountPolicyAttachmentResource{}
}

// ServiceAccountPolicyAttachmentResource defines the resource implementation
type ServiceAccountPolicyAttachmentResource struct {
	client *client.Client
}

// ServiceAccountPolicyAttachmentResourceModel describes the resource data model
type ServiceAccountPolicyAttachmentResourceModel struct {
	ID               types.String `tfsdk:"id"`
	ServiceAccountID types.String `tfsdk:"service_account_id"`
	PolicyID         types.String `tfsdk:"policy_id"`
}

// Metadata returns the resource type name
func (r *ServiceAccountPolicyAttachmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account_policy_attachment"
}

// Schema defines the schema for the resource
func (r *ServiceAccountPolicyAttachmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Attaches a policy to a service account, granting the permissions defined in the policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Attachment identifier (format: service_account_id/policy_id).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_account_id": schema.StringAttribute{
				MarkdownDescription: "Service account ID to attach the policy to. Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_id": schema.StringAttribute{
				MarkdownDescription: "Policy ID to attach. Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *ServiceAccountPolicyAttachmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state
func (r *ServiceAccountPolicyAttachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ServiceAccountPolicyAttachmentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Attach policy to service account
	err := r.client.AttachPolicyToServiceAccount(ctx, plan.ServiceAccountID.ValueString(), plan.PolicyID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Attaching Policy to Service Account",
			"Could not attach policy "+plan.PolicyID.ValueString()+" to service account "+plan.ServiceAccountID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Set ID
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", plan.ServiceAccountID.ValueString(), plan.PolicyID.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data
func (r *ServiceAccountPolicyAttachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ServiceAccountPolicyAttachmentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get service account to verify the attachment still exists
	sa, err := r.client.GetServiceAccount(ctx, state.ServiceAccountID.ValueString())
	if err != nil {
		// If the service account is not found (404), remove the attachment from state
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Service Account",
			"Could not read service account "+state.ServiceAccountID.ValueString()+": "+err.Error(),
		)
		return
	}

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

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is not supported for policy attachments
func (r *ServiceAccountPolicyAttachmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Policy attachments cannot be updated. To change an attachment, delete the existing one and create a new one.",
	)
}

// Delete deletes the resource and removes the Terraform state on success
func (r *ServiceAccountPolicyAttachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ServiceAccountPolicyAttachmentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Detach policy from service account
	err := r.client.DetachPolicyFromServiceAccount(ctx, state.ServiceAccountID.ValueString(), state.PolicyID.ValueString())
	if err != nil {
		// If the attachment is already deleted (404), consider it a success
		if client.IsNotFoundError(err) {
			return
		}

		resp.Diagnostics.AddError(
			"Error Detaching Policy from Service Account",
			"Could not detach policy "+state.PolicyID.ValueString()+" from service account "+state.ServiceAccountID.ValueString()+": "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource into Terraform state
func (r *ServiceAccountPolicyAttachmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: service_account_id/policy_id
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be in the format: service_account_id/policy_id",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_account_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}