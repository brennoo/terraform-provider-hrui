package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/provider"
)

// Configure sets up the client and prepares the provider with credentials, URLs, and optional configuration overrides.
func (p *hruiProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider configuration from the request into the hruiProviderModel struct.
	var config hruiProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Load environment variables for overrides if present.
	url, urlOk := os.LookupEnv("HRUI_URL")
	username, usernameOk := os.LookupEnv("HRUI_USERNAME")
	password, passwordOk := os.LookupEnv("HRUI_PASSWORD")
	autosaveEnv, autosaveEnvOk := os.LookupEnv("HRUI_AUTOSAVE")

	// Determine the correct URL, either from the config or environment variable.
	if !config.URL.IsNull() {
		url = config.URL.ValueString()
	} else if !urlOk {
		resp.Diagnostics.AddError(
			"Missing URL Configuration",
			"'url' must be provided either in the configuration or via the 'HRUI_URL' environment variable.",
		)
		return
	}

	// Determine the correct username, either from the config or environment variable.
	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	} else if !usernameOk {
		resp.Diagnostics.AddError(
			"Missing Username Configuration",
			"'username' must be provided either in the configuration or via the 'HRUI_USERNAME' environment variable.",
		)
		return
	}

	// Determine the correct password, either from the config or environment variable.
	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	} else if !passwordOk {
		resp.Diagnostics.AddError(
			"Missing Password Configuration",
			"'password' must be provided either in the configuration or via the 'HRUI_PASSWORD' environment variable.",
		)
		return
	}

	// Handle Autosave: default to true, support environment variable override.
	autosave := true
	if !config.Autosave.IsNull() {
		autosave = config.Autosave.ValueBool()
	} else if autosaveEnvOk {
		var err error
		autosave, err = strconv.ParseBool(autosaveEnv)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid HRUI_AUTOSAVE Environment Variable",
				fmt.Sprintf("HRUI_AUTOSAVE must be set to a valid boolean, got: %s", autosaveEnv),
			)
			return
		}
	}

	// Create a new HRUI client using the resolved configuration and credentials.
	hruiClient, err := sdk.NewClient(ctx, url, username, password, autosave, p.testHttpClient)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Creation Error",
			fmt.Sprintf("Failed to create the HRUI client: %s", err.Error()),
		)
		return
	}

	// Test connectivity with a basic request to validate the client setup.
	_, err = hruiClient.Request(ctx, "GET", hruiClient.URL, nil, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Connection Error",
			"Unable to connect to the HRUI API: "+err.Error(),
		)
		return
	}

	// Provide the HRUI client to the data sources and resources.
	resp.DataSourceData = hruiClient
	resp.ResourceData = hruiClient
}
