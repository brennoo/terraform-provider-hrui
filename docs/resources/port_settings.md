---
page_title: "hrui_port_settings (Resource)"
description: |-
  
---

# hrui_port_settings (Resource)



## Example Usage

```terraform
# Configure port 1
resource "hrui_port_settings" "port1" {
  port_id      = 1
  enabled      = true
  speed_duplex = "Auto"
  flow_control = "Off"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `port_id` (Number) The ID of the port to configure.

### Optional

- `enabled` (Boolean) Whether the port is enabled.
- `flow_control` (String) The flow control setting of the port.
- `speed_duplex` (String) The speed and duplex mode of the port.

### Read-Only

- `id` (String) The ID of the port setting data source.


