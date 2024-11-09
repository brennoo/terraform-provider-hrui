package qos_queue_weight

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// qosQueueWeightModel maps the resource schema data.
type qosQueueWeightModel struct {
	QueueID types.Int64  `tfsdk:"queue_id"`
	Weight  types.String `tfsdk:"weight"`
}
