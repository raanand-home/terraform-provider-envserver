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
var _ datasource.DataSource = &ConnectTokenDataSource{}

// NewConnectTokenDataSource creates a new connect token data source
func NewConnectTokenDataSource() datasource.DataSource {
	return &ConnectTokenDataSource{}
}

// ConnectTokenDataSource defines the data source implementation
type ConnectTokenDataSource struct {
	client *client.Client
}

// ConnectTokenDataSourceModel describes the data source data model
type ConnectTokenDataSourceModel struct {
	ProjectID types.String `tfsdk:"project_id"`
	EnvID     types.String `tfsdk:"env_id"`
	Audience  types.String `tfsdk:"audience"`
	Exp       types.Int64  `tfsdk:"exp"`
	Token     types.String `tfsdk:"token"`
}

// Metadata returns the data source type name
func (d *ConnectTokenDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_connect_token"
}

// Schema defines the schema for the data source
func (d *ConnectTokenDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Generates an OIDC token for connecting to external services from an environment. This token can be used to authenticate with external services (e.g., AWS via IAM OIDC Identity Provider).",
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
			"audience": schema.StringAttribute{
				MarkdownDescription: "The audience for the OIDC token (e.g., the AWS account ID or ARN).",
				Required:            true,
			},
			"exp": schema.Int64Attribute{
				MarkdownDescription: "Token expiration time in seconds. Defaults to 3600 (1 hour).",
				Optional:            true,
				Computed:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "The generated OIDC token.",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source
func (d *ConnectTokenDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ConnectTokenDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ConnectTokenDataSourceModel

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

	// Set default expiration if not specified
	exp := int(data.Exp.ValueInt64())
	if exp == 0 {
		exp = 3600
		data.Exp = types.Int64Value(3600)
	}

	// Build request
	tokenRequest := &client.ConnectTokenRequest{
		Audience: data.Audience.ValueString(),
		Exp:      exp,
	}

	// Get connect token from API
	tokenResponse, err := d.client.GetConnectToken(ctx, projectID, envID, tokenRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Connect Token",
			"Could not get connect token for project "+projectID+" environment "+envID+": "+err.Error(),
		)
		return
	}

	// Map response to model
	data.Token = types.StringValue(tokenResponse.Token)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
