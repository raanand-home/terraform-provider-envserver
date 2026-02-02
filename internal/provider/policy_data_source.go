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
var _ datasource.DataSource = &PolicyDataSource{}

// NewPolicyDataSource creates a new policy data source
func NewPolicyDataSource() datasource.DataSource {
	return &PolicyDataSource{}
}

// PolicyDataSource defines the data source implementation
type PolicyDataSource struct {
	client *client.Client
}

// PolicyDataSourceModel describes the data source data model
type PolicyDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	Policy      types.String `tfsdk:"policy"`
	Managed     types.Bool   `tfsdk:"managed"`
}

// Metadata returns the data source type name
func (d *PolicyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

// Schema defines the schema for the data source
func (d *PolicyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about an existing Environment Server policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Policy identifier.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Policy description.",
				Computed:            true,
			},
			"policy": schema.StringAttribute{
				MarkdownDescription: "Policy document in YAML or JSON format.",
				Computed:            true,
			},
			"managed": schema.BoolAttribute{
				MarkdownDescription: "Whether this is a managed policy (cannot be modified or deleted).",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source
func (d *PolicyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *PolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PolicyDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get policy from API
	policy, err := d.client.GetPolicy(ctx, data.ID.ValueString())
	if err != nil {
		// If the policy is not found (404), remove it from state so Terraform will recreate it
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		
		// Other errors still fail as before
		resp.Diagnostics.AddError(
			"Error Reading Policy",
			"Could not read policy ID "+data.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response to model
	data.ID = types.StringValue(policy.ID)
	data.Description = types.StringValue(policy.Description)
	data.Policy = types.StringValue(policy.Policy)
	data.Managed = types.BoolValue(policy.Managed)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}