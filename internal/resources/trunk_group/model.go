package trunk_group

import "github.com/hashicorp/terraform-plugin-framework/types"

// trunkGroupModel represents the resource schema state.
type trunkGroupModel struct {
	ID    types.Int64  `tfsdk:"id"`
	Type  types.String `tfsdk:"type"`
	Ports types.List   `tfsdk:"ports"`
}
