package qos_port_queue

import (
	"context"
	"fmt"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ datasource.DataSource              = &qosPortQueueDataSource{}
	_ datasource.DataSourceWithConfigure = &qosPortQueueDataSource{}
)

// qosPortQueueDataSource is the data source implementation
type qosPortQueueDataSource struct {
	client *sdk.HRUIClient
}

// NewDataSource is a helper function to simplify the provider implementation
func NewDataSource() datasource.DataSource {
	return &qosPortQueueDataSource{}
}

// Metadata returns the data source type name
func (d *qosPortQueueDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_qos_port_queue"
}

// Schema defines the schema for the data source
func (d *qosPortQueueDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"port_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The port number for which the queue is being configured.",
			},
			"queue": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The QoS queue setting for the specific port.",
			},
		},
	}
}

// Configure adds the provider configured client to the data source
func (d *qosPortQueueDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*sdk.HRUIClient)
}

// Read refreshes the Terraform state with the latest data
func (d *qosPortQueueDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Use the shared qosPortQueueModel
	var state qosPortQueueModel

	// Get config
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the port ID from the configuration
	portID := int(state.PortID.ValueInt64())

	// Fetch the specific QoS port queue data from the API using the SDK
	portQueue, err := d.client.GetQOSPortQueue(portID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch QoS Port Queue for port %d: %s", portID, err))
		return
	}

	// Update the state with the data fetched from the API
	state.Queue = types.Int64Value(int64(portQueue.Queue))

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
