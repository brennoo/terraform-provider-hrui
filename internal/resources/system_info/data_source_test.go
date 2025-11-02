package system_info_test

import (
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSystemInfoDataSource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "system_info_data_source_test")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSystemInfoDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hrui_system_info.test", "id"),
					resource.TestCheckResourceAttrSet("data.hrui_system_info.test", "device_model"),
					resource.TestCheckResourceAttrSet("data.hrui_system_info.test", "mac_address"),
					resource.TestCheckResourceAttrSet("data.hrui_system_info.test", "ip_address"),
					resource.TestCheckResourceAttrSet("data.hrui_system_info.test", "netmask"),
					resource.TestCheckResourceAttrSet("data.hrui_system_info.test", "gateway"),
					resource.TestCheckResourceAttrSet("data.hrui_system_info.test", "firmware_version"),
					resource.TestCheckResourceAttrSet("data.hrui_system_info.test", "firmware_date"),
					resource.TestCheckResourceAttrSet("data.hrui_system_info.test", "hardware_version"),
				),
			},
		},
	})
}

func testAccSystemInfoDataSourceConfig() string {
	return `
provider "hrui" {}

data "hrui_system_info" "test" {}
`
}
