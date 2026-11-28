package dhcp_snooping_test

import (
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/dhcp_snooping"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDHCPSnoopingDataSource_Schema(t *testing.T) {
	d := dhcp_snooping.NewDataSource()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}
	d.Schema(t.Context(), req, resp)

	if resp.Schema.Attributes["enabled"] == nil || resp.Schema.Attributes["trusted_ports"] == nil {
		t.Fatal("expected enabled and trusted_ports attributes in data source schema")
	}
}

func TestAccDHCPSnoopingDataSource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "dhcp_snooping_data_source")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "hrui" {}

resource "hrui_dhcp_snooping" "setup" {
  enabled       = true
  trusted_ports = ["Port 1"]
}

data "hrui_dhcp_snooping" "test" {
  depends_on = [hrui_dhcp_snooping.setup]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hrui_dhcp_snooping.test", "enabled", "true"),
					resource.TestCheckResourceAttr("data.hrui_dhcp_snooping.test", "trusted_ports.#", "1"),
					resource.TestCheckResourceAttr("data.hrui_dhcp_snooping.test", "trusted_ports.0", "Port 1"),
				),
			},
		},
	})
}
