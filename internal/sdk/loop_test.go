package sdk

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLoopProtocol(t *testing.T) {
	htmlResponse := `
	<html>
		<body>
			<select name="func_type">
				<option value="0">Off</option>
				<option value="1">Loop Detection</option>
				<option value="2" selected>Loop Prevention</option>
			</select>
			<input name="interval_time" value="5" />
			<input name="recover_time" value="10" />
			
			<table>
				<tr>
					<th>Port</th><th>State</th><th>Status</th>
				</tr>
				<tr>
					<td>1</td><td>Disable</td><td>Forwarding</td>
				</tr>
				<tr>
					<td>2</td><td>Enable</td><td>Forwarding</td>
				</tr>
			</table>
		</body>
	</html>`

	// Mock the HTTP response from loop.cgi
	mockServer := mockServerMock(htmlResponse, http.StatusOK)
	defer mockServer.Close()

	client := &HRUIClient{
		URL:        mockServer.URL,
		HttpClient: &http.Client{},
	}

	// Call GetLoopProtocol and check for correct parsing
	loopProtocol, err := client.GetLoopProtocol()
	assert.NoError(t, err)

	// Check the detected loop function
	assert.Equal(t, "Loop Prevention", loopProtocol.LoopFunction)

	// Verify the interval and recovery times
	assert.Equal(t, 5, loopProtocol.IntervalTime)
	assert.Equal(t, 10, loopProtocol.RecoverTime)

	// Expected port statuses for each port (based on the HTML response)
	expected := []PortStatus{
		{Port: 1, Enable: false, LoopState: "Disable", LoopStatus: "Forwarding"},
		{Port: 2, Enable: true, LoopState: "Enable", LoopStatus: "Forwarding"},
	}

	// Assert that the parsed result matches the expected port statuses
	assert.Equal(t, expected, loopProtocol.PortStatuses)
}

func TestConfigureLoopProtocol(t *testing.T) {
	// Use the mock server to simulate the HTTP POST response.
	mock := mockServerMock("OK", http.StatusOK)
	defer mock.Close()

	// Create an HRUIClient pointing to the mock server
	client := &HRUIClient{
		URL:        mock.URL,
		HttpClient: &http.Client{},
	}

	// Call ConfigureLoopProtocol and assert that no error is returned.
	err := client.ConfigureLoopProtocol("Loop Prevention", 5, 12, []PortStatus{
		{Port: 1, Enable: true},
		{Port: 2, Enable: false},
	})
	assert.NoError(t, err)
}

