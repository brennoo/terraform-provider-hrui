---
page_title: "hrui_vlan_8021q (Resource)"
description: |-
  
---

# hrui_vlan_8021q (Resource)



## Example Usage

```terraform
resource "hrui_vlan_8021q" "example" {

  vlan_id = 10
  name    = "vlan10"

  untagged_ports = ["Port 1", "Port 2"]
  tagged_ports   = ["Trunk1"]

}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The VLAN name assigned to the VLAN ID.
- `tagged_ports` (List of String) The list of tagged ports assigned to the VLAN.
- `untagged_ports` (List of String) The list of untagged ports assigned to the VLAN (e.g., 'Port 1', 'Trunk1').
- `vlan_id` (Number) VLAN ID (1-4094). The unique identifier for the VLAN.

### Read-Only

- `member_ports` (List of String) The list of all ports assigned to the VLAN.


