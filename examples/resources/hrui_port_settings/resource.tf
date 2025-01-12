resource "hrui_port_settings" "example" {
  port    = "Port 1"
  enabled = true

  speed = {
    config = "Auto"
  }

  flow_control = {
    config = "On"
  }
}
