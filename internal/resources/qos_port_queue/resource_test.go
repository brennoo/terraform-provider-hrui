package qos_port_queue_test

import (
	"fmt"
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccQosPortQueueResource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "qos_port_queue_resource_test")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQosPortQueueResourceConfig("Port 1", 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_qos_port_queue.test", "port", "Port 1"),
					resource.TestCheckResourceAttr("hrui_qos_port_queue.test", "queue", "1"),
				),
			},
			{
				Config: testAccQosPortQueueResourceConfig("Port 1", 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_qos_port_queue.test", "queue", "2"),
				),
			},
			{
				Config: testAccQosPortQueueResourceConfig("Port 1", 8),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_qos_port_queue.test", "queue", "8"),
				),
			},
		},
	})
}

func testAccQosPortQueueResourceConfig(port string, queue int64) string {
	return fmt.Sprintf(`
provider "hrui" {}

resource "hrui_qos_port_queue" "test" {
  port  = "%s"
  queue = %d
}
`, port, queue)
}
