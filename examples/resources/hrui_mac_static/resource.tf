resource "hrui_mac_static" "example" {
  mac_address = "AA:BB:CC:DD:EE:FF"
  vlan_id     = 100
  port        = "Port 1"
}
