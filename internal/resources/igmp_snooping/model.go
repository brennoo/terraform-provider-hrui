package igmp_snooping

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// igmpSnoopingModel maps the schema to Go types.
type igmpSnoopingModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}
