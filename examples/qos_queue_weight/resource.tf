# Resource to set the weight of a specific QoS queue (e.g., Queue 2)
resource "hrui_qos_queue_weight" "queue_2" {
  queue_id = 2
  weight   = 10
}
