package mac_limit_test

import (
	"fmt"
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMacLimitResource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "mac_limit_resource_test")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMacLimitResourceConfig("Port 1", true, 100),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_mac_limit.test", "port", "Port 1"),
					resource.TestCheckResourceAttr("hrui_mac_limit.test", "enabled", "true"),
					resource.TestCheckResourceAttr("hrui_mac_limit.test", "limit", "100"),
				),
			},
			{
				Config: testAccMacLimitResourceConfig("Port 1", true, 200),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_mac_limit.test", "limit", "200"),
				),
			},
			{
				Config: testAccMacLimitResourceConfigDisabled("Port 1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_mac_limit.test", "enabled", "false"),
				),
			},
		},
	})
}

func testAccMacLimitResourceConfig(port string, enabled bool, limit int64) string {
	return fmt.Sprintf(`
provider "hrui" {}

resource "hrui_mac_limit" "test" {
  port    = "%s"
  enabled = %t
  limit   = %d
}
`, port, enabled, limit)
}

func testAccMacLimitResourceConfigDisabled(port string) string {
	return fmt.Sprintf(`
provider "hrui" {}

resource "hrui_mac_limit" "test" {
  port    = "%s"
  enabled = false
}
`, port)
}
