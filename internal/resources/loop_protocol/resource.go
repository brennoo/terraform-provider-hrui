package loop_protocol

import (
	"context"
	"fmt"

	"github.com/brennoo/terraform-provider-hrui/internal/providerutil"
	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the Terraform resource interfaces.
var (
	_ resource.Resource                = &loopProtocolResource{}
	_ resource.ResourceWithConfigure   = &loopProtocolResource{}
	_ resource.ResourceWithImportState = &loopProtocolResource{}
)

// loopProtocolResource implements the resource for Loop Protocol configuration.
type loopProtocolResource struct {
	client *sdk.HRUIClient
}

// NewResource instantiates the resource.
func NewResource() resource.Resource {
	return &loopProtocolResource{}
}

// Metadata provides the resource type name.
func (r *loopProtocolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loop_protocol"
}

// isTimingRelevant determines if "interval_time" and "recover_time" attributes are required based on the "loop_function".
func isTimingRelevant(loopFunction string) bool {
	return loopFunction == "Loop Detection" || loopFunction == "Loop Prevention"
}

// Schema defines the attributes and schema for the resource.
func (r *loopProtocolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages loop protocol settings.",
		Attributes: map[string]schema.Attribute{
			"loop_function": schema.StringAttribute{
				Description: "Specifies the loop function mode. Valid options are 'Off', 'Loop Detection', 'Loop Prevention', and 'Spanning Tree'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("Off"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(), // Changing this value requires resource replacement.
				},
			},
			"interval_time": schema.Int64Attribute{
				Description: "The time interval in seconds for Loop Detection or Loop Prevention modes. Valid range is 1-32767 seconds.",
				Optional:    true,
				PlanModifiers: []planmodifier.Int64{
					clearIfIrrelevantTimingPlanModifier(), // Clears value if not relevant for the selected loop function.
				},
			},
			"recover_time": schema.Int64Attribute{
				Description: "Recovery time in seconds for detection/prevention modes. Must be 0 or between 4-255 seconds.",
				Optional:    true,
				PlanModifiers: []planmodifier.Int64{
					clearIfIrrelevantTimingPlanModifier(), // Clears value if not relevant for the selected loop function.
				},
			},
		},
	}
}

// clearIfIrrelevantTimingPlanModifier is a helper to clear timing values when they are unnecessary.
func clearIfIrrelevantTimingPlanModifier() planmodifier.Int64 {
	return &clearTimingIfIrrelevant{}
}

// clearTimingIfIrrelevant resets timing-related fields if they aren't applicable based on "loop_function".
type clearTimingIfIrrelevant struct{}

// Description provides a simple string description for this plan modifier.
func (m *clearTimingIfIrrelevant) Description(ctx context.Context) string {
	return "Clears 'interval_time' and 'recover_time' to null when they are irrelevant for the selected 'loop_function'."
}

// MarkdownDescription provides a markdown-friendly description.
func (m *clearTimingIfIrrelevant) MarkdownDescription(ctx context.Context) string {
	return "Sets `interval_time` and `recover_time` to null when they are unnecessary for the selected `loop_function`."
}

// PlanModifyInt64 analyzes the context and loop function to clear timing fields if unnecessary.
func (m *clearTimingIfIrrelevant) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	var loopFunction types.String

	diags := req.Plan.GetAttribute(ctx, path.Root("loop_function"), &loopFunction)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fallback to `state` if `loop_function` isn't in the plan.
	if loopFunction.IsNull() || loopFunction.IsUnknown() {
		diags = req.State.GetAttribute(ctx, path.Root("loop_function"), &loopFunction)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Clear timing fields if they are not relevant.
	if !isTimingRelevant(loopFunction.ValueString()) {
		// Nullify values only if they are known in the plan.
		if !req.PlanValue.IsUnknown() {
			resp.PlanValue = types.Int64Null()
		}
	}
}

// Configure sets the SDK client for the resource.
func (r *loopProtocolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Create provisions or initializes the resource with the specified settings.
func (r *loopProtocolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Creating loop protocol settings")

	var plan loopProtocolModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	loopFunction := plan.LoopFunction.ValueString()
	intervalTime, recoverTime := 0, 0

	// Set interval and recovery times only if they are specified.
	if !plan.IntervalTime.IsNull() {
		intervalTime = int(plan.IntervalTime.ValueInt64())
	}
	if !plan.RecoverTime.IsNull() {
		recoverTime = int(plan.RecoverTime.ValueInt64())
	}

	// Avoid managing port statuses since they are not handled in this resource.
	portStatuses := []sdk.PortStatus{}

	err := r.client.ConfigureLoopProtocol(ctx, loopFunction, intervalTime, recoverTime, portStatuses)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Loop Protocol", fmt.Sprintf("Unable to create Loop Protocol: %s", err.Error()))
		return
	}

	// Read back from the device to ensure state matches what was actually applied
	loopProtocol, err := r.client.GetLoopProtocol(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Loop Protocol", fmt.Sprintf("Unable to read Loop Protocol after creation: %s", err.Error()))
		return
	}

	// Update state with values from device
	state := loopProtocolModel{
		LoopFunction: types.StringValue(loopProtocol.LoopFunction),
	}

	if isTimingRelevant(loopProtocol.LoopFunction) {
		state.IntervalTime = types.Int64Value(int64(loopProtocol.IntervalTime))
		state.RecoverTime = types.Int64Value(int64(loopProtocol.RecoverTime))
	} else {
		state.IntervalTime = types.Int64Null()
		state.RecoverTime = types.Int64Null()
	}

	// Set the state based on what was read from the device
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	tflog.Debug(ctx, "Loop protocol settings created")
}

