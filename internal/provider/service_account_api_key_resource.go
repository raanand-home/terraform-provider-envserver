package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/raanand-home/terraform-provider-envserver/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ServiceAccountAPIKeyResource{}

// NewServiceAccountAPIKeyResource creates a new service account API key resource
func NewServiceAccountAPIKeyResource() resource.Resource {
	return &ServiceAccountAPIKeyResource{}
}

// ServiceAccountAPIKeyResource defines the resource implementation
type ServiceAccountAPIKeyResource struct {
	client *client.Client
}

// ServiceAccountAPIKeyResourceModel describes the resource data model
type ServiceAccountAPIKeyResourceModel struct {
	ServiceAccountID types.String `tfsdk:"service_account_id"`
	KeyID            types.String `tfsdk:"key_id"`
	FullKey          types.String `tfsdk:"full_key"`
	LastLogin        types.String `tfsdk:"last_login"`
}

// Metadata returns the resource type name
func (r *ServiceAccountAPIKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account_api_key"
}

// Schema defines the schema for the resource
func (r *ServiceAccountAPIKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an API key for an Environment Server service account. " +
			"**Important:** The full API key is only available during creation and cannot be retrieved later. " +
			"Store it securely immediately after creation.",
		Attributes: map[string]schema.Attribute{
			"service_account_id": schema.StringAttribute{
				MarkdownDescription: "Service account ID to create the API key for. Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key_id": schema.StringAttribute{
				MarkdownDescription: "API key identifier.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"full_key": schema.StringAttribute{
				MarkdownDescription: "Full API key. **Only available during creation.** Store this securely as it cannot be retrieved later.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_login": schema.StringAttribute{
				MarkdownDescription: "Timestamp of the last login using this API key.",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *ServiceAccountAPIKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *ServiceAccountAPIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ServiceAccountAPIKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create API key
	apiKey, err := r.client.CreateServiceAccountAPIKey(ctx, plan.ServiceAccountID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Service Account API Key",
			"Could not create API key for service account "+plan.ServiceAccountID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response to model
	plan.KeyID = types.StringValue(apiKey.KeyID)
	plan.FullKey = types.StringValue(apiKey.FullKey)
	if apiKey.LastLogin != "" {
		plan.LastLogin = types.StringValue(apiKey.LastLogin)
	} else {
		plan.LastLogin = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data
func (r *ServiceAccountAPIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ServiceAccountAPIKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state from API
	apiKey, err := r.client.GetServiceAccountAPIKey(ctx, state.KeyID.ValueString())
	if err != nil {
		// If the API key is not found (404), remove it from state so Terraform will recreate it
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		
		resp.Diagnostics.AddError(
			"Error Reading Service Account API Key3",
			"Could not read API key "+state.KeyID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update state (note: full_key is not returned by GET, so we keep the existing value)
	if apiKey.LastLogin != "" {
		state.LastLogin = types.StringValue(apiKey.LastLogin)
	} else {
		state.LastLogin = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is not supported for API keys
func (r *ServiceAccountAPIKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"API keys cannot be updated. To change an API key, delete the existing one and create a new one.",
	)
}
// Delete deletes the resource and removes the Terraform state on success
func (r *ServiceAccountAPIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ServiceAccountAPIKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the API key using the key ID
	err := r.client.DeleteServiceAccountAPIKey(ctx, state.KeyID.ValueString())
	if err != nil {
		// If the API key is already deleted (404), we can safely remove it from state
		if client.IsNotFoundError(err) {
			return
		}
		
		resp.Diagnostics.AddError(
			"Error Deleting Service Account API Key",
			"Could not delete API key "+state.KeyID.ValueString()+": "+err.Error(),
		)
		return
	}
}