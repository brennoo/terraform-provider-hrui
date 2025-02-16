# Example usage of the hrui_mac_static data source to query all static MAC entries
data "hrui_mac_static" "all" {}

# Example usage of the hrui_mac_static data source with filters
data "hrui_mac_static" "filtered" {
  mac_address = "AA:BB:CC:DD:EE:FF"
  vlan_id     = 100
}
