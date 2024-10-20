# Configure port 1
resource "hrui_port_settings" "port1" {
  port_id       = 1
  enabled       = true
  speed_duplex  = "Auto"
  flow_control = "Off"
}
