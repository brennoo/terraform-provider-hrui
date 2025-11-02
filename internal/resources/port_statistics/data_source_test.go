package port_statistics_test

import (
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPortStatisticsDataSource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "port_statistics_data_source_test")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPortStatisticsDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hrui_port_statistics.test", "port_statistics.#"),
				),
			},
		},
	})
}

func testAccPortStatisticsDataSourceConfig() string {
	return `
provider "hrui" {}

data "hrui_port_statistics" "test" {}
`
}
