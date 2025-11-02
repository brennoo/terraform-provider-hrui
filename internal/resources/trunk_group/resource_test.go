package trunk_group_test

import (
	"fmt"
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTrunkGroupResource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "trunk_group_resource_test")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTrunkGroupResourceConfig(1, "static", []int64{1, 2}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_trunk_group.test", "id", "1"),
					resource.TestCheckResourceAttr("hrui_trunk_group.test", "type", "static"),
					resource.TestCheckResourceAttr("hrui_trunk_group.test", "ports.#", "2"),
					resource.TestCheckResourceAttr("hrui_trunk_group.test", "ports.0", "1"),
					resource.TestCheckResourceAttr("hrui_trunk_group.test", "ports.1", "2"),
				),
			},
			{
				Config: testAccTrunkGroupResourceConfig(1, "LACP", []int64{1, 2, 3}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_trunk_group.test", "type", "LACP"),
					resource.TestCheckResourceAttr("hrui_trunk_group.test", "ports.#", "3"),
					resource.TestCheckResourceAttr("hrui_trunk_group.test", "ports.2", "3"),
				),
			},
		},
	})
}

func testAccTrunkGroupResourceConfig(id int64, trunkType string, ports []int64) string {
	portsStr := ""
	for i, port := range ports {
		if i > 0 {
			portsStr += ", "
		}
		portsStr += fmt.Sprintf("%d", port)
	}

	return fmt.Sprintf(`
provider "hrui" {}

resource "hrui_trunk_group" "test" {
  id    = %d
  type  = "%s"
  ports = [%s]
}
`, id, trunkType, portsStr)
}
