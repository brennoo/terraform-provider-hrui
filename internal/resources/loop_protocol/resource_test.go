package loop_protocol_test

import (
	"fmt"
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLoopProtocolResource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "loop_protocol_resource_test")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLoopProtocolResourceConfig("Loop Detection", 5, 10),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_loop_protocol.test", "loop_function", "Loop Detection"),
					resource.TestCheckResourceAttr("hrui_loop_protocol.test", "interval_time", "5"),
					resource.TestCheckResourceAttr("hrui_loop_protocol.test", "recover_time", "10"),
				),
			},
			{
				Config: testAccLoopProtocolResourceConfig("Loop Prevention", 10, 20),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_loop_protocol.test", "loop_function", "Loop Prevention"),
					resource.TestCheckResourceAttr("hrui_loop_protocol.test", "interval_time", "10"),
					resource.TestCheckResourceAttr("hrui_loop_protocol.test", "recover_time", "20"),
				),
			},
			{
				Config: testAccLoopProtocolResourceConfigOff(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_loop_protocol.test", "loop_function", "Off"),
				),
			},
		},
	})
}

func testAccLoopProtocolResourceConfig(loopFunction string, intervalTime, recoverTime int64) string {
	return fmt.Sprintf(`
provider "hrui" {}

resource "hrui_loop_protocol" "test" {
  loop_function = "%s"
  interval_time  = %d
  recover_time   = %d
}
`, loopFunction, intervalTime, recoverTime)
}

func testAccLoopProtocolResourceConfigOff() string {
	return `
provider "hrui" {}

resource "hrui_loop_protocol" "test" {
  loop_function = "Off"
}
`
}
