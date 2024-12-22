package igmp_snooping_static

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type igmpSnoopingStaticModel struct {
	Port    types.Int64 `tfsdk:"port"`
	Enabled types.Bool  `tfsdk:"enabled"`
}
