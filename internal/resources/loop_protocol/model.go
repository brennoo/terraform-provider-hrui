package loop_protocol

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// loopProtocolModel maps the resource schema data.
type loopProtocolModel struct {
	LoopFunction types.String `tfsdk:"loop_function"`
	IntervalTime types.Int64  `tfsdk:"interval_time"`
	RecoverTime  types.Int64  `tfsdk:"recover_time"`
}
