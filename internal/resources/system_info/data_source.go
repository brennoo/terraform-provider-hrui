package system_info

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &systemInfoDataSource{}
	_ datasource.DataSourceWithConfigure = &systemInfoDataSource{}
)

// systemInfoDataSource is the data source implementation that now uses *sdk.HRUIClient.
type systemInfoDataSource struct {
	client *sdk.HRUIClient
}

// NewDataSourceSystemInfo is a helper function to simplify the provider implementation.
func NewDataSourceSystemInfo() datasource.DataSource {
	return &systemInfoDataSource{}
}

// Configure assigns the provider-configured client to the data source.
func (d *systemInfoDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Ensure that the client exists before making assignments.
	if req.ProviderData == nil {
		return
	}

	// Cast req.ProviderData to sdk.HRUIClient instead of client.Client.
	client, ok := req.ProviderData.(*sdk.HRUIClient)
	if !ok || client == nil {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *sdk.HRUIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	// Assign the client instance to the data source.
	d.client = client
}

// Metadata defines the schema type name.
func (d *systemInfoDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_info"
}

// Read retrieves data from the HRUI system using the HRUIClient and parses the system information.
func (d *systemInfoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data systemInfoModel

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Client Not Configured",
			"The HRUI client was not properly configured. Ensure the provider is set up correctly.",
		)
		return
	}

	systemInfo, err := d.client.GetSystemInfo()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read HRUI system info, got error: %s", err))
		return
	}

	// Map the systemInfo data to your data model
	data.DeviceModel = types.StringValue(systemInfo["Device Model"])
	data.MACAddress = types.StringValue(systemInfo["MAC Address"])
	data.IPAddress = types.StringValue(systemInfo["IP Address"])
	data.Netmask = types.StringValue(systemInfo["Netmask"])
	data.Gateway = types.StringValue(systemInfo["Gateway"])
	data.FirmwareVersion = types.StringValue(systemInfo["Firmware Version"])
	data.FirmwareDate = types.StringValue(systemInfo["Firmware Date"])
	data.HardwareVersion = types.StringValue(systemInfo["Hardware Version"])

	// Set the ID (using MAC address as the unique identifier)
	id := data.MACAddress.ValueString()
	data.ID = types.StringValue(id)

	// Store the parsed data into Terraform state
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
