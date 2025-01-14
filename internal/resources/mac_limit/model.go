package mac_limit

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// macLimitModel defines the schema model for the MAC limit resource.
type macLimitModel struct {
	Port    types.String `tfsdk:"port"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Limit   types.Int64  `tfsdk:"limit"`
}
