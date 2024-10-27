# Assign VLAN ID 10 to port 4 with "All" frame type
resource "hrui_vlan_vid" "port_4_vlan_10" {
  port              = 4
  vlan_id           = 10
  accept_frame_type = "All"
}
