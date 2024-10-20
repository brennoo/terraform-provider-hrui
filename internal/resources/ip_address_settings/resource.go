package ip_address_settings

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/brennoo/terraform-provider-hrui/internal/client"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource                = &ipAddressResource{}
	_ resource.ResourceWithConfigure   = &ipAddressResource{}
	_ resource.ResourceWithImportState = &ipAddressResource{}
)

// ipAddressResource is the resource implementation.
type ipAddressResource struct {
	client *client.Client
}

// NewResourceIPAddressSetting is a helper function to simplify the provider implementation.
func NewResourceIPAddressSetting() resource.Resource {
	return &ipAddressResource{}
}

// Configure adds the provider configured client to the resource.
func (r *ipAddressResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ipAddressResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_address_settings"
}

func (r *ipAddressResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ipAddressModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	form := url.Values{}
	form.Set("cmd", "ip")

	dhcpEnabled := data.DHCPEnabled.ValueBool()

	if dhcpEnabled {
		form.Set("dhcp_state", "1")
		// You don't need to set ip, netmask, and gateway when DHCP is enabled
	} else {
		form.Set("dhcp_state", "0")
		form.Set("ip", data.IPAddress.ValueString())
		form.Set("netmask", data.Netmask.ValueString())
		form.Set("gateway", data.Gateway.ValueString())
	}

	// Make POST request to ip.cgi
	url := fmt.Sprintf("%s/ip.cgi", r.client.URL)
	httpReq, err := http.NewRequest(http.MethodPost, url, strings.NewReader(form.Encode()))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create HRUI IP address settings, got error: %s", err))
		return
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpResp, err := r.client.HttpClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create HRUI IP address settings, got error: %s", err))
		return
	}

	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Unexpected HTTP Status Code", fmt.Sprintf("Unable to create HRUI IP address settings, got HTTP status code: %d", httpResp.StatusCode))
		return
	}
	// If DHCP is enabled, wait for IP address settings to become available
	if data.DHCPEnabled.ValueBool() {
		for i := 0; i < 5; i++ {
			// Make a GET request to ip.cgi to fetch the updated settings
			url := fmt.Sprintf("%s/ip.cgi", r.client.URL)
			httpResp, err := r.client.MakeRequest(url)
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read HRUI IP address settings, got error: %s", err))
				return
			}
			defer httpResp.Body.Close()

			if httpResp.StatusCode != http.StatusOK {
				resp.Diagnostics.AddError("Unexpected HTTP Status Code", fmt.Sprintf("Unable to read HRUI IP address settings, got HTTP status code: %d", httpResp.StatusCode))
				return
			}

			// Parse the HTML response
			doc, err := goquery.NewDocumentFromReader(httpResp.Body)
			if err != nil {
				resp.Diagnostics.AddError("HTML Parsing Error", fmt.Sprintf("Unable to parse HRUI IP address settings HTML response, got error: %s", err))
				return
			}

			ipAddress := doc.Find("input[name='ip']").AttrOr("value", "")
			netmask := doc.Find("input[name='netmask']").AttrOr("value", "")
			gateway := doc.Find("input[name='gateway']").AttrOr("value", "")

			if ipAddress != "" && netmask != "" && gateway != "" {
				data.IPAddress = types.StringValue(ipAddress)
				data.Netmask = types.StringValue(netmask)
				data.Gateway = types.StringValue(gateway)
				break
			}

			time.Sleep(2 * time.Second)
		}
	}

	// Save the configuration if autosave is enabled
	if r.client.Autosave {
		if err := r.client.SaveConfiguration(); err != nil {
			resp.Diagnostics.AddError("Save Error", fmt.Sprintf("Unable to save HRUI configuration, got error: %s", err))
			return
		}
	}

	// Set the ID
	id := "ip_address_settings"
	data.ID = types.StringValue(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *ipAddressResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ipAddressModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/ip.cgi", r.client.URL)
	httpResp, err := r.client.MakeRequest(url)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read HRUI IP address settings, got error: %s", err))
		return
	}

	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Unexpected HTTP Status Code", fmt.Sprintf("Unable to read HRUI IP address settings, got HTTP status code: %d", httpResp.StatusCode))
		return
	}

	// Parse HTML with goquery
	doc, err := goquery.NewDocumentFromReader(httpResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("HTML Parsing Error", fmt.Sprintf("Unable to parse HRUI IP address settings HTML response, got error: %s", err))
		return
	}

	// Extract DHCP setting
	dhcpEnabled := false
	doc.Find("select[name='dhcp_state'] option").Each(func(i int, s *goquery.Selection) {
		if _, ok := s.Attr("selected"); ok {
			dhcpEnabledStr, _ := s.Attr("value")
			dhcpEnabled, _ = strconv.ParseBool(dhcpEnabledStr)
		}
	})

	data.DHCPEnabled = types.BoolValue(dhcpEnabled)

	// Extract IP address, netmask, and gateway
	ipAddress := doc.Find("input[name='ip']").AttrOr("value", "")
	netmask := doc.Find("input[name='netmask']").AttrOr("value", "")
	gateway := doc.Find("input[name='gateway']").AttrOr("value", "")

	// Set state for IP address, netmask, and gateway
	data.IPAddress = types.StringValue(ipAddress)
	data.Netmask = types.StringValue(netmask)
	data.Gateway = types.StringValue(gateway)

	// Set the ID
	id := "ip_address_settings"
	data.ID = types.StringValue(id)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *ipAddressResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ipAddressModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	form := url.Values{}
	form.Set("cmd", "ip")

	// Check if the resource is being "deleted" (dhcp_enabled set to true)
	if data.DHCPEnabled.ValueBool() {
		// If dhcp_enabled is true, it means the resource is being "deleted"
		// So, you only need to set dhcp_state to 1
		form.Set("dhcp_state", "1")
	} else {
		// Otherwise, set the static IP configuration
		form.Set("dhcp_state", "0")
		form.Set("ip", data.IPAddress.ValueString())
		form.Set("netmask", data.Netmask.ValueString())
		form.Set("gateway", data.Gateway.ValueString())
	}

	// Make POST request to ip.cgi
	url := fmt.Sprintf("%s/ip.cgi", r.client.URL)
	httpReq, err := http.NewRequest(http.MethodPost, url, strings.NewReader(form.Encode()))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update HRUI IP address settings, got error: %s", err))
		return
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpResp, err := r.client.HttpClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update HRUI IP address settings, got error: %s", err))
		return
	}

	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Unexpected HTTP Status Code", fmt.Sprintf("Unable to update HRUI IP address settings, got HTTP status code: %d", httpResp.StatusCode))
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

func (r *ipAddressResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// No-op
}

func (r *ipAddressResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("port_id"), req.ID)...)
}
