---
page_title: "hrui_mac_table (Data Source)"
description: |-
  Data source to fetch the MAC address table from the switch.
---

# hrui_mac_table (Data Source)

Data source to fetch the MAC address table from the switch.

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `mac_table` (Attributes List) List of MAC table entries retrieved from the switch. (see [below for nested schema](#nestedatt--mac_table))

<a id="nestedatt--mac_table"></a>
### Nested Schema for `mac_table`

Read-Only:

- `id` (Number) The sequence number of the entry.
- `mac_address` (String) The MAC address in the format xx:xx:xx:xx:xx:xx.
- `port` (Number) The port number where the MAC address is located.
- `type` (String) The type of the MAC address entry (e.g., dynamic or static).
- `vlan_id` (Number) The VLAN ID associated with the MAC address.


