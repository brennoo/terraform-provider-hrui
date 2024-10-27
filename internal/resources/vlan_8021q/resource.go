package vlan_8021q

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
)

// vlan8021qResource defines the VLAN resource using *sdk.HRUIClient
type vlan8021qResource struct {
	client *sdk.HRUIClient
}

// Ensure that vlan8021qResource implements the resource.Resource interface
var _ resource.Resource = &vlan8021qResource{}

// NewResource creates a new VLAN resource instance
func NewResource() resource.Resource {
	return &vlan8021qResource{}
}

// Metadata sets the resource type name in Terraform
func (r *vlan8021qResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan_8021q"
}

// Schema defines the schema for the resource, similar to the data source
func (r *vlan8021qResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"vlan_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "VLAN ID (1-4094). The unique identifier for the VLAN to be created or managed.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The VLAN name assigned to the given VLAN ID.",
			},
			"untagged_ports": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.Int64Type,
				MarkdownDescription: "The list of untagged ports assigned to the VLAN.",
			},
			"tagged_ports": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.Int64Type,
				MarkdownDescription: "The list of tagged ports assigned to the VLAN.",
			},
			"member_ports": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.Int64Type,
				MarkdownDescription: "The list of all ports assigned to the VLAN.",
			},
		},
	}
}

// Configure assigns the configured HRUIClient to the resource
func (r *vlan8021qResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*sdk.HRUIClient)
	if !ok || client == nil {
		resp.Diagnostics.AddError(
			"Missing HRUI Client",
			"The client has not been properly initialized in the Configure method.",
		)
		return
	}

	// Assign the client to the resource
	r.client = client
}

// Create method
func (r *vlan8021qResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model vlan8021qModel

	// Get the user-provided configuration from Plan
	diags := req.Plan.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the VLAN
	vlan := sdk.Vlan{
		VlanID:        int(model.VlanID.ValueInt64()),
		Name:          model.Name.ValueString(),
		UntaggedPorts: sdk.ConvertToNativeIntList(model.UntaggedPorts),
		TaggedPorts:   sdk.ConvertToNativeIntList(model.TaggedPorts),
	}

	allPorts, err := r.client.GetAllPorts()
	if err != nil {
		resp.Diagnostics.AddError("Error creating VLAN", fmt.Sprintf("Could not get port count: %s", err.Error()))
		return
	}
	totalPorts := len(allPorts)
	err = r.client.CreateVLAN(&vlan, totalPorts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating VLAN",
			fmt.Sprintf("Failed to create VLAN with ID %d: %s", vlan.VlanID, err.Error()),
		)
		return
	}

	// Fetch the created VLAN data
	_, err = r.client.GetVLAN(vlan.VlanID)
	if err != nil {
		resp.Diagnostics.AddError("Error fetching VLAN after creation", fmt.Sprintf("Could not get VLAN with ID %d: %s", vlan.VlanID, err.Error()))
		return
	}

	// Set the updated state
	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
}

// Read fetches the current VLAN data from the backend system using the SDK
func (r *vlan8021qResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model vlan8021qModel

	// Read the current state from Terraform
	diags := req.State.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch VLAN data from the backend using the SDK
	vlanID := model.VlanID.ValueInt64()
	vlan, err := r.client.GetVLAN(int(vlanID))
	if err != nil {
		resp.Diagnostics.AddError("Error fetching VLAN", fmt.Sprintf("Could not get VLAN with ID %d: %s", vlanID, err.Error()))
		return
	}

	// Update the model with the fresh data fetched from the server
	model.VlanID = types.Int64Value(int64(vlan.VlanID))
	model.Name = types.StringValue(vlan.Name)
	model.UntaggedPorts = sdk.FlattenInt64List(vlan.UntaggedPorts)
	model.TaggedPorts = sdk.FlattenInt64List(vlan.TaggedPorts)
	model.MemberPorts = sdk.FlattenInt64List(vlan.MemberPorts)

	// Set the updated state
	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
}

// Update updates an existing VLAN using the SDK
func (r *vlan8021qResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model vlan8021qModel

	// Get the updated configuration from Plan
	diags := req.Plan.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform types to Go native types using the new helper function
	untaggedPorts := sdk.ConvertToNativeIntList(model.UntaggedPorts)
	taggedPorts := sdk.ConvertToNativeIntList(model.TaggedPorts)

	// Create the VLAN object for the SDK
	vlan := sdk.Vlan{
		VlanID:        int(model.VlanID.ValueInt64()),
		Name:          model.Name.ValueString(),
		UntaggedPorts: untaggedPorts,
		TaggedPorts:   taggedPorts,
	}

	// Adjust total ports accordingly (sample value)
	totalPorts := 6

	// Call SDK to update the VLAN (ensure totalPorts is passed)
	err := r.client.CreateVLAN(&vlan, totalPorts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating VLAN",
			fmt.Sprintf("Failed to update VLAN with ID %d: %s", vlan.VlanID, err.Error()),
		)
		return
	}

	// Update the Terraform state
	model.UntaggedPorts = sdk.FlattenInt64List(untaggedPorts)
	model.TaggedPorts = sdk.FlattenInt64List(taggedPorts)

	// Set the updated state
	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes an existing VLAN using the SDK
func (r *vlan8021qResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model vlan8021qModel

	// Read the current state to get the VLAN ID
	diags := req.State.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call the SDK to delete the VLAN
	err := r.client.DeleteVLAN(int(model.VlanID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting VLAN", fmt.Sprintf("Failed to delete VLAN with ID %d: %s", model.VlanID.ValueInt64(), err.Error()))
		return
	}

	// Remove resource from the state
	resp.State.RemoveResource(ctx)
}
