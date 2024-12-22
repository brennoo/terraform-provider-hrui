# Enable static IGMP snooping for port 2
resource "hrui_igmp_snooping_static" "port_2" {
  port    = 2
  enabled = true
}
