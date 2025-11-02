package sdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var mockPortResponse = `
<html>
<head>
<title>Port Setting</title>
<link rel="stylesheet" type="text/css" href="/style.css">
<script type="text/javascript"></script>
</head>
<body>
<center>

<fieldset>
<legend>Port Setting</legend>

<table border="1"> </table>
<table border="1"> </table>
<br>
<table border="1">
  <tr>
    <th rowspan="2" width="90">Port</th>
    <th rowspan="2" width="90">State</th>
    <th colspan="2">Speed/Duplex</th>
    <th colspan="2">Flow Control</th>
  </tr>
  <tr>
    <th width="90">Config</th>
    <th width="90">Actual</th>
    <th width="90">Config</th>
    <th width="90">Actual</th>
  </tr>
  <tr>
    <td>Port 1</td>
    <td>Enable</td>
    <td>Auto</td>
    <td>1000Full</td>
    <td>Off</td>
    <td>Off</td>
  </tr>
  <tr>
    <td>Port 2</td>
    <td>Enable</td>
    <td>Auto</td>
    <td>Link Down</td>
    <td>Off</td>
    <td>Off</td>
  </tr>
  <tr>
    <td>Port 3</td>
    <td>Enable</td>
    <td>Auto</td>
    <td>Link Down</td>
    <td>Off</td>
    <td>Off</td>
  </tr>
  <tr>
    <td>Port 4</td>
    <td>Disable</td>
    <td>Auto</td>
    <td>Link Down</td>
    <td>Off</td>
    <td>Off</td>
  </tr>
  <tr>
    <td>Trunk1</td>
    <td>Enable</td>
    <td>Auto</td>
    <td>Link Down</td>
    <td>Off</td>
    <td>Off</td>
  </tr>
</table>
<br>

</fieldset>
</center>
</body>
</html>
`

func TestListPorts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockPortResponse))
	}))
	defer server.Close()

	client := &HRUIClient{
		URL:        server.URL,
		HttpClient: server.Client(),
	}

	ports, err := client.ListPorts(context.Background())
	require.NoError(t, err)
	require.Len(t, ports, 5)

	// Updated expected ports
	expectedPorts := []*Port{
		{
			ID:                "Port 1",
			IsTrunk:           false,
			State:             1,
			SpeedDuplexConfig: "Auto",
			SpeedDuplexActual: "1000Full",
			FlowControlConfig: "Off",
			FlowControlActual: "Off",
		},
		{
			ID:                "Port 2",
			IsTrunk:           false,
			State:             1,
			SpeedDuplexConfig: "Auto",
			SpeedDuplexActual: "Link Down",
			FlowControlConfig: "Off",
			FlowControlActual: "Off",
		},
		{
			ID:                "Port 3",
			IsTrunk:           false,
			State:             1,
			SpeedDuplexConfig: "Auto",
			SpeedDuplexActual: "Link Down",
			FlowControlConfig: "Off",
			FlowControlActual: "Off",
		},
		{
			ID:                "Port 4",
			IsTrunk:           false,
			State:             0,
			SpeedDuplexConfig: "Auto",
			SpeedDuplexActual: "Link Down",
			FlowControlConfig: "Off",
			FlowControlActual: "Off",
		},
		{
			ID:                "Trunk1",
			IsTrunk:           true,
			State:             1,
			SpeedDuplexConfig: "Auto",
			SpeedDuplexActual: "Link Down",
			FlowControlConfig: "Off",
			FlowControlActual: "Off",
		},
	}

	assert.Equal(t, expectedPorts, ports)
}

func TestGetPort(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockPortResponse))
	}))
	defer server.Close()

	client := &HRUIClient{
		URL:        server.URL,
		HttpClient: server.Client(),
	}

	// Test for a physical port
	port, err := client.GetPort(context.Background(), "Port 1")
	require.NoError(t, err)

	expectedPort := &Port{
		ID:                "Port 1",
		IsTrunk:           false,
		State:             1,
		SpeedDuplexConfig: "Auto",
		SpeedDuplexActual: "1000Full",
		FlowControlConfig: "Off",
		FlowControlActual: "Off",
	}

	assert.Equal(t, expectedPort, port)

	// Test for a trunk port
	port, err = client.GetPort(context.Background(), "Trunk1")
	require.NoError(t, err)

	expectedPort = &Port{
		ID:                "Trunk1",
		IsTrunk:           true,
		State:             1,
		SpeedDuplexConfig: "Auto",
		SpeedDuplexActual: "Link Down",
		FlowControlConfig: "Off",
		FlowControlActual: "Off",
	}

	assert.Equal(t, expectedPort, port)
}

