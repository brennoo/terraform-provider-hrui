package mac_table

import "github.com/hashicorp/terraform-plugin-framework/types"

// macTableModel represents a single MAC table entry for Terraform's state.
type macTableModel struct {
	ID         types.Int64  `tfsdk:"id"`
	MACAddress types.String `tfsdk:"mac_address"`
	VLANID     types.Int64  `tfsdk:"vlan_id"`
	Type       types.String `tfsdk:"type"`
	Port       types.String `tfsdk:"port"`
}

// macTableDataSourceModel represents the entire MAC table data source for Terraform's state.
type macTableDataSourceModel struct {
	MacTable []macTableModel `tfsdk:"mac_table"`
}
