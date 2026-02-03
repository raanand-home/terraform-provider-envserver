package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/raanand-home/terraform-provider-envserver/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ApplicationResource{}
var _ resource.ResourceWithImportState = &ApplicationResource{}

// NewApplicationResource creates a new application resource
func NewApplicationResource() resource.Resource {
	return &ApplicationResource{}
}

// ApplicationResource defines the resource implementation
type ApplicationResource struct {
	client *client.Client
}

// ApplicationResourceModel describes the resource data model
// Based on AppsPublic schema from OpenAPI spec
type ApplicationResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Tags types.Map    `tfsdk:"tags"`
}

// Metadata returns the resource type name
func (r *ApplicationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

// Schema defines the schema for the resource
func (r *ApplicationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an application in Environment Server. Applications represent deployable units that can have instances registered in environments.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique application identifier. Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tags": schema.MapAttribute{
				MarkdownDescription: "Key-value tags associated with the application.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *ApplicationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured
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
func (r *ApplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApplicationResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert tags from Terraform types to Go map
	var tags map[string]string
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		tags = make(map[string]string)
		diags := plan.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Create API request
	app := &client.ApplicationCreate{
		ID:   plan.ID.ValueString(),
		Tags: tags,
	}

	created, err := r.client.CreateApplication(ctx, app)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Application",
			"Could not create application, unexpected error: "+err.Error(),
		)
		return
	}

	// Set computed values
	plan.ID = types.StringValue(created.ID)

	// Convert tags back to Terraform types
	if created.Tags != nil {
		tagsValue, diags := types.MapValueFrom(ctx, types.StringType, created.Tags)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Tags = tagsValue
	} else {
		plan.Tags = types.MapNull(types.StringType)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data
func (r *ApplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApplicationResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state from API
	app, err := r.client.GetApplication(ctx, state.ID.ValueString())
	if err != nil {
		// If the application is not found (404), remove it from state so Terraform will recreate it
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Application",
			"Could not read application "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update state with values from API
	state.ID = types.StringValue(app.ID)

	// Convert tags to Terraform types
	if app.Tags != nil {
		tagsValue, diags := types.MapValueFrom(ctx, types.StringType, app.Tags)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Tags = tagsValue
	} else {
		state.Tags = types.MapNull(types.StringType)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success
func (r *ApplicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ApplicationResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert tags from Terraform types to Go map
	var tags map[string]string
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		tags = make(map[string]string)
		diags := plan.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Build patch request
	patch := &client.ApplicationPatch{
		Tags: tags,
	}

	updated, err := r.client.UpdateApplication(ctx, plan.ID.ValueString(), patch)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Application",
			"Could not update application "+plan.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update state with values from API
	plan.ID = types.StringValue(updated.ID)

	// Convert tags to Terraform types
	if updated.Tags != nil {
		tagsValue, diags := types.MapValueFrom(ctx, types.StringType, updated.Tags)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Tags = tagsValue
	} else {
		plan.Tags = types.MapNull(types.StringType)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success
func (r *ApplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApplicationResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete API request
	err := r.client.DeleteApplication(ctx, state.ID.ValueString())
	if err != nil {
		// If the application is already deleted (404), consider it a success
		if client.IsNotFoundError(err) {
			return
		}

		resp.Diagnostics.AddError(
			"Error Deleting Application",
			"Could not delete application "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource into Terraform state
// Import ID format: app_id
func (r *ApplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
