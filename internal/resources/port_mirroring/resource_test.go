package port_mirroring_test

import (
	"fmt"
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPortMirroringResource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "port_mirroring_resource_test")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPortMirroringResourceConfig("BOTH", "Port 1", "Port 2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_port_mirroring.test", "mirror_direction", "BOTH"),
					resource.TestCheckResourceAttr("hrui_port_mirroring.test", "mirroring_port", "Port 1"),
					resource.TestCheckResourceAttr("hrui_port_mirroring.test", "mirrored_port", "Port 2"),
				),
			},
			{
				Config: testAccPortMirroringResourceConfig("Rx", "Port 1", "Port 2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_port_mirroring.test", "mirror_direction", "Rx"),
				),
			},
			{
				Config: testAccPortMirroringResourceConfig("Tx", "Port 1", "Port 3"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_port_mirroring.test", "mirror_direction", "Tx"),
					resource.TestCheckResourceAttr("hrui_port_mirroring.test", "mirrored_port", "Port 3"),
				),
			},
		},
	})
}

func testAccPortMirroringResourceConfig(mirrorDirection, mirroringPort, mirroredPort string) string {
	return fmt.Sprintf(`
provider "hrui" {}

resource "hrui_port_mirroring" "test" {
  mirror_direction = "%s"
  mirroring_port   = "%s"
  mirrored_port    = "%s"
}
`, mirrorDirection, mirroringPort, mirroredPort)
}
