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

// Metadata sets the data source type name in Terraform
func (d *vlan8021qDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan_8021q"
}

// Schema defines the schema for the data source.
func (d *vlan8021qDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
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
				ElementType:         types.Int64Type,
				MarkdownDescription: "List of untagged ports for the queried VLAN.",
			},
			"tagged_ports": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.Int64Type,
				MarkdownDescription: "List of tagged ports for the queried VLAN.",
			},
			"member_ports": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.Int64Type,
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
	// Assign the client to the data source struct for further use.
	d.client = client
}

func (d *vlan8021qDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model vlan8021qModel

	// Retrieve the VLAN ID from the user request.
	diags := req.Config.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure that the HRUIClient has been initialized
	if d.client == nil {
		resp.Diagnostics.AddError("Missing HRUI Client", "The HRUI client has not been properly initialized in the Configure method.")
		return
	}

	// Fetch VLAN data
	vlanID := model.VlanID.ValueInt64()
	vlan, err := d.client.GetVLAN(int(vlanID))
	if err != nil {
		resp.Diagnostics.AddError("Error fetching VLAN", fmt.Sprintf("Could not fetch VLAN with ID %d: %s", vlanID, err.Error()))
		return
	}

	// Update the model with sanitized data returned from the SDK
	model.VlanID = types.Int64Value(int64(vlan.VlanID))
	model.Name = types.StringValue(vlan.Name)
	model.UntaggedPorts = sdk.FlattenInt64List(vlan.UntaggedPorts)
	model.TaggedPorts = sdk.FlattenInt64List(vlan.TaggedPorts)
	model.MemberPorts = sdk.FlattenInt64List(vlan.MemberPorts)

	// Set the updated model back into the Terraform state
	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
}
