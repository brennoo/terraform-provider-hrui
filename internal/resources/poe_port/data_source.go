package poe_port

import (
	"context"
	"fmt"

	"github.com/brennoo/terraform-provider-hrui/internal/providerutil"
	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &poePortDataSource{}
	_ datasource.DataSourceWithConfigure = &poePortDataSource{}
)

type poePortDataSource struct {
	client *sdk.HRUIClient
}

// NewDataSource creates a new PoE port data source.
func NewDataSource() datasource.DataSource {
	return &poePortDataSource{}
}

func (d *poePortDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_poe_port"
}

func (d *poePortDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving Power over Ethernet (PoE) port configuration and live PSE telemetry.",
		Attributes: map[string]schema.Attribute{
			"port": schema.StringAttribute{
				Required:    true,
				Description: "The PoE port name (e.g., 'Port 1').",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether PoE power delivery is administratively enabled on the port.",
			},
			"pse_type": schema.StringAttribute{
				Computed:    true,
				Description: "PoE protocol type configured on the port ('BT', 'HiPoE', 'AT', or 'AF').",
			},
			"watchdog_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the PoE watchdog (PoEDog) is enabled on the port.",
			},
			"operation": schema.StringAttribute{
				Computed:    true,
				Description: "Current PoE operational state ('On' or 'Off').",
			},
			"pd_class": schema.StringAttribute{
				Computed:    true,
				Description: "Detected Powered Device (PD) IEEE 802.3 class (e.g., 'AT(Class4)', 'AF(Class3)', or '-').",
			},
			"power_watts": schema.Float64Attribute{
				Computed:    true,
				Description: "Live PoE power draw on the port in Watts.",
			},
			"voltage_volts": schema.Float64Attribute{
				Computed:    true,
				Description: "Live PoE bus voltage on the port in Volts.",
			},
			"current_ma": schema.Int64Attribute{
				Computed:    true,
				Description: "Live PoE current draw on the port in milliamps (mA).",
			},
		},
	}
}

func (d *poePortDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

func (d *poePortDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data poePortModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if d.client == nil {
		resp.Diagnostics.AddError("Client Not Configured", "The HRUI client was not properly configured.")
		return
	}

	status, err := d.client.GetPoEPortStatus(ctx, data.Port.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading PoE Port", fmt.Sprintf("Unable to read PoE port %s: %s", data.Port.ValueString(), err))
		return
	}

	populateModelFromStatus(&data, status)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
