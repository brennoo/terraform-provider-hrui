package vlan_8021q

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// vlan8021qModel represents the data structure for the VLAN resource and data source within Terraform.
type vlan8021qModel struct {
	VlanID        types.Int64   `tfsdk:"vlan_id"`
	Name          types.String  `tfsdk:"name"`
	UntaggedPorts []types.Int64 `tfsdk:"untagged_ports"`
	TaggedPorts   []types.Int64 `tfsdk:"tagged_ports"`
	MemberPorts   []types.Int64 `tfsdk:"member_ports",computed`
}
