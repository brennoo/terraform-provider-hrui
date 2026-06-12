package bandwidth_control

import (
	"context"
	"fmt"
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
	_ resource.Resource                = &bandwidthControlResource{}
	_ resource.ResourceWithImportState = &bandwidthControlResource{}
)

// bandwidthControlResource is the implementation of the resource.
type bandwidthControlResource struct {
	client *sdk.HRUIClient
}

// NewResource creates a new instance of the bandwidth control resource.
func NewResource() resource.Resource {
	return &bandwidthControlResource{}
}

// Metadata sets the resource type name.
func (r *bandwidthControlResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bandwidth_control"
}

// Schema defines the schema for the bandwidth control resource.
func (r *bandwidthControlResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Configures bandwidth control on a specific port.",
		Attributes: map[string]schema.Attribute{
			"port": schema.StringAttribute{
				Description: "Port where bandwidth control is configured (e.g., 'Port 1', 'Trunk2').",
				Required:    true,
			},
			"ingress_rate": schema.StringAttribute{
				Description: "Ingress bandwidth rate in kbps. Use '0' or 'Unlimited' to disable limitation.",
				Required:    true,
			},
			"egress_rate": schema.StringAttribute{
				Description: "Egress bandwidth rate in kbps. Use '0' or 'Unlimited' to disable limitation.",
				Required:    true,
			},
		},
	}
}

// Configure assigns the SDK client from provider configuration.
func (r *bandwidthControlResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Create sets bandwidth control on a port.
func (r *bandwidthControlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data bandwidthControlModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating bandwidth control", map[string]any{"port": data.Port.ValueString()})

	// Configure ingress rate
	if err := r.client.ConfigureBandwidthControl(ctx,
		data.Port.ValueString(),
		true, // isIngress
		true, // enable
		normalizeRate(data.IngressRate.ValueString()),
	); err != nil {
		resp.Diagnostics.AddError("Error Creating Bandwidth Control", err.Error())
		return
	}

	// Configure egress rate
	if err := r.client.ConfigureBandwidthControl(ctx,
		data.Port.ValueString(),
		false, // isIngress
		true,  // enable
		normalizeRate(data.EgressRate.ValueString()),
	); err != nil {
		resp.Diagnostics.AddError("Error Creating Bandwidth Control", err.Error())
		return
	}

	// Normalize state values: if user set "0", normalize to "Unlimited" in state
	if data.IngressRate.ValueString() == "0" {
		data.IngressRate = types.StringValue("Unlimited")
	}
	if data.EgressRate.ValueString() == "0" {
		data.EgressRate = types.StringValue("Unlimited")
	}

	// Save state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

	tflog.Debug(ctx, "Bandwidth control created", map[string]any{"port": data.Port.ValueString()})
}

// Read retrieves the current state for the resource.
func (r *bandwidthControlResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bandwidthControlModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading bandwidth control", map[string]any{"port": state.Port.ValueString()})

	// Fetch all bandwidth control configurations
	controls, err := r.client.GetBandwidthControl(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Bandwidth Control", err.Error())
		return
	}

	// Use a copy of the struct to avoid implicit memory aliasing
	var foundConfig sdk.BandwidthControl
	for _, control := range controls {
		if control.Port == state.Port.ValueString() {
			foundConfig = control
			break
		}
	}

	// If no matching port was found, return an error
	if foundConfig.Port == "" {
		resp.Diagnostics.AddError(
			"Error Reading Bandwidth Control",
			fmt.Sprintf("Could not find port '%s' in bandwidth control table", state.Port.ValueString()),
		)
		return
	}

	// Update state with the live values
	// Normalize "0" to "Unlimited" since the device stores unlimited rates as "0"
	ingressRate := foundConfig.IngressRate
	if ingressRate == "0" {
		ingressRate = "Unlimited"
	}
	egressRate := foundConfig.EgressRate
	if egressRate == "0" {
		egressRate = "Unlimited"
	}
	state.IngressRate = types.StringValue(ingressRate)
	state.EgressRate = types.StringValue(egressRate)

	// Save updated state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	tflog.Debug(ctx, "Bandwidth control read", map[string]any{"port": state.Port.ValueString()})
}

// Update modifies bandwidth control settings.
func (r *bandwidthControlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bandwidthControlModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating bandwidth control", map[string]any{"port": plan.Port.ValueString()})

	// Update ingress rate
	if err := r.client.ConfigureBandwidthControl(ctx,
		plan.Port.ValueString(),
		true, // isIngress
		true, // enable
		normalizeRate(plan.IngressRate.ValueString()),
	); err != nil {
		resp.Diagnostics.AddError("Error Updating Bandwidth Control", err.Error())
		return
	}

	// Update egress rate
	if err := r.client.ConfigureBandwidthControl(ctx,
		plan.Port.ValueString(),
		false, // isIngress
		true,  // enable
		normalizeRate(plan.EgressRate.ValueString()),
	); err != nil {
		resp.Diagnostics.AddError("Error Updating Bandwidth Control", err.Error())
		return
	}

	// Normalize state values: if user set "0", normalize to "Unlimited" in state
	if plan.IngressRate.ValueString() == "0" {
		plan.IngressRate = types.StringValue("Unlimited")
	}
	if plan.EgressRate.ValueString() == "0" {
		plan.EgressRate = types.StringValue("Unlimited")
	}

	// Save state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	tflog.Debug(ctx, "Bandwidth control updated", map[string]any{"port": plan.Port.ValueString()})
}

// Delete disables bandwidth control on a port.
func (r *bandwidthControlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state bandwidthControlModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting bandwidth control", map[string]any{"port": state.Port.ValueString()})

	// Disable ingress rate
	if err := r.client.ConfigureBandwidthControl(ctx,
		state.Port.ValueString(),
		true,  // isIngress
		false, // disable
		"0",   // reset rate to 0
	); err != nil {
		resp.Diagnostics.AddError("Error Deleting Bandwidth Control", err.Error())
		return
	}

	// Disable egress rate
	if err := r.client.ConfigureBandwidthControl(ctx,
		state.Port.ValueString(),
		false, // isIngress
		false, // disable
		"0",   // reset rate to 0
	); err != nil {
		resp.Diagnostics.AddError("Error Deleting Bandwidth Control", err.Error())
		return
	}
}

// normalizeRate ensures the rate is formatted correctly.
// Normalizes "0" to "Unlimited" since both mean disable limitation.
func normalizeRate(rate string) string {
	rateLower := strings.ToLower(rate)
	if rateLower == "unlimited" || rate == "0" {
		return "Unlimited"
	}
	return rate
}

// ImportState imports an existing Bandwidth Control resource by port name.
func (r *bandwidthControlResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Importing bandwidth control", map[string]any{"id": req.ID})

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("port"), req.ID)...)
}
