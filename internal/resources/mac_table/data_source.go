package mac_table

import (
	"context"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// macTableDataSource defines the structure of the MAC table data source.
type macTableDataSource struct {
	client *sdk.HRUIClient
}

// NewDataSource creates a new instance of the MAC table data source.
func NewDataSource() datasource.DataSource {
	return &macTableDataSource{}
}

// Metadata sets the data source type name.
func (d *macTableDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mac_table"
}

// Schema defines the schema for the MAC table data source.
func (d *macTableDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source to fetch the MAC address table from the switch.",
		Attributes: map[string]schema.Attribute{
			"mac_table": schema.ListNestedAttribute{
				Description: "List of MAC table entries retrieved from the switch.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "The sequence number of the entry.",
							Computed:    true,
						},
						"mac_address": schema.StringAttribute{
							Description: "The MAC address in the format xx:xx:xx:xx:xx:xx.",
							Computed:    true,
						},
						"vlan_id": schema.Int64Attribute{
							Description: "The VLAN ID associated with the MAC address.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The type of the MAC address entry (e.g., dynamic or static).",
							Computed:    true,
						},
						"port": schema.Int64Attribute{
							Description: "The port number where the MAC address is located.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure associates the client to the data source.
func (d *macTableDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		client, ok := req.ProviderData.(*sdk.HRUIClient)
		if !ok {
			resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *sdk.HRUIClient")
			return
		}
		d.client = client
	}
}

// Read fetches the MAC address table data from the switch.
func (d *macTableDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Call the SDK to fetch the MAC address table
	macTableEntries, err := d.client.GetMACAddressTable()
	if err != nil {
		resp.Diagnostics.AddError("Unable to fetch MAC address table", err.Error())
		return
	}

	// Transform the fetched data into Terraform state representation
	macTable := make([]macTableModel, len(macTableEntries))
	for i, entry := range macTableEntries {
		macTable[i] = macTableModel{
			ID:         types.Int64Value(int64(entry.ID)),
			MACAddress: types.StringValue(entry.MAC),
			VLANID:     types.Int64Value(int64(entry.VLANID)),
			Type:       types.StringValue(entry.Type),
			Port:       types.Int64Value(int64(entry.Port)),
		}
	}

	// Set the state with the retrieved MAC table
	state := macTableDataSourceModel{
		MacTable: macTable,
	}
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
