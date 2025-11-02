package vlan_8021q_test

import (
	"fmt"
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVlan8021qResource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "vlan_8021q_resource_test")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVlan8021qResourceConfig(100, "test-vlan", []string{"Port 1"}, []string{"Port 2"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_vlan_8021q.test", "vlan_id", "100"),
					resource.TestCheckResourceAttr("hrui_vlan_8021q.test", "name", "test-vlan"),
					resource.TestCheckResourceAttr("hrui_vlan_8021q.test", "untagged_ports.#", "1"),
					resource.TestCheckResourceAttr("hrui_vlan_8021q.test", "untagged_ports.0", "Port 1"),
					resource.TestCheckResourceAttr("hrui_vlan_8021q.test", "tagged_ports.#", "1"),
					resource.TestCheckResourceAttr("hrui_vlan_8021q.test", "tagged_ports.0", "Port 2"),
				),
			},
			{
				Config: testAccVlan8021qResourceConfig(100, "test-vlan-upd", []string{"Port 1", "Port 3"}, []string{"Port 2"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_vlan_8021q.test", "name", "test-vlan-upd"),
					resource.TestCheckResourceAttr("hrui_vlan_8021q.test", "untagged_ports.#", "2"),
					resource.TestCheckResourceAttr("hrui_vlan_8021q.test", "untagged_ports.1", "Port 3"),
				),
			},
			{
				Config: testAccVlan8021qResourceConfig(100, "test-vlan-last", []string{"Port 2"}, []string{"Port 1", "Port 3"}),
				Check: resource.ComposeTestCheckFunc(
					// Device truncates VLAN names to 14 characters, "test-vlan-last" is 14 chars
					resource.TestCheckResourceAttr("hrui_vlan_8021q.test", "name", "test-vlan-last"),
					resource.TestCheckResourceAttr("hrui_vlan_8021q.test", "untagged_ports.#", "1"),
					resource.TestCheckResourceAttr("hrui_vlan_8021q.test", "tagged_ports.#", "2"),
				),
			},
		},
	})
}

func testAccVlan8021qResourceConfig(vlanID int64, name string, untaggedPorts, taggedPorts []string) string {
	untaggedStr := ""
	for i, port := range untaggedPorts {
		if i > 0 {
			untaggedStr += ", "
		}
		untaggedStr += fmt.Sprintf(`"%s"`, port)
	}

	taggedStr := ""
	for i, port := range taggedPorts {
		if i > 0 {
			taggedStr += ", "
		}
		taggedStr += fmt.Sprintf(`"%s"`, port)
	}

	return fmt.Sprintf(`
provider "hrui" {}

resource "hrui_vlan_8021q" "test" {
  vlan_id        = %d
  name           = "%s"
  untagged_ports = [%s]
  tagged_ports   = [%s]
}
`, vlanID, name, untaggedStr, taggedStr)
}
