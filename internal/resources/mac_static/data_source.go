package mac_static

import (
	"context"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure implementation satisfies the datasource.DataSource interface.
var _ datasource.DataSource = &macStaticDataSource{}

// macStaticDataSource retrieves static MAC addresses from the HRUI device.
type macStaticDataSource struct {
	client *sdk.HRUIClient
}

// NewDataSource initializes a new instance of the data source.
func NewDataSource() datasource.DataSource {
	return &macStaticDataSource{}
}

// Metadata sets the type name for the data source.
func (d *macStaticDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mac_static"
}

// Schema defines the schema for the data source.
func (d *macStaticDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for querying static MAC addresses from the HRUI device.",
		Attributes: map[string]schema.Attribute{
			"mac_address": schema.StringAttribute{
				Description: "Filter results by a specific MAC address in the format xx:xx:xx:xx:xx:xx.",
				Optional:    true,
			},
			"vlan_id": schema.Int64Attribute{
				Description: "Filter results by a specific VLAN ID.",
				Optional:    true,
			},
			"port": schema.StringAttribute{
				Description: "Filter results by a specific port (e.g., 'Port 1' or 'Trunk2').",
				Optional:    true,
			},
			"entries": schema.ListNestedAttribute{
				Description: "List of matching static MAC entries.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"mac_address": schema.StringAttribute{
							Description: "The MAC address.",
							Computed:    true,
						},
						"vlan_id": schema.Int64Attribute{
							Description: "The VLAN ID associated with the MAC address.",
							Computed:    true,
						},
						"port": schema.StringAttribute{
							Description: "The port associated with the MAC address (e.g., 'Port 1' or 'Trunk2').",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure assigns the SDK client from provider configuration.
func (d *macStaticDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		client, ok := req.ProviderData.(*sdk.HRUIClient)
		if !ok {
			resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *sdk.HRUIClient")
			return
		}
		d.client = client
	}
}

// Read queries static MAC addresses and returns the filtered results.
func (d *macStaticDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Extract filters from the request
	var filters macStaticDataSourceModel
	diags := req.Config.Get(ctx, &filters)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve the static MAC address table from the SDK
	macTable, err := d.client.GetStaticMACAddressTable()
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch static MAC table", err.Error())
		return
	}

	// Apply filters to the MAC table
	var matchingEntries []macStaticEntryModel
	for _, entry := range macTable {
		// Check if MAC Address filter is set and doesn't match
		if !filters.MACAddress.IsNull() && filters.MACAddress.ValueString() != entry.MACAddress {
			continue
		}

		// Check if VLAN ID filter is set and doesn't match
		if !filters.VLANID.IsNull() && filters.VLANID.ValueInt64() != int64(entry.VLANID) {
			continue
		}

		// Check if Port filter is set and doesn't match
		if !filters.Port.IsNull() && filters.Port.ValueString() != entry.Port {
			continue
		}

		// Add the matching entry to the results list
		matchingEntries = append(matchingEntries, macStaticEntryModel{
			MACAddress: types.StringValue(entry.MACAddress),
			VLANID:     types.Int64Value(int64(entry.VLANID)),
			Port:       types.StringValue(entry.Port),
		})
	}

	// Populate computed state
	var state macStaticDataSourceModel
	state.MACAddress = filters.MACAddress
	state.VLANID = filters.VLANID
	state.Port = filters.Port
	state.Entries = matchingEntries

	// Set the state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
