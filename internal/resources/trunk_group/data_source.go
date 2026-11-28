package trunk_group

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
	_ datasource.DataSource              = &trunkGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &trunkGroupDataSource{}
)

type trunkGroupDataSource struct {
	client *sdk.HRUIClient
}

// NewDataSource creates a new instance of the trunk group data source.
func NewDataSource() datasource.DataSource {
	return &trunkGroupDataSource{}
}

func (d *trunkGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trunk_group"
}

func (d *trunkGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a configured trunk group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The trunk group ID to query.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Type of the trunk group ('static' or 'LACP').",
				Computed:    true,
			},
			"ports": schema.ListAttribute{
				Description: "List of ports in the trunk group (1-indexed).",
				ElementType: types.Int64Type,
				Computed:    true,
			},
		},
	}
}

func (d *trunkGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

func (d *trunkGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data trunkGroupModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	trunkGroup, err := d.client.GetTrunk(ctx, int(data.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Trunk Group",
			fmt.Sprintf("Could not read trunk group %d: %s", data.ID.ValueInt64(), err),
		)
		return
	}

	data.Type = types.StringValue(trunkGroup.Type)
	trunkPorts := make([]int64, len(trunkGroup.Ports))
	for i, port := range trunkGroup.Ports {
		trunkPorts[i] = int64(port)
	}
	data.Ports, diags = types.ListValueFrom(ctx, types.Int64Type, trunkPorts)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
