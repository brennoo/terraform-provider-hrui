package qos_port_queue

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
	_ datasource.DataSource              = &qosPortQueueDataSource{}
	_ datasource.DataSourceWithConfigure = &qosPortQueueDataSource{}
)

// qosPortQueueDataSource implements the QoS Port Queue data source.
type qosPortQueueDataSource struct {
	client *sdk.HRUIClient
}

// NewDataSource initializes a new QoS Port Queue data source.
func NewDataSource() datasource.DataSource {
	return &qosPortQueueDataSource{}
}

// Metadata returns the data source type name.
func (d *qosPortQueueDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_qos_port_queue"
}

// Schema defines the schema for the data source.
func (d *qosPortQueueDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving QoS port queue settings.",
		Attributes: map[string]schema.Attribute{
			"port": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The port name for which the QoS queue is being fetched.",
			},
			"queue": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The QoS queue setting for the specified port.",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *qosPortQueueDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = req.ProviderData.(*sdk.HRUIClient)
	}
}

// Read fetches the latest data for the QoS Port Queue data source.
func (d *qosPortQueueDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state qosPortQueueModel

	// Get config.
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate the port string from the state.
	portName := state.Port.ValueString()
	if portName == "" {
		resp.Diagnostics.AddError("Invalid Port", "The provided port name cannot be empty.")
		return
	}

	// Resolve the port name to its numeric port ID
	portID, err := d.client.GetPortByName(ctx, portName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to resolve port name '%s' to a port ID: %s", portName, err))
		return
	}

	// Fetch the QoS Port Queue data for the resolved port ID.
	portQueue, err := d.client.GetQoSPortQueue(ctx, portID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch QoS Port Queue for port '%s': %s", portName, err))
		return
	}

	// Update and set the state with the fetched data.
	state.Queue = types.Int64Value(int64(portQueue.Queue))
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
