package poe_system

import "github.com/hashicorp/terraform-plugin-framework/types"

type poeSystemModel struct {
	PowerLimitWatts       types.Int64   `tfsdk:"power_limit_watts"`
	PowerConsumptionWatts types.Float64 `tfsdk:"power_consumption_watts"`
	PowerUsagePercent     types.Int64   `tfsdk:"power_usage_percent"`
}
