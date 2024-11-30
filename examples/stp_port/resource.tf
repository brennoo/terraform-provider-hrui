# Example usage of the stp_port resource
resource "stp_port" "example" {
  # The port ID for the physical switch port
  port = 10

  # Desired STP path cost for the port
  path_cost = 200

  # The STP port priority
  priority = 128

  # STP Edge port configuration
  edge = "False"

  # Other attributes like `state`, and `role` are computed
}
