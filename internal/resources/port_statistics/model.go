package port_statistics

import "github.com/hashicorp/terraform-plugin-framework/types"

// portStatisticsModel represents a single port's statistics schema.
type portStatisticsModel struct {
	Port       types.String      `tfsdk:"port"`
	State      types.String      `tfsdk:"state"`
	LinkStatus types.String      `tfsdk:"link_status"`
	TxPackets  *packetStatistics `tfsdk:"tx_packets"`
	RxPackets  *packetStatistics `tfsdk:"rx_packets"`
}

// packetStatistics is a helper struct for Tx/Rx packet statistics.
type packetStatistics struct {
	Good types.Int64 `tfsdk:"good"`
	Bad  types.Int64 `tfsdk:"bad"`
}

// portStatisticsDataSourceModel represents the data source schema.
type portStatisticsDataSourceModel struct {
	PortStatistics []portStatisticsModel `tfsdk:"port_statistics"`
}
