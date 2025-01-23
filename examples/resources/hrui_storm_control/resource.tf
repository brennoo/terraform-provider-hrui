resource "hrui_storm_control" "example" {
  port       = "Port 1"
  storm_type = "Broadcast"
  state      = true
  rate       = 100000
}
