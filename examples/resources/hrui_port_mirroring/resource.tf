resource "hrui_port_mirroring" "example" {
  mirror_direction = "BOTH"
  mirroring_port   = "Port 1"
  mirrored_port    = "Port 2"
}
