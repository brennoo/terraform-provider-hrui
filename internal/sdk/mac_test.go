package sdk_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
)

func TestGetMACAddressTable(t *testing.T) {
	// Mock HTML response with anonymized MAC addresses
	htmlResponse := `
	<html>
	<body>
	<table>
		<tr>
			<th>No.</th>
			<th>MAC Address</th>
			<th>VLAN ID</th>
			<th>Type</th>
			<th>Port</th>
		</tr>
		<tr>
			<td>1</td>
			<td>AA:BB:CC:DD:EE:FF</td>
			<td>1</td>
			<td>dynamic</td>
			<td>1</td>
		</tr>
		<tr>
			<td>2</td>
			<td>11:22:33:44:55:66</td>
			<td>1</td>
			<td>dynamic</td>
			<td>1</td>
		</tr>
		<tr>
			<td>3</td>
			<td>77:88:99:AA:BB:CC</td>
			<td>1</td>
			<td>dynamic</td>
			<td>1</td>
		</tr>
		<tr>
			<td>4</td>
			<td>00:11:22:33:44:55</td>
			<td>1</td>
			<td>static</td>
			<td>6</td>
		</tr>
	</table>
	</body>
	</html>`

	// Set up mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(htmlResponse))
	}))
	defer server.Close()

	// Create client and call GetMACAddressTable
	client := &sdk.HRUIClient{
		URL:        server.URL,
		HttpClient: &http.Client{},
	}

	macTable, err := client.GetMACAddressTable()
	assert.NoError(t, err)

	// Validate results with anonymized expected MAC addresses
	expected := []sdk.MACAddressEntry{
		{ID: 1, MAC: "AA:BB:CC:DD:EE:FF", VLANID: 1, Type: "dynamic", Port: 1},
		{ID: 2, MAC: "11:22:33:44:55:66", VLANID: 1, Type: "dynamic", Port: 1},
		{ID: 3, MAC: "77:88:99:AA:BB:CC", VLANID: 1, Type: "dynamic", Port: 1},
		{ID: 4, MAC: "00:11:22:33:44:55", VLANID: 1, Type: "static", Port: 6},
	}
	assert.Equal(t, expected, macTable)
}
