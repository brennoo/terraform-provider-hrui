package stp_global_test

import (
	"fmt"
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStpGlobalResource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "stp_global_resource_test")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStpGlobalResourceConfig("RSTP", 32768, 20, 2, 15),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_stp_global.test", "force_version", "RSTP"),
					resource.TestCheckResourceAttr("hrui_stp_global.test", "priority", "32768"),
					resource.TestCheckResourceAttr("hrui_stp_global.test", "max_age", "20"),
					resource.TestCheckResourceAttr("hrui_stp_global.test", "hello_time", "2"),
					resource.TestCheckResourceAttr("hrui_stp_global.test", "forward_delay", "15"),
				),
			},
			{
				Config: testAccStpGlobalResourceConfig("STP", 32768, 20, 2, 15),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_stp_global.test", "force_version", "STP"),
				),
			},
			{
				Config: testAccStpGlobalResourceConfig("RSTP", 4096, 20, 2, 15),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_stp_global.test", "force_version", "RSTP"),
					resource.TestCheckResourceAttr("hrui_stp_global.test", "priority", "4096"),
				),
			},
		},
	})
}

func testAccStpGlobalResourceConfig(forceVersion string, priority, maxAge, helloTime, forwardDelay int64) string {
	return fmt.Sprintf(`
provider "hrui" {}

resource "hrui_stp_global" "test" {
  force_version = "%s"
  priority      = %d
  max_age       = %d
  hello_time    = %d
  forward_delay = %d
}
`, forceVersion, priority, maxAge, helloTime, forwardDelay)
}
