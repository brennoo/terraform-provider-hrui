package ip_address_settings

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ipAddressModel struct {
	DHCPEnabled types.Bool   `tfsdk:"dhcp_enabled"`
	IPAddress   types.String `tfsdk:"ip_address"`
	Netmask     types.String `tfsdk:"netmask"`
	Gateway     types.String `tfsdk:"gateway"`
}
