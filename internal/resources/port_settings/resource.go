package port_settings

import (
	"context"
	"fmt"
	"strconv"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &portSettingResource{}
	_ resource.ResourceWithConfigure   = &portSettingResource{}
	_ resource.ResourceWithImportState = &portSettingResource{}
)

// portSettingResource is the resource implementation.
type portSettingResource struct {
	client *sdk.HRUIClient
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &portSettingResource{}
}

// Configure adds the provider configured client to the resource.
func (r *portSettingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *portSettingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_port_settings"
}

func (r *portSettingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan portSettingModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get port ID and state
	portID := int(plan.PortID.ValueInt64())
	state := 0
	if plan.Enabled.ValueBool() {
		state = 1
	}

	// Prepare API request
	port := &sdk.Port{
		ID:          portID,
		State:       state,
		SpeedDuplex: plan.Speed.Config.ValueString(),
		FlowControl: plan.FlowControl.Config.ValueString(),
	}

	// Make API request to create the VLAN
	err := r.client.UpdatePortSettings(port)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create HRUI port settings, got error: %s", err))
		return
	}

	// Save the configuration if autosave is enabled
	if r.client.Autosave {
		if err := r.client.SaveConfiguration(); err != nil {
			resp.Diagnostics.AddError("Save Error", fmt.Sprintf("Unable to save HRUI configuration, got error: %s", err))
			return
		}
	}

	// Set the ID
	plan.ID = types.StringValue(strconv.FormatInt(plan.PortID.ValueInt64(), 10))

	// Set the state with the ID included
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *portSettingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state portSettingModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	portID := int(state.PortID.ValueInt64())

	port, err := r.client.GetPort(portID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read HRUI port settings, got error: %s", err))
		return
	}

	// Update the model with the fresh data fetched from the server
	// model.PortID = types.Int64Value(int64(port.ID))
	state.Enabled = types.BoolValue(port.State == 1)
	state.Speed.Config = types.StringValue(port.SpeedDuplex)
	state.FlowControl.Config = types.StringValue(port.FlowControl)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *portSettingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan portSettingModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare API request
	portID := int(plan.PortID.ValueInt64())
	state := 0
	if plan.Enabled.ValueBool() {
		state = 1
	}

	port := &sdk.Port{
		ID:          portID,
		State:       state,
		SpeedDuplex: plan.Speed.Config.ValueString(),
		FlowControl: plan.FlowControl.Config.ValueString(),
	}

	// Make API request to create the VLAN
	err := r.client.UpdatePortSettings(port)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create HRUI port settings, got error: %s", err))
		return
	}

	// Save the configuration if autosave is enabled
	if r.client.Autosave {
		if err := r.client.SaveConfiguration(); err != nil {
			resp.Diagnostics.AddError("Save Error", fmt.Sprintf("Unable to save HRUI configuration, got error: %s", err))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *portSettingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// No-op
}

func (r *portSettingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Since the API does not return an ID, we will always
	// import the ID as is, without lookup.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("port_id"), req.ID)...)
}
