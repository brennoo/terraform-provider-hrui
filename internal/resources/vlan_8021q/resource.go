package vlan_8021q

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
)

var (
	_ resource.Resource              = &vlan8021qResource{}
	_ resource.ResourceWithConfigure = &vlan8021qResource{}
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
		Description: "Configures 802.1Q VLAN settings.",
		Attributes: map[string]schema.Attribute{
			"vlan_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "VLAN ID (1-4094). The unique identifier for the VLAN.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The VLAN name assigned to the VLAN ID.",
			},
			"untagged_ports": schema.ListAttribute{
				ElementType:         types.StringType,
				Required:            true,
				MarkdownDescription: "The list of untagged ports assigned to the VLAN (e.g., 'Port 1', 'Trunk1').",
			},
			"tagged_ports": schema.ListAttribute{
				ElementType:         types.StringType,
				Required:            true,
				MarkdownDescription: "The list of tagged ports assigned to the VLAN.",
			},
			"member_ports": schema.ListAttribute{
				ElementType:         types.StringType,
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

func extractStringList(ctx context.Context, list types.List) ([]string, diag.Diagnostics) {
	var result []string
	diags := list.ElementsAs(ctx, &result, false)
	if diags.HasError() {
		return nil, diags
	}
	return result, nil
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
	untaggedPorts, diags := extractStringList(ctx, model.UntaggedPorts)
	resp.Diagnostics.Append(diags...)
	taggedPorts, diags := extractStringList(ctx, model.TaggedPorts)
	resp.Diagnostics.Append(diags...)

	memberPorts := mergeStringPorts(taggedPorts, untaggedPorts)

	vlan := &sdk.Vlan{
		VlanID:        int(model.VlanID.ValueInt64()),
		Name:          model.Name.ValueString(),
		UntaggedPorts: untaggedPorts,
		TaggedPorts:   taggedPorts,
		MemberPorts:   memberPorts,
	}

	if err := r.client.AddVLAN(ctx, vlan); err != nil {
		resp.Diagnostics.AddError("Error creating VLAN", fmt.Sprintf("Failed to create VLAN: %s", err))
		return
	}

	model.MemberPorts, diags = types.ListValueFrom(ctx, types.StringType, memberPorts)
	resp.Diagnostics.Append(diags...)

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

	vlan, err := r.client.GetVLAN(ctx, int(state.VlanID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading VLAN", fmt.Sprintf("Could not read VLAN ID %d: %s", state.VlanID.ValueInt64(), err))
		return
	}

	state.VlanID = types.Int64Value(int64(vlan.VlanID))
	state.Name = types.StringValue(vlan.Name)

	allPorts := mergeStringPorts(vlan.TaggedPorts, vlan.UntaggedPorts)

	state.TaggedPorts, diags = types.ListValueFrom(ctx, types.StringType, vlan.TaggedPorts)
	resp.Diagnostics.Append(diags...)
	state.UntaggedPorts, diags = types.ListValueFrom(ctx, types.StringType, vlan.UntaggedPorts)
	resp.Diagnostics.Append(diags...)
	state.MemberPorts, diags = types.ListValueFrom(ctx, types.StringType, allPorts)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *vlan8021qResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan vlan8021qModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract port lists (now using string type instead of int)
	untaggedPorts, diags := extractStringList(ctx, plan.UntaggedPorts)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	taggedPorts, diags := extractStringList(ctx, plan.TaggedPorts)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Compute the full list of member ports
	memberPorts := mergeStringPorts(taggedPorts, untaggedPorts)

	// Prepare the updated VLAN data to send to the device
	vlan := &sdk.Vlan{
		VlanID:        int(plan.VlanID.ValueInt64()),
		Name:          plan.Name.ValueString(),
		UntaggedPorts: untaggedPorts,
		TaggedPorts:   taggedPorts,
		MemberPorts:   memberPorts,
	}

	if err := r.client.AddVLAN(ctx, vlan); err != nil {
		resp.Diagnostics.AddError("Error updating VLAN", fmt.Sprintf("Failed to update VLAN ID %d: %s", plan.VlanID.ValueInt64(), err))
		return
	}

	// Set the state with computed fields
	plan.MemberPorts, diags = types.ListValueFrom(ctx, types.StringType, memberPorts)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update Terraform state with the new VLAN details
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *vlan8021qResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vlan8021qModel

	// Retrieve the current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the VLAN using the SDK
	if err := r.client.RemoveVLAN(ctx, int(state.VlanID.ValueInt64())); err != nil {
		resp.Diagnostics.AddError("Error deleting VLAN", fmt.Sprintf("Failed to delete VLAN ID %d: %s", state.VlanID.ValueInt64(), err))
		return
	}
}

func mergeStringPorts(taggedPorts, untaggedPorts []string) []string {
	portSet := make(map[string]bool)

	for _, port := range taggedPorts {
		portSet[port] = true
	}

	for _, port := range untaggedPorts {
		portSet[port] = true
	}

	mergedPorts := make([]string, 0, len(portSet))
	for port := range portSet {
		mergedPorts = append(mergedPorts, port)
	}

	return mergedPorts
}
