# Configure port 1
resource "hrui_port_settings" "example" {
  port         = "Port 1"
  enabled      = true
  speed_duplex = "Auto"
  flow_control = "Off"
}
