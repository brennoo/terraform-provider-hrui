# Data source to fetch the weight of a specific QoS queue by its ID
data "hrui_qos_queue_weight" "queue_1" {
  queue_id = 1
}

# Output the retrieved queue weight
output "queue_1_weight" {
  value = data.hrui_qos_queue_weight.queue_1.weight
}
