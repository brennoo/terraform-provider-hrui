package mac_table_test

import (
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMacTableDataSource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "mac_table_data_source_test")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMacTableDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hrui_mac_table.test", "mac_table.#"),
				),
			},
		},
	})
}

func testAccMacTableDataSourceConfig() string {
	return `
provider "hrui" {}

data "hrui_mac_table" "test" {}
`
}
