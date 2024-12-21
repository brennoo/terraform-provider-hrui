package sdk_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
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
    <td>Port 5</td>
    <td>Enable</td>
    <td>Auto</td>
    <td>Link Down</td>
    <td>Off</td>
    <td>Off</td>
  </tr>
  <tr>
    <td>Port 6</td>
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

func TestGetAllPorts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockPortResponse))
	}))
	defer server.Close()

	client := &sdk.HRUIClient{
		URL:        server.URL,
		HttpClient: server.Client(),
	}

	ports, err := client.GetAllPorts()
	require.NoError(t, err)
	require.Len(t, ports, 6)

	expectedPorts := []*sdk.Port{
		{ID: 1, State: 1, SpeedDuplex: "Auto", FlowControl: "Off"},
		{ID: 2, State: 1, SpeedDuplex: "Auto", FlowControl: "Off"},
		{ID: 3, State: 1, SpeedDuplex: "Auto", FlowControl: "Off"},
		{ID: 4, State: 0, SpeedDuplex: "Auto", FlowControl: "Off"},
		{ID: 5, State: 1, SpeedDuplex: "Auto", FlowControl: "Off"},
		{ID: 6, State: 1, SpeedDuplex: "Auto", FlowControl: "Off"},
	}

	assert.Equal(t, expectedPorts, ports)
}

func TestGetPort(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockPortResponse))
	}))
	defer server.Close()

	client := &sdk.HRUIClient{
		URL:        server.URL,
		HttpClient: server.Client(),
	}

	port, err := client.GetPort(1)
	require.NoError(t, err)

	expectedPort := &sdk.Port{
		ID:          1,
		State:       1,
		SpeedDuplex: "Auto",
		FlowControl: "Off",
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
							<td>Port 2</td>
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

	client := &sdk.HRUIClient{
		URL:        server.URL,
		HttpClient: server.Client(),
	}

	stats, err := client.GetPortStatistics()
	require.NoError(t, err)

	expectedStats := []*sdk.PortStatistics{
		{Port: 1, State: 1, LinkStatus: "Link Up", TxGoodPkt: 1539909, TxBadPkt: 0, RxGoodPkt: 6063265, RxBadPkt: 0},
		{Port: 2, State: 1, LinkStatus: "Link Down", TxGoodPkt: 173978, TxBadPkt: 0, RxGoodPkt: 2140448, RxBadPkt: 0},
	}

	assert.Equal(t, expectedStats, stats)
}
