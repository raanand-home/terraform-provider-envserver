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
var _ datasource.DataSource = &EnvironmentConfigsDataSource{}

// NewEnvironmentConfigsDataSource creates a new environment configs data source
func NewEnvironmentConfigsDataSource() datasource.DataSource {
	return &EnvironmentConfigsDataSource{}
}

// EnvironmentConfigsDataSource defines the data source implementation
type EnvironmentConfigsDataSource struct {
	client *client.Client
}

// EnvironmentConfigModel represents a single config in the list
type EnvironmentConfigModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

// EnvironmentConfigsDataSourceModel describes the data source data model
type EnvironmentConfigsDataSourceModel struct {
	ProjectID types.String             `tfsdk:"project_id"`
	EnvID     types.String             `tfsdk:"env_id"`
	Configs   []EnvironmentConfigModel `tfsdk:"configs"`
}

// Metadata returns the data source type name
func (d *EnvironmentConfigsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_configs"
}

// Schema defines the schema for the data source
func (d *EnvironmentConfigsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves all environment configuration values.",
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
			"configs": schema.ListNestedAttribute{
				MarkdownDescription: "List of environment configurations.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							MarkdownDescription: "The configuration key.",
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "The configuration value.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source
func (d *EnvironmentConfigsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *EnvironmentConfigsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EnvironmentConfigsDataSourceModel

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

	// Get environment configs from API
	configsResponse, err := d.client.GetEnvironmentConfigs(ctx, projectID, envID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Environment Configs",
			"Could not get environment configs for project "+projectID+" environment "+envID+": "+err.Error(),
		)
		return
	}

	// Map response to model
	configs := make([]EnvironmentConfigModel, len(configsResponse))
	for i, config := range configsResponse {
		configs[i] = EnvironmentConfigModel{
			Key:   types.StringValue(config.Key),
			Value: types.StringValue(config.Value),
		}
	}
	data.Configs = configs

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
