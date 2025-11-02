package eee_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	// Import the test helper package.
	"github.com/brennoo/terraform-provider-hrui/internal/provider"
)

// TestAccEeeResource provides a complete acceptance test for the
// hrui_eee resource, covering Create, Update, and Delete.
func TestAccEeeResource(t *testing.T) {
	// The cassette name should be unique for this test.
	// We pass `t` and the cassette name to our VCR helper.
	providerFactories := provider.TestAccProtoV6ProviderFactories(t, "eee_resource_test")

	resource.Test(t, resource.TestCase{
		// Tell the test harness to use our VCR-configured provider
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			// Step 1: Create the resource (enabled = true)
			{
				// HCL configuration for this step
				Config: testAccEeeResourceConfig(true),
				// Check functions to verify the state after apply
				Check: resource.ComposeTestCheckFunc(
					// Check that the resource "hrui_eee.test" exists
					resource.TestCheckResourceAttr("hrui_eee.test", "id", "eee"), // Assuming a static ID
					// Check that the 'enabled' attribute is 'true'
					resource.TestCheckResourceAttr("hrui_eee.test", "enabled", "true"),
				),
			},
			// Step 2: Update the resource (enabled = false)
			{
				// HCL for the update
				Config: testAccEeeResourceConfig(false),
				Check: resource.ComposeTestCheckFunc(
					// Check that the 'enabled' attribute is now 'false'
					resource.TestCheckResourceAttr("hrui_eee.test", "enabled", "false"),
				),
			},
			// Step 3: Revert to original (enabled = true)
			// This verifies that updating back works correctly.
			{
				Config: testAccEeeResourceConfig(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hrui_eee.test", "enabled", "true"),
				),
			},
			// The `resource.Test` harness automatically runs a
			// `terraform destroy` at the end of all steps, so
			// the Delete lifecycle is implicitly tested.
		},
	})
}

// testAccEeeResourceConfig is a helper function to generate
// the HCL for the hrui_eee resource.
func testAccEeeResourceConfig(enabled bool) string {
	// We use minimal provider config here.
	// The URL/user/pass are only needed for recording,
	// where they'll be set by environment variables.
	return fmt.Sprintf(`
provider "hrui" {}

resource "hrui_eee" "test" {
  enabled = %t
}
`, enabled)
}
