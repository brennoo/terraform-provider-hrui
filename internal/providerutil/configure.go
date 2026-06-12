package providerutil

import (
	"fmt"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ConfigureClient extracts *sdk.HRUIClient from ProviderData for use in resource/datasource Configure methods.
// Returns nil without error when ProviderData is nil (framework calls Configure before the provider is ready
// during planning/validation phases).
func ConfigureClient(providerData any, diags *diag.Diagnostics) *sdk.HRUIClient {
	if providerData == nil {
		return nil
	}
	client, ok := providerData.(*sdk.HRUIClient)
	if !ok {
		diags.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.HRUIClient, got: %T. Please report this issue to the provider developers.", providerData),
		)
		return nil
	}
	return client
}
