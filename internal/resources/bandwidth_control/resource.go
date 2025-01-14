package bandwidth_control

import (
	"context"
	"fmt"
	"strings"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure implementation satisfies the resource.Resource interface.
var _ resource.Resource = &bandwidthControlResource{}

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
		Description: "Resource for configuring bandwidth control on a specific port in the device.",
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
	if req.ProviderData != nil {
		client, ok := req.ProviderData.(*sdk.HRUIClient)
		if !ok {
			resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *sdk.HRUIClient")
			return
		}
		r.client = client
	}
}

// Create sets bandwidth control on a port.
func (r *bandwidthControlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data bandwidthControlModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configure ingress rate
	if err := r.client.ConfigureBandwidthControl(
		data.Port.ValueString(),
		true, // isIngress
		true, // enable
		normalizeRate(data.IngressRate.ValueString()),
	); err != nil {
		resp.Diagnostics.AddError("Failed to configure Ingress Bandwidth Control", err.Error())
		return
	}

	// Configure egress rate
	if err := r.client.ConfigureBandwidthControl(
		data.Port.ValueString(),
		false, // isIngress
		true,  // enable
		normalizeRate(data.EgressRate.ValueString()),
	); err != nil {
		resp.Diagnostics.AddError("Failed to configure Egress Bandwidth Control", err.Error())
		return
	}

	// Save state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// Read retrieves the current state for the resource.
func (r *bandwidthControlResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bandwidthControlModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch all bandwidth control configurations
	controls, err := r.client.GetBandwidthControl()
	if err != nil {
		resp.Diagnostics.AddError("Error fetching bandwidth control configurations", err.Error())
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
			"Port not found",
			fmt.Sprintf("Could not find port '%s' in bandwidth control table", state.Port.ValueString()),
		)
		return
	}

	// Update state with the live values
	state.IngressRate = types.StringValue(foundConfig.IngressRate)
	state.EgressRate = types.StringValue(foundConfig.EgressRate)

	// Save updated state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update modifies bandwidth control settings.
func (r *bandwidthControlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bandwidthControlModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update ingress rate
	if err := r.client.ConfigureBandwidthControl(
		plan.Port.ValueString(),
		true, // isIngress
		true, // enable
		normalizeRate(plan.IngressRate.ValueString()),
	); err != nil {
		resp.Diagnostics.AddError("Failed to update Ingress Bandwidth Control", err.Error())
		return
	}

	// Update egress rate
	if err := r.client.ConfigureBandwidthControl(
		plan.Port.ValueString(),
		false, // isIngress
		true,  // enable
		normalizeRate(plan.EgressRate.ValueString()),
	); err != nil {
		resp.Diagnostics.AddError("Failed to update Egress Bandwidth Control", err.Error())
		return
	}

	// Save state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete disables bandwidth control on a port.
func (r *bandwidthControlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state bandwidthControlModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Disable ingress rate
	if err := r.client.ConfigureBandwidthControl(
		state.Port.ValueString(),
		true,  // isIngress
		false, // disable
		"0",   // reset rate to 0
	); err != nil {
		resp.Diagnostics.AddError("Failed to disable Ingress Bandwidth Control", err.Error())
		return
	}

	// Disable egress rate
	if err := r.client.ConfigureBandwidthControl(
		state.Port.ValueString(),
		false, // isIngress
		false, // disable
		"0",   // reset rate to 0
	); err != nil {
		resp.Diagnostics.AddError("Failed to disable Egress Bandwidth Control", err.Error())
		return
	}
}

// normalizeRate ensures the rate is formatted correctly.
func normalizeRate(rate string) string {
	if strings.ToLower(rate) == "unlimited" {
		return "unlimited"
	}
	return rate
}
