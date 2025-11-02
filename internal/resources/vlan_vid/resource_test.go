package vlan_vid_test

import (
	"fmt"
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVlanVidResource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "vlan_vid_resource_test")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVlanVidResourceConfig("Port 1", 10, "Tagged"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_vlan_vid.test", "port", "Port 1"),
					resource.TestCheckResourceAttr("hrui_vlan_vid.test", "vlan_id", "10"),
					resource.TestCheckResourceAttr("hrui_vlan_vid.test", "accept_frame_type", "Tagged"),
				),
			},
			{
				Config: testAccVlanVidResourceConfig("Port 1", 20, "Untagged"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_vlan_vid.test", "vlan_id", "20"),
					resource.TestCheckResourceAttr("hrui_vlan_vid.test", "accept_frame_type", "Untagged"),
				),
			},
			{
				Config: testAccVlanVidResourceConfig("Port 1", 20, "All"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_vlan_vid.test", "accept_frame_type", "All"),
				),
			},
		},
	})
}

func testAccVlanVidResourceConfig(port string, vlanID int64, acceptFrameType string) string {
	return fmt.Sprintf(`
provider "hrui" {}

resource "hrui_vlan_vid" "test" {
  port             = "%s"
  vlan_id          = %d
  accept_frame_type = "%s"
}
`, port, vlanID, acceptFrameType)
}
