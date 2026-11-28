resource "hrui_dhcp_snooping" "example" {
  enabled       = true
  trusted_ports = ["Port 1"]
}
