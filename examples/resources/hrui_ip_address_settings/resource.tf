# Enable DHCP on the switch
resource "hrui_ip_address_settings" "dhcp" {
  dhcp_enabled = true
}

# Disable DHCP and set static IP address settings
resource "hrui_ip_address_settings" "static" {
  dhcp_enabled = false
  ip_address   = "192.168.1.100"
  netmask      = "255.255.255.0"
  gateway      = "192.168.1.1"
}
