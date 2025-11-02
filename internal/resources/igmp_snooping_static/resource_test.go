package igmp_snooping_static_test

import (
	"fmt"
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIgmpSnoopingStaticResource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "igmp_snooping_static_resource_test")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIgmpSnoopingStaticResourceConfig("Port 1", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_igmp_snooping_static.test", "port", "Port 1"),
					resource.TestCheckResourceAttr("hrui_igmp_snooping_static.test", "enabled", "true"),
				),
			},
			{
				Config: testAccIgmpSnoopingStaticResourceConfig("Port 1", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_igmp_snooping_static.test", "enabled", "false"),
				),
			},
		},
	})
}

func testAccIgmpSnoopingStaticResourceConfig(port string, enabled bool) string {
	return fmt.Sprintf(`
provider "hrui" {}

resource "hrui_igmp_snooping_static" "test" {
  port    = "%s"
  enabled = %t
}
`, port, enabled)
}
