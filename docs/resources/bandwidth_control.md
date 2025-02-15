---
page_title: "hrui_bandwidth_control (Resource)"
description: |-
  Resource for configuring bandwidth control on a specific port in the device.
---

# hrui_bandwidth_control (Resource)

Resource for configuring bandwidth control on a specific port in the device.

## Example Usage

```terraform
resource "hrui_bandwidth_control" "example" {
  port         = "Port 1"
  ingress_rate = "992"
  egress_rate  = "2000"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `egress_rate` (String) Egress bandwidth rate in kbps. Use '0' or 'Unlimited' to disable limitation.
- `ingress_rate` (String) Ingress bandwidth rate in kbps. Use '0' or 'Unlimited' to disable limitation.
- `port` (String) Port where bandwidth control is configured (e.g., 'Port 1', 'Trunk2').


