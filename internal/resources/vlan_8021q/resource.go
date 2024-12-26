package vlan_8021q

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
)

var (
	_ resource.Resource                = &vlan8021qResource{}
	_ resource.ResourceWithConfigure   = &vlan8021qResource{}
	_ resource.ResourceWithImportState = &vlan8021qResource{}
)

// vlan8021qResource defines the VLAN resource using *sdk.HRUIClient.
type vlan8021qResource struct {
	client *sdk.HRUIClient
}

// NewResource creates a new VLAN resource instance.
func NewResource() resource.Resource {
	return &vlan8021qResource{}
}

func (r *vlan8021qResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan_8021q"
}

func (r *vlan8021qResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"vlan_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "VLAN ID (1-4094). The unique identifier for the VLAN to be created or managed.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The VLAN name assigned to the given VLAN ID.",
			},
			"untagged_ports": schema.ListAttribute{
				ElementType:         types.Int64Type,
				Required:            true,
				MarkdownDescription: "The list of untagged ports assigned to the VLAN.",
			},
			"tagged_ports": schema.ListAttribute{
				ElementType:         types.Int64Type,
				Required:            true,
				MarkdownDescription: "The list of tagged ports assigned to the VLAN.",
			},
			"member_ports": schema.ListAttribute{
				ElementType:         types.Int64Type,
				Computed:            true,
				MarkdownDescription: "The list of all ports assigned to the VLAN.",
			},
		},
	}
}

func (r *vlan8021qResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*sdk.HRUIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *sdk.HRUIClient, got: %T", req.ProviderData))
		return
	}

	r.client = client
}

func extractInt64List(ctx context.Context, l types.List) ([]int, diag.Diagnostics) {
	var result []int
	var diags diag.Diagnostics
	err := l.ElementsAs(ctx, &result, false)
	if err != nil {
		diags.AddError("Error extracting list", fmt.Sprintf("Failed to extract list: %s", err))
	}

	return result, diags
}

// Create configures the VLAN with the given ID, name, and ports.
func (r *vlan8021qResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model vlan8021qModel

	// Extract the plan
	diags := req.Plan.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract the untagged and tagged ports from the model
	untaggedPorts, diags := extractInt64List(ctx, model.UntaggedPorts)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	taggedPorts, diags := extractInt64List(ctx, model.TaggedPorts)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Merge untagged and tagged to compute member_ports
	memberPorts := mergePorts(taggedPorts, untaggedPorts)

	// Retrieve the total number of ports from the backend
	totalPorts, err := r.client.GetTotalPorts()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting ports",
			"Could not get ports from device: "+err.Error(),
		)
		return
	}

	// Create VLAN structure with computed member_ports
	vlan := &sdk.Vlan{
		VlanID:        int(model.VlanID.ValueInt64()),
		Name:          model.Name.ValueString(),
		UntaggedPorts: untaggedPorts,
		TaggedPorts:   taggedPorts,
		MemberPorts:   memberPorts,
	}

	// Call AddVLAN with totalPorts
	if err := r.client.AddVLAN(vlan, totalPorts); err != nil {
		resp.Diagnostics.AddError("Error creating VLAN", fmt.Sprintf("Failed to create VLAN: %s", err))
		return
	}

	// Set computed member_ports in state
	model.MemberPorts, diags = types.ListValueFrom(ctx, types.Int64Type, memberPorts)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the final state
	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
}

func (r *vlan8021qResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vlan8021qModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch the current VLAN data from the SDK using the state VLAN ID
	vlan, err := r.client.GetVLAN(int(state.VlanID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading VLAN", fmt.Sprintf("Could not read VLAN ID %d: %s", state.VlanID.ValueInt64(), err))
		return
	}

	// Update the state with the latest fetched data
	state.VlanID = types.Int64Value(int64(vlan.VlanID))
	state.Name = types.StringValue(vlan.Name)

	// Combine tagged_ports and untagged_ports into member_ports
	allPorts := mergePorts(vlan.TaggedPorts, vlan.UntaggedPorts)

	// Set the computed field for member_ports
	state.TaggedPorts, diags = types.ListValueFrom(ctx, types.Int64Type, vlan.TaggedPorts)
	resp.Diagnostics.Append(diags...)
	state.UntaggedPorts, diags = types.ListValueFrom(ctx, types.Int64Type, vlan.UntaggedPorts)
	resp.Diagnostics.Append(diags...)
	state.MemberPorts, diags = types.ListValueFrom(ctx, types.Int64Type, allPorts)
	resp.Diagnostics.Append(diags...)

	// Set the state again with the changes
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *vlan8021qResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan vlan8021qModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract values from tagged and untagged ports
	untaggedPorts, diags := extractInt64List(ctx, plan.UntaggedPorts)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	taggedPorts, diags := extractInt64List(ctx, plan.TaggedPorts)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Recompute member_ports
	memberPorts := mergePorts(taggedPorts, untaggedPorts)

	// Create the updated VLAN in the backend SDK
	vlan := &sdk.Vlan{
		VlanID:        int(plan.VlanID.ValueInt64()),
		Name:          plan.Name.ValueString(),
		UntaggedPorts: untaggedPorts,
		TaggedPorts:   taggedPorts,
		MemberPorts:   memberPorts,
	}

	totalPorts, err := r.client.GetTotalPorts()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting ports",
			"Could not get ports from device: "+err.Error(),
		)
		return
	}

	if err := r.client.AddVLAN(vlan, totalPorts); err != nil {
		resp.Diagnostics.AddError("Error updating VLAN", fmt.Sprintf("Failed to update VLAN ID %d: %s", plan.VlanID.ValueInt64(), err))
		return
	}

	// Set the state with computed fields
	plan.MemberPorts, diags = types.ListValueFrom(ctx, types.Int64Type, memberPorts)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the state with the new data
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *vlan8021qResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vlan8021qModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the VLAN using the SDK
	if err := r.client.RemoveVLAN(int(state.VlanID.ValueInt64())); err != nil {
		resp.Diagnostics.AddError("Error deleting VLAN", err.Error())
	}
}

func (r *vlan8021qResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Use path.Root() to convert "vlan_id" to a path.Path value
	resource.ImportStatePassthroughID(ctx, path.Root("vlan_id"), req, resp)
}

func mergePorts(taggedPorts, untaggedPorts []int) []int {
	// Create a set of ports to avoid duplicates
	portSet := make(map[int]bool)

	// Add all tagged ports to the set
	for _, port := range taggedPorts {
		portSet[port] = true
	}

	// Add all untagged ports to the set
	for _, port := range untaggedPorts {
		portSet[port] = true
	}

	// Convert set back to a list
	mergedPorts := make([]int, 0, len(portSet))
	for port := range portSet {
		mergedPorts = append(mergedPorts, port)
	}

	return mergedPorts
}