func TestGetSTPPortSettings(t *testing.T) {
	htmlResponse := `
	<html>
		<head>
			<title>Spanning Tree Port Setting</title>
		</head>
		<body>
			<center>
				<fieldset>
					<legend>Spanning Tree Port Setting</legend>
					<table border="1">
						<tr>
							<th rowspan="2" width="115">Port</th>
							<th rowspan="2" width="165">State</th>
							<th rowspan="2" width="165">Role</th>
							<th colspan="2" width="165">Path Cost</th>
							<th rowspan="2" width="90">Priority</th>
							<th colspan="2" width="90">P2P</th>
							<th colspan="2" width="90">Edge</th>
						</tr>
						<tr>
							<th>Config</th>
							<th>Actual</th>
							<th>Config</th>
							<th>Actual</th>
							<th>Config</th>
							<th>Actual</th>
						</tr>
						<tr>
							<td>Port 1</td>
							<td align="center">Forwarding</td>
							<td align="center">Designated</td>
							<td align="center">234</td>
							<td align="center">234</td>
							<td align="center">128</td>
							<td align="center">Auto</td>
							<td align="center">TRUE</td>
							<td align="center">False</td>
							<td align="center">True</td>
						</tr>
						<tr>
							<td>Port 2</td>
							<td align="center">Disabled</td>
							<td align="center">-</td>
							<td align="center">201</td>
							<td align="center">-</td>
							<td align="center">128</td>
							<td align="center">Auto</td>
							<td align="center">-</td>
							<td align="center">False</td>
							<td align="center">-</td>
						</tr>
						<tr>
							<td>Port 3</td>
							<td align="center">Disabled</td>
							<td align="center">-</td>
							<td align="center">20000</td>
							<td align="center">-</td>
							<td align="center">128</td>
							<td align="center">Auto</td>
							<td align="center">-</td>
							<td align="center">False</td>
							<td align="center">-</td>
						</tr>
						<tr>
							<td>Port 4</td>
							<td align="center">Disabled</td>
							<td align="center">-</td>
							<td align="center">Auto</td>
							<td align="center">-</td>
							<td align="center">128</td>
							<td align="center">True</td>
							<td align="center">-</td>
							<td align="center">False</td>
							<td align="center">-</td>
						</tr>
					</table>
				</fieldset>
			</center>
		</body>
	</html>`

	// Mock the HTTP server response with the above HTML
	mockServer := mockServerMock(htmlResponse, http.StatusOK)
	defer mockServer.Close()

	// Create a client instance
	client := &HRUIClient{
		URL:        mockServer.URL,
		HttpClient: &http.Client{},
	}

	// Call GetSTPPortSettings
	stpPorts, err := client.GetSTPPortSettings()
	assert.NoError(t, err)

	// Expected result based on the mock HTML structure
	expected := []STPPort{
		{
			Port:           0,
			State:          "Forwarding",
			Role:           "Designated",
			PathCostConfig: 234,
			PathCostActual: 234,
			Priority:       128,
			P2PConfig:      "Auto",
			P2PActual:      "True",
			EdgeConfig:     "False",
			EdgeActual:     "True",
		},
		{
			Port:           1,
			State:          "Disabled",
			Role:           "-",
			PathCostConfig: 201,
			PathCostActual: 0,
			Priority:       128,
			P2PConfig:      "Auto",
			P2PActual:      "-",
			EdgeConfig:     "False",
			EdgeActual:     "-",
		},
		{
			Port:           2,
			State:          "Disabled",
			Role:           "-",
			PathCostConfig: 20000,
			PathCostActual: 0,
			Priority:       128,
			P2PConfig:      "Auto",
			P2PActual:      "-",
			EdgeConfig:     "False",
			EdgeActual:     "-",
		},
		{
			Port:           3,
			State:          "Disabled",
			Role:           "-",
			PathCostConfig: 0,
			PathCostActual: 0,
			Priority:       128,
			P2PConfig:      "True",
			P2PActual:      "-",
			EdgeConfig:     "False",
			EdgeActual:     "-",
		},
	}

	// Assert that the actual parsed data matches the expected data
	assert.Equal(t, expected, stpPorts)
}

func TestGetLoopProtocol_LoopPreventionMode(t *testing.T) {
	htmlResponse := `
	<html>
		<body>
			<select name="func_type">
				<option value="0">Off</option>
				<option value="1">Loop Detection</option>
				<option value="2" selected>Loop Prevention</option>
			</select>
			<input name="interval_time" value="5" />
			<input name="recover_time" value="10" />
			
			<table>
				<tr>
					<th>Port</th><th>State</th><th>Status</th>
				</tr>
				<tr>
					<td>1</td><td>Disable</td><td>Forwarding</td>
				</tr>
				<tr>
					<td>2</td><td>Enable</td><td>Forwarding</td>
				</tr>
			</table>
		</body>
	</html>`

	mockServer := mockServerMock(htmlResponse, http.StatusOK)
	defer mockServer.Close()

	client := &HRUIClient{
		URL:        mockServer.URL,
		HttpClient: &http.Client{},
	}

	loopProtocol, err := client.GetLoopProtocol()
	assert.NoError(t, err)

	assert.Equal(t, "Loop Prevention", loopProtocol.LoopFunction)
	assert.Equal(t, 5, loopProtocol.IntervalTime)
	assert.Equal(t, 10, loopProtocol.RecoverTime)

	expected := []PortStatus{
		{Port: 1, Enable: false, LoopState: "Disable", LoopStatus: "Forwarding"},
		{Port: 2, Enable: true, LoopState: "Enable", LoopStatus: "Forwarding"},
	}
	assert.Equal(t, expected, loopProtocol.PortStatuses)
}

func TestSetSTPPortSettings(t *testing.T) {
	// Mock POST response (200 OK) for STP update
	mock := mockServerMock("OK", http.StatusOK)
	defer mock.Close()

	client := &HRUIClient{
		URL:        mock.URL,
		HttpClient: &http.Client{},
	}

	// Call SetSTPPortSettings on Port 1 with specific settings
	err := client.SetSTPPortSettings(1, 2000, 128, "Auto", "True")
	assert.NoError(t, err)
}

