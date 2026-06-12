package mac_static

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/brennoo/terraform-provider-hrui/internal/providerutil"
	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure implementation satisfies the resource.Resource interface.
var (
	_ resource.Resource                = &macStaticResource{}
	_ resource.ResourceWithConfigure   = &macStaticResource{}
	_ resource.ResourceWithImportState = &macStaticResource{}
)

// macStaticResource manages static MAC entries on the switch.
type macStaticResource struct {
	client *sdk.HRUIClient
}

// NewResource initializes a new instance of the `macStaticResource`.
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
		Description: "Manages static MAC address entries.",
		Attributes: map[string]schema.Attribute{
			"mac_address": schema.StringAttribute{
				Description: "The MAC address in the format xx:xx:xx:xx:xx:xx.",
				Required:    true,
			},
			"vlan_id": schema.Int64Attribute{
				Description: "The VLAN ID to associate with the MAC address.",
				Required:    true,
			},
			"port": schema.StringAttribute{
				Description: "The port to associate with the MAC address (e.g., 'Port 1', 'Trunk2').",
				Required:    true,
			},
		},
	}
}

// Configure assigns the SDK client from provider configuration.
func (r *macStaticResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
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

	tflog.Debug(ctx, "Creating static MAC entry", map[string]any{"mac_address": data.MACAddress.ValueString()})

	// Use the SDK to add the static MAC entry
	err := r.client.AddStaticMACEntry(ctx, data.MACAddress.ValueString(), int(data.VLANID.ValueInt64()), data.Port.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Static MAC Entry", err.Error())
		return
	}

	// Save the state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

	tflog.Debug(ctx, "Static MAC entry created", map[string]any{"mac_address": data.MACAddress.ValueString()})
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

	tflog.Debug(ctx, "Reading static MAC entry", map[string]any{"mac_address": state.MACAddress.ValueString()})

	// Use MAC address and VLAN ID directly from state
	macAddress := state.MACAddress.ValueString()
	vlanIDStr := fmt.Sprintf("%d", state.VLANID.ValueInt64())

	// Fetch current MAC table from the SDK
	macTable, err := r.client.GetStaticMACAddressTable(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Static MAC Entry", err.Error())
		return
	}

	// Look for the specific entry
	for _, entry := range macTable {
		if entry.MACAddress == macAddress && fmt.Sprintf("%d", entry.VLANID) == vlanIDStr {
			state.MACAddress = types.StringValue(entry.MACAddress)
			state.VLANID = types.Int64Value(int64(entry.VLANID))
			state.Port = types.StringValue(entry.Port)

			// Update the state
			diags = resp.State.Set(ctx, &state)
			resp.Diagnostics.Append(diags...)

			tflog.Debug(ctx, "Static MAC entry read", map[string]any{"mac_address": state.MACAddress.ValueString()})
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

	tflog.Debug(ctx, "Updating static MAC entry", map[string]any{"mac_address": plan.MACAddress.ValueString()})

	// Check if attributes have changed
	if state.MACAddress.ValueString() != plan.MACAddress.ValueString() ||
		state.VLANID.ValueInt64() != plan.VLANID.ValueInt64() ||
		state.Port.ValueString() != plan.Port.ValueString() {
		// Delete the existing entry
		err := r.client.RemoveStaticMACEntries(ctx, []sdk.StaticMACEntry{
			{
				MACAddress: state.MACAddress.ValueString(),
				VLANID:     int(state.VLANID.ValueInt64()),
			},
		})
		if err != nil {
			resp.Diagnostics.AddError("Error Updating Static MAC Entry", err.Error())
			return
		}

		// Add the updated entry
		err = r.client.AddStaticMACEntry(ctx, plan.MACAddress.ValueString(), int(plan.VLANID.ValueInt64()), plan.Port.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error Updating Static MAC Entry", err.Error())
			return
		}
	}

	// Save the updated state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	tflog.Debug(ctx, "Static MAC entry updated", map[string]any{"mac_address": plan.MACAddress.ValueString()})
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

	tflog.Debug(ctx, "Deleting static MAC entry", map[string]any{"mac_address": state.MACAddress.ValueString()})

	// Use the SDK to delete the static MAC entry
	entry := sdk.StaticMACEntry{
		MACAddress: state.MACAddress.ValueString(),
		VLANID:     int(state.VLANID.ValueInt64()),
	}
	err := r.client.RemoveStaticMACEntries(ctx, []sdk.StaticMACEntry{entry})
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Static MAC Entry", err.Error())
	}
}

// ImportState imports an existing Static MAC Entry resource by mac_address/vlan_id composite ID.
func (r *macStaticResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Importing static MAC entry", map[string]any{"id": req.ID})

	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Error Importing Static MAC Entry", `Expected import ID format: "<mac_address>/<vlan_id>"`)
		return
	}
	vlanID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing Static MAC Entry", fmt.Sprintf("Invalid VLAN ID %q: %s", parts[1], err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("mac_address"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("vlan_id"), types.Int64Value(vlanID))...)
}
