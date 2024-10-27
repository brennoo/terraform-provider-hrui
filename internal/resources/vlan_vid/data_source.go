package vlan_vid

import (
	"context"
	"fmt"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// vlanVIDDataSource defines the VLAN data source.
type vlanVIDDataSource struct {
	client *sdk.HRUIClient
}

// Ensure that vlanVIDDataSource implements the datasource.DataSource interface.
var _ datasource.DataSource = &vlanVIDDataSource{}

// NewDataSource creates a new instance of the VID data source.
func NewDataSource() datasource.DataSource {
	return &vlanVIDDataSource{}
}

// Metadata sets the data source type name in Terraform
func (d *vlanVIDDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan_vid"
}

// Schema defines the schema for the data source.
func (d *vlanVIDDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"port": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Port number used to query the VLAN configuration.",
			},
			"vlan_id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "VLAN ID assigned to the port.",
			},
			"accept_frame_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Accepted frame type: 'All', 'Tagged', or 'Untagged'.",
			},
		},
	}
}

func (d *vlanVIDDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read refreshes the Terraform state with the latest data.
func (d *vlanVIDDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model vlanVIDModel

	// Retrieve the port number from the user request.
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
	port := int(model.Port.ValueInt64())
	config, err := d.client.GetPortVLANConfig(port)
	if err != nil {
		resp.Diagnostics.AddError("Error fetching VLAN configuration", fmt.Sprintf("Could not fetch VLAN configuration: %s", err.Error()))
		return
	}

	// Assign values to the model
	model.VlanID = types.Int64Value(int64(config.PVID))

	acceptFrameTypeMap := map[string]string{
		"All":        "All",
		"Tag-only":   "Tagged",
		"Untag-only": "Untagged",
	}
	model.AcceptFrameType = types.StringValue(acceptFrameTypeMap[config.AcceptFrameType])

	// Set the updated model back into the Terraform state
	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
}
