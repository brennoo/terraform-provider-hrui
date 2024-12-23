# Get VLAN VID configuration for port 1
data "hrui_vlan_vid" "port_1_config" {
  port = 1
}

# Output the VLAN ID and accepted frame type
output "vlan_id" {
  value = data.hrui_vlan_vid.port_1_config.vlan_id
}

output "accept_frame_type" {
  value = data.hrui_vlan_vid.port_1_config.accept_frame_type
}
