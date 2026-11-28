package dhcp_snooping_test

import (
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/dhcp_snooping"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDHCPSnoopingResource_Schema(t *testing.T) {
	r := dhcp_snooping.NewResource()
	req := fwresource.SchemaRequest{}
	resp := &fwresource.SchemaResponse{}
	r.Schema(t.Context(), req, resp)

	if resp.Schema.Attributes["enabled"] == nil || resp.Schema.Attributes["trusted_ports"] == nil {
		t.Fatal("expected enabled and trusted_ports attributes in schema")
	}
}

func TestAccDHCPSnoopingResource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "dhcp_snooping_resource")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "hrui" {}

resource "hrui_dhcp_snooping" "test" {
  enabled       = true
  trusted_ports = ["Port 1", "Port 3"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_dhcp_snooping.test", "enabled", "true"),
					resource.TestCheckResourceAttr("hrui_dhcp_snooping.test", "trusted_ports.#", "2"),
					resource.TestCheckResourceAttr("hrui_dhcp_snooping.test", "trusted_ports.0", "Port 1"),
					resource.TestCheckResourceAttr("hrui_dhcp_snooping.test", "trusted_ports.1", "Port 3"),
				),
			},
			{
				Config: `
provider "hrui" {}

resource "hrui_dhcp_snooping" "test" {
  enabled       = true
  trusted_ports = ["Port 2"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_dhcp_snooping.test", "enabled", "true"),
					resource.TestCheckResourceAttr("hrui_dhcp_snooping.test", "trusted_ports.#", "1"),
					resource.TestCheckResourceAttr("hrui_dhcp_snooping.test", "trusted_ports.0", "Port 2"),
				),
			},
			{
				Config: `
provider "hrui" {}

resource "hrui_dhcp_snooping" "test" {
  enabled       = false
  trusted_ports = []
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_dhcp_snooping.test", "enabled", "false"),
					resource.TestCheckResourceAttr("hrui_dhcp_snooping.test", "trusted_ports.#", "0"),
				),
			},
		},
	})
}
