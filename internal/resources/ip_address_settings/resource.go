package ip_address_settings

import (
	"context"
	"fmt"

	"github.com/brennoo/terraform-provider-hrui/internal/providerutil"
	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &ipAddressResource{}
	_ resource.ResourceWithConfigure   = &ipAddressResource{}
	_ resource.ResourceWithImportState = &ipAddressResource{}
)

// ipAddressResource is the resource implementation.
type ipAddressResource struct {
	client *sdk.HRUIClient
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &ipAddressResource{}
}

// Configure adds the provider configured client to the resource.
func (r *ipAddressResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Metadata defines the schema type name.
func (r *ipAddressResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_address_settings"
}

// Create operation for IP Address Settings.
func (r *ipAddressResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ipAddressModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating IP address settings")

	settings := &sdk.IPAddressSettings{
		DHCPEnabled: data.DHCPEnabled.ValueBool(),
		IPAddress:   data.IPAddress.ValueString(),
		Netmask:     data.Netmask.ValueString(),
		Gateway:     data.Gateway.ValueString(),
	}

	if err := r.client.SetIPAddressSettings(ctx, settings); err != nil {
		resp.Diagnostics.AddError("Error Creating IP Address Settings", fmt.Sprintf("Unable to create HRUI IP address settings, got error: %s", err))
		return
	}

	// Fetch updated data to sync with the latest state
	updatedSettings, err := r.client.GetIPAddressSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading IP Address Settings", fmt.Sprintf("Unable to read the latest HRUI IP address settings, got error: %s", err))
		return
	}

	data.DHCPEnabled = types.BoolValue(updatedSettings.DHCPEnabled)
	data.IPAddress = types.StringValue(updatedSettings.IPAddress)
	data.Netmask = types.StringValue(updatedSettings.Netmask)
	data.Gateway = types.StringValue(updatedSettings.Gateway)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Debug(ctx, "IP address settings created")
}

// Read function for IP Address settings.
func (r *ipAddressResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Reading IP address settings")

	settings, err := r.client.GetIPAddressSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading IP Address Settings", fmt.Sprintf("Unable to read the latest HRUI IP address settings, got error: %s", err))
		return
	}

	var data ipAddressModel
	data.DHCPEnabled = types.BoolValue(settings.DHCPEnabled)
	data.IPAddress = types.StringValue(settings.IPAddress)
	data.Netmask = types.StringValue(settings.Netmask)
	data.Gateway = types.StringValue(settings.Gateway)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Debug(ctx, "IP address settings read")
}

// Update function for IP Address settings.
func (r *ipAddressResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ipAddressModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating IP address settings")

	settings := &sdk.IPAddressSettings{
		DHCPEnabled: data.DHCPEnabled.ValueBool(),
		IPAddress:   data.IPAddress.ValueString(),
		Netmask:     data.Netmask.ValueString(),
		Gateway:     data.Gateway.ValueString(),
	}

	if err := r.client.SetIPAddressSettings(ctx, settings); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating IP Address Settings",
			fmt.Sprintf("Failed to update IP address settings: %s", err),
		)
		return
	}

	updatedSettings, err := r.client.GetIPAddressSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading IP Address Settings",
			fmt.Sprintf("Failed to fetch updated IP address settings: %s", err),
		)
		return
	}
	data.DHCPEnabled = types.BoolValue(updatedSettings.DHCPEnabled)
	data.IPAddress = types.StringValue(updatedSettings.IPAddress)
	data.Netmask = types.StringValue(updatedSettings.Netmask)
	data.Gateway = types.StringValue(updatedSettings.Gateway)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Debug(ctx, "IP address settings updated")
}

// Now implement the missing Delete method to ensure compliance with the Resource interface.
func (r *ipAddressResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Deleting IP address settings")

	// Handle the deletion logic if needed, or leave it empty if the resource cannot be deleted via API
	// For demonstration purposes, we can leave it as a no-op if IP Address settings can't be deleted from the backend.

	// For now, we'll simply remove the state without making any API calls
	resp.State.RemoveResource(ctx)
}

func (r *ipAddressResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Importing IP address settings", map[string]any{"id": req.ID})

	// ImportState is not needed for singleton resources
	// The resource will be read from the device on the next refresh
	var data ipAddressModel
	settings, err := r.client.GetIPAddressSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing IP Address Settings", fmt.Sprintf("Unable to read IP address settings during import: %s", err))
		return
	}
	data.DHCPEnabled = types.BoolValue(settings.DHCPEnabled)
	data.IPAddress = types.StringValue(settings.IPAddress)
	data.Netmask = types.StringValue(settings.Netmask)
	data.Gateway = types.StringValue(settings.Gateway)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