// Read synchronizes the Terraform state with current API schema settings.
func (r *loopProtocolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Reading loop protocol settings")

	var state loopProtocolModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	loopProtocol, err := r.client.GetLoopProtocol(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Loop Protocol", fmt.Sprintf("Unable to read Loop Protocol: %s", err.Error()))
		return
	}

	// Update the state according to the current values from the API.
	state.LoopFunction = types.StringValue(loopProtocol.LoopFunction)

	if isTimingRelevant(loopProtocol.LoopFunction) {
		state.IntervalTime = types.Int64Value(int64(loopProtocol.IntervalTime))
		state.RecoverTime = types.Int64Value(int64(loopProtocol.RecoverTime))
	} else {
		state.IntervalTime = types.Int64Null()
		state.RecoverTime = types.Int64Null()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	tflog.Debug(ctx, "Loop protocol settings read")
}

// Update modifies the existing resource configuration.
func (r *loopProtocolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Updating loop protocol settings")

	var plan loopProtocolModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	loopFunction := plan.LoopFunction.ValueString()
	intervalTime, recoverTime := 0, 0

	// Set interval and recovery times based on the plan.
	if !plan.IntervalTime.IsNull() {
		intervalTime = int(plan.IntervalTime.ValueInt64())
	}
	if !plan.RecoverTime.IsNull() {
		recoverTime = int(plan.RecoverTime.ValueInt64())
	}

	// Avoid managing port statuses since they are not handled in this resource.
	portStatuses := []sdk.PortStatus{}

	err := r.client.ConfigureLoopProtocol(ctx, loopFunction, intervalTime, recoverTime, portStatuses)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Loop Protocol", fmt.Sprintf("Unable to update Loop Protocol: %s", err.Error()))
		return
	}

	// Read back from the device to ensure state matches what was actually applied
	loopProtocol, err := r.client.GetLoopProtocol(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Loop Protocol", fmt.Sprintf("Unable to read Loop Protocol after update: %s", err.Error()))
		return
	}

	// Update state with values from device
	state := loopProtocolModel{
		LoopFunction: types.StringValue(loopProtocol.LoopFunction),
	}

	if isTimingRelevant(loopProtocol.LoopFunction) {
		state.IntervalTime = types.Int64Value(int64(loopProtocol.IntervalTime))
		state.RecoverTime = types.Int64Value(int64(loopProtocol.RecoverTime))
	} else {
		state.IntervalTime = types.Int64Null()
		state.RecoverTime = types.Int64Null()
	}

	// Update the state to match what was read from the device
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	tflog.Debug(ctx, "Loop protocol settings updated")
}

// Delete deactivates the protocol and removes the resource from state.
func (r *loopProtocolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Deleting loop protocol settings")

	// Use an empty slice since port statuses are not handled.
	portStatuses := []sdk.PortStatus{}

	err := r.client.ConfigureLoopProtocol(ctx, "Off", 0, 0, portStatuses)
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Loop Protocol", fmt.Sprintf("Unable to delete Loop Protocol: %s", err.Error()))
		return
	}

	// Remove resource from Terraform state.
	resp.State.RemoveResource(ctx)
}

// ImportState imports an existing Loop Protocol resource by fetching the current state.
func (r *loopProtocolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Importing loop protocol settings", map[string]any{"id": req.ID})

	loopProtocol, err := r.client.GetLoopProtocol(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing Loop Protocol", fmt.Sprintf("Unable to import loop protocol settings: %s", err))
		return
	}
	state := loopProtocolModel{
		LoopFunction: types.StringValue(loopProtocol.LoopFunction),
	}
	if isTimingRelevant(loopProtocol.LoopFunction) {
		state.IntervalTime = types.Int64Value(int64(loopProtocol.IntervalTime))
		state.RecoverTime = types.Int64Value(int64(loopProtocol.RecoverTime))
	} else {
		state.IntervalTime = types.Int64Null()
		state.RecoverTime = types.Int64Null()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
