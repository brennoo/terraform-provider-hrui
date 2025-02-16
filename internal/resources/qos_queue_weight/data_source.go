package qos_queue_weight

import (
	"context"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// qosQueueWeightDataSource defines the data source structure for a single queue.
type qosQueueWeightDataSource struct {
	client *sdk.HRUIClient
}

// NewDataSource creates a new instance of the data source for a single queue.
func NewDataSource() datasource.DataSource {
	return &qosQueueWeightDataSource{}
}

// Metadata sets the data source type name.
func (d *qosQueueWeightDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_qos_queue_weight"
}

// Schema defines the schema for the data source of a single queue.
func (d *qosQueueWeightDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source to fetch a specific QoS queue by its ID.",
		Attributes: map[string]schema.Attribute{
			"queue_id": schema.Int64Attribute{
				Description: "The ID of the queue.",
				Required:    true,
			},
			"weight": schema.StringAttribute{
				Description: "The weight of the queue. Can be a numerical value or 'Strict priority'.",
				Computed:    true,
			},
		},
	}
}

// Configure associates the client to the data source.
func (d *qosQueueWeightDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		// Set the SDK client
		client, ok := req.ProviderData.(*sdk.HRUIClient)
		if !ok {
			resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *sdk.HRUIClient")
			return
		}
		d.client = client
	}
}

// Read retrieves the weight of a specific queue based on the queue_id.
func (d *qosQueueWeightDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Retrieve queue_id from current configuration
	var state qosQueueWeightModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call the SDK to fetch all QoS Queue Weights
	fetchedQueues, err := d.client.ListQoSQueueWeights()
	if err != nil {
		resp.Diagnostics.AddError("Unable to list QoS Queue Weights", err.Error())
		return
	}

	// Find the queue that matches the given queue_id
	var found bool
	for _, qw := range fetchedQueues {
		if qw.Queue == int(state.QueueID.ValueInt64()) {
			// Set the weight in the state (this updates Terraform state)
			state.Weight = types.StringValue(qw.Weight)
			found = true
			break
		}
	}

	// If the queue was not found, return an error.
	if !found {
		resp.Diagnostics.AddError(
			"Queue Not Found",
			"Could not find a queue with the specified ID.",
		)
		return
	}

	// Set the state with the updated queue information
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
