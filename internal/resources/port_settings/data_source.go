package port_settings

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/brennoo/terraform-provider-hrui/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &portSettingDataSource{}
	_ datasource.DataSourceWithConfigure = &portSettingDataSource{}
)

// portSettingDataSource is the data source implementation.
type portSettingDataSource struct {
	client *client.Client
}

// NewDataSourcePortSetting is a helper function to simplify the provider implementation.
func NewDataSourcePortSetting() datasource.DataSource {
	return &portSettingDataSource{}
}

// Configure adds the provider configured client to the data source.
func (d *portSettingDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *portSettingDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_port_settings"
}

func (d *portSettingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data portSettingModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/port.cgi", d.client.URL)
	httpResp, err := d.client.MakeRequest(url)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read HRUI port settings, got error: %s", err))
		return
	}

	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Unexpected HTTP Status Code", fmt.Sprintf("Unable to read HRUI port settings, got HTTP status code: %d", httpResp.StatusCode))
		return
	}

	// Parse HTML with goquery
	doc, err := goquery.NewDocumentFromReader(httpResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("HTML Parsing Error", fmt.Sprintf("Unable to parse HRUI port settings HTML response, got error: %s", err))
		return
	}

	// Find the row for the given port ID
	portRow := doc.Find("table:last-of-type tbody tr").FilterFunction(func(i int, s *goquery.Selection) bool {
		return s.Find("td:nth-child(1)").Text() == fmt.Sprintf("Port %d", data.PortID.ValueInt64()+1) // Exact match
	})

	// Extract values
	enabled := portRow.Find("td:nth-child(2)").Text() == "Enable"
	speedDuplex := portRow.Find("td:nth-child(3)").Text()
	actualSpeedDuplex := portRow.Find("td:nth-child(4)").Text()
	flowControl := portRow.Find("td:nth-child(5)").Text()
	actualFlowControl := portRow.Find("td:nth-child(6)").Text()

	// Set state
	// Set state
	data.Enabled = types.BoolValue(enabled)
	data.Speed = &portSettingSpeed{ // Assign to a pointer to the struct
		Config: types.StringValue(speedDuplex),
		Actual: types.StringValue(actualSpeedDuplex),
	}
	data.FlowControl = &portSettingFlowControl{ // Assign to a pointer to the struct
		Config: types.StringValue(flowControl),
		Actual: types.StringValue(actualFlowControl),
	}

	data.ID = types.StringValue(strconv.FormatInt(data.PortID.ValueInt64(), 10))

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
