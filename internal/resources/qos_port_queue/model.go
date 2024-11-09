package qos_port_queue

import "github.com/hashicorp/terraform-plugin-framework/types"

// qosPortQueueModel defines the schema model for qos_port_queue.
type qosPortQueueModel struct {
	PortID types.Int64 `tfsdk:"port_id"`
	Queue  types.Int64 `tfsdk:"queue"`
}
