package sdk

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetIPAddressSettings(t *testing.T) {
	tests := []struct {
		name       string
		html       string
		expected   *IPAddressSettings
		shouldFail bool
		httpStatus int
	}{
		{
			"Valid IP settings",
			`<select name='dhcp_state'>
				<option value='0' selected>Static</option>
				<option value='1'>DHCP</option>
			</select>
			<input name='ip' value='192.168.1.100'>
			<input name='netmask' value='255.255.255.0'>
			<input name='gateway' value='192.168.1.1'>`,
			&IPAddressSettings{false, "192.168.1.100", "255.255.255.0", "192.168.1.1"},
			false,
			http.StatusOK,
		},
		{
			"DHCP enabled",
			`<select name='dhcp_state'>
				<option value='0'>Static</option>
				<option value='1' selected>DHCP</option>
			</select>
			<input name='ip' value=''>
			<input name='netmask' value=''>
			<input name='gateway' value=''>`,
			&IPAddressSettings{true, "", "", ""},
			false,
			http.StatusOK,
		},
		{
			"Request error",
			"",
			nil,
			true,
			http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.httpStatus)
				if _, err := w.Write([]byte(tt.html)); err != nil {
					t.Errorf("failed to write response: %v", err)
				}
			}))
			defer server.Close()

			client := &HRUIClient{URL: server.URL, HttpClient: server.Client()}
			settings, err := client.GetIPAddressSettings()

			if tt.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, settings)
			}
		})
	}
}

func TestSetIPAddressSettings(t *testing.T) {
	tests := []struct {
		name       string
		settings   *IPAddressSettings
		httpStatus int
		shouldFail bool
	}{
		{
			"Successful update",
			&IPAddressSettings{false, "192.168.1.200", "255.255.255.0", "192.168.1.1"},
			http.StatusOK,
			false,
		},
		{
			"DHCP enabled",
			&IPAddressSettings{true, "192.168.1.200", "255.255.255.0", "192.168.1.1"},
			http.StatusOK,
			false,
		},
		{
			"Request error",
			&IPAddressSettings{false, "192.168.1.200", "255.255.255.0", "192.168.1.1"},
			http.StatusInternalServerError,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedForm url.Values

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				err := r.ParseForm()
				if err != nil {
					t.Errorf("failed to parse form data: %v", err)
				}
				receivedForm = r.Form

				w.WriteHeader(tt.httpStatus)
			}))
			defer server.Close()

			client := &HRUIClient{URL: server.URL, HttpClient: server.Client()}
			err := client.SetIPAddressSettings(tt.settings)

			if tt.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Check that correct form values were sent
				assert.Equal(t, strconv.Itoa(btoi(tt.settings.DHCPEnabled)), receivedForm.Get("dhcp_state"))
				assert.Equal(t, tt.settings.IPAddress, receivedForm.Get("ip"))
				assert.Equal(t, tt.settings.Netmask, receivedForm.Get("netmask"))
				assert.Equal(t, tt.settings.Gateway, receivedForm.Get("gateway"))
			}
		})
	}
}

// Helper function to convert bool to int (1 for true, 0 for false).
func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}
