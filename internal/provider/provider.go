// provider.go - Main provider implementation for KeepHQ
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/keephq/terraform-provider-keep/internal/client"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &keepProvider{}
)

// New is a helper function to simplify provider server implementation
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &keepProvider{
			version: version,
		}
	}
}

// keepProvider is the provider implementation
type keepProvider struct {
	version string
}

// Metadata returns the provider type name.
func (p *keepProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "keep"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *keepProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The API key for KeepHQ API. Can also be set with the KEEP_API_KEY environment variable.",
			},
			"api_url": schema.StringAttribute{
				Optional:    true,
				Description: "The URL of the KeepHQ API. Defaults to http://localhost:8080. Can also be set with the KEEP_API_URL environment variable.",
			},
		},
	}
}

// Configure prepares a KeepHQ API client for data sources and resources.
func (p *keepProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Add debug logging
	ctx = tflog.SetField(ctx, "provider", "keep")
	tflog.Debug(ctx, "Configuring KeepHQ client")

	// Retrieve provider data from configuration
	var config providerModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Error retrieving provider configuration")
		return
	}

	// If practitioner provided a configuration value for any of the attributes, it must be a known value.
	if config.APIKey.IsUnknown() {
		errMsg := "The provider cannot create the KeepHQ API client as there is an unknown configuration value for the KeepHQ API key. " +
			"Either target apply the source of the value first, set the value statically in the configuration, or use the KEEP_API_KEY environment variable."
		tflog.Error(ctx, errMsg)
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown KeepHQ API Key",
			errMsg,
		)
	}

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Configuration has errors, cannot create client")
		return
	}

	// Default values
	apiKey := ""
	apiURL := "http://localhost:8080"

	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	if !config.APIURL.IsNull() {
		apiURL = config.APIURL.ValueString()
	}

	tflog.Debug(ctx, "Provider configuration", 
		map[string]interface{}{
			"api_key_set": apiKey != "",
			"api_url":     apiURL,
	})

	// Create a new KeepHQ client using the configuration values
	client, err := client.NewClient(apiURL, apiKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create KeepHQ client",
			fmt.Sprintf("Unable to create KeepHQ client: %s", err),
		)
		return
	}

	// Make the KeepHQ client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider.
func (p *keepProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Add data sources here
	}
}

// Resources defines the resources implemented in the provider.
func (p *keepProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProviderResource,
		NewExtractionRuleResource,
		NewAlertResource,
		NewMappingRuleResource,
	}
}

// providerModel maps provider schema data to a Go type
type providerModel struct {
	APIKey types.String `tfsdk:"api_key"`
	APIURL types.String `tfsdk:"api_url"`
}
