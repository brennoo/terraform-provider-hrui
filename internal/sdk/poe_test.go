package sdk

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

const samplePoESystemHTML = `<html><body>
<center><fieldset><legend>PoE Port Configuration</legend>
<form method="post" action="pse_system.cgi">
  <table border="1">
    <tr><th>Total PoE Power(W)</th><td><input type="text" name="pse_sys_pwr" value="100"></td></tr>
  </table>
  <table border="1">
    <tr><th>Port</th><th>PoE Admin Status</th><th>PoE Protocol Type</th><th>PoEDog</th></tr>
    <tr>
      <td><select name="portid" multiple size="8"><option value="0">Port 1</option><option value="1">Port 2</option></select></td>
      <td><select name="state"><option value="1">Enable</option><option value="0">Disable</option></select></td>
      <td><select name="pse_type"><option value="0">BT</option><option value="1">HiPoE</option><option value="2">AT</option><option value="3">AF</option></select></td>
      <td><select name="watchdog_state"><option value="0">Disable</option><option value="1">Enable</option></select></td>
    </tr>
  </table>
  <input type="hidden" name="cmd" value="poe">
</form>
<table border="1">
  <tr><th>Port</th><th>PoE Admin Status</th><th>PoE Protocol Type</th><th>PoEDog</th></tr>
  <tr><td>Port 1</td><td>Enable</td><td>BT</td><td>Disable</td></tr>
  <tr><td>Port 2</td><td>Disable</td><td>AF</td><td>Enable</td></tr>
</table>
</fieldset></center></body></html>`

const samplePoEPortHTML = `<html><body>
<center><fieldset><legend>PoE Port Status</legend>
<table>
  <tr><th>Consumption (W)</th><th>Power Usage (%)</th></tr>
  <tr>
    <td><input type="text" readonly name="pse_con_pwr" value="30.85"></td>
    <td><input type="text" readonly name="pse_con_pwr_percent" value="30"></td>
  </tr>
</table>
<table>
  <tr><th>Port</th><th>Admin Status</th><th>PSE Type</th><th>PoEDog</th><th>Operation</th><th>Class</th><th>Power(W)</th><th>Voltage(V)</th><th>Current(mA)</th><th>Reset</th></tr>
  <tr><td>Port 1</td><td>Enable</td><td>BT</td><td>Disable</td><td>On</td><td>AT(Class4)</td><td>13.30</td><td>53.20</td><td>250</td><td><input type="button" value="Reset" onclick="pse_port_submit(0)"></td></tr>
  <tr><td>Port 2</td><td>Disable</td><td>AF</td><td>Enable</td><td>Off</td><td>-</td><td>-</td><td>-</td><td>-</td><td><input type="button" value="Reset" onclick="pse_port_submit(1)"></td></tr>
</table>
</fieldset></center></body></html>`

func TestGetPoESystemAndPortStatuses(t *testing.T) {
	client := &HRUIClient{
		URL: "http://switch.local",
		HttpClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.Method == http.MethodPost && req.URL.Path == "/pse_system.cgi" {
					b, _ := io.ReadAll(req.Body)
					vals, err := url.ParseQuery(string(b))
					if err != nil {
						t.Fatalf("ParseQuery failed: %v", err)
					}
					if vals.Get("cmd") != "poe" {
						t.Errorf("expected cmd=poe, got %q", vals.Get("cmd"))
					}
				}
				body := samplePoESystemHTML
				if req.URL.Path == "/pse_port.cgi" {
					body = samplePoEPortHTML
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(body)),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}
	ctx := t.Context()

	sys, err := client.GetPoESystem(ctx)
	if err != nil {
		t.Fatalf("GetPoESystem failed: %v", err)
	}
	if sys.PowerLimitWatts != 100 || sys.PowerConsumptionWatts != 30.85 || sys.PowerUsagePercent != 30 {
		t.Fatalf("unexpected PoESystem: %+v", sys)
	}

	if err := client.SetPoESystemPowerLimit(ctx, 120); err != nil {
		t.Fatalf("SetPoESystemPowerLimit failed: %v", err)
	}

	if err := client.ConfigurePoEPort(ctx, PoEPortConfig{
		Port:          "Port 1",
		Enabled:       true,
		PSEType:       "AT",
		WatchdogState: true,
	}); err != nil {
		t.Fatalf("ConfigurePoEPort failed: %v", err)
	}

	p1, err := client.GetPoEPortStatus(ctx, "Port 1")
	if err != nil {
		t.Fatalf("GetPoEPortStatus(Port 1) failed: %v", err)
	}
	if !p1.Enabled || p1.PSEType != "BT" || p1.WatchdogState || p1.Operation != "On" || p1.Class != "AT(Class4)" || p1.PowerWatts != 13.30 || p1.VoltageVolts != 53.20 || p1.CurrentMA != 250 {
		t.Fatalf("unexpected Port 1 status: %+v", p1)
	}

	p2, err := client.GetPoEPortStatus(ctx, "Port 2")
	if err != nil {
		t.Fatalf("GetPoEPortStatus(Port 2) failed: %v", err)
	}
	if p2.Enabled || p2.PSEType != "AF" || !p2.WatchdogState || p2.Operation != "Off" || p2.PowerWatts != 0 {
		t.Fatalf("unexpected Port 2 status: %+v", p2)
	}

	if err := client.ResetPoEPort(ctx, "Port 1"); err != nil {
		t.Fatalf("ResetPoEPort failed: %v", err)
	}
}
