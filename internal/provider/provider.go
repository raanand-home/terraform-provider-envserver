package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/your-org/terraform-provider-envserver/internal/client"
)

// Ensure EnvServerProvider satisfies various provider interfaces
var _ provider.Provider = &EnvServerProvider{}

// EnvServerProvider defines the provider implementation
type EnvServerProvider struct {
	version string
}

// EnvServerProviderModel describes the provider data model
type EnvServerProviderModel struct {
	Endpoint           types.String `tfsdk:"endpoint"`
	APIKey             types.String `tfsdk:"api_key"`
	Username           types.String `tfsdk:"username"`
	Password           types.String `tfsdk:"password"`
	OktaToken          types.String `tfsdk:"okta_token"`
	Timeout            types.Int64  `tfsdk:"timeout"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
}

// New returns a new provider instance
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &EnvServerProvider{
			version: version,
		}
	}
}

// Metadata returns the provider type name
func (p *EnvServerProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "envserver"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data
func (p *EnvServerProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for Environment Server API. Manage projects, applications, environments, and access control.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Description: "Environment Server API endpoint URL. Can also be set via ENVSERVER_ENDPOINT environment variable.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "Service account API key for authentication. Can also be set via ENVSERVER_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"username": schema.StringAttribute{
				Description: "Username for authentication. Can also be set via ENVSERVER_USERNAME environment variable.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password for authentication. Can also be set via ENVSERVER_PASSWORD environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"okta_token": schema.StringAttribute{
				Description: "Okta token for authentication. Can also be set via ENVSERVER_OKTA_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"timeout": schema.Int64Attribute{
				Description: "API request timeout in seconds. Defaults to 30.",
				Optional:    true,
			},
			"insecure_skip_verify": schema.BoolAttribute{
				Description: "Skip TLS certificate verification. Not recommended for production use.",
				Optional:    true,
			},
		},
	}
}

// Configure prepares the provider for data sources and resources
func (p *EnvServerProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config EnvServerProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get values from environment variables if not set in config
	endpoint := getConfigValue(config.Endpoint, "ENVSERVER_ENDPOINT", "")
	apiKey := getConfigValue(config.APIKey, "ENVSERVER_API_KEY", "")
	username := getConfigValue(config.Username, "ENVSERVER_USERNAME", "")
	password := getConfigValue(config.Password, "ENVSERVER_PASSWORD", "")
	oktaToken := getConfigValue(config.OktaToken, "ENVSERVER_OKTA_TOKEN", "")

	timeout := 30
	if !config.Timeout.IsNull() {
		timeout = int(config.Timeout.ValueInt64())
	}

	// Validate required configuration
	if endpoint == "" {
		resp.Diagnostics.AddError(
			"Missing Endpoint Configuration",
			"The provider requires an endpoint URL. Set it in the provider configuration or via the ENVSERVER_ENDPOINT environment variable.",
		)
		return
	}

	// Create API client
	clientConfig := &client.AuthConfig{
		Endpoint:  endpoint,
		APIKey:    apiKey,
		Username:  username,
		Password:  password,
		OktaToken: oktaToken,
		Timeout:   timeout,
	}

	apiClient, err := client.NewClient(clientConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Environment Server API Client",
			"An unexpected error occurred when creating the Environment Server API client: "+err.Error(),
		)
		return
	}

	// Make the client available to data sources and resources
	resp.DataSourceData = apiClient
	resp.ResourceData = apiClient
}

// getConfigValue returns the config value if set, otherwise checks environment variable, otherwise returns default
func getConfigValue(configValue types.String, envVar, defaultValue string) string {
	if !configValue.IsNull() && configValue.ValueString() != "" {
		return configValue.ValueString()
	}
	if val := os.Getenv(envVar); val != "" {
		return val
	}
	return defaultValue
}

// Resources defines the resources implemented in the provider
func (p *EnvServerProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProjectResource,
		NewServiceAccountResource,
		NewServiceAccountAPIKeyResource,
		NewPolicyResource,
		NewServiceAccountPolicyAttachmentResource,
	}
}

// DataSources defines the data sources implemented in the provider
func (p *EnvServerProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewProjectDataSource,
		NewPolicyDataSource,
	}
}