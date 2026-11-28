package sdk

import (
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"
)

const sampleDHCPSnoopingV1HTML = `<html><body>
<form method="post" name="dhcp" action="/dhcp_snooping.cgi?page=enable_dhcpsnooping">
  <table><tr><th>DHCP Snooping Global Enable</th><td><input type="checkbox" name="enable_dhcpsnooping" checked></td></tr></table>
</form>
<form method="post" action="/dhcp_snooping.cgi?page=static" name="formDhcpsnoopingStatic">
  <table>
    <tr><th>Port</th><th>1</th><th>2</th><th>3</th></tr>
    <tr>
      <td>DHCP Server</td>
      <td><input type="radio" name="vlanPort_0" value="1" checked></td>
      <td><input type="radio" name="vlanPort_1" value="1"></td>
      <td><input type="radio" name="vlanPort_2" value="1" checked></td>
    </tr>
    <tr>
      <td>DHCP Clinet</td>
      <td><input type="radio" name="vlanPort_0" value="0"></td>
      <td><input type="radio" name="vlanPort_1" value="0" checked></td>
      <td><input type="radio" name="vlanPort_2" value="0"></td>
    </tr>
  </table>
</form>
</body></html>`

const sampleDHCPSnoopingV2HTML = `<html><body>
<form method="post" id="dhcpsnpfrm" action="/dhcp_snooping.cgi?page=enable_dhcpsnooping">
  <input name="enable_dhcpsnp" type="checkbox" class="check" id="toggle">
</form>
<form method="post" action="/dhcp_snooping.cgi?page=static">
  <input type="checkbox" id="port1" name="port1" class="chkp">
  <input type="checkbox" id="port2" name="port2" class="chkp" checked>
  <input type="hidden" name="cmd" value="dhcpsnpfrm">
</form>
</body></html>`

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestDHCPSnoopingV1AndV2(t *testing.T) {
	ctx := t.Context()

	// Test v1.x radio button dialect
	var lastV1Post string
	v1Client := &HRUIClient{
		URL: "http://switch.local",
		HttpClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.Method == http.MethodPost {
					b, _ := io.ReadAll(req.Body)
					lastV1Post = string(b)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(sampleDHCPSnoopingV1HTML)),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}

	cfgV1, err := v1Client.GetDHCPSnooping(ctx)
	if err != nil {
		t.Fatalf("GetDHCPSnooping(v1) failed: %v", err)
	}
	if !cfgV1.Enabled || !slices.Equal(cfgV1.TrustedPorts, []string{"Port 1", "Port 3"}) {
		t.Fatalf("unexpected v1 config: %+v", cfgV1)
	}

	if err := v1Client.SetDHCPSnoopingPorts(ctx, []string{"Port 2"}); err != nil {
		t.Fatalf("SetDHCPSnoopingPorts(v1) failed: %v", err)
	}
	if !strings.Contains(lastV1Post, "vlanPort_0=0") || !strings.Contains(lastV1Post, "vlanPort_1=1") || !strings.Contains(lastV1Post, "vlanPort_2=0") {
		t.Fatalf("unexpected v1 POST body: %q", lastV1Post)
	}

	// Test v2.x checkbox dialect
	var lastV2Post string
	v2Client := &HRUIClient{
		URL: "http://switch.local",
		HttpClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.Method == http.MethodPost {
					b, _ := io.ReadAll(req.Body)
					lastV2Post = string(b)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(sampleDHCPSnoopingV2HTML)),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}

	cfgV2, err := v2Client.GetDHCPSnooping(ctx)
	if err != nil {
		t.Fatalf("GetDHCPSnooping(v2) failed: %v", err)
	}
	if cfgV2.Enabled || !slices.Equal(cfgV2.TrustedPorts, []string{"Port 2"}) {
		t.Fatalf("unexpected v2 config: %+v", cfgV2)
	}

	if err := v2Client.SetDHCPSnoopingGlobal(ctx, true); err != nil {
		t.Fatalf("SetDHCPSnoopingGlobal(v2) failed: %v", err)
	}
	if lastV2Post != "enable_dhcpsnp=on" {
		t.Fatalf("expected enable_dhcpsnp=on, got %q", lastV2Post)
	}

	if err := v2Client.SetDHCPSnoopingPorts(ctx, []string{"Port 1"}); err != nil {
		t.Fatalf("SetDHCPSnoopingPorts(v2) failed: %v", err)
	}
	if !strings.Contains(lastV2Post, "cmd=dhcpsnpfrm") || !strings.Contains(lastV2Post, "port1=on") || strings.Contains(lastV2Post, "port2=on") {
		t.Fatalf("unexpected v2 POST body: %q", lastV2Post)
	}
}
