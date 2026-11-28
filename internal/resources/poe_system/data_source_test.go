package poe_system_test

import (
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/poe_system"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPoESystemDataSource_Schema(t *testing.T) {
	d := poe_system.NewDataSource()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}
	d.Schema(t.Context(), req, resp)

	if resp.Schema.Attributes["power_limit_watts"] == nil {
		t.Fatal("power_limit_watts attribute should exist in schema")
	}
}

func TestAccPoESystemDataSource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "poe_system_data_source")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "hrui" {}

data "hrui_poe_system" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hrui_poe_system.test", "power_limit_watts", "100"),
					resource.TestCheckResourceAttrSet("data.hrui_poe_system.test", "power_consumption_watts"),
					resource.TestCheckResourceAttrSet("data.hrui_poe_system.test", "power_usage_percent"),
				),
			},
		},
	})
}
