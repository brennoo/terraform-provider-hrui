package port_settings

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/brennoo/terraform-provider-hrui/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource                = &portSettingResource{}
	_ resource.ResourceWithConfigure   = &portSettingResource{}
	_ resource.ResourceWithImportState = &portSettingResource{}
)

// portSettingResource is the resource implementation.
type portSettingResource struct {
	client *client.Client
}

// NewResourcePortSetting is a helper function to simplify the provider implementation.
func NewResourcePortSetting() resource.Resource {
	return &portSettingResource{}
}

// Configure adds the provider configured client to the resource.
func (r *portSettingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *portSettingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_port_settings"
}

func (r *portSettingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data portSettingModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	form := url.Values{}
	form.Set("cmd", "port")

	form.Set("portid", strconv.FormatInt(data.PortID.ValueInt64(), 10))
	if data.Enabled.ValueBool() {
		form.Set("state", "1")
	} else {
		form.Set("state", "0")
	}
	form.Set(speedDuplexField, data.Speed.Config.ValueString())
	form.Set(flowControlField, data.FlowControl.Config.ValueString())

	// Make POST request to ip.cgi
	url := fmt.Sprintf("%s/port.cgi", r.client.URL)
	httpReq, err := http.NewRequest(http.MethodPost, url, strings.NewReader(form.Encode()))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create HRUI port settings, got error: %s", err))
		return
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpResp, err := r.client.HttpClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create HRUI port settings, got error: %s", err))
		return
	}

	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Unexpected HTTP Status Code", fmt.Sprintf("Unable to create HRUI port settings, got HTTP status code: %d", httpResp.StatusCode))
		return
	}

	// Save the configuration if autosave is enabled
	if r.client.Autosave {
		if err := r.client.SaveConfiguration(); err != nil {
			resp.Diagnostics.AddError("Save Error", fmt.Sprintf("Unable to save HRUI configuration, got error: %s", err))
			return
		}
	}

	// Set the ID
	id := strconv.FormatInt(data.PortID.ValueInt64(), 10)
	data.ID = types.StringValue(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *portSettingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state portSettingModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	portID := state.PortID.ValueInt64()

	url := fmt.Sprintf("%s/port.cgi", r.client.URL)
	httpResp, err := r.client.MakeRequest(url)
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
		return s.Find("td:nth-child(1)").Text() == fmt.Sprintf("Port %d", portID+1) // Exact match
	})

	// Extract values
	enabled := portRow.Find("td:nth-child(2)").Text() == "Enable"
	speedDuplex := portRow.Find("td:nth-child(3)").Text()
	actualSpeedDuplex := portRow.Find("td:nth-child(4)").Text()
	flowControl := portRow.Find("td:nth-child(5)").Text()
	actualFlowControl := portRow.Find("td:nth-child(6)").Text()

	state.ID = types.StringValue(strconv.FormatInt(portID, 10))

	// Set state
	state.Enabled = types.BoolValue(enabled)
	state.Speed.Config = types.StringValue(speedDuplex)
	state.Speed.Actual = types.StringValue(actualSpeedDuplex)
	state.FlowControl.Config = types.StringValue(flowControl)
	state.FlowControl.Actual = types.StringValue(actualFlowControl)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *portSettingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data portSettingModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	form := url.Values{}
	form.Set("cmd", "port")

	form.Set("portid", strconv.FormatInt(data.PortID.ValueInt64(), 10))
	if data.Enabled.ValueBool() {
		form.Set("state", "1")
	} else {
		form.Set("state", "0")
	}
	form.Set(speedDuplexField, data.Speed.Config.ValueString())
	form.Set(flowControlField, data.FlowControl.Config.ValueString())

	// Make POST request to ip.cgi
	url := fmt.Sprintf("%s/port.cgi", r.client.URL)
	httpReq, err := http.NewRequest(http.MethodPost, url, strings.NewReader(form.Encode()))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update HRUI port settings, got error: %s", err))
		return
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpResp, err := r.client.HttpClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update HRUI port settings, got error: %s", err))
		return
	}

	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Unexpected HTTP Status Code", fmt.Sprintf("Unable to update HRUI port settings, got HTTP status code: %d", httpResp.StatusCode))
		return
	}

	// Save the configuration if autosave is enabled
	if r.client.Autosave {
		if err := r.client.SaveConfiguration(); err != nil {
			resp.Diagnostics.AddError("Save Error", fmt.Sprintf("Unable to save HRUI configuration, got error: %s", err))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *portSettingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// No-op
}

func (r *portSettingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Since the API does not return an ID, we will always
	// import the ID as is, without lookup.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("port_id"), req.ID)...)
}
