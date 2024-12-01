# Example usage of the hrui_mac_static data source to query all static MAC entries

data "hrui_mac_static" "all" {}

output "all_static_mac_entries" {
  value = data.hrui_mac_static.all.entries
}

# Example usage of the hrui_mac_static data source with filters

data "hrui_mac_static" "filtered" {
  mac_address = "AA:BB:CC:DD:EE:FF"
  vlan_id     = 100
}

output "filtered_static_mac_entries" {
  value = data.hrui_mac_static.filtered.entries
}
