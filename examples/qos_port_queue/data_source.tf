data "hrui_qos_port_queue" "port_queue_1" {
  port_id = 1
}

output "queue_value_1" {
  value = data.hrui_qos_port_queue.port_queue_1.queue
}

data "hrui_qos_port_queue" "port_queue_3" {
  port_id = 3
}

output "queue_value_3" {
  value = data.hrui_qos_port_queue.port_queue_3.queue
}
