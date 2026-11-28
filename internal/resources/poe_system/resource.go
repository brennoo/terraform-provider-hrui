package poe_system

import (
	"context"
	"fmt"

	"github.com/brennoo/terraform-provider-hrui/internal/providerutil"
	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &poeSystemResource{}
	_ resource.ResourceWithConfigure   = &poeSystemResource{}
	_ resource.ResourceWithImportState = &poeSystemResource{}
)

type poeSystemResource struct {
	client *sdk.HRUIClient
}

// NewResource creates a new PoE system resource.
func NewResource() resource.Resource {
	return &poeSystemResource{}
}

func (r *poeSystemResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_poe_system"
}

func (r *poeSystemResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages global Power over Ethernet (PoE) system power budget settings and reports total PoE consumption telemetry.",
		Attributes: map[string]schema.Attribute{
			"power_limit_watts": schema.Int64Attribute{
				Required:    true,
				Description: "Total PoE power budget in Watts configured on the switch.",
			},
			"power_consumption_watts": schema.Float64Attribute{
				Computed:    true,
				Description: "Current total PoE power consumption in Watts reported by the PSE controller.",
			},
			"power_usage_percent": schema.Int64Attribute{
				Computed:    true,
				Description: "Current total PoE power usage percentage reported by the switch.",
			},
		},
	}
}

func (r *poeSystemResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

func (r *poeSystemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan poeSystemModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating PoE system settings", map[string]any{"power_limit_watts": plan.PowerLimitWatts.ValueInt64()})

	if err := r.client.SetPoESystemPowerLimit(ctx, int(plan.PowerLimitWatts.ValueInt64())); err != nil {
		resp.Diagnostics.AddError("Error Creating PoE System Settings", fmt.Sprintf("Failed to set PoE system power limit: %s", err))
		return
	}

	sys, err := r.client.GetPoESystem(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading PoE System Settings", fmt.Sprintf("Failed to read PoE system settings after creation: %s", err))
		return
	}

	plan.PowerLimitWatts = types.Int64Value(int64(sys.PowerLimitWatts))
	plan.PowerConsumptionWatts = types.Float64Value(sys.PowerConsumptionWatts)
	plan.PowerUsagePercent = types.Int64Value(int64(sys.PowerUsagePercent))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *poeSystemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state poeSystemModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sys, err := r.client.GetPoESystem(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading PoE System Settings", fmt.Sprintf("Failed to read PoE system settings: %s", err))
		return
	}

	state.PowerLimitWatts = types.Int64Value(int64(sys.PowerLimitWatts))
	state.PowerConsumptionWatts = types.Float64Value(sys.PowerConsumptionWatts)
	state.PowerUsagePercent = types.Int64Value(int64(sys.PowerUsagePercent))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *poeSystemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan poeSystemModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetPoESystemPowerLimit(ctx, int(plan.PowerLimitWatts.ValueInt64())); err != nil {
		resp.Diagnostics.AddError("Error Updating PoE System Settings", fmt.Sprintf("Failed to update PoE system power limit: %s", err))
		return
	}

	sys, err := r.client.GetPoESystem(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading PoE System Settings", fmt.Sprintf("Failed to read PoE system settings after update: %s", err))
		return
	}

	plan.PowerLimitWatts = types.Int64Value(int64(sys.PowerLimitWatts))
	plan.PowerConsumptionWatts = types.Float64Value(sys.PowerConsumptionWatts)
	plan.PowerUsagePercent = types.Int64Value(int64(sys.PowerUsagePercent))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *poeSystemResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	if err := r.client.SetPoESystemPowerLimit(ctx, 100); err != nil {
		resp.Diagnostics.AddError("Error Deleting PoE System Settings", fmt.Sprintf("Failed to reset PoE system power limit to default: %s", err))
	}
}

func (r *poeSystemResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	sys, err := r.client.GetPoESystem(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing PoE System Settings", fmt.Sprintf("Unable to import PoE system settings: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &poeSystemModel{
		PowerLimitWatts:       types.Int64Value(int64(sys.PowerLimitWatts)),
		PowerConsumptionWatts: types.Float64Value(sys.PowerConsumptionWatts),
		PowerUsagePercent:     types.Int64Value(int64(sys.PowerUsagePercent)),
	})...)
}
