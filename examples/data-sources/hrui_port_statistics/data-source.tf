data "hrui_port_statistics" "ports" {}

output "enabled_ports" {
  description = "Port statistics for enabled ports"
  value       = [for port in data.hrui_port_statistics.ports.port_statistics : port if port.state == "Enable"]
}

output "ports_with_link_down" {
  description = "Port statistics for ports where the link is down"
  value       = [for port in data.hrui_port_statistics.ports.port_statistics : port if port.link_status == "Down"]
}
