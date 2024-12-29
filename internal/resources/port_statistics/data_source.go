package port_statistics

import (
	"context"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// portStatisticsDataSource defines the Port Statistics data source.
type portStatisticsDataSource struct {
	client *sdk.HRUIClient
}

// Ensure the data source implements necessary interfaces.
var (
	_ datasource.DataSource              = &portStatisticsDataSource{}
	_ datasource.DataSourceWithConfigure = &portStatisticsDataSource{}
)

// NewDataSource creates a new instance of the Port Statistics data source.
func NewDataSource() datasource.DataSource {
	return &portStatisticsDataSource{}
}

// Metadata sets the data source type name.
func (d *portStatisticsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_port_statistics"
}

// Schema defines the schema for the Port Statistics data source.
func (d *portStatisticsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source to fetch port statistics from the switch.",
		Attributes: map[string]schema.Attribute{
			"port_statistics": schema.ListNestedAttribute{
				Description: "List of port statistics retrieved from the switch.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"port": schema.StringAttribute{
							Description: "The port ID (e.g., 'Port 1', 'Trunk1').",
							Computed:    true,
						},
						"state": schema.StringAttribute{
							Description: "The state of the port (Enable/Disable).",
							Computed:    true,
						},
						"link_status": schema.StringAttribute{
							Description: "The link status of the port (Up/Down).",
							Computed:    true,
						},
						"tx_packets": schema.SingleNestedAttribute{
							Description: "Transmitter packet statistics (good/bad).",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"good": schema.Int64Attribute{
									Description: "The number of successfully transmitted packets.",
									Computed:    true,
								},
								"bad": schema.Int64Attribute{
									Description: "The number of bad transmitted packets.",
									Computed:    true,
								},
							},
						},
						"rx_packets": schema.SingleNestedAttribute{
							Description: "Receiver packet statistics (good/bad).",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"good": schema.Int64Attribute{
									Description: "The number of successfully received packets.",
									Computed:    true,
								},
								"bad": schema.Int64Attribute{
									Description: "The number of bad received packets.",
									Computed:    true,
								},
							},
						},
					},
				},
			},
		},
	}
}

// Configure associates the client to the data source.
func (d *portStatisticsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*sdk.HRUIClient)
	if !ok || client == nil {
		resp.Diagnostics.AddError(
			"Invalid Provider Client",
			"The provider client could not be configured. Please ensure the provider is correctly set up.",
		)
		return
	}

	// Assign the configured client to the data source
	d.client = client
}

// Read fetches the port statistics from the switch.
func (d *portStatisticsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Ensure the client is not nil
	if d.client == nil {
		resp.Diagnostics.AddError(
			"Client Not Configured",
			"The client has not been configured. Ensure the provider is properly set up.",
		)
		return
	}

	// Fetch port statistics from the SDK client
	portStats, err := d.client.GetPortStatistics()
	if err != nil {
		resp.Diagnostics.AddError("Unable to fetch port statistics", err.Error())
		return
	}

	// Transform the fetched data into Terraform state representation
	portStatistics := make([]portStatisticsModel, len(portStats))
	for i, stat := range portStats {
		// Convert state (int) to string representation
		state := "Disable"
		if stat.State == 1 {
			state = "Enable"
		}

		portStatistics[i] = portStatisticsModel{
			Port:       types.StringValue(stat.Port),
			State:      types.StringValue(state),
			LinkStatus: types.StringValue(stat.LinkStatus),
			TxPackets: &packetStatistics{
				Good: types.Int64Value(stat.TxGoodPkt),
				Bad:  types.Int64Value(stat.TxBadPkt),
			},
			RxPackets: &packetStatistics{
				Good: types.Int64Value(stat.RxGoodPkt),
				Bad:  types.Int64Value(stat.RxBadPkt),
			},
		}
	}

	// Set the state with the transformed data
	state := portStatisticsDataSourceModel{
		PortStatistics: portStatistics,
	}
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
