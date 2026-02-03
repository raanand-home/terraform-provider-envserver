package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/raanand-home/terraform-provider-envserver/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &EnvironmentConfigDataSource{}

// NewEnvironmentConfigDataSource creates a new environment config data source
func NewEnvironmentConfigDataSource() datasource.DataSource {
	return &EnvironmentConfigDataSource{}
}

// EnvironmentConfigDataSource defines the data source implementation
type EnvironmentConfigDataSource struct {
	client *client.Client
}

// EnvironmentConfigDataSourceModel describes the data source data model
type EnvironmentConfigDataSourceModel struct {
	ProjectID types.String `tfsdk:"project_id"`
	EnvID     types.String `tfsdk:"env_id"`
	Key       types.String `tfsdk:"key"`
	Value     types.String `tfsdk:"value"`
}

// Metadata returns the data source type name
func (d *EnvironmentConfigDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_config"
}

// Schema defines the schema for the data source
func (d *EnvironmentConfigDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a single environment configuration value by key.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project identifier. If not specified, uses the provider's default_project_id.",
				Optional:            true,
				Computed:            true,
			},
			"env_id": schema.StringAttribute{
				MarkdownDescription: "Environment identifier. If not specified, uses the provider's default_env_id.",
				Optional:            true,
				Computed:            true,
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "The configuration key to retrieve.",
				Required:            true,
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "The configuration value.",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source
func (d *EnvironmentConfigDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Read refreshes the Terraform state with the latest data
func (d *EnvironmentConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EnvironmentConfigDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use provider defaults if not specified
	projectID := data.ProjectID.ValueString()
	if projectID == "" {
		projectID = d.client.DefaultProjectID
		if projectID == "" {
			resp.Diagnostics.AddError(
				"Missing Project ID",
				"project_id must be specified either in the data source or as default_project_id in the provider configuration.",
			)
			return
		}
		data.ProjectID = types.StringValue(projectID)
	}

	envID := data.EnvID.ValueString()
	if envID == "" {
		envID = d.client.DefaultEnvID
		if envID == "" {
			resp.Diagnostics.AddError(
				"Missing Environment ID",
				"env_id must be specified either in the data source or as default_env_id in the provider configuration.",
			)
			return
		}
		data.EnvID = types.StringValue(envID)
	}

	key := data.Key.ValueString()

	// Get environment config from API
	configResponse, err := d.client.GetEnvironmentConfig(ctx, projectID, envID, key)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Environment Config",
			"Could not get environment config for project "+projectID+" environment "+envID+" key "+key+": "+err.Error(),
		)
		return
	}

	// Map response to model
	data.Value = types.StringValue(configResponse.Value)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
