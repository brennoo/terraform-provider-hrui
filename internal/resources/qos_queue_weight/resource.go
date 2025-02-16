package qos_queue_weight

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
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies expected interfaces.
var (
	_ resource.Resource                = &qosQueueWeightResource{}
	_ resource.ResourceWithConfigure   = &qosQueueWeightResource{}
	_ resource.ResourceWithImportState = &qosQueueWeightResource{}
)

// qosQueueWeightResource defines the resource implementation for queue weights.
type qosQueueWeightResource struct {
	client *sdk.HRUIClient
}

// NewResource creates a new resource instance.
func NewResource() resource.Resource {
	return &qosQueueWeightResource{}
}

// Metadata sets the resource type name in Terraform.
func (r *qosQueueWeightResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_qos_queue_weight"
}

// Schema defines the attributes for this resource.
func (r *qosQueueWeightResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Configures QoS queue weight settings.",
		Attributes: map[string]schema.Attribute{
			"queue_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The queue ID for which the weight is being configured.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"weight": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: `The weight for the queue. Can be a numerical weight from "1" to "15" or the string value "Strict priority".`,
				Validators: []validator.String{
					NewWeightValidator(),
				},
			},
		},
	}
}

// Configure stores the provider's configured SDK client.
func (r *qosQueueWeightResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates the resource by posting necessary data to apply a queue weight.
func (r *qosQueueWeightResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan qosQueueWeightModel

	// Read plan configuration values
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	queueID := int(plan.QueueID.ValueInt64())
	weightStr := plan.Weight.ValueString()

	// Convert weight to appropriate format (backing API uses integers, so normalize)
	var weight int
	if weightStr == "Strict priority" {
		weight = 0 // Use 0 for "Strict priority"
	} else {
		// Convert string weight (which should be numeric) to an integer
		parsedWeight, err := strconv.Atoi(weightStr)
		if err != nil {
			resp.Diagnostics.AddError("Weight Error", fmt.Sprintf("Unable to parse queue weight '%s' as an integer.", weightStr))
			return
		}
		weight = parsedWeight
	}

	// Apply the weight configuration using the SDK client
	err := r.client.SetQoSQueueWeight(queueID, weight)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set QoS Queue Weight: %s", err))
		return
	}

	// Set the new state of the resource in Terraform
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads the current state of a queue weight.
func (r *qosQueueWeightResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state qosQueueWeightModel

	// Fetch current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use SDK to get current queue weights and update state
	queueID := int(state.QueueID.ValueInt64())
	queueWeights, err := r.client.ListQoSQueueWeights()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch QoS Queue Weights: %s", err))
		return
	}

	for _, qWeight := range queueWeights {
		if qWeight.Queue == queueID {
			if qWeight.Weight == "Strict priority" {
				state.Weight = types.StringValue("Strict priority")
			} else {
				// Otherwise treat it as a numeric string
				state.Weight = types.StringValue(qWeight.Weight)
			}
			break
		}
	}

	// Persist new state in resource
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the queue weight.
func (r *qosQueueWeightResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan qosQueueWeightModel

	// Get the updated configuration from the Plan
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	queueID := int(plan.QueueID.ValueInt64())
	weightStr := plan.Weight.ValueString()

	var weight int
	if weightStr == "Strict priority" {
		weight = 0
	} else {
		parsedWeight, err := strconv.Atoi(weightStr)
		if err != nil {
			resp.Diagnostics.AddError("Weight Error", fmt.Sprintf("Unable to parse queue weight '%s' as an integer.", weightStr))
			return
		}
		weight = parsedWeight
	}

	// Apply the update using the SDK
	err := r.client.SetQoSQueueWeight(queueID, weight)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update QoS Queue Weight: %s", err))
		return
	}

	// Persist the latest state to Terraform
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete resets a queue to default weight.
func (r *qosQueueWeightResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state qosQueueWeightModel

	// Read current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	queueID := int(state.QueueID.ValueInt64())

	// Reset the weight to default (strict priority)
	err := r.client.SetQoSQueueWeight(queueID, 0)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to reset QoS Queue Weight: %s", err))
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

// ImportState allows importing existing configurations via "terraform import".
func (r *qosQueueWeightResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	queueID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID format", fmt.Sprintf("Expected an integer for queue_id, got: %s", req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("queue_id"), queueID)...)
}
