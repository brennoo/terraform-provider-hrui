package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/provider"
)

// Configure prepares the provider for use.
func (p *hruiProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config hruiProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get values from environment variables
	url, urlOk := os.LookupEnv("HRUI_URL")
	username, usernameOk := os.LookupEnv("HRUI_USERNAME")
	password, passwordOk := os.LookupEnv("HRUI_PASSWORD")
	autosaveEnv, autosaveEnvOk := os.LookupEnv("HRUI_AUTOSAVE")

	// Check if URL is set via config or environment variable
	if !config.URL.IsNull() {
		url = config.URL.ValueString()
	} else if !urlOk {
		resp.Diagnostics.AddError(
			"Missing URL Configuration",
			"The provider cannot connect to the HRUI API without a valid 'url' being provided.",
		)
	}

	// Check if Username is set via config or environment variable
	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	} else if !usernameOk {
		resp.Diagnostics.AddError(
			"Missing Username Configuration",
			"The provider cannot connect to the HRUI API without a valid 'username' being provided.",
		)
	}

	// Check if Password is set via config or environment variable
	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	} else if !passwordOk {
		resp.Diagnostics.AddError(
			"Missing Password Configuration",
			"The provider cannot connect to the HRUI API without a valid 'password' being provided.",
		)
	}

	// Check if Autosave is set via config or environment variable
	autosave := true
	if !config.Autosave.IsNull() {
		autosave = config.Autosave.ValueBool()
	} else if autosaveEnvOk {
		var err error
		autosave, err = strconv.ParseBool(autosaveEnv)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid HRUI_AUTOSAVE Environment Variable",
				fmt.Sprintf("The HRUI_AUTOSAVE environment variable must be a boolean value, got: %s", autosaveEnv),
			)
			return
		}
	}

	// Exit if we encountered any configuration errors
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new client using the SDK
	// Replace `sdk.NewClient` with the correct method from your SDK package
	hruiClient, err := sdk.NewClient(url, username, password, autosave)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Creation Error",
			fmt.Sprintf("Failed to create SDK client: %s", err.Error()),
		)
		return
	}

	// Make a sample request to validate the API connection
	_, err = hruiClient.MakeRequest(hruiClient.URL)
	if err != nil {
		resp.Diagnostics.AddError(
			"Connection Error",
			"The provider cannot connect to the HRUI API: "+err.Error(),
		)
		return
	}

	// Assign the newly created client instance to DataSourceData and ResourceData
	// Make sure you're passing the pointer `hruiClient` as needed by the SDK
	resp.DataSourceData = hruiClient
	resp.ResourceData = hruiClient
}
