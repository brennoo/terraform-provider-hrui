package poe_system_test

import (
	"fmt"
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/poe_system"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPoESystemResource_Schema(t *testing.T) {
	r := poe_system.NewResource()
	req := fwresource.SchemaRequest{}
	resp := &fwresource.SchemaResponse{}
	r.Schema(t.Context(), req, resp)

	if resp.Schema.Attributes["power_limit_watts"] == nil {
		t.Fatal("power_limit_watts attribute should exist in schema")
	}
	if resp.Schema.Attributes["power_consumption_watts"] == nil {
		t.Fatal("power_consumption_watts attribute should exist in schema")
	}
}

func TestAccPoESystemResource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "poe_system_resource")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPoESystemResourceConfig(90),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_poe_system.test", "power_limit_watts", "90"),
					resource.TestCheckResourceAttrSet("hrui_poe_system.test", "power_consumption_watts"),
					resource.TestCheckResourceAttrSet("hrui_poe_system.test", "power_usage_percent"),
				),
			},
			{
				Config: testAccPoESystemResourceConfig(100),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_poe_system.test", "power_limit_watts", "100"),
				),
			},
		},
	})
}

func testAccPoESystemResourceConfig(powerLimit int) string {
	return fmt.Sprintf(`
provider "hrui" {}

resource "hrui_poe_system" "test" {
  power_limit_watts = %d
}
`, powerLimit)
}
