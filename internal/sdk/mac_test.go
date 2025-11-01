package sdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMACAddressTable(t *testing.T) {
	// Mock HTML response
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
			<td>2</td>
		</tr>
		<tr>
			<td>3</td>
			<td>77:88:99:AA:BB:CC</td>
			<td>1</td>
			<td>dynamic</td>
			<td>Trunk1</td>
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

	mock := mockServerMock(htmlResponse, http.StatusOK)
	defer mock.Close()

	// Create an HRUIClient pointing to the mock server
	client := &HRUIClient{
		URL:        mock.URL,
		HttpClient: &http.Client{},
	}

	macTable, err := client.GetMACAddressTable(context.Background())
	assert.NoError(t, err)

	// Validate results with anonymized expected MAC addresses
	expected := []MACAddressEntry{
		{ID: 1, MAC: "AA:BB:CC:DD:EE:FF", VLANID: 1, Type: "dynamic", Port: "Port 1"},
		{ID: 2, MAC: "11:22:33:44:55:66", VLANID: 1, Type: "dynamic", Port: "Port 2"},
		{ID: 3, MAC: "77:88:99:AA:BB:CC", VLANID: 1, Type: "dynamic", Port: "Trunk1"},
		{ID: 4, MAC: "00:11:22:33:44:55", VLANID: 1, Type: "static", Port: "Port 6"},
	}
	assert.Equal(t, expected, macTable)
}

func TestGetStaticMACAddressTable(t *testing.T) {
	// Mock server HTML response
	htmlResponse := `
	<html>
		<head><title>Static MAC Addresses</title></head>
		<body>
			<form action="/mac.cgi?page=staticdel">
				<table>
					<tr>
						<th>No.</th>
						<th>MAC Address</th>
						<th>VLAN ID</th>
						<th>Port</th>
						<th>Select</th>
					</tr>
					<tr>
						<td>1</td>
						<td>A8:80:55:59:E9:72</td>
						<td>1</td>
						<td>6</td>
						<td><input type="checkbox" name="del" value="A8:80:55:59:E9:72_1"></td>
					</tr>
					<tr>
						<td>2</td>
						<td>BB:44:55:66:77:88</td>
						<td>2</td>
						<td>Trunk2</td>
						<td><input type="checkbox" name="del" value="BB:44:55:66:77:88_2"></td>
					</tr>
				</table>
			</form>
		</body>
	</html>`

	mock := mockServerMock(htmlResponse, http.StatusOK)
	defer mock.Close()

	// Create client pointing to the mock server
	client := &HRUIClient{
		URL:        mock.URL,
		HttpClient: &http.Client{},
	}

	// Test the `GetStaticMACAddressTable` method
	entries, err := client.GetStaticMACAddressTable(context.Background())
	assert.NoError(t, err)
	assert.Len(t, entries, 2)

	// Validate the parsed entries
	expected := []StaticMACEntry{
		{ID: 1, MACAddress: "A8:80:55:59:E9:72", VLANID: 1, Port: "Port 6"},
		{ID: 2, MACAddress: "BB:44:55:66:77:88", VLANID: 2, Port: "Trunk2"},
	}
	assert.Equal(t, expected, entries)
}

func TestAddStaticMACEntry(t *testing.T) {
	// Mock HTML response for port.cgi
	portHTMLResponse := `
    <html>
    <body>
        <form action="/port.cgi" method="get">
            <select name="portid">
                <option value="1">Port 1</option>
                <option value="2">Port 2</option>
                <option value="3">Port 3</option>
                <option value="4">Port 4</option>
                <option value="5">Port 5</option>
                <option value="6">Port 6</option>
            </select>
        </form>
    </body>
    </html>
    `

	// Use a conditional handler as it also calls GetPortByName
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/port.cgi" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(portHTMLResponse))
			return
		}

		if r.URL.Path == "/mac.cgi" && r.Method == http.MethodPost {
			// Handle POST request for adding the static MAC entry
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
			return
		}

		http.NotFound(w, r)
	}))
	defer mock.Close()

	// Create client pointing to the mock server
	client := &HRUIClient{
		URL:        mock.URL,
		HttpClient: &http.Client{},
	}

	// Test the `AddStaticMACEntry` method
	err := client.AddStaticMACEntry(context.Background(), "01:23:45:67:89:AB", 10, "Port 1")
	assert.NoError(t, err, "AddStaticMACEntry should succeed for a valid port name")

	// Test with another port name
	err = client.AddStaticMACEntry(context.Background(), "01:23:45:67:89:AC", 20, "Port 6")
	assert.NoError(t, err, "AddStaticMACEntry should succeed for a valid port name")
}

func TestRemoveStaticMACEntries(t *testing.T) {
	// Use the mock server to simulate POST response
	mock := mockServerMock("OK", http.StatusOK)
	defer mock.Close()

	// Create client pointing to the mock server
	client := &HRUIClient{
		URL:        mock.URL,
		HttpClient: &http.Client{},
	}

	// Test `RemoveStaticMACEntries` with two entries
	entries := []StaticMACEntry{
		{MACAddress: "01:23:45:67:89:AB", VLANID: 10, Port: "Port 1"},
		{MACAddress: "02:33:44:55:66:77", VLANID: 20, Port: "Trunk3"},
	}

	err := client.RemoveStaticMACEntries(context.Background(), entries)
	assert.NoError(t, err)
}
