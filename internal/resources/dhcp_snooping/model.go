package dhcp_snooping

import "github.com/hashicorp/terraform-plugin-framework/types"

type dhcpSnoopingModel struct {
	Enabled      types.Bool `tfsdk:"enabled"`
	TrustedPorts types.List `tfsdk:"trusted_ports"`
}
