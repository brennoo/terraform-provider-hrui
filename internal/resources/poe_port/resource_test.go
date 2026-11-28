package poe_port_test

import (
	"fmt"
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/provider"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/poe_port"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPoEPortResource_Schema(t *testing.T) {
	r := poe_port.NewResource()
	req := fwresource.SchemaRequest{}
	resp := &fwresource.SchemaResponse{}
	r.Schema(t.Context(), req, resp)

	for _, attr := range []string{"port", "enabled", "pse_type", "watchdog_enabled", "operation", "pd_class", "power_watts", "voltage_volts", "current_ma"} {
		if resp.Schema.Attributes[attr] == nil {
			t.Fatalf("%s attribute should exist in schema", attr)
		}
	}
}

func TestAccPoEPortResource(t *testing.T) {
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "poe_port_resource")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPoEPortResourceConfig("Port 1", true, "AT", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_poe_port.test", "port", "Port 1"),
					resource.TestCheckResourceAttr("hrui_poe_port.test", "enabled", "true"),
					resource.TestCheckResourceAttr("hrui_poe_port.test", "pse_type", "AT"),
					resource.TestCheckResourceAttr("hrui_poe_port.test", "watchdog_enabled", "true"),
					resource.TestCheckResourceAttr("hrui_poe_port.test", "operation", "On"),
					resource.TestCheckResourceAttr("hrui_poe_port.test", "voltage_volts", "53.2"),
				),
			},
			{
				Config: testAccPoEPortResourceConfig("Port 1", false, "AF", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_poe_port.test", "port", "Port 1"),
					resource.TestCheckResourceAttr("hrui_poe_port.test", "enabled", "false"),
					resource.TestCheckResourceAttr("hrui_poe_port.test", "pse_type", "AF"),
					resource.TestCheckResourceAttr("hrui_poe_port.test", "watchdog_enabled", "false"),
					resource.TestCheckResourceAttr("hrui_poe_port.test", "operation", "Off"),
				),
			},
		},
	})
}

func testAccPoEPortResourceConfig(port string, enabled bool, pseType string, watchdog bool) string {
	return fmt.Sprintf(`
provider "hrui" {}

resource "hrui_poe_port" "test" {
  port             = %q
  enabled          = %t
  pse_type         = %q
  watchdog_enabled = %t
}
`, port, enabled, pseType, watchdog)
}
