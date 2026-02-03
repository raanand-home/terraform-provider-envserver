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
var _ datasource.DataSource = &ApplicationDataSource{}

// NewApplicationDataSource creates a new application data source
func NewApplicationDataSource() datasource.DataSource {
	return &ApplicationDataSource{}
}

// ApplicationDataSource defines the data source implementation
type ApplicationDataSource struct {
	client *client.Client
}

// ApplicationDataSourceModel describes the data source data model
type ApplicationDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Tags types.Map    `tfsdk:"tags"`
}

// Metadata returns the data source type name
func (d *ApplicationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

// Schema defines the schema for the data source
func (d *ApplicationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about an existing Environment Server application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Application identifier (app_id).",
				Required:            true,
			},
			"tags": schema.MapAttribute{
				MarkdownDescription: "Tags associated with the application.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

// Configure adds the provider configured client to the data source
func (d *ApplicationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ApplicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ApplicationDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get application from API
	app, err := d.client.GetApplication(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Application",
			"Could not read application ID "+data.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response to model
	data.ID = types.StringValue(app.ID)

	// Convert tags map to types.Map
	if app.Tags != nil {
		tagsMap, diags := types.MapValueFrom(ctx, types.StringType, app.Tags)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Tags = tagsMap
	} else {
		data.Tags = types.MapNull(types.StringType)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
