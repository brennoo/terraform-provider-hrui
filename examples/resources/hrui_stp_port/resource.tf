resource "stp_port" "example" {
  port      = "Port 1"
  path_cost = 200
  priority  = 128
  edge      = "False"
}
