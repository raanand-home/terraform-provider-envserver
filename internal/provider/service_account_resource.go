package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/your-org/terraform-provider-envserver/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ServiceAccountResource{}
var _ resource.ResourceWithImportState = &ServiceAccountResource{}

// NewServiceAccountResource creates a new service account resource
func NewServiceAccountResource() resource.Resource {
	return &ServiceAccountResource{}
}

// ServiceAccountResource defines the resource implementation
type ServiceAccountResource struct {
	client *client.Client
}

// ServiceAccountResourceModel describes the resource data model
type ServiceAccountResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	CreatedTime types.String `tfsdk:"created_time"`
	CreatedBy   types.String `tfsdk:"created_by"`
}

// Metadata returns the resource type name
func (r *ServiceAccountResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

// Schema defines the schema for the resource
func (r *ServiceAccountResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Environment Server service account. Service accounts are used for programmatic access to the API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Service account identifier. Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Service account description.",
				Optional:            true,
			},
			"created_time": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the service account was created.",
				Computed:            true,
			},
			"created_by": schema.StringAttribute{
				MarkdownDescription: "Email of the user who created the service account.",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *ServiceAccountResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *ServiceAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ServiceAccountResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create API request
	createReq := &client.CreateServiceAccountRequest{
		ID:          plan.ID.ValueString(),
		Description: plan.Description.ValueString(),
	}

	created, err := r.client.CreateServiceAccount(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Service Account",
			"Could not create service account, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to model
	plan.ID = types.StringValue(created.ID)
	// Only set description if it's not empty, otherwise keep the plan value (which may be null)
	if created.Description != "" {
		plan.Description = types.StringValue(created.Description)
	} else if plan.Description.IsNull() {
		plan.Description = types.StringNull()
	}
	plan.CreatedTime = types.StringValue(created.CreatedTime.Time.Format(time.RFC3339))
	plan.CreatedBy = types.StringValue(created.CreatedBy)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data
func (r *ServiceAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ServiceAccountResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state from API
	sa, err := r.client.GetServiceAccount(ctx, state.ID.ValueString())
	if err != nil {
		// If the service account is not found (404), remove it from state so Terraform will recreate it
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Service Account",
			"Could not read service account ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update state
	if sa.Description != "" {
		state.Description = types.StringValue(sa.Description)
	} else {
		state.Description = types.StringNull()
	}
	state.CreatedTime = types.StringValue(sa.CreatedTime.Time.Format(time.RFC3339))
	state.CreatedBy = types.StringValue(sa.CreatedBy)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success
func (r *ServiceAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ServiceAccountResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update API request
	updateReq := &client.PatchServiceAccountRequest{
		Description: plan.Description.ValueString(),
	}

	updated, err := r.client.UpdateServiceAccount(ctx, plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Service Account",
			"Could not update service account ID "+plan.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update state
	if updated.Description != "" {
		plan.Description = types.StringValue(updated.Description)
	} else if plan.Description.IsNull() {
		plan.Description = types.StringNull()
	}
	plan.CreatedTime = types.StringValue(updated.CreatedTime.Time.Format(time.RFC3339))
	plan.CreatedBy = types.StringValue(updated.CreatedBy)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success
func (r *ServiceAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ServiceAccountResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteServiceAccount(ctx, state.ID.ValueString())
	if err != nil {
		// If the service account is already deleted (404), consider it a success
		if client.IsNotFoundError(err) {
			return
		}

		resp.Diagnostics.AddError(
			"Error Deleting Service Account",
			"Could not delete service account ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource into Terraform state
func (r *ServiceAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}