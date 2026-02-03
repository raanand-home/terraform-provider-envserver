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
var _ datasource.DataSource = &ApplicationVersionDataSource{}

// NewApplicationVersionDataSource creates a new application version data source
func NewApplicationVersionDataSource() datasource.DataSource {
	return &ApplicationVersionDataSource{}
}

// ApplicationVersionDataSource defines the data source implementation
type ApplicationVersionDataSource struct {
	client *client.Client
}

// ApplicationVersionDataSourceModel describes the data source data model
type ApplicationVersionDataSourceModel struct {
	AppID       types.String `tfsdk:"app_id"`
	RefID       types.String `tfsdk:"ref_id"`
	VersionID   types.String `tfsdk:"version_id"`
	ID          types.String `tfsdk:"id"`
	Ref         types.String `tfsdk:"ref"`
	VersionData types.Map    `tfsdk:"version_data"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

// Metadata returns the data source type name
func (d *ApplicationVersionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_version"
}

// Schema defines the schema for the data source
func (d *ApplicationVersionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a specific version of an Environment Server application.",
		Attributes: map[string]schema.Attribute{
			"app_id": schema.StringAttribute{
				MarkdownDescription: "Application identifier.",
				Required:            true,
			},
			"ref_id": schema.StringAttribute{
				MarkdownDescription: "Reference identifier for the version.",
				Required:            true,
			},
			"version_id": schema.StringAttribute{
				MarkdownDescription: "Version identifier.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the application version.",
				Computed:            true,
			},
			"ref": schema.StringAttribute{
				MarkdownDescription: "The reference of the application version.",
				Computed:            true,
			},
			"version_data": schema.MapAttribute{
				MarkdownDescription: "Version data associated with the application version.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the version was created.",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source
func (d *ApplicationVersionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ApplicationVersionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ApplicationVersionDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get application version from API
	version, err := d.client.GetApplicationVersion(
		ctx,
		data.AppID.ValueString(),
		data.RefID.ValueString(),
		data.VersionID.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Application Version",
			"Could not read application version for app_id "+data.AppID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response to model
	data.ID = types.StringValue(version.ID)
	data.Ref = types.StringValue(version.Ref)
	data.CreatedAt = types.StringValue(version.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"))

	// Convert version_data map to types.Map
	if version.VersionData != nil {
		versionDataMap, diags := types.MapValueFrom(ctx, types.StringType, version.VersionData)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.VersionData = versionDataMap
	} else {
		data.VersionData = types.MapNull(types.StringType)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
