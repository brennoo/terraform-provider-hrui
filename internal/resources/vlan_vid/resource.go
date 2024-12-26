package vlan_vid

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
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
		Attributes: map[string]schema.Attribute{
			"port": schema.Int64Attribute{
				Required:    true,
				Description: "Port number.",
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

// Configure adds the provider configured client to  the resource.
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

// Create configures the port with the given VLAN ID and accepted frame type.
func (r *vlanVIDResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vlanVIDModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map the accepted frame type string to the corresponding option value
	acceptFrameTypeMap := map[string]string{
		"All":      "All",
		"Tagged":   "Tag-only",
		"Untagged": "Untag-only",
	}

	portConfig := &sdk.PortVLANConfig{
		Port:            int(plan.Port.ValueInt64()),
		PVID:            int(plan.VlanID.ValueInt64()),
		AcceptFrameType: acceptFrameTypeMap[plan.AcceptFrameType.ValueString()],
	}

	if err := r.client.SetPortVLANConfig(portConfig); err != nil {
		resp.Diagnostics.AddError(
			"Error creating VLAN VID configuration",
			"Could not create VLAN VID configuration, unexpected error: "+err.Error(),
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

	port := int(state.Port.ValueInt64())
	configs, err := r.client.ListPortVLANConfigs()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading VLAN VID configuration",
			"Could not read VLAN VID configuration: "+err.Error(),
		)
		return
	}

	for _, config := range configs {
		if config.Port == port {
			state.VlanID = types.Int64Value(int64(config.PVID))

			// Reverse the mapping from SetPortVLANConfig
			acceptFrameTypeMap := map[string]string{
				"All":        "All",
				"Tag-only":   "Tagged",
				"Untag-only": "Untagged",
			}
			state.AcceptFrameType = types.StringValue(acceptFrameTypeMap[config.AcceptFrameType])
			break
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *vlanVIDResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state vlanVIDModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map the accepted frame type string to the corresponding option value
	acceptFrameTypeMap := map[string]string{
		"All":      "All",
		"Tagged":   "Tag-only",
		"Untagged": "Untag-only",
	}

	portConfig := &sdk.PortVLANConfig{
		Port:            int(plan.Port.ValueInt64()),
		PVID:            int(plan.VlanID.ValueInt64()),
		AcceptFrameType: acceptFrameTypeMap[plan.AcceptFrameType.ValueString()],
	}

	if err := r.client.SetPortVLANConfig(portConfig); err != nil {
		resp.Diagnostics.AddError(
			"Error updating VLAN VID configuration",
			"Could not update VLAN VID configuration, unexpected error: "+err.Error(),
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

	port := int(state.Port.ValueInt64())

	// Reset the port configuration to default (PVID = 1, AcceptFrameType = "All")
	portConfig := &sdk.PortVLANConfig{
		Port:            port,
		PVID:            1,
		AcceptFrameType: "All",
	}

	if err := r.client.SetPortVLANConfig(portConfig); err != nil {
		resp.Diagnostics.AddError(
			"Error resetting VLAN VID configuration",
			"Could not reset VLAN VID configuration, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *vlanVIDResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("port"),
		req, resp)
}
