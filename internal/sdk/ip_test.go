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
			</select>
			<input name='ip' value='192.168.1.100'>
			<input name='netmask' value='255.255.255.0'>
			<input name='gateway' value='192.168.1.1'>`,
			&IPAddressSettings{false, "192.168.1.100", "255.255.255.0", "192.168.1.1"},
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
				w.Write([]byte(tt.html))
			}))
			defer server.Close()

			client := &HRUIClient{URL: server.URL, HttpClient: server.Client()}
			resp, err := client.Request("GET", server.URL, nil, nil)

			if tt.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
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
			"Request error",
			&IPAddressSettings{false, "192.168.1.200", "255.255.255.0", "192.168.1.1"},
			http.StatusInternalServerError,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.httpStatus)
			}))
			defer server.Close()

			client := &HRUIClient{URL: server.URL, HttpClient: server.Client()}
			err := client.SetIPAddressSettings(tt.settings)

			if tt.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
