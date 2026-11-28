package poe_port

import (
	"cmp"
	"context"
	"fmt"

	"github.com/brennoo/terraform-provider-hrui/internal/providerutil"
	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &poePortResource{}
	_ resource.ResourceWithConfigure   = &poePortResource{}
	_ resource.ResourceWithImportState = &poePortResource{}
)

type poePortResource struct {
	client *sdk.HRUIClient
}

// NewResource creates a new PoE port resource.
func NewResource() resource.Resource {
	return &poePortResource{}
}

func (r *poePortResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_poe_port"
}

func (r *poePortResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Power over Ethernet (PoE) administrative configuration on a switch port and exposes live PSE telemetry.",
		Attributes: map[string]schema.Attribute{
			"port": schema.StringAttribute{
				Required:    true,
				Description: "The PoE port name (e.g., 'Port 1').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether PoE power delivery is administratively enabled on the port.",
			},
			"pse_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "PoE protocol type configured on the port ('BT', 'HiPoE', 'AT', or 'AF').",
				Validators: []validator.String{
					stringvalidator.OneOf("BT", "HiPoE", "AT", "AF"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"watchdog_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the PoE watchdog (PoEDog) is enabled on the port.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"operation": schema.StringAttribute{
				Computed:    true,
				Description: "Current PoE operational state ('On' or 'Off').",
			},
			"pd_class": schema.StringAttribute{
				Computed:    true,
				Description: "Detected Powered Device (PD) IEEE 802.3 class (e.g., 'AT(Class4)', 'AF(Class3)', or '-').",
			},
			"power_watts": schema.Float64Attribute{
				Computed:    true,
				Description: "Live PoE power draw on the port in Watts.",
			},
			"voltage_volts": schema.Float64Attribute{
				Computed:    true,
				Description: "Live PoE bus voltage on the port in Volts.",
			},
			"current_ma": schema.Int64Attribute{
				Computed:    true,
				Description: "Live PoE current draw on the port in milliamps (mA).",
			},
		},
	}
}

func (r *poePortResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

func (r *poePortResource) applyAndRead(ctx context.Context, plan *poePortModel) error {
	pseType := "AT"
	if !plan.PSEType.IsNull() && !plan.PSEType.IsUnknown() {
		pseType = cmp.Or(plan.PSEType.ValueString(), "AT")
	} else if cur, err := r.client.GetPoEPortConfig(ctx, plan.Port.ValueString()); err == nil && cur.PSEType != "" {
		pseType = cur.PSEType
	}

	watchdog := false
	if !plan.WatchdogEnabled.IsNull() && !plan.WatchdogEnabled.IsUnknown() {
		watchdog = plan.WatchdogEnabled.ValueBool()
	}

	cfg := sdk.PoEPortConfig{
		Port:          plan.Port.ValueString(),
		Enabled:       plan.Enabled.ValueBool(),
		PSEType:       pseType,
		WatchdogState: watchdog,
	}

	if err := r.client.ConfigurePoEPort(ctx, cfg); err != nil {
		return err
	}

	status, err := r.client.GetPoEPortStatus(ctx, plan.Port.ValueString())
	if err != nil {
		return err
	}

	populateModelFromStatus(plan, status)
	return nil
}

func populateModelFromStatus(m *poePortModel, s *sdk.PoEPortStatus) {
	m.Port = types.StringValue(s.Port)
	m.Enabled = types.BoolValue(s.Enabled)
	m.PSEType = types.StringValue(s.PSEType)
	m.WatchdogEnabled = types.BoolValue(s.WatchdogState)
	m.Operation = types.StringValue(s.Operation)
	m.PDClass = types.StringValue(s.Class)
	m.PowerWatts = types.Float64Value(s.PowerWatts)
	m.VoltageVolts = types.Float64Value(s.VoltageVolts)
	m.CurrentMA = types.Int64Value(int64(s.CurrentMA))
}

func (r *poePortResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan poePortModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating PoE port configuration", map[string]any{"port": plan.Port.ValueString()})

	if err := r.applyAndRead(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Error Creating PoE Port Settings", fmt.Sprintf("Failed to configure PoE port %s: %s", plan.Port.ValueString(), err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *poePortResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state poePortModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := r.client.GetPoEPortStatus(ctx, state.Port.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading PoE Port Settings", fmt.Sprintf("Failed to read PoE port %s: %s", state.Port.ValueString(), err))
		return
	}

	populateModelFromStatus(&state, status)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *poePortResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan poePortModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyAndRead(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Error Updating PoE Port Settings", fmt.Sprintf("Failed to update PoE port %s: %s", plan.Port.ValueString(), err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *poePortResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state poePortModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	defaultCfg := sdk.PoEPortConfig{
		Port:          state.Port.ValueString(),
		Enabled:       true,
		PSEType:       "AT",
		WatchdogState: false,
	}
	if err := r.client.ConfigurePoEPort(ctx, defaultCfg); err != nil {
		resp.Diagnostics.AddError("Error Deleting PoE Port Settings", fmt.Sprintf("Failed to reset PoE port %s to default: %s", state.Port.ValueString(), err))
		return
	}
}

func (r *poePortResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("port"), req.ID)...)
}