func TestGetPortStatistics(t *testing.T) {
	mockResponse := `
	<html>
		<head>
			<title>Port Statistics</title>
			<link rel="stylesheet" type="text/css" href="/style.css">
		</head>
		<body>
			<center>
				<fieldset>
					<legend>Port Statistics</legend>
					<table>
						<tr>
							<th>Port</th>
							<th>State</th>
							<th>Link Status</th>
							<th>TxGoodPkt</th>
							<th>TxBadPkt</th>
							<th>RxGoodPkt</th>
							<th>RxBadPkt</th>
						</tr>
						<tr>
							<td>Port 1</td>
							<td>Enable</td>
							<td>Link Up</td>
							<td>1539909</td>
							<td>0</td>
							<td>6063265</td>
							<td>0</td>
						</tr>
						<tr>
							<td>Trunk1</td>
							<td>Enable</td>
							<td>Link Down</td>
							<td>173978</td>
							<td>0</td>
							<td>2140448</td>
							<td>0</td>
						</tr>
					</table>
				</fieldset>
			</center>
		</body>
	</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	client := &HRUIClient{
		URL:        server.URL,
		HttpClient: server.Client(),
	}

	stats, err := client.GetPortStatistics(context.Background())
	require.NoError(t, err)

	expectedStats := []*PortStatistics{
		{Port: "Port 1", State: 1, LinkStatus: "Link Up", TxGoodPkt: 1539909, TxBadPkt: 0, RxGoodPkt: 6063265, RxBadPkt: 0},
		{Port: "Trunk1", State: 1, LinkStatus: "Link Down", TxGoodPkt: 173978, TxBadPkt: 0, RxGoodPkt: 2140448, RxBadPkt: 0},
	}

	assert.Equal(t, expectedStats, stats)
}

func TestPortMirroring(t *testing.T) {
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
                <option value="7">Trunk2</option>
            </select>
        </form>
        </body>
        </html>
    `

	portMirrorHTMLResponse := `
	<html>
	<body>
	    <center>
	        <fieldset>
	            <legend>Port Mirroring Setting</legend>
	            <form method="post" action="/port.cgi?page=delete_mirror">
	                <table border="1">
	                    <tr>
	                        <th align="center" width="120">Mirror Direction</th>
	                        <th align="center" width="120">Mirroring Port</th>
	                        <th align="center" width="200">Mirrored Port List</th>
	                    </tr>
	                    <tr>
	                        <td align="center">Rx</td>
	                        <td align="center">1</td>
	                        <td align="center">Trunk2</td>
	                    </tr>
	                </table>
	            </form>
	        </fieldset>
	    </center>
	</body>
	</html>
	`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/port.cgi" && r.URL.RawQuery == "" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(portHTMLResponse))
		} else if r.URL.Path == "/port.cgi" && r.URL.RawQuery == "page=mirroring" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(portMirrorHTMLResponse))
		} else if r.URL.Path == "/port.cgi" && r.URL.RawQuery == "page=delete_mirror" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<html><body>Mirror deleted.</body></html>`))
		} else if r.URL.Path == "/login.cgi" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<html><body>Login successful</body></html>`))
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewClient(context.Background(), server.URL, "testuser", "testpassword", false, nil)
	require.NoError(t, err)
	require.NotNil(t, client)

	portMirror, err := client.GetPortMirror(context.Background())
	require.NoError(t, err)
	expectedPortMirror := &PortMirror{
		MirrorDirection: "Rx",
		MirroringPort:   "Port 1",
		MirroredPort:    "Trunk2",
	}
	assert.Equal(t, expectedPortMirror, portMirror)

	newPortMirror := &PortMirror{
		MirrorDirection: "Tx",
		MirroringPort:   "Port 2",
		MirroredPort:    "Port 3",
	}
	err = client.ConfigurePortMirror(context.Background(), newPortMirror)
	require.NoError(t, err)

	err = client.DeletePortMirror(context.Background())
	require.NoError(t, err)
}

func TestGetPortIsolation(t *testing.T) {
	// Mock HTML response for port isolation page
	mockPortIsolationResponse := `
<html>
<head>
<title>Port Isolation</title>
</head>
<body>
<center>
<fieldset>
<legend>Port Isolation Setting</legend>
<table border="1">
    <tr>
        <th align="center" width="80">Port</th>
        <th align="center" width="200">Port Isolation List</th>
    </tr>
    <tr>
        <td align="center">Port 1</td>
        <td align="center">Port 2,Port 3,Trunk1</td>
    </tr>
    <tr>
        <td align="center">Trunk1</td>
        <td align="center">Port 1,Port 4</td>
    </tr>
    <tr>
        <td align="center">Port 4</td>
        <td align="center"></td>
    </tr>
</table>
</fieldset>
</center>
</body>
</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Respond with mock HTML when port isolation page is requested
		if r.URL.Path == "/port.cgi" && r.URL.Query().Get("page") == "isolation" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(mockPortIsolationResponse))
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Create new client
	client := &HRUIClient{
		URL:        server.URL,
		HttpClient: server.Client(),
	}

	// Call GetPortIsolation
	isolations, err := client.GetPortIsolation(context.Background())
	require.NoError(t, err)

	// Assert that the number of isolations matches the number of rows in the mock HTML
	require.Len(t, isolations, 3)

	// Define the expected result
	expectedIsolationConfig := []PortIsolation{
		{
			Port:          "Port 1",
			IsolationList: []string{"Port 2", "Port 3", "Trunk1"},
		},
		{
			Port:          "Trunk1",
			IsolationList: []string{"Port 1", "Port 4"},
		},
		{
			Port:          "Port 4",
			IsolationList: []string{}, // Empty isolation list
		},
	}

	// Assert that the parsed result matches the expected configuration
	assert.Equal(t, expectedIsolationConfig, isolations)
}
