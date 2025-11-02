package storm_control_test

import (
	"fmt"
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStormControlResource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "storm_control_resource_test")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStormControlResourceConfig("Port 1", "Broadcast", true, 1000),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_storm_control.test", "port", "Port 1"),
					resource.TestCheckResourceAttr("hrui_storm_control.test", "storm_type", "Broadcast"),
					resource.TestCheckResourceAttr("hrui_storm_control.test", "state", "true"),
					resource.TestCheckResourceAttr("hrui_storm_control.test", "rate", "1000"),
				),
			},
			{
				Config: testAccStormControlResourceConfig("Port 1", "Broadcast", true, 2000),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_storm_control.test", "rate", "2000"),
				),
			},
			{
				Config: testAccStormControlResourceConfig("Port 1", "Broadcast", false, 0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_storm_control.test", "state", "false"),
				),
			},
		},
	})
}

func testAccStormControlResourceConfig(port, stormType string, state bool, rate int64) string {
	return fmt.Sprintf(`
provider "hrui" {}

resource "hrui_storm_control" "test" {
  port       = "%s"
  storm_type = "%s"
  state      = %t
  rate       = %d
}
`, port, stormType, state, rate)
}
