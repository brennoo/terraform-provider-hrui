package ip_address_settings_test

import (
	"fmt"
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIPAddressSettingsResource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "ip_address_settings_resource_test")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAddressSettingsResourceConfig(false, "192.168.178.30", "255.255.255.0", "192.168.178.1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_ip_address_settings.test", "dhcp_enabled", "false"),
					resource.TestCheckResourceAttr("hrui_ip_address_settings.test", "ip_address", "192.168.178.30"),
					resource.TestCheckResourceAttr("hrui_ip_address_settings.test", "netmask", "255.255.255.0"),
					resource.TestCheckResourceAttr("hrui_ip_address_settings.test", "gateway", "192.168.178.1"),
				),
			},
			{
				Config: testAccIPAddressSettingsResourceConfig(true, "192.168.178.30", "255.255.255.0", "192.168.178.1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_ip_address_settings.test", "dhcp_enabled", "true"),
				),
			},
			{
				Config: testAccIPAddressSettingsResourceConfig(false, "192.168.178.30", "255.255.255.0", "192.168.178.1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_ip_address_settings.test", "ip_address", "192.168.178.30"),
					resource.TestCheckResourceAttr("hrui_ip_address_settings.test", "dhcp_enabled", "false"),
				),
			},
		},
	})
}

func testAccIPAddressSettingsResourceConfig(dhcpEnabled bool, ipAddress, netmask, gateway string) string {
	return fmt.Sprintf(`
provider "hrui" {}

resource "hrui_ip_address_settings" "test" {
  dhcp_enabled = %t
  ip_address    = "%s"
  netmask      = "%s"
  gateway      = "%s"
}
`, dhcpEnabled, ipAddress, netmask, gateway)
}
