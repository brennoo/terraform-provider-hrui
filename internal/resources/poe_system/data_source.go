package poe_system

import (
	"context"
	"fmt"

	"github.com/brennoo/terraform-provider-hrui/internal/providerutil"
	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &poeSystemDataSource{}
	_ datasource.DataSourceWithConfigure = &poeSystemDataSource{}
)

type poeSystemDataSource struct {
	client *sdk.HRUIClient
}

// NewDataSource creates a new PoE system data source.
func NewDataSource() datasource.DataSource {
	return &poeSystemDataSource{}
}

func (d *poeSystemDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_poe_system"
}

func (d *poeSystemDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving global Power over Ethernet (PoE) system power budget and consumption telemetry.",
		Attributes: map[string]schema.Attribute{
			"power_limit_watts": schema.Int64Attribute{
				Computed:    true,
				Description: "Total PoE power budget in Watts configured on the switch.",
			},
			"power_consumption_watts": schema.Float64Attribute{
				Computed:    true,
				Description: "Current total PoE power consumption in Watts reported by the PSE controller.",
			},
			"power_usage_percent": schema.Int64Attribute{
				Computed:    true,
				Description: "Current total PoE power usage percentage reported by the switch.",
			},
		},
	}
}

func (d *poeSystemDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

func (d *poeSystemDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client Not Configured", "The HRUI client was not properly configured.")
		return
	}

	sys, err := d.client.GetPoESystem(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading PoE System", fmt.Sprintf("Unable to read PoE system settings: %s", err))
		return
	}

	state := poeSystemModel{
		PowerLimitWatts:       types.Int64Value(int64(sys.PowerLimitWatts)),
		PowerConsumptionWatts: types.Float64Value(sys.PowerConsumptionWatts),
		PowerUsagePercent:     types.Int64Value(int64(sys.PowerUsagePercent)),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
