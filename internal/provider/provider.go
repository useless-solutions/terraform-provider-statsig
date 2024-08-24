package provider

import (
	"context"
	"os"
	"regexp"

	"github.com/useless-solutions/terraform-provider-statsig/internal/service/tags"
	"github.com/useless-solutions/terraform-provider-statsig/internal/service/target_apps"
	client "github.com/useless-solutions/terraform-provider-statsig/internal/statsig"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var _ provider.Provider = &StatsigProvider{}

// StatsigProvider is the provider implementation.
type StatsigProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// StatsigProviderModel describes the provider data model.
type StatsigProviderModel struct {
	ConsoleKey types.String `tfsdk:"console_api_key"`
}

// Metadata returns the provider type name.
func (p *StatsigProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "statsig"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
// This should include an API token and endpoint.
func (p *StatsigProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"console_api_key": schema.StringAttribute{
				Required:    true,
				Description: "A Statsig Console API Key",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("^console-[a-zA-Z0-9]{3,}"), "Provided key is not a valid Console API key"),
				},
			},
		},
	}
}

// Configure prepares a Statsig API client for data sources and resources.
func (p *StatsigProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Statsig provider")
	var config StatsigProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.ConsoleKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("console_api_key"),
			"Unknown Console API Key",
			"The provider cannot create the Statsig API client as there is an unknown configuration value for the Statsig Console API Key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the STATSIG_CONSOLE_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	consoleAPIKey := os.Getenv("STATSIG_CONSOLE_KEY")
	if !config.ConsoleKey.IsNull() {
		consoleAPIKey = config.ConsoleKey.ValueString()
	}

	if consoleAPIKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("console_api_key"),
			"Missing Statsig Console API Key",
			"The provider cannot create the Statsig API client as there is a missing or empty value for the Statsig Console API Key. "+
				"Set the key value in the configuration or use the STATSIG_CONSOLE_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "console_api_key", consoleAPIKey)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "console_api_key")
	tflog.Debug(ctx, "Creating Statsig API Client")

	// Create a new Statsig client using the configuration values
	client, err := client.NewClient(ctx, consoleAPIKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Statsig API Client",
			"An unexpected error occurred when creating the Statsig API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Statsig Client Error: "+err.Error(),
		)
		return
	}

	// Make the Statsig client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Statsig provider", map[string]any{"success": true})
}

// Resources defines the resources implemented in the provider.
func (p *StatsigProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		tags.NewTagResource,
		target_apps.NewTargetAppResource,
	}
}

// DataSources defines the data sources implemented in the provider.
func (p *StatsigProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		tags.NewTagsDataSource,
		target_apps.NewTargetAppsDataSource,
	}
}

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &StatsigProvider{
			version: version,
		}
	}
}
