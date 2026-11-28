package poe_port_test

import (
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/poe_port"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPoEPortDataSource_Schema(t *testing.T) {
	d := poe_port.NewDataSource()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}
	d.Schema(t.Context(), req, resp)

	if resp.Schema.Attributes["port"] == nil || resp.Schema.Attributes["voltage_volts"] == nil {
		t.Fatal("expected attributes in poe_port data source schema")
	}
}

func TestAccPoEPortDataSource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "poe_port_data_source")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "hrui" {}

data "hrui_poe_port" "test" {
  port = "Port 1"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hrui_poe_port.test", "port", "Port 1"),
					resource.TestCheckResourceAttr("data.hrui_poe_port.test", "enabled", "true"),
					resource.TestCheckResourceAttr("data.hrui_poe_port.test", "operation", "On"),
					resource.TestCheckResourceAttr("data.hrui_poe_port.test", "voltage_volts", "53.2"),
					resource.TestCheckResourceAttr("data.hrui_poe_port.test", "current_ma", "250"),
				),
			},
		},
	})
}
