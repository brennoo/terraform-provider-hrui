resource "hrui_vlan_8021q" "example" {

  vlan_id = 10
  name    = "vlan10"

  untagged_ports = [2, 3]
  tagged_ports   = [4, 5]

}
