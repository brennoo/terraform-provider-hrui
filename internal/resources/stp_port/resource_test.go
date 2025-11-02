package stp_port_test

import (
	"fmt"
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStpPortResource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "stp_port_resource_test")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStpPortResourceConfig("Port 1", 200000, 128, "True"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_stp_port.test", "port", "Port 1"),
					resource.TestCheckResourceAttr("hrui_stp_port.test", "path_cost", "200000"),
					resource.TestCheckResourceAttr("hrui_stp_port.test", "priority", "128"),
					resource.TestCheckResourceAttr("hrui_stp_port.test", "edge", "True"),
				),
			},
			{
				Config: testAccStpPortResourceConfig("Port 1", 100000, 64, "False"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_stp_port.test", "path_cost", "100000"),
					resource.TestCheckResourceAttr("hrui_stp_port.test", "priority", "64"),
					resource.TestCheckResourceAttr("hrui_stp_port.test", "edge", "False"),
				),
			},
		},
	})
}

func testAccStpPortResourceConfig(port string, pathCost, priority int64, edge string) string {
	return fmt.Sprintf(`
provider "hrui" {}

resource "hrui_stp_port" "test" {
  port      = "%s"
  path_cost = %d
  priority  = %d
  edge      = "%s"
}
`, port, pathCost, priority, edge)
}
