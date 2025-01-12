resource "hrui_vlan_vid" "example" {
  port              = "Trunk1"
  vlan_id           = 10
  accept_frame_type = "All"
}
