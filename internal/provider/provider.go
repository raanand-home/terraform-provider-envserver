package provider

import (
	"context"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/raanand-home/terraform-provider-envserver/internal/client"
)

// envServerConfigFile represents the structure of ~/.env_server.toml
type envServerConfigFile struct {
	Username string `toml:"username"`
	URL      string `toml:"url"`
	Password string `toml:"password"`
}

// loadConfigFile attempts to load configuration from ~/.env_server.toml
func loadConfigFile() *envServerConfigFile {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	configPath := filepath.Join(homeDir, ".env_server.toml")
	var config envServerConfigFile
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil
	}

	return &config
}

// Ensure EnvServerProvider satisfies various provider interfaces
var _ provider.Provider = &EnvServerProvider{}

// EnvServerProvider defines the provider implementation
type EnvServerProvider struct {
	version string
}

// EnvServerProviderModel describes the provider data model
type EnvServerProviderModel struct {
	Endpoint           types.String `tfsdk:"endpoint"`
	Token              types.String `tfsdk:"token"`
	APIKey             types.String `tfsdk:"api_key"`
	Username           types.String `tfsdk:"username"`
	Password           types.String `tfsdk:"password"`
	OktaToken          types.String `tfsdk:"okta_token"`
	Timeout            types.Int64  `tfsdk:"timeout"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
	DefaultProjectID   types.String `tfsdk:"default_project_id"`
	DefaultEnvID       types.String `tfsdk:"default_env_id"`
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
				Description: "Environment Server API endpoint URL. Can also be set via ENVSERVER_ENDPOINT environment variable or 'url' in ~/.env_server.toml.",
				Optional:    true,
			},
			"token": schema.StringAttribute{
				Description: "Pre-authenticated token for API access. Can also be set via ENV_SERVER_TOKEN environment variable. When set, this takes precedence over other authentication methods.",
				Optional:    true,
				Sensitive:   true,
			},
			"api_key": schema.StringAttribute{
				Description: "Service account API key for authentication. Can also be set via ENVSERVER_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"username": schema.StringAttribute{
				Description: "Username for authentication. Can also be set via ENVSERVER_USERNAME environment variable or 'username' in ~/.env_server.toml.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password for authentication. Can also be set via ENVSERVER_PASSWORD environment variable or 'password' in ~/.env_server.toml.",
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
			"default_project_id": schema.StringAttribute{
				Description: "Default project ID to use for resources and data sources. Can also be set via ENVSERVER_DEFAULT_PROJECT_ID environment variable.",
				Optional:    true,
			},
			"default_env_id": schema.StringAttribute{
				Description: "Default environment ID to use for resources and data sources. Can also be set via ENVSERVER_DEFAULT_ENV_ID environment variable.",
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

	// Load config file as fallback (lowest priority)
	fileConfig := loadConfigFile()

	// Get values from environment variables if not set in config, then fall back to config file
	endpoint := getConfigValueWithFileFallback(config.Endpoint, "ENVSERVER_ENDPOINT", fileConfig, "url")
	token := getConfigValue(config.Token, "ENV_SERVER_TOKEN", "")
	apiKey := getConfigValue(config.APIKey, "ENVSERVER_API_KEY", "")
	username := getConfigValueWithFileFallback(config.Username, "ENVSERVER_USERNAME", fileConfig, "username")
	password := getConfigValueWithFileFallback(config.Password, "ENVSERVER_PASSWORD", fileConfig, "password")
	oktaToken := getConfigValue(config.OktaToken, "ENVSERVER_OKTA_TOKEN", "")
	defaultProjectID := getConfigValue(config.DefaultProjectID, "ENVSERVER_DEFAULT_PROJECT_ID", "")
	defaultEnvID := getConfigValue(config.DefaultEnvID, "ENVSERVER_DEFAULT_ENV_ID", "")

	timeout := 30
	if !config.Timeout.IsNull() {
		timeout = int(config.Timeout.ValueInt64())
	}

	// Validate required configuration
	if endpoint == "" {
		resp.Diagnostics.AddError(
			"Missing Endpoint Configuration",
			"The provider requires an endpoint URL. Set it in the provider configuration, via the ENVSERVER_ENDPOINT environment variable, or in ~/.env_server.toml.",
		)
		return
	}

	// Create API client
	clientConfig := &client.AuthConfig{
		Endpoint:         endpoint,
		Token:            token,
		APIKey:           apiKey,
		Username:         username,
		Password:         password,
		OktaToken:        oktaToken,
		Timeout:          timeout,
		DefaultProjectID: defaultProjectID,
		DefaultEnvID:     defaultEnvID,
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

// getConfigValueWithFileFallback returns the config value if set, otherwise checks environment variable,
// otherwise checks the config file, otherwise returns empty string
func getConfigValueWithFileFallback(configValue types.String, envVar string, fileConfig *envServerConfigFile, fileField string) string {
	// First priority: explicit provider config
	if !configValue.IsNull() && configValue.ValueString() != "" {
		return configValue.ValueString()
	}
	// Second priority: environment variable
	if val := os.Getenv(envVar); val != "" {
		return val
	}
	// Third priority: config file (~/.env_server.toml)
	if fileConfig != nil {
		switch fileField {
		case "url":
			if fileConfig.URL != "" {
				return fileConfig.URL
			}
		case "username":
			if fileConfig.Username != "" {
				return fileConfig.Username
			}
		case "password":
			if fileConfig.Password != "" {
				return fileConfig.Password
			}
		}
	}
	return ""
}

// Resources defines the resources implemented in the provider
func (p *EnvServerProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProjectResource,
		NewServiceAccountResource,
		NewServiceAccountAPIKeyResource,
		NewPolicyResource,
		NewServiceAccountPolicyAttachmentResource,
		NewApplicationInstanceResource,
		NewApplicationResource,
	}
}

// DataSources defines the data sources implemented in the provider
func (p *EnvServerProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewProjectDataSource,
		NewPolicyDataSource,
		NewConnectTokenDataSource,
		NewApplicationDataSource,
		NewApplicationVersionDataSource,
		NewEnvironmentConfigDataSource,
		NewEnvironmentConfigsDataSource,
	}
}