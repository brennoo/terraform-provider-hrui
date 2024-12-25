package qos_port_queue

import (
	"context"
	"fmt"
	"strconv"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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
			"port_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The port number for which the queue is being configured.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"queue": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The QoS queue setting for the specific port.",
			},
		},
	}
}

// Configure configures the resource client.
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

// Create creates a new QoS port queue.
func (r *qosPortQueueResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan qosPortQueueModel

	diags := req.Plan.Get(ctx, &plan)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	portID := int(plan.PortID.ValueInt64())
	queue := int(plan.Queue.ValueInt64())

	err := r.client.UpdateQOSPortQueue(portID, queue)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create QoS Port Queue: %s", err))
		return
	}

	if r.client.Autosave {
		err = r.client.SaveConfiguration()
		if err != nil {
			resp.Diagnostics.AddError("Save Error", fmt.Sprintf("Unable to save configuration: %s", err))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Update updates an existing QoS port queue.
func (r *qosPortQueueResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan qosPortQueueModel

	diags := req.Plan.Get(ctx, &plan)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	portID := int(plan.PortID.ValueInt64())
	queue := int(plan.Queue.ValueInt64())

	err := r.client.UpdateQOSPortQueue(portID, queue)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update QoS Port Queue: %s", err))
		return
	}

	if r.client.Autosave {
		err = r.client.SaveConfiguration()
		if err != nil {
			resp.Diagnostics.AddError("Save Error", fmt.Sprintf("Unable to save configuration: %s", err))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads the current state of the QoS port queue.
func (r *qosPortQueueResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state qosPortQueueModel

	diags := req.State.Get(ctx, &state)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	portID := int(state.PortID.ValueInt64())

	portQueue, err := r.client.GetQOSPortQueue(portID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch QoS Port Queue for port %d: %s", portID, err))
		return
	}

	state.Queue = types.Int64Value(int64(portQueue.Queue))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete deletes the QoS port queue.
func (r *qosPortQueueResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state qosPortQueueModel

	diags := req.State.Get(ctx, &state)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	portID := int(state.PortID.ValueInt64())

	err := r.client.UpdateQOSPortQueue(portID, 1) // Reset to queue 1
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to reset QoS Port Queue for port %d: %s", portID, err))
		return
	}

	if r.client.Autosave {
		err = r.client.SaveConfiguration()
		if err != nil {
			resp.Diagnostics.AddError("Save Error", fmt.Sprintf("Unable to save configuration: %s", err))
			return
		}
	}

	resp.State.RemoveResource(ctx)
}

// ImportState imports the resource state.
func (r *qosPortQueueResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	portID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID format", fmt.Sprintf("Expected an integer for port_id, got: %s", req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("port_id"), portID)...)
}
