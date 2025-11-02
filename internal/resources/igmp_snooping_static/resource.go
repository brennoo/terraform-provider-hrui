package igmp_snooping_static

import (
	"context"
	"fmt"
	"time"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
			"port": schema.StringAttribute{
				Description: "The port name for which IGMP snooping static configuration is managed.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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

	// Ensure global IGMP snooping is enabled before configuring port-level settings
	igmpConfig, err := r.client.FetchIGMPConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading IGMP Configuration",
			fmt.Sprintf("Failed to fetch IGMP configuration: %s", err),
		)
		return
	}

	if !igmpConfig.Enabled {
		// Enable global IGMP snooping first
		if err := r.client.EnableIGMPSnooping(ctx); err != nil {
			resp.Diagnostics.AddError(
				"Error Enabling Global IGMP Snooping",
				fmt.Sprintf("Global IGMP snooping must be enabled before configuring port-level settings. Failed to enable: %s", err),
			)
			return
		}
		// Wait a moment for the global setting to take effect
		time.Sleep(1 * time.Second)
	}

	// Enable the specified port while preserving other ports
	if err := r.client.UpdatePortIGMPSnoopingByName(ctx, plan.Port.ValueString(), plan.Enabled.ValueBool()); err != nil {
		resp.Diagnostics.AddError(
			"Error Configuring IGMP Snooping",
			fmt.Sprintf("Failed to configure IGMP snooping for port %s: %s", plan.Port.ValueString(), err),
		)
		return
	}

	// Wait for the device to reflect the changes
	time.Sleep(2 * time.Second)

	// Read back the actual state from the device to ensure consistency
	enabled, err := r.client.GetPortIGMPSnoopingByName(ctx, plan.Port.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading IGMP Snooping State",
			fmt.Sprintf("Failed to read IGMP snooping state for port %s after creation: %s", plan.Port.ValueString(), err),
		)
		return
	}

	// Update state with the actual device value
	state := igmpSnoopingStaticModel{
		Port:    plan.Port,
		Enabled: types.BoolValue(enabled),
	}

	// Save the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
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
	enabled, err := r.client.GetPortIGMPSnoopingByName(ctx, state.Port.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading IGMP Snooping State",
			fmt.Sprintf("Failed to read IGMP snooping state for port %s: %s", state.Port.ValueString(), err),
		)
		return
	}

	// Update the Terraform state
	state.Enabled = types.BoolValue(enabled)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
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

	// Ensure global IGMP snooping is enabled before configuring port-level settings
	igmpConfig, err := r.client.FetchIGMPConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading IGMP Configuration",
			fmt.Sprintf("Failed to fetch IGMP configuration: %s", err),
		)
		return
	}

	if !igmpConfig.Enabled {
		// Enable global IGMP snooping first
		if err := r.client.EnableIGMPSnooping(ctx); err != nil {
			resp.Diagnostics.AddError(
				"Error Enabling Global IGMP Snooping",
				fmt.Sprintf("Global IGMP snooping must be enabled before configuring port-level settings. Failed to enable: %s", err),
			)
			return
		}
		// Wait a moment for the global setting to take effect
		time.Sleep(1 * time.Second)
	}

	// Update the specified port while preserving other ports
	if err := r.client.UpdatePortIGMPSnoopingByName(ctx, plan.Port.ValueString(), plan.Enabled.ValueBool()); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating IGMP Snooping",
			fmt.Sprintf("Failed to update IGMP snooping for port %s: %s", plan.Port.ValueString(), err),
		)
		return
	}

	// Wait for the device to reflect the changes
	time.Sleep(2 * time.Second)

	// Read back the actual state from the device to ensure consistency
	enabled, err := r.client.GetPortIGMPSnoopingByName(ctx, plan.Port.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading IGMP Snooping State",
			fmt.Sprintf("Failed to read IGMP snooping state for port %s after update: %s", plan.Port.ValueString(), err),
		)
		return
	}

	// Update state with the actual device value
	state := igmpSnoopingStaticModel{
		Port:    plan.Port,
		Enabled: types.BoolValue(enabled),
	}

	// Save the updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
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
	if err := r.client.UpdatePortIGMPSnoopingByName(ctx, state.Port.ValueString(), false); err != nil {
		resp.Diagnostics.AddError(
			"Error Disabling IGMP Snooping",
			fmt.Sprintf("Failed to disable IGMP snooping for port %s: %s", state.Port.ValueString(), err),
		)
		return
	}

	// Remove the resource from the state
	resp.State.RemoveResource(ctx)
}
