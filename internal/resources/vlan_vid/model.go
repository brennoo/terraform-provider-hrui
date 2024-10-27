package vlan_vid

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// vlanVIDModel maps the data source schema data.
type vlanVIDModel struct {
	Port            types.Int64  `tfsdk:"port"`
	VlanID          types.Int64  `tfsdk:"vlan_id"`
	AcceptFrameType types.String `tfsdk:"accept_frame_type"`
}
