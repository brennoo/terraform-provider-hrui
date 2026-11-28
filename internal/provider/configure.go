package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/brennoo/terraform-provider-hrui/internal/providerutil"
	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
)

// ConfigureClient extracts *sdk.HRUIClient from ProviderData for use in resource/datasource Configure methods.
// Returns nil without error when ProviderData is nil (framework calls Configure before the provider is ready
// during planning/validation phases).
//
// Deprecated: Use providerutil.ConfigureClient directly. Resources and data sources must not import the provider
// package due to the import cycle it creates.
func ConfigureClient(providerData any, diags *diag.Diagnostics) *sdk.HRUIClient {
	return providerutil.ConfigureClient(providerData, diags)
}

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

	// Determine the correct URL, either from the config, environment variable, or test harness.
	if !config.URL.IsNull() {
		url = config.URL.ValueString()
	} else if !urlOk {
		if p.testURL != "" {
			url = p.testURL
		} else {
			resp.Diagnostics.AddError(
				"Missing URL Configuration",
				"'url' must be provided either in the configuration or via the 'HRUI_URL' environment variable.",
			)
			return
		}
	}

	// Determine the correct username, either from the config, environment variable, or test default.
	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	} else if !usernameOk {
		if p.testURL != "" {
			username = "admin"
		} else {
			resp.Diagnostics.AddError(
				"Missing Username Configuration",
				"'username' must be provided either in the configuration or via the 'HRUI_USERNAME' environment variable.",
			)
			return
		}
	}

	// Determine the correct password, either from the config, environment variable, or test default.
	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	} else if !passwordOk {
		if p.testURL != "" {
			password = "admin"
		} else {
			resp.Diagnostics.AddError(
				"Missing Password Configuration",
				"'password' must be provided either in the configuration or via the 'HRUI_PASSWORD' environment variable.",
			)
			return
		}
	}

	// Handle Autosave: default to true (or false in emulator test harness), support environment variable override.
	autosave := p.testURL == ""
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

	// Test connectivity with an authenticated request to /info.cgi to validate the client session.
	// Note: requesting GET / or GET /login.cgi after POST /login.cgi resets the active session on some firmwares.
	_, err = hruiClient.Request(ctx, "GET", hruiClient.URL+"/info.cgi", nil, nil)
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
