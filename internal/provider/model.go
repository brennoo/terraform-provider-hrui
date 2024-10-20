package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// hruiProviderModel describes the provider data model.
type hruiProviderModel struct {
	URL      types.String `tfsdk:"url"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Autosave types.Bool   `tfsdk:"autosave"`
}
