package vlan_8021q

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
)

// vlan8021qDataSource defines the VLAN data source.
type vlan8021qDataSource struct {
	client *sdk.HRUIClient
}

// Ensure that vlan8021qDataSource implements the datasource.DataSource interface.
var _ datasource.DataSource = &vlan8021qDataSource{}

// NewDataSource creates a new instance of the VLAN data source.
func NewDataSource() datasource.DataSource {
	return &vlan8021qDataSource{}
}

// Metadata sets the data source type name in Terraform.
func (d *vlan8021qDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan_8021q"
}

// Schema defines the schema for the data source.
func (d *vlan8021qDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving 802.1Q VLAN settings.",
		Attributes: map[string]schema.Attribute{
			"vlan_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "VLAN ID (1-4094) used to query the VLAN.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the queried VLAN.",
			},
			"untagged_ports": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of untagged ports for the queried VLAN.",
			},
			"tagged_ports": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of tagged ports for the queried VLAN.",
			},
			"member_ports": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of all member ports for the queried VLAN.",
			},
		},
	}
}

func (d *vlan8021qDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*sdk.HRUIClient)
	if !ok || client == nil {
		resp.Diagnostics.AddError(
			"Missing HRUI Client",
			"The client has not been properly initialized in the Configure method.",
		)
		return
	}

	// Set the client in the data source.
	d.client = client
}

// Read fetches the VLAN information and sets it in the Terraform state.
func (d *vlan8021qDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model vlan8021qModel

	// Retrieve the VLAN ID from the user input.
	diags := req.Config.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure the client is initialized.
	if d.client == nil {
		resp.Diagnostics.AddError("Missing HRUI Client", "The HRUI client has not been properly initialized in the Configure method.")
		return
	}

	// Fetch the VLAN data
	vlanID := model.VlanID.ValueInt64()
	vlan, err := d.client.GetVLAN(ctx, int(vlanID))
	if err != nil {
		resp.Diagnostics.AddError("Error fetching VLAN", fmt.Sprintf("Could not fetch VLAN with ID %d: %s", vlanID, err.Error()))
		return
	}

	// Map the fetched data from the SDK to the Terraform model.
	model.VlanID = types.Int64Value(int64(vlan.VlanID))
	model.Name = types.StringValue(vlan.Name)

	// Convert untagged_ports and tagged_ports to types.List.
	model.UntaggedPorts, diags = types.ListValueFrom(ctx, types.StringType, vlan.UntaggedPorts)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.TaggedPorts, diags = types.ListValueFrom(ctx, types.StringType, vlan.TaggedPorts)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Concatenate tagged_ports and untagged_ports directly to form member_ports.
	allPorts := append(vlan.TaggedPorts, vlan.UntaggedPorts...)
	model.MemberPorts, diags = types.ListValueFrom(ctx, types.StringType, allPorts)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Finally, set the updated values back into the state.
	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
}
