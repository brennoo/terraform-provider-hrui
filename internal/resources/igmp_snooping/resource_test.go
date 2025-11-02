package igmp_snooping_test

import (
	"fmt"
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIgmpSnoopingResource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "igmp_snooping_resource_test")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIgmpSnoopingResourceConfig(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_igmp_snooping.test", "enabled", "true"),
				),
			},
			{
				Config: testAccIgmpSnoopingResourceConfig(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_igmp_snooping.test", "enabled", "false"),
				),
			},
			{
				Config: testAccIgmpSnoopingResourceConfig(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_igmp_snooping.test", "enabled", "true"),
				),
			},
		},
	})
}

func testAccIgmpSnoopingResourceConfig(enabled bool) string {
	return fmt.Sprintf(`
provider "hrui" {}

resource "hrui_igmp_snooping" "test" {
  enabled = %t
}
`, enabled)
}
