package igmp_snooping_static

import (
	"context"
	"fmt"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure implementation satisfies resource interfaces.
var (
	_ resource.Resource              = &igmpSnoopingStaticResource{}
	_ resource.ResourceWithConfigure = &igmpSnoopingStaticResource{}
)

type igmpSnoopingStaticResource struct {
	client *sdk.HRUIClient
}

// New creates a new resource instance.
func NewResource() resource.Resource {
	return &igmpSnoopingStaticResource{}
}

// Metadata defines the resource type name.
func (r *igmpSnoopingStaticResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_igmp_snooping_static"
}

// Schema defines the resource schema.
func (r *igmpSnoopingStaticResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages IGMP snooping static settings for a specific port.",
		Attributes: map[string]schema.Attribute{
			"port": schema.Int64Attribute{
				Description: "The port for which IGMP snooping static configuration is managed.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Specifies whether IGMP snooping is enabled (true) or disabled (false) for the given port.",
				Required:    true,
			},
		},
	}
}

// Configure sets up the resource client.
func (r *igmpSnoopingStaticResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*sdk.HRUIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Configure Type",
			fmt.Sprintf("Expected *sdk.HRUIClient. Got: %T. Please contact the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create enables IGMP snooping for a specific port.
func (r *igmpSnoopingStaticResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan igmpSnoopingStaticModel

	// Retrieve the desired state from the plan
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Enable the specified port while preserving other ports
	if err := r.client.ConfigurePortIGMPSnooping(int(plan.Port.ValueInt64()), plan.Enabled.ValueBool()); err != nil {
		resp.Diagnostics.AddError(
			"Error Configuring IGMP Snooping for Port",
			fmt.Sprintf("Failed to configure IGMP snooping for port %d: %s", plan.Port.ValueInt64(), err.Error()),
		)
		return
	}

	// Save the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read retrieves the current IGMP snooping static configuration for a port.
func (r *igmpSnoopingStaticResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state igmpSnoopingStaticModel

	// Get current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Query the current IGMP snooping status for the specified port
	enabled, err := r.client.GetPortIGMPSnooping(int(state.Port.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading IGMP Snooping Static Port",
			fmt.Sprintf("Failed to fetch IGMP snooping status for port %d: %s", state.Port.ValueInt64(), err.Error()),
		)
		return
	}

	// Update the Terraform state
	state.Enabled = types.BoolValue(enabled)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies the IGMP snooping static configuration for a specific port.
func (r *igmpSnoopingStaticResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan igmpSnoopingStaticModel

	// Get the updated state from the plan
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the specified port while preserving other ports
	if err := r.client.ConfigurePortIGMPSnooping(int(plan.Port.ValueInt64()), plan.Enabled.ValueBool()); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating IGMP Snooping for Port",
			fmt.Sprintf("Failed to update IGMP snooping for port %d: %s", plan.Port.ValueInt64(), err.Error()),
		)
		return
	}

	// Save the updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete disables IGMP snooping for a specific port.
func (r *igmpSnoopingStaticResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state igmpSnoopingStaticModel

	// Retrieve the current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Disable the specified port while preserving other ports
	if err := r.client.ConfigurePortIGMPSnooping(int(state.Port.ValueInt64()), false); err != nil {
		resp.Diagnostics.AddError(
			"Error Removing IGMP Snooping for Port",
			fmt.Sprintf("Failed to disable IGMP snooping for port %d: %s", state.Port.ValueInt64(), err.Error()),
		)
		return
	}

	// Remove the resource from the state
	resp.State.RemoveResource(ctx)
}
