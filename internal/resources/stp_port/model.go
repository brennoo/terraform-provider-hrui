package stp_port

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// stpPortModel represents the configuration for a spanning tree protocol (STP) port.
type stpPortModel struct {
	Port     types.Int64  `tfsdk:"port"`
	PathCost types.Int64  `tfsdk:"path_cost"`
	Priority types.Int64  `tfsdk:"priority"`
	P2P      types.String `tfsdk:"p2p"`
	Edge     types.String `tfsdk:"edge"`
	State    types.String `tfsdk:"state"`
	Role     types.String `tfsdk:"role"`
}
