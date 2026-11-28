package poe_port

import "github.com/hashicorp/terraform-plugin-framework/types"

type poePortModel struct {
	Port            types.String  `tfsdk:"port"`
	Enabled         types.Bool    `tfsdk:"enabled"`
	PSEType         types.String  `tfsdk:"pse_type"`
	WatchdogEnabled types.Bool    `tfsdk:"watchdog_enabled"`
	Operation       types.String  `tfsdk:"operation"`
	PDClass         types.String  `tfsdk:"pd_class"`
	PowerWatts      types.Float64 `tfsdk:"power_watts"`
	VoltageVolts    types.Float64 `tfsdk:"voltage_volts"`
	CurrentMA       types.Int64   `tfsdk:"current_ma"`
}
