package mac_limit

import (
	"context"
	"fmt"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the required interfaces.
var (
	_ resource.Resource              = &macLimitResource{}
	_ resource.ResourceWithConfigure = &macLimitResource{}
)

// macLimitResource is the implementation of the MAC Limit Terraform resource.
type macLimitResource struct {
	client *sdk.HRUIClient
}

// NewResource creates a new instance of the MAC Limit resource.
func NewResource() resource.Resource {
	return &macLimitResource{}
}

// Metadata sets the resource name.
func (r *macLimitResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mac_limit"
}

// Schema defines the schema for the MAC Limit resource.
func (r *macLimitResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages MAC entry limits for specific ports on the HRUI system.",
		Attributes: map[string]schema.Attribute{
			"port": schema.StringAttribute{
				Required:    true,
				Description: "The name of the port to configure (e.g., 'Port 1', 'Trunk2').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the MAC limit is enabled on this port.",
			},
			"limit": schema.Int64Attribute{
				Optional:    true,
				Description: "The MAC limit value, between 0 and 4160. This is required when `enabled` is `true`.",
			},
		},
	}
}

// Configure assigns the provider-configured client to the resource.
func (r *macLimitResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// retrievePortID resolves the port name to a numeric PortID using the SDK.
func (r *macLimitResource) retrievePortID(portName string, diagnostics *diag.Diagnostics) (int, bool) {
	portID, err := r.client.GetPortByName(portName)
	if err != nil {
		diagnostics.AddError(
			"Port Resolution Error",
			fmt.Sprintf("Failed to resolve port name '%s': %s", portName, err),
		)
		return 0, false
	}
	return portID, true
}

// Create sets the MAC limit for the specified port.
func (r *macLimitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan macLimitModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve the port name to a numeric PortID.
	portID, ok := r.retrievePortID(plan.Port.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}

	// Prepare the limit value.
	var limit *int
	if !plan.Limit.IsNull() {
		value := int(plan.Limit.ValueInt64())
		limit = &value
	}

	// Call the SDK to set the MAC limit.
	err := r.client.SetMACLimit(portID, plan.Enabled.ValueBool(), limit)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Failed to set MAC limit for port '%s' (ID: %d): %s", plan.Port.ValueString(), portID, err),
		)
		return
	}

	// Set the Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read fetches the current MAC limit configuration for the resource and updates the state.
func (r *macLimitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state macLimitModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the MAC limits.
	macLimits, err := r.client.GetMACLimits()
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Failed to fetch MAC limits: %s", err),
		)
		return
	}

	// Find the limit for the specific port by name (no PortID in MACLimit).
	var found bool
	for _, limit := range macLimits {
		if limit.Port == state.Port.ValueString() { // Match on port name directly.
			state.Enabled = types.BoolValue(limit.Enabled)
			if limit.Limit != nil {
				state.Limit = types.Int64Value(int64(*limit.Limit))
			} else {
				state.Limit = types.Int64Null()
			}
			found = true
			break
		}
	}

	if !found {
		// Resource does not exist anymore.
		resp.State.RemoveResource(ctx)
		return
	}

	// Update the state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies the MAC limit for the specified port.
func (r *macLimitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan macLimitModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve the port name to a numeric PortID.
	portID, ok := r.retrievePortID(plan.Port.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}

	// Prepare the limit value.
	var limit *int
	if !plan.Limit.IsNull() {
		value := int(plan.Limit.ValueInt64())
		limit = &value
	}

	// Call the SDK to update the MAC limit.
	err := r.client.SetMACLimit(portID, plan.Enabled.ValueBool(), limit)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Failed to update MAC limit for port '%s' (ID: %d): %s", plan.Port.ValueString(), portID, err),
		)
		return
	}

	// Update the state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete disables the MAC limit for the specified port.
func (r *macLimitResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state macLimitModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve the port name to a numeric PortID.
	portID, ok := r.retrievePortID(state.Port.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}

	// Call the SDK to disable the MAC limit.
	err := r.client.SetMACLimit(portID, false, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Failed to disable MAC limit for port '%s' (ID: %d): %s", state.Port.ValueString(), portID, err),
		)
	}
}
