package qos_port_queue

import (
	"context"
	"fmt"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &qosPortQueueResource{}
	_ resource.ResourceWithConfigure   = &qosPortQueueResource{}
	_ resource.ResourceWithImportState = &qosPortQueueResource{}
)

// qosPortQueueResource defines the resource implementation.
type qosPortQueueResource struct {
	client *sdk.HRUIClient
}

// NewResource creates a new resource instance.
func NewResource() resource.Resource {
	return &qosPortQueueResource{}
}

// Metadata returns the resource type name.
func (r *qosPortQueueResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_qos_port_queue"
}

// Schema defines the schema for the resource.
func (r *qosPortQueueResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"port": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The port name for which the QoS queue is being configured (e.g., 'Port 1', 'Trunk2').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"queue": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The QoS queue setting for the specified port.",
			},
		},
	}
}

// Configure sets up the client for the resource.
func (r *qosPortQueueResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*sdk.HRUIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.HRUIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

// Create creates a new QoS Port Queue resource.
func (r *qosPortQueueResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan qosPortQueueModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map the port name to its numeric ID.
	portID, err := r.client.GetPortByName(plan.Port.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not resolve port name '%s' to ID: %s", plan.Port.ValueString(), err))
		return
	}

	// Configure the QoS queue for the resolved port ID.
	err = r.client.SetQoSPortQueue(portID, int(plan.Queue.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create QoS Port Queue: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Update updates the QoS Port Queue resource.
func (r *qosPortQueueResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan qosPortQueueModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map the port name to its numeric ID.
	portID, err := r.client.GetPortByName(plan.Port.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not resolve port name '%s' to ID: %s", plan.Port.ValueString(), err))
		return
	}

	// Update the QoS queue for the resolved port ID.
	err = r.client.SetQoSPortQueue(portID, int(plan.Queue.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update QoS Port Queue: %s", err))
		return
	}

	// Save the updated state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the state of the QoS Port Queue resource.
func (r *qosPortQueueResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state qosPortQueueModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map the port name to its numeric ID.
	portID, err := r.client.GetPortByName(state.Port.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not resolve port name '%s' to ID: %s", state.Port.ValueString(), err))
		return
	}

	// Query the current QoS queue for the resolved port ID.
	portQueue, err := r.client.GetQoSPortQueue(portID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch QoS Port Queue for port '%s': %s", state.Port.ValueString(), err))
		return
	}

	// Update the state with the fetched queue value.
	state.Queue = types.Int64Value(int64(portQueue.Queue))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete resets the QoS Port Queue resource to its default state.
func (r *qosPortQueueResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state qosPortQueueModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map the port name to its numeric ID.
	portID, err := r.client.GetPortByName(state.Port.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not resolve port name '%s' to ID: %s", state.Port.ValueString(), err))
		return
	}

	// Reset the QoS queue to its default value (e.g., 1).
	err = r.client.SetQoSPortQueue(portID, 1)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to reset QoS Port Queue for port '%s': %s", state.Port.ValueString(), err))
		return
	}

	// Remove the resource from the state.
	resp.State.RemoveResource(ctx)
}

// ImportState imports the resource state.
func (r *qosPortQueueResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	port := req.ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("port"), port)...)
}
