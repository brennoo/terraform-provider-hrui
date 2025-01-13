resource "hrui_port_isolation" "example" {
  port           = "Port 1"
  isolation_list = ["Port 2", "Port 3"]
}
