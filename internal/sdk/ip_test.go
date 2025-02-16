package sdk

import (
	"net/http"
	"net/http/httptest"
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
