resource "hrui_vlan_8021q" "example" {

  vlan_id = 10
  name    = "vlan10"

  untagged_ports = ["Port 1", "Port 2"]
  tagged_ports   = ["Trunk1"]

}
