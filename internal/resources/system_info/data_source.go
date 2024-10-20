package system_info

import (
	"context"
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/brennoo/terraform-provider-hrui/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &systemInfoDataSource{}
	_ datasource.DataSourceWithConfigure = &systemInfoDataSource{}
)

// systemInfoDataSource is the data source implementation.
type systemInfoDataSource struct {
	client *client.Client
}

// NewDataSourceSystemInfo is a helper function to simplify the provider implementation.
func NewDataSourceSystemInfo() datasource.DataSource {
	return &systemInfoDataSource{}
}

// Configure adds the provider configured client to the data source.
func (d *systemInfoDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *systemInfoDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_info"
}

func (d *systemInfoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data systemInfoModel

	url := fmt.Sprintf("%s/info.cgi", d.client.URL)
	httpResp, err := d.client.MakeRequest(url)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read HRUI system info, got error: %s", err))
		return
	}

	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Unexpected HTTP Status Code", fmt.Sprintf("Unable to read HRUI system info, got HTTP status code: %d", httpResp.StatusCode))
		return
	}

	// Parse HTML with goquery
	doc, err := goquery.NewDocumentFromReader(httpResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("HTML Parsing Error", fmt.Sprintf("Unable to parse HRUI system info HTML response, got error: %s", err))
		return
	}

	// Extract data using goquery selectors
	data.DeviceModel = types.StringValue(doc.Find("th:contains('Device Model') + td").Text())
	data.MACAddress = types.StringValue(doc.Find("th:contains('MAC Address') + td").Text())
	data.IPAddress = types.StringValue(doc.Find("th:contains('IP Address') + td").Text())
	data.Netmask = types.StringValue(doc.Find("th:contains('Netmask') + td").Text())
	data.Gateway = types.StringValue(doc.Find("th:contains('Gateway') + td").Text())
	data.FirmwareVersion = types.StringValue(doc.Find("th:contains('Firmware Version') + td").Text())
	data.FirmwareDate = types.StringValue(doc.Find("th:contains('Firmware Date') + td").Text())
	data.HardwareVersion = types.StringValue(doc.Find("th:contains('Hardware Version') + td").Text())

	// Set the ID
	id := data.MACAddress.ValueString()

	// Set the state with the ID included
	data.ID = types.StringValue(id)

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
