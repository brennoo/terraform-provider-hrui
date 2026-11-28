package dhcp_snooping

import (
	"context"
	"fmt"

	"github.com/brennoo/terraform-provider-hrui/internal/providerutil"
	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &dhcpSnoopingResource{}
	_ resource.ResourceWithConfigure   = &dhcpSnoopingResource{}
	_ resource.ResourceWithImportState = &dhcpSnoopingResource{}
)

type dhcpSnoopingResource struct {
	client *sdk.HRUIClient
}

// NewResource creates a new DHCP snooping resource.
func NewResource() resource.Resource {
	return &dhcpSnoopingResource{}
}

func (r *dhcpSnoopingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcp_snooping"
}

func (r *dhcpSnoopingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages global DHCP Snooping state and per-port trusted DHCP server ports.",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether global DHCP Snooping is enabled on the switch.",
			},
			"trusted_ports": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "List of switch ports configured as trusted DHCP server ports (e.g., ['Port 1', 'Port 5']). Untrusted ports only allow DHCP client traffic.",
			},
		},
	}
}

func (r *dhcpSnoopingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

func extractTrustedPorts(ctx context.Context, listVal types.List) ([]string, diag.Diagnostics) {
	if listVal.IsNull() || listVal.IsUnknown() {
		return nil, nil
	}
	var ports []string
	diags := listVal.ElementsAs(ctx, &ports, false)
	return ports, diags
}

func toTrustedPortsList(ports []string) (types.List, diag.Diagnostics) {
	if ports == nil {
		ports = []string{}
	}
	elems := make([]attr.Value, len(ports))
	for i, p := range ports {
		elems[i] = types.StringValue(p)
	}
	return types.ListValue(types.StringType, elems)
}

func (r *dhcpSnoopingResource) applyAndRead(ctx context.Context, plan *dhcpSnoopingModel, diags *diag.Diagnostics) {
	var trusted []string
	if !plan.TrustedPorts.IsNull() && !plan.TrustedPorts.IsUnknown() {
		var d diag.Diagnostics
		trusted, d = extractTrustedPorts(ctx, plan.TrustedPorts)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
		if err := r.client.ConfigureDHCPSnooping(ctx, sdk.DHCPSnoopingConfig{
			Enabled:      plan.Enabled.ValueBool(),
			TrustedPorts: trusted,
		}); err != nil {
			diags.AddError("Error Configuring DHCP Snooping", fmt.Sprintf("Failed to configure DHCP snooping: %s", err))
			return
		}
	} else {
		if err := r.client.SetDHCPSnoopingGlobal(ctx, plan.Enabled.ValueBool()); err != nil {
			diags.AddError("Error Configuring DHCP Snooping", fmt.Sprintf("Failed to set DHCP snooping global state: %s", err))
			return
		}
	}

	cfg, err := r.client.GetDHCPSnooping(ctx)
	if err != nil {
		diags.AddError("Error Reading DHCP Snooping", fmt.Sprintf("Failed to read DHCP snooping state: %s", err))
		return
	}

	plan.Enabled = types.BoolValue(cfg.Enabled)
	listVal, d := toTrustedPortsList(cfg.TrustedPorts)
	diags.Append(d...)
	plan.TrustedPorts = listVal
}

func (r *dhcpSnoopingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dhcpSnoopingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating DHCP snooping configuration", map[string]any{"enabled": plan.Enabled.ValueBool()})

	r.applyAndRead(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dhcpSnoopingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dhcpSnoopingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.client.GetDHCPSnooping(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading DHCP Snooping", fmt.Sprintf("Failed to read DHCP snooping state: %s", err))
		return
	}

	state.Enabled = types.BoolValue(cfg.Enabled)
	listVal, d := toTrustedPortsList(cfg.TrustedPorts)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.TrustedPorts = listVal

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dhcpSnoopingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dhcpSnoopingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.applyAndRead(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dhcpSnoopingResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	if err := r.client.ConfigureDHCPSnooping(ctx, sdk.DHCPSnoopingConfig{
		Enabled:      false,
		TrustedPorts: nil,
	}); err != nil {
		resp.Diagnostics.AddError("Error Deleting DHCP Snooping", fmt.Sprintf("Failed to reset DHCP snooping to default: %s", err))
	}
}

func (r *dhcpSnoopingResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	cfg, err := r.client.GetDHCPSnooping(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing DHCP Snooping", fmt.Sprintf("Unable to import DHCP snooping settings: %s", err))
		return
	}

	listVal, d := toTrustedPortsList(cfg.TrustedPorts)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dhcpSnoopingModel{
		Enabled:      types.BoolValue(cfg.Enabled),
		TrustedPorts: listVal,
	})...)
}
