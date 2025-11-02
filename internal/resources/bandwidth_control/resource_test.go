package bandwidth_control_test

import (
	"fmt"
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBandwidthControlResource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "bandwidth_control_resource_test")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBandwidthControlResourceConfig("Port 1", "992", "2000"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_bandwidth_control.test", "port", "Port 1"),
					resource.TestCheckResourceAttr("hrui_bandwidth_control.test", "ingress_rate", "992"),
					resource.TestCheckResourceAttr("hrui_bandwidth_control.test", "egress_rate", "2000"),
				),
			},
			{
				Config: testAccBandwidthControlResourceConfig("Port 1", "512", "1536"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_bandwidth_control.test", "ingress_rate", "512"),
					resource.TestCheckResourceAttr("hrui_bandwidth_control.test", "egress_rate", "1536"),
				),
			},
			{
				Config: testAccBandwidthControlResourceConfig("Port 1", "Unlimited", "Unlimited"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_bandwidth_control.test", "ingress_rate", "Unlimited"),
					resource.TestCheckResourceAttr("hrui_bandwidth_control.test", "egress_rate", "Unlimited"),
				),
			},
		},
	})
}

func testAccBandwidthControlResourceConfig(port, ingressRate, egressRate string) string {
	return fmt.Sprintf(`
provider "hrui" {}

resource "hrui_bandwidth_control" "test" {
  port         = "%s"
  ingress_rate = "%s"
  egress_rate  = "%s"
}
`, port, ingressRate, egressRate)
}
