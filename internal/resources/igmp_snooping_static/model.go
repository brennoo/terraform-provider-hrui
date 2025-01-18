package igmp_snooping_static

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type igmpSnoopingStaticModel struct {
	Port    types.String `tfsdk:"port"`
	Enabled types.Bool   `tfsdk:"enabled"`
}
