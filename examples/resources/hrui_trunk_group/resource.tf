resource "hrui_trunk_group" "example" {
  id    = 2
  type  = "LACP"
  ports = [4, 5, 6]
}
