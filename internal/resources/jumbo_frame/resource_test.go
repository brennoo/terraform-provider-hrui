package jumbo_frame_test

import (
	"fmt"
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccJumboFrameResource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "jumbo_frame_resource_test")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJumboFrameResourceConfig(1522),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_jumbo_frame.test", "size", "1522"),
				),
			},
			{
				Config: testAccJumboFrameResourceConfig(9216),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_jumbo_frame.test", "size", "9216"),
				),
			},
			{
				Config: testAccJumboFrameResourceConfig(16383),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_jumbo_frame.test", "size", "16383"),
				),
			},
		},
	})
}

func testAccJumboFrameResourceConfig(size int64) string {
	return fmt.Sprintf(`
provider "hrui" {}

resource "hrui_jumbo_frame" "test" {
  size = %d
}
`, size)
}