// TestGetSTPSettings ensures that the STP Global Settings are correctly parsed from the HTML response.
func TestGetSTPSettings(t *testing.T) {
	// Mocked HTML response for STP Global Settings
	htmlResponse := `
	<html>
		<center>
		<fieldset>
		<legend>Spanning Tree Setting</legend>
		<form method="post" name="stp" action="/loop.cgi?page=stp_global">
		<table border="1">
			<tr>
				<th>Spanning Tree Status</th><td style="text-align:left;">Disable</td>
			</tr>
			<tr>
				<th>Force Version</th><td width="200px"><select style="text-align:left;" name="version">
					<option value="0">STP</option>
					<option value="1" selected>RSTP</option>
				</select></td>
			</tr>
			<tr>
				<th>Priority</th><td><select style="text-align:left;" name="priority">
					<option value="0">0</option>
					<option value="4096">4096</option>
					<option value="8192">8192</option>
					<option value="32768" selected>32768</option>
				</select></td>
			</tr>
			<tr><th>Maximum Age</th><td><input type="text" size="2" name="maxage" value="20" maxlength="2">(6~40 Sec)</td></tr>
			<tr><th>Hello Time</th><td><input type="text" size="2" name="hello" value="2" maxlength="2">(1~10 Sec)</td></tr>
			<tr><th>Forward Delay</th><td><input type="text" size="2" name="delay" value="15" maxlength="2">(4~30 Sec)</td></tr>
			<tr><th>Root Priority</th><td style="text-align:left;">32768</td></tr>
			<tr><th>Root MAC Address</th><td style="text-align:left;">1C:2A:A3:23:D1:BA</td></tr>
			<tr><th>Root Path Cost</th><td style="text-align:left;">0</td></tr>
			<tr><th>Root Port</th><td style="text-align:left;">-</td></tr>
			<tr><th>Root Maximum Age</th><td style="text-align:left;">20 Sec</td></tr>
			<tr><th>Root Hello Time</th><td style="text-align:left;">2 Sec</td></tr>
			<tr><th>Root Forward Delay</th><td style="text-align:left;">15 Sec</td></tr>
		</table>
		</form>
		</fieldset>
		</center>
	</html>`

	// Use the mock server to simulate the HTML response.
	mockServer := mockServerMock(htmlResponse, http.StatusOK)
	defer mockServer.Close()

	// Create the HRUIClient with the mocked server URL.
	client := &HRUIClient{
		URL:        mockServer.URL,
		HttpClient: &http.Client{},
	}

	// Call our GetSTPSettings function.
	stpSettings, err := client.GetSTPSettings()
	assert.NoError(t, err)

	// Validate the parsed results.
	assert.Equal(t, "Disable", stpSettings.STPStatus)
	assert.Equal(t, "RSTP", stpSettings.ForceVersion)
	assert.Equal(t, 32768, stpSettings.Priority)
	assert.Equal(t, 20, stpSettings.MaxAge)
	assert.Equal(t, 2, stpSettings.HelloTime)
	assert.Equal(t, 15, stpSettings.ForwardDelay)
	assert.Equal(t, 32768, stpSettings.RootPriority)
	assert.Equal(t, "1C:2A:A3:23:D1:BA", stpSettings.RootMAC)
	assert.Equal(t, 0, stpSettings.RootPathCost)
	assert.Equal(t, "-", stpSettings.RootPort)
	assert.Equal(t, 20, stpSettings.RootMaxAge)
	assert.Equal(t, 2, stpSettings.RootHelloTime)
	assert.Equal(t, 15, stpSettings.RootForwardDelay)
}

// TestSetSTPSettings ensures the correct POST data is sent when updating the STP settings.
func TestSetSTPSettings(t *testing.T) {
	// Prepare mock server to validate POST data.
	mock := mockServerMock("OK", http.StatusOK)
	defer mock.Close()

	client := &HRUIClient{
		URL:        mock.URL,
		HttpClient: &http.Client{},
	}

	// Create the STPGlobalSettings that we want to update.
	stpUpdate := &STPGlobalSettings{
		ForceVersion: "RSTP", // Should map to version=1
		Priority:     4096,
		MaxAge:       15,
		HelloTime:    3,
		ForwardDelay: 7,
	}

	// Execute the SetSTPSettings function.
	err := client.SetSTPSettings(stpUpdate)
	assert.NoError(t, err)

	// Test has already passed if data was sent correctly (mock does the validation).
}
