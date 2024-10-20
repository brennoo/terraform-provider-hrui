package system_info

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type systemInfoModel struct {
	ID              types.String `tfsdk:"id"`
	DeviceModel     types.String `tfsdk:"device_model"`
	MACAddress      types.String `tfsdk:"mac_address"`
	IPAddress       types.String `tfsdk:"ip_address"`
	Netmask         types.String `tfsdk:"netmask"`
	Gateway         types.String `tfsdk:"gateway"`
	FirmwareVersion types.String `tfsdk:"firmware_version"`
	FirmwareDate    types.String `tfsdk:"firmware_date"`
	HardwareVersion types.String `tfsdk:"hardware_version"`
}
