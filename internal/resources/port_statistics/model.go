package port_statistics

import "github.com/hashicorp/terraform-plugin-framework/types"

// portStatisticsModel maps the resource and data source schema data.
type portStatisticsModel struct {
	Port       types.Int64       `tfsdk:"port"`
	State      types.String      `tfsdk:"state"`
	LinkStatus types.String      `tfsdk:"link_status"`
	TxPackets  *packetStatistics `tfsdk:"tx_packets"`
	RxPackets  *packetStatistics `tfsdk:"rx_packets"`
}

type packetStatistics struct {
	Good types.Int64 `tfsdk:"good"`
	Bad  types.Int64 `tfsdk:"bad"`
}

// portStatisticsDataSourceModel represents the data source containing a list of port statistics.
type portStatisticsDataSourceModel struct {
	PortStatistics []portStatisticsModel `tfsdk:"port_statistics"`
}
