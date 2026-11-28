package trunk_group_test

import (
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTrunkGroupDataSource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "trunk_group_data_source_test")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "hrui" {}

resource "hrui_trunk_group" "test" {
  id    = 1
  type  = "static"
  ports = [1, 2]
}

data "hrui_trunk_group" "test" {
  id         = hrui_trunk_group.test.id
  depends_on = [hrui_trunk_group.test]
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.hrui_trunk_group.test", "id", "1"),
					resource.TestCheckResourceAttr("data.hrui_trunk_group.test", "type", "static"),
					resource.TestCheckResourceAttr("data.hrui_trunk_group.test", "ports.#", "2"),
					resource.TestCheckResourceAttr("data.hrui_trunk_group.test", "ports.0", "1"),
					resource.TestCheckResourceAttr("data.hrui_trunk_group.test", "ports.1", "2"),
				),
			},
		},
	})
}
