package port_settings_test

import (
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPortSettingsDataSource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "port_settings_data_source_test")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPortSettingsDataSourceConfig("Port 1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.hrui_port_settings.test", "port", "Port 1"),
					resource.TestCheckResourceAttrSet("data.hrui_port_settings.test", "enabled"),
					resource.TestCheckResourceAttrSet("data.hrui_port_settings.test", "speed.config"),
					resource.TestCheckResourceAttrSet("data.hrui_port_settings.test", "speed.actual"),
					resource.TestCheckResourceAttrSet("data.hrui_port_settings.test", "flow_control.config"),
					resource.TestCheckResourceAttrSet("data.hrui_port_settings.test", "flow_control.actual"),
				),
			},
			{
				Config: testAccPortSettingsDataSourceConfig("Port 2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.hrui_port_settings.test", "port", "Port 2"),
				),
			},
		},
	})
}

func testAccPortSettingsDataSourceConfig(port string) string {
	return `
provider "hrui" {}

data "hrui_port_settings" "test" {
  port = "` + port + `"
}
`
}
