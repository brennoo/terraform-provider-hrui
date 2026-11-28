package dhcp_snooping

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
	_ datasource.DataSource              = &dhcpSnoopingDataSource{}
	_ datasource.DataSourceWithConfigure = &dhcpSnoopingDataSource{}
)

type dhcpSnoopingDataSource struct {
	client *sdk.HRUIClient
}

// NewDataSource creates a new DHCP snooping data source.
func NewDataSource() datasource.DataSource {
	return &dhcpSnoopingDataSource{}
}

func (d *dhcpSnoopingDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcp_snooping"
}

func (d *dhcpSnoopingDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving global DHCP Snooping status and trusted DHCP server ports.",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether global DHCP Snooping is enabled on the switch.",
			},
			"trusted_ports": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "List of switch ports configured as trusted DHCP server ports.",
			},
		},
	}
}

func (d *dhcpSnoopingDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

func (d *dhcpSnoopingDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client Not Configured", "The HRUI client was not properly configured.")
		return
	}

	cfg, err := d.client.GetDHCPSnooping(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading DHCP Snooping", fmt.Sprintf("Unable to read DHCP snooping settings: %s", err))
		return
	}

	listVal, diags := toTrustedPortsList(cfg.TrustedPorts)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := dhcpSnoopingModel{
		Enabled:      types.BoolValue(cfg.Enabled),
		TrustedPorts: listVal,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
