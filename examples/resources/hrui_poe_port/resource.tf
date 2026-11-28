resource "hrui_poe_port" "example" {
  port             = "Port 1"
  enabled          = true
  pse_type         = "AT"
  watchdog_enabled = false
}
