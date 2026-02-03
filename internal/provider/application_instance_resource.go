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
var _ resource.Resource = &ApplicationInstanceResource{}
var _ resource.ResourceWithImportState = &ApplicationInstanceResource{}

// NewApplicationInstanceResource creates a new application instance resource
func NewApplicationInstanceResource() resource.Resource {
	return &ApplicationInstanceResource{}
}

// ApplicationInstanceResource defines the resource implementation
type ApplicationInstanceResource struct {
	client *client.Client
}

// ApplicationInstanceResourceModel describes the resource data model
type ApplicationInstanceResourceModel struct {
	ID         types.String `tfsdk:"id"`
	ProjectID  types.String `tfsdk:"project_id"`
	EnvID      types.String `tfsdk:"env_id"`
	AppID      types.String `tfsdk:"app_id"`
	InstanceID types.String `tfsdk:"instance_id"`
	Tags       types.Map    `tfsdk:"tags"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

// Metadata returns the resource type name
func (r *ApplicationInstanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_instance"
}

// Schema defines the schema for the resource
func (r *ApplicationInstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an application instance registration in an Environment Server environment. Application instances represent running instances of applications within a specific environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Composite identifier in the format `project_id/env_id/app_id/instance_id`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The project ID where the environment exists. If not specified, uses the provider's default_project_id.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"env_id": schema.StringAttribute{
				MarkdownDescription: "The environment ID where the application instance will be registered. If not specified, uses the provider's default_env_id.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app_id": schema.StringAttribute{
				MarkdownDescription: "The application identifier. Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"instance_id": schema.StringAttribute{
				MarkdownDescription: "The unique instance identifier within the application. Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tags": schema.MapAttribute{
				MarkdownDescription: "Key-value tags associated with the application instance.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the application instance was created.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the application instance was last updated.",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *ApplicationInstanceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *ApplicationInstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApplicationInstanceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use provider defaults if not specified
	projectID := plan.ProjectID.ValueString()
	if projectID == "" {
		projectID = r.client.DefaultProjectID
		if projectID == "" {
			resp.Diagnostics.AddError(
				"Missing Project ID",
				"project_id must be specified either in the resource or as default_project_id in the provider configuration.",
			)
			return
		}
		plan.ProjectID = types.StringValue(projectID)
	}

	envID := plan.EnvID.ValueString()
	if envID == "" {
		envID = r.client.DefaultEnvID
		if envID == "" {
			resp.Diagnostics.AddError(
				"Missing Environment ID",
				"env_id must be specified either in the resource or as default_env_id in the provider configuration.",
			)
			return
		}
		plan.EnvID = types.StringValue(envID)
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
	instance := &client.ApplicationInstanceCreate{
		AppID:      plan.AppID.ValueString(),
		InstanceID: plan.InstanceID.ValueString(),
		Tags:       tags,
	}

	created, err := r.client.CreateApplicationInstance(
		ctx,
		projectID,
		envID,
		instance,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Application Instance",
			"Could not create application instance, unexpected error: "+err.Error(),
		)
		return
	}

	// Set computed values
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s/%s/%s",
		projectID,
		envID,
		created.AppID,
		created.InstanceID,
	))

	if created.CreatedAt != nil {
		plan.CreatedAt = types.StringValue(created.CreatedAt.String())
	}
	if created.UpdatedAt != nil {
		plan.UpdatedAt = types.StringValue(created.UpdatedAt.String())
	}

	// Convert tags back to Terraform types
	if created.Tags != nil {
		tagsValue, diags := types.MapValueFrom(ctx, types.StringType, created.Tags)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Tags = tagsValue
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data
func (r *ApplicationInstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApplicationInstanceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state from API
	instance, err := r.client.GetApplicationInstance(
		ctx,
		state.ProjectID.ValueString(),
		state.EnvID.ValueString(),
		state.AppID.ValueString(),
		state.InstanceID.ValueString(),
	)
	if err != nil {
		// If the instance is not found (404), remove it from state so Terraform will recreate it
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Application Instance",
			"Could not read application instance "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update state with values from API
	state.AppID = types.StringValue(instance.AppID)
	state.InstanceID = types.StringValue(instance.InstanceID)

	if instance.CreatedAt != nil {
		state.CreatedAt = types.StringValue(instance.CreatedAt.String())
	}
	if instance.UpdatedAt != nil {
		state.UpdatedAt = types.StringValue(instance.UpdatedAt.String())
	}

	// Convert tags to Terraform types
	if instance.Tags != nil {
		tagsValue, diags := types.MapValueFrom(ctx, types.StringType, instance.Tags)
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
func (r *ApplicationInstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ApplicationInstanceResourceModel

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
	} else {
		// API requires tags field, use empty map if not provided
		tags = make(map[string]string)
	}

	// Update API request
	patch := &client.ApplicationInstancePatch{
		Tags: tags,
	}

	updated, err := r.client.UpdateApplicationInstance(
		ctx,
		plan.ProjectID.ValueString(),
		plan.EnvID.ValueString(),
		plan.AppID.ValueString(),
		plan.InstanceID.ValueString(),
		patch,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Application Instance",
			"Could not update application instance "+plan.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update state with values from API
	if updated.UpdatedAt != nil {
		plan.UpdatedAt = types.StringValue(updated.UpdatedAt.String())
	}

	// Convert tags to Terraform types
	if updated.Tags != nil {
		tagsValue, diags := types.MapValueFrom(ctx, types.StringType, updated.Tags)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Tags = tagsValue
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success
func (r *ApplicationInstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApplicationInstanceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete API request
	err := r.client.DeleteApplicationInstance(
		ctx,
		state.ProjectID.ValueString(),
		state.EnvID.ValueString(),
		state.AppID.ValueString(),
		state.InstanceID.ValueString(),
	)
	if err != nil {
		// If the instance is already deleted (404), consider it a success
		if client.IsNotFoundError(err) {
			return
		}

		resp.Diagnostics.AddError(
			"Error Deleting Application Instance",
			"Could not delete application instance "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource into Terraform state
// Import ID format: project_id/env_id/app_id/instance_id
func (r *ApplicationInstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 4 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'project_id/env_id/app_id/instance_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("env_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app_id"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("instance_id"), parts[3])...)
}
