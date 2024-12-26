package sdk

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockIGMPHTMLStaticPorts(enabledPorts []int) string {
	html := `
<html>  
<head>  
<title>IGMP</title>  
<link rel="stylesheet" type="text/css" href="/style.css">  
<script type="text/javascript"></script>  
</head>

<body>  
<center>  
<fieldset>  
<legend>IGMP Enable Setting</legend>  
<form method="post" name="igmp" action="/igmp.cgi?page=enable_igmp">  
  <table border="1">  
    <tr>  
      <th width="180">Enable</th>  
      <td style="text-align:left;" width="180"><input type="checkbox" name="enable_igmp" checked></td>  
    </tr>  
  </table>  
  <br style="line-height:50%">  
  <input type="submit" value="   Apply   ">  
</form>  
<hr>  
<form method="post" action="/igmp.cgi?page=igmp_static_router">  
  <table border="0" style="font-family:Sans-serif">  
    <tr>  
      <td colspan="2" style="text-align:left">  
        <table border="1" style="font-size:12px">  
          <tr>  
            <th width="120">Router Port</th>  
            <td>1</td>  
            <td>2</td>  
            <td>3</td>  
            <td>4</td>  
            <td>5</td>  
            <td>6</td>  
          </tr>  
          <tr>  
            <th width="120">static</th>
`

	// Add the static IGMP port data dynamically
	for i := 0; i < 6; i++ {
		portIndex := i + 1 // Ports are 1-based in the HTML
		if contains(enabledPorts, portIndex) {
			html += fmt.Sprintf(`<td align="center"><input type="checkbox" name="lPort_%d" checked></td>`, i)
		} else {
			html += fmt.Sprintf(`<td align="center"><input type="checkbox" name="lPort_%d"></td>`, i)
		}
	}

	html += `
          </tr>  
          <tr>  
            <th width="120">dynamic</th>  
            <td align="center"><input type="checkbox" name="lPort_0" checked disabled></td>  
            <td align="center"><input type="checkbox" name="lPort_1" disabled></td>  
            <td align="center"><input type="checkbox" name="lPort_2" disabled></td>  
            <td align="center"><input type="checkbox" name="lPort_3" disabled></td>  
            <td align="center"><input type="checkbox" name="lPort_4" disabled></td>  
            <td align="center"><input type="checkbox" name="lPort_5" disabled></td>  
          </tr>  
        </table>  
      </td>  
    </tr>  
  </table>  
  <br>  
  <input type="hidden" name="cmd" value="set">  
  <input type="submit" value="Add / Modify">  
</form>  
</fieldset>  
</center>  
</body>  
</html>  
`
	return html
}

func TestGetPortIGMPSnooping(t *testing.T) {
	tests := []struct {
		name           string
		port           int
		mockHTML       string
		expectedStatus bool
		expectedError  bool
	}{
		{
			name:           "Port 3 IGMP Enabled",
			port:           3,
			mockHTML:       mockIGMPHTMLStaticPorts([]int{3}),
			expectedStatus: true,
			expectedError:  false,
		},
		{
			name:           "Port 3 IGMP Disabled",
			port:           3,
			mockHTML:       mockIGMPHTMLStaticPorts([]int{1, 2, 4, 5, 6}),
			expectedStatus: false,
			expectedError:  false,
		},
		{
			name:           "Invalid Port Number",
			port:           7,
			mockHTML:       mockIGMPHTMLStaticPorts([]int{1, 3}),
			expectedStatus: false,
			expectedError:  true,
		},
		{
			name:           "Port 2 IGMP Enabled",
			port:           2,
			mockHTML:       mockIGMPHTMLStaticPorts([]int{2}),
			expectedStatus: true,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Printf("Received request: Path=%s, Query=%s\n", r.URL.Path, r.URL.RawQuery)

				if r.URL.Path == "/igmp.cgi" && r.URL.Query().Get("page") == "dump" {
					w.WriteHeader(http.StatusOK)
					if _, err := w.Write([]byte(tt.mockHTML)); err != nil {
						t.Fatalf("failed to write response for /igmp.cgi: %v", err)
					}
					return
				}
				if r.URL.Path == "/port.cgi" {
					w.WriteHeader(http.StatusOK)
					if _, err := w.Write([]byte(mockPortHTML())); err != nil {
						t.Fatalf("failed to write response for /port.cgi: %v", err)
					}
					return
				}

				http.Error(w, "not found", http.StatusNotFound)
			}))
			defer server.Close()

			// Create the mock client
			client := &HRUIClient{
				URL:        server.URL,
				HttpClient: server.Client(),
			}

			// Call the method to test
			status, err := client.GetPortIGMPSnooping(tt.port)

			// Validate the results
			if tt.expectedError {
				if assert.Error(t, err) {
					fmt.Printf("Expected error occurred: %v\n", err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, status)
			}
		})
	}
}

// Helper function to create a mock response HTML for ListPorts.
func mockPortHTML() string {
	return `
<html>

<head>
<title>Port Setting</title>
<link rel="stylesheet" type="text/css" href="/style.css">
<script type="text/javascript">
</script>
</head>

<body>
<center>

<fieldset>
<legend>Port Setting</legend>
<form method="post" action="port.cgi">
	<table border="1">
	<tr>
		<th class=MidSize>Port</th>
		<th>State</th>
		<th>Speed/Duplex</th>
		<th>Flow Control</th>
	</tr>
	<tr>
		<td align="center">
			<select name="portid" multiple size="4" >
			<option value="0">Port 1
			<option value="1">Port 2
			<option value="2">Port 3
			<option value="3">Port 4
		</td>
		<td align="center">
			<select name="state" class=MidSize>
			<option value="1">Enable
			<option value="0">Disable
			</select>
		</td>
		<td align="center">
			<select name="speed_duplex" class=MidSize>
			<option value="0">Auto
			<option value="1">10M/Half
			<option value="2">10M/Full
			<option value="3">100M/Half
			<option value="4">100M/Full
			<option value="5">1000M/Full
			<option value="6">2500M/Full
			</select>
		</td>
		<td align="center">
			<select name="flow" class=MidSize>
			<option value="0">Off
			<option value="1">On
			</select>
		</td>
	</tr>
	</table>
<br style="line-height:50%">
<input type="submit" name="submit" value="   Apply   ">
<input type="hidden" name="cmd" value="port">
</form>
<hr>
<form method="post" action="port.cgi">
	<table border="1">
	<tr>
		<th class=MidSize>Port</th>
		<th>State</th>
		<th>Speed/Duplex</th>
		<th>Flow Control</th>
	</tr>
	<tr>
		<td align="center">
			<select name="portid" multiple size="2" >
			<option value="4">Port 5
			<option value="5">Port 6
		</td>
		<td align="center">
			<select name="state" class=MidSize>
			<option value="1">Enable
			<option value="0">Disable
			</select>
		</td>
		<td align="center">
			<select name="speed_duplex" class=MidSize>
			<option value="0">Auto
			<option value="4">100M/Full
			<option value="5">1000M/Full
			<option value="6">2500M/Full
			<option value="8">10G/Full
			</select>
		</td>
		<td align="center">
			<select name="flow" class=MidSize>
			<option value="0">Off
			<option value="1">On
			</select>
		</td>
	</tr>
	</table>
<br style="line-height:50%">
<input type="submit" name="submit" value="   Apply   ">
<input type="hidden" name="cmd" value="port">
</form>
<hr>
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
}
