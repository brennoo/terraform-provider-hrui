package storm_control

import "github.com/hashicorp/terraform-plugin-framework/types"

// stormControlModel represents the resource's state.
type stormControlModel struct {
	Port      types.Int64  `tfsdk:"port"`
	StormType types.String `tfsdk:"storm_type"`
	State     types.Bool   `tfsdk:"state"`
	Rate      types.Int64  `tfsdk:"rate"`
}
