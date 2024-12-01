package mac_static

import (
	"context"
	"fmt"
	"strings"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure implementation satisfies the resource.Resource interface
var _ resource.Resource = &macStaticResource{}

// macStaticResource manages static MAC entries on the switch
type macStaticResource struct {
	client *sdk.HRUIClient
}

// NewResource initializes a new instance of the `macStaticResource`
func NewResource() resource.Resource {
	return &macStaticResource{}
}

// Metadata sets the resource name/type.
func (r *macStaticResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mac_static"
}

// Schema defines the schema for the resource.
func (r *macStaticResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Resource for managing static MAC addresses on the HRUI device.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the resource, used internally. Format: mac_address_vlan_id",
				Computed:    true,
			},
			"mac_address": schema.StringAttribute{
				Description: "The MAC address in the format xx:xx:xx:xx:xx:xx.",
				Required:    true,
			},
			"vlan_id": schema.Int64Attribute{
				Description: "The VLAN ID to associate with the MAC address.",
				Required:    true,
			},
			"port": schema.Int64Attribute{
				Description: "The port to associate with the MAC address.",
				Required:    true,
			},
		},
	}
}

// Configure assigns the SDK client from provider configuration.
func (r *macStaticResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		client, ok := req.ProviderData.(*sdk.HRUIClient)
		if !ok {
			resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *sdk.HRUIClient")
			return
		}
		r.client = client
	}
}

// Create a new static MAC entry.
func (r *macStaticResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Extract input configuration
	var data macStaticModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use the SDK to add the static MAC entry
	err := r.client.AddStaticMACAddress(data.MACAddress.ValueString(), int(data.VLANID.ValueInt64()), int(data.Port.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create static MAC entry", err.Error())
		return
	}

	// Set resource ID to a composite of MAC and VLAN ID
	data.ID = types.StringValue(fmt.Sprintf("%s_%d", data.MACAddress.ValueString(), data.VLANID.ValueInt64()))

	// Save the state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// Read retrieves the static MAC entry from the device.
func (r *macStaticResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state macStaticModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Split the ID to fetch MAC address and VLAN info
	parts := strings.Split(state.ID.ValueString(), "_")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid Resource ID", "Expected format <mac_address>_<vlan_id>")
		return
	}

	macAddress := parts[0]
	vlanID := parts[1] // VLAN is extracted as string for validation later.

	// Fetch current MAC table from the SDK
	macTable, err := r.client.GetStaticMACAddressTable()
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch static MAC table", err.Error())
		return
	}

	// Look for the specific entry
	for _, entry := range macTable {
		if entry.MACAddress == macAddress && fmt.Sprintf("%d", entry.VLANID) == vlanID {
			state.MACAddress = types.StringValue(entry.MACAddress)
			state.VLANID = types.Int64Value(int64(entry.VLANID))
			state.Port = types.Int64Value(int64(entry.Port))

			// Update the state
			diags = resp.State.Set(ctx, &state)
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	// If entry is not found, mark the resource as removed
	resp.State.RemoveResource(ctx)
}

// Update modifies an existing static MAC entry.
func (r *macStaticResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state macStaticModel
	// Get the existing state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan macStaticModel
	// Get the desired state
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if attributes have changed
	if state.MACAddress.ValueString() != plan.MACAddress.ValueString() ||
		state.VLANID.ValueInt64() != plan.VLANID.ValueInt64() ||
		state.Port.ValueInt64() != plan.Port.ValueInt64() {
		// Delete the existing entry
		err := r.client.DeleteStaticMACAddress([]sdk.StaticMACEntry{
			{
				MACAddress: state.MACAddress.ValueString(),
				VLANID:     int(state.VLANID.ValueInt64()),
			},
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed to delete old static MAC entry", err.Error())
			return
		}

		// Add the updated entry
		err = r.client.AddStaticMACAddress(plan.MACAddress.ValueString(), int(plan.VLANID.ValueInt64()), int(plan.Port.ValueInt64()))
		if err != nil {
			resp.Diagnostics.AddError("Failed to create updated static MAC entry", err.Error())
			return
		}

		// Update the ID to reflect changes
		plan.ID = types.StringValue(fmt.Sprintf("%s_%d", plan.MACAddress.ValueString(), plan.VLANID.ValueInt64()))
	}

	// Save the updated state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete removes a static MAC entry using the SDK.
func (r *macStaticResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Extract state
	var state macStaticModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use the SDK to delete the static MAC entry
	entry := sdk.StaticMACEntry{
		MACAddress: state.MACAddress.ValueString(),
		VLANID:     int(state.VLANID.ValueInt64()),
	}
	err := r.client.DeleteStaticMACAddress([]sdk.StaticMACEntry{entry})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete static MAC entry", err.Error())
	}
}
