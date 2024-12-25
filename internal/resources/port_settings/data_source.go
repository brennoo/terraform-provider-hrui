package port_settings

import (
	"context"
	"fmt"
	"strconv"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &portSettingDataSource{}
	_ datasource.DataSourceWithConfigure = &portSettingDataSource{}
)

// portSettingDataSource is the data source implementation.
type portSettingDataSource struct {
	client *sdk.HRUIClient
}

// NewDataSource is a helper function to instantiate the data source.
func NewDataSource() datasource.DataSource {
	return &portSettingDataSource{}
}

// Configure assigns the provider-configured client to the data source.
func (d *portSettingDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*sdk.HRUIClient)
	if !ok || client == nil {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *sdk.HRUIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	// Assign the client to the data source
	d.client = client
}

// Metadata defines the schema type name.
func (d *portSettingDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_port_settings"
}

// Read reads the current port settings from the HRUI system.
func (d *portSettingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data portSettingModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Make sure the client is set up before making any request
	if d.client == nil {
		resp.Diagnostics.AddError("Client Not Configured", "The HRUI client was not properly configured. Ensure the provider is set up correctly.")
		return
	}

	port, err := d.client.GetPort(int(data.PortID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read HRUI port settings, got error: %s", err))
		return
	}

	// Assign the state
	data.Enabled = types.BoolValue(port.State == 1)
	data.Speed = &portSettingSpeed{
		Config: types.StringValue(port.SpeedDuplex),
		Actual: types.StringValue(port.SpeedDuplex),
	}
	data.FlowControl = &portSettingFlowControl{
		Config: types.StringValue(port.FlowControl),
		Actual: types.StringValue(port.FlowControl),
	}
	data.ID = types.StringValue(strconv.FormatInt(data.PortID.ValueInt64(), 10))

	// Save the state back to Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
