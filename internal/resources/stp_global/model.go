package stp_global

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// stpGlobalModel defines the resource schema data, excluding the duplicates.
type stpGlobalModel struct {
	STPStatus    types.String `tfsdk:"stp_status"` // Defined at loop_protocol
	ForceVersion types.String `tfsdk:"force_version"`
	Priority     types.Int64  `tfsdk:"priority"`
	MaxAge       types.Int64  `tfsdk:"max_age"`
	HelloTime    types.Int64  `tfsdk:"hello_time"`
	ForwardDelay types.Int64  `tfsdk:"forward_delay"`

	// Computed attributes
	RootMAC      types.String `tfsdk:"root_mac"`
	RootPathCost types.Int64  `tfsdk:"root_path_cost"`
	RootPort     types.String `tfsdk:"root_port"`
}
