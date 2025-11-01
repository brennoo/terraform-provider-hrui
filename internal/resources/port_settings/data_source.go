package port_settings

import (
	"context"
	"fmt"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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

func (d *portSettingDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving port settings.",
		Attributes: map[string]schema.Attribute{
			"port": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The port name or ID (e.g., 'Port 1', 'Trunk1').",
			},
			"enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the port is enabled.",
			},
			"speed": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Speed and duplex settings of the port.",
				Attributes: map[string]schema.Attribute{
					"config": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Configured speed and duplex mode retrieved from the system.",
					},
					"actual": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Actual speed and duplex mode returned by the system.",
					},
				},
			},
			"flow_control": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Flow control configuration of the port.",
				Attributes: map[string]schema.Attribute{
					"config": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Configured flow control setting retrieved from the system.",
					},
					"actual": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Actual flow control setting returned by the system.",
					},
				},
			},
		},
	}
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
		resp.Diagnostics.AddError(
			"Client Not Configured",
			"The HRUI client was not properly configured. Ensure the provider is set up correctly.",
		)
		return
	}

	port, err := d.client.GetPort(ctx, data.Port.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read HRUI port settings, got error: %s", err),
		)
		return
	}

	// Map the port settings into the Terraform data source state
	data.Enabled = types.BoolValue(port.State == 1)
	data.Speed = &portSettingSpeed{
		Config: types.StringValue(port.SpeedDuplexConfig),
		Actual: types.StringValue(port.SpeedDuplexActual),
	}
	data.FlowControl = &portSettingFlowControl{
		Config: types.StringValue(port.FlowControlConfig),
		Actual: types.StringValue(port.FlowControlActual),
	}

	// Save the state back to Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
