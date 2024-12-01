package mac_static

import "github.com/hashicorp/terraform-plugin-framework/types"

// macStaticModel represents the resource schema state.
type macStaticModel struct {
	ID         types.String `tfsdk:"id"`
	MACAddress types.String `tfsdk:"mac_address"`
	VLANID     types.Int64  `tfsdk:"vlan_id"`
	Port       types.Int64  `tfsdk:"port"`
}

// macStaticDataSourceModel represents the filter inputs and computed outputs for the data source
type macStaticDataSourceModel struct {
	MACAddress types.String          `tfsdk:"mac_address"`
	VLANID     types.Int64           `tfsdk:"vlan_id"`
	Port       types.Int64           `tfsdk:"port"`
	Entries    []macStaticEntryModel `tfsdk:"entries"`
}

// macStaticEntryModel represents an individual static MAC entry in the output
type macStaticEntryModel struct {
	MACAddress types.String `tfsdk:"mac_address"`
	VLANID     types.Int64  `tfsdk:"vlan_id"`
	Port       types.Int64  `tfsdk:"port"`
}
