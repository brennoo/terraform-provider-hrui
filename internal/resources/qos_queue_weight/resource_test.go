package qos_queue_weight_test

import (
	"fmt"
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccQosQueueWeightResource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "qos_queue_weight_resource_test")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQosQueueWeightResourceConfig(0, "1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_qos_queue_weight.test", "queue_id", "0"),
					resource.TestCheckResourceAttr("hrui_qos_queue_weight.test", "weight", "1"),
				),
			},
			{
				Config: testAccQosQueueWeightResourceConfig(0, "5"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_qos_queue_weight.test", "weight", "5"),
				),
			},
			{
				Config: testAccQosQueueWeightResourceConfig(0, "Strict priority"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_qos_queue_weight.test", "weight", "Strict priority"),
				),
			},
		},
	})
}

func testAccQosQueueWeightResourceConfig(queueID int64, weight string) string {
	return fmt.Sprintf(`
provider "hrui" {}

resource "hrui_qos_queue_weight" "test" {
  queue_id = %d
  weight   = "%s"
}
`, queueID, weight)
}
