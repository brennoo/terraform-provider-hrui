package vlan_vid

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

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &vlanVIDResource{}
	_ resource.ResourceWithConfigure   = &vlanVIDResource{}
	_ resource.ResourceWithImportState = &vlanVIDResource{}
)

// vlanVIDResource defines the resource implementation.
type vlanVIDResource struct {
	client *sdk.HRUIClient
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &vlanVIDResource{}
}

// Metadata returns the resource type name.
func (r *vlanVIDResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan_vid"
}

// Schema defines the schema for the resource.
func (r *vlanVIDResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Configures VLAN ID settings.",
		Attributes: map[string]schema.Attribute{
			"port": schema.StringAttribute{
				Required:    true,
				Description: "The name of the port (e.g., 'Port 1', 'Trunk2').",
			},
			"vlan_id": schema.Int64Attribute{
				Required:    true,
				Description: "VLAN ID to assign to the port.",
			},
			"accept_frame_type": schema.StringAttribute{
				Required:    true,
				Description: "Accepted frame type: 'All', 'Tagged', or 'Untagged'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *vlanVIDResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Helper function to resolve PortID from Port Name.
func (r *vlanVIDResource) resolvePortID(portName string) (int, error) {
	if portName == "" {
		return 0, fmt.Errorf("port name cannot be empty")
	}

	portID, err := r.client.GetPortByName(portName)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve Port ID for '%s': %w", portName, err)
	}

	// Validate PortID
	if portID <= 0 {
		return 0, fmt.Errorf("invalid Port ID '%d' resolved for port '%s'", portID, portName)
	}

	return portID, nil
}

// Create configures the port with the given VLAN ID and accepted frame type.
func (r *vlanVIDResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vlanVIDModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	portID, err := r.resolvePortID(plan.Port.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Resolving Port",
			err.Error(),
		)
		return
	}

	portConfig := &sdk.PortVLANConfig{
		PortID:          portID,
		PVID:            int(plan.VlanID.ValueInt64()),
		AcceptFrameType: plan.AcceptFrameType.ValueString(),
	}

	err = r.client.SetPortVLANConfig(portConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating VLAN VID Configuration",
			"Failed to configure the port. Details: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *vlanVIDResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vlanVIDModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	portID, err := r.resolvePortID(state.Port.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Port",
			err.Error(),
		)
		return
	}

	configs, err := r.client.ListPortVLANConfigs()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading VLAN Configuration",
			"Failed to retrieve port VLAN configurations. Details: "+err.Error(),
		)
		return
	}

	for _, config := range configs {
		if config.PortID == portID {
			state.VlanID = types.Int64Value(int64(config.PVID))
			state.AcceptFrameType = types.StringValue(config.AcceptFrameType)
			break
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *vlanVIDResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan vlanVIDModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	portID, err := r.resolvePortID(plan.Port.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Resolving Port",
			err.Error(),
		)
		return
	}

	portConfig := &sdk.PortVLANConfig{
		PortID:          portID,
		PVID:            int(plan.VlanID.ValueInt64()),
		AcceptFrameType: plan.AcceptFrameType.ValueString(),
	}

	err = r.client.SetPortVLANConfig(portConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating VLAN VID Configuration",
			"Failed to update the port. Details: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete resets the port configuration to its default state (PVID = 1).
func (r *vlanVIDResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vlanVIDModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	portName := state.Port.ValueString()

	// Reset the port configuration to default (PVID = 1, AcceptFrameType = "All")
	portConfig := &sdk.PortVLANConfig{
		PortName:        portName,
		PVID:            1,
		AcceptFrameType: "All",
	}

	if err := r.client.SetPortVLANConfig(portConfig); err != nil {
		resp.Diagnostics.AddError(
			"Error resetting VLAN VID configuration",
			"Could not reset VLAN VID configuration, unexpected error: "+err.Error(),
		)
	}
}

// ImportState imports an existing resource into Terraform.
func (r *vlanVIDResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("port"),
		req, resp)
}
