---
page_title: "hrui_jumbo_frame (Resource)"
description: |-
  Manages Jumbo Frame settings for the HRUI system.
---

# hrui_jumbo_frame (Resource)

Manages Jumbo Frame settings for the HRUI system.

## Example Usage

```terraform
resource "hrui_jumbo_frame" "example" {
  size = 16383
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `size` (Number) Size of the Jumbo Frame in bytes. Valid options are 1522, 1536, 1552, 9216, and 16383.


