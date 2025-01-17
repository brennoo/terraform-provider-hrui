package qos_port_queue

import "github.com/hashicorp/terraform-plugin-framework/types"

// qosPortQueueModel defines the schema model for qos_port_queue.
type qosPortQueueModel struct {
	Port  types.String `tfsdk:"port"`
	Queue types.Int64  `tfsdk:"queue"`
}
