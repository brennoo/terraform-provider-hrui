package mac_static_test

import (
	"fmt"
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMacStaticResource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "mac_static_resource_test")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMacStaticResourceConfig("00:11:22:33:44:55", 10, "Port 1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_mac_static.test", "mac_address", "00:11:22:33:44:55"),
					resource.TestCheckResourceAttr("hrui_mac_static.test", "vlan_id", "10"),
					resource.TestCheckResourceAttr("hrui_mac_static.test", "port", "Port 1"),
				),
			},
			{
				Config: testAccMacStaticResourceConfig("00:11:22:33:44:55", 10, "Port 2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_mac_static.test", "port", "Port 2"),
				),
			},
			{
				Config: testAccMacStaticResourceConfig("00:11:22:33:44:66", 20, "Port 1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_mac_static.test", "mac_address", "00:11:22:33:44:66"),
					resource.TestCheckResourceAttr("hrui_mac_static.test", "vlan_id", "20"),
				),
			},
		},
	})
}

func testAccMacStaticResourceConfig(macAddress string, vlanID int64, port string) string {
	return fmt.Sprintf(`
provider "hrui" {}

resource "hrui_mac_static" "test" {
  mac_address = "%s"
  vlan_id     = %d
  port        = "%s"
}
`, macAddress, vlanID, port)
}
