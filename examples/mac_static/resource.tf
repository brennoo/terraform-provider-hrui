resource "hrui_mac_static" "example" {
  mac_address = "AA:BB:CC:DD:EE:FF"
  vlan_id     = 100
  port        = 2
}

output "mac_static_resource" {
  value = {
    id          = hrui_mac_static.example.id
    mac_address = hrui_mac_static.example.mac_address
    vlan_id     = hrui_mac_static.example.vlan_id
    port        = hrui_mac_static.example.port
  }
}
