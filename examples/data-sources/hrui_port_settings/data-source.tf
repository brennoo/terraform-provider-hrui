data "hrui_port_settings" "port1" {
  port_id = 1
}

output "port1_settings" {
  value = data.hrui_port_settings.port1
}
