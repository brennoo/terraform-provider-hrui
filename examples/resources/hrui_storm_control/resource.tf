resource "hrui_storm_control" "example" {
  port       = 4
  storm_type = "Broadcast"
  state      = true
  rate       = 2490000
}
