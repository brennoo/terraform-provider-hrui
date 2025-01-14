package bandwidth_control

import "github.com/hashicorp/terraform-plugin-framework/types"

// bandwidthControlModel represents the resource schema state.
type bandwidthControlModel struct {
	Port        types.String `tfsdk:"port"`
	IngressRate types.String `tfsdk:"ingress_rate"`
	EgressRate  types.String `tfsdk:"egress_rate"`
}
