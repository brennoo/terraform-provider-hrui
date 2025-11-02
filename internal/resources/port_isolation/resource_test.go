package port_isolation_test

import (
	"fmt"
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPortIsolationResource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "port_isolation_resource_test")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPortIsolationResourceConfig("Port 1", []string{"Port 2", "Port 3"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_port_isolation.test", "port", "Port 1"),
					resource.TestCheckResourceAttr("hrui_port_isolation.test", "isolation_list.#", "2"),
					resource.TestCheckResourceAttr("hrui_port_isolation.test", "isolation_list.0", "Port 2"),
					resource.TestCheckResourceAttr("hrui_port_isolation.test", "isolation_list.1", "Port 3"),
				),
			},
			{
				Config: testAccPortIsolationResourceConfig("Port 1", []string{"Port 2", "Port 3", "Port 4"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_port_isolation.test", "isolation_list.#", "3"),
					resource.TestCheckResourceAttr("hrui_port_isolation.test", "isolation_list.2", "Port 4"),
				),
			},
			{
				Config: testAccPortIsolationResourceConfig("Port 1", []string{"Port 2"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_port_isolation.test", "isolation_list.#", "1"),
				),
			},
		},
	})
}

func testAccPortIsolationResourceConfig(port string, isolationList []string) string {
	listStr := ""
	for i, item := range isolationList {
		if i > 0 {
			listStr += ", "
		}
		listStr += fmt.Sprintf(`"%s"`, item)
	}

	return fmt.Sprintf(`
provider "hrui" {}

resource "hrui_port_isolation" "test" {
  port          = "%s"
  isolation_list = [%s]
}
`, port, listStr)
}
