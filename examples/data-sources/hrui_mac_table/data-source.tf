data "hrui_mac_table" "example" {}

# Define a local variable to filter the MAC address
locals {
  target_mac_address = "AA:BB:CC:DD:EE:FF"
  filtered_mac_table = tomap({
    for entry in data.hrui_mac_table.example.mac_table :
    entry.mac_address => entry if entry.mac_address == local.target_mac_address
  })
}

output "filtered_mac_entry" {
  value = local.filtered_mac_table
}
