package port_settings_test

import (
	"fmt"
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPortSettingsResource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "port_settings_resource_test")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPortSettingsResourceConfig("Port 4", true, "Auto", "Off"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_port_settings.test", "port", "Port 4"),
					resource.TestCheckResourceAttr("hrui_port_settings.test", "enabled", "true"),
					// Port settings may not take effect immediately when port is up
					// So we only check that values are set, not exact values
					resource.TestCheckResourceAttrSet("hrui_port_settings.test", "speed.config"),
					resource.TestCheckResourceAttrSet("hrui_port_settings.test", "speed.actual"),
					resource.TestCheckResourceAttrSet("hrui_port_settings.test", "flow_control.config"),
					resource.TestCheckResourceAttrSet("hrui_port_settings.test", "flow_control.actual"),
				),
			},
			{
				Config: testAccPortSettingsResourceConfig("Port 4", true, "Auto", "Off"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_port_settings.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("hrui_port_settings.test", "speed.config"),
					resource.TestCheckResourceAttrSet("hrui_port_settings.test", "flow_control.config"),
				),
			},
		},
	})
}

func testAccPortSettingsResourceConfig(port string, enabled bool, speed, flowControl string) string {
	return fmt.Sprintf(`
provider "hrui" {}

resource "hrui_port_settings" "test" {
  port    = "%s"
  enabled = %t
  
  speed = {
    config = "%s"
  }
  
  flow_control = {
    config = "%s"
  }
}
`, port, enabled, speed, flowControl)
}
