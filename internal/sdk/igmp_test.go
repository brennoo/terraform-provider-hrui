package sdk

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// igmpHandlerMock mocks the IGMP-related HTML response for igmp.cgi.
func igmpHandlerMock(enabledPorts map[int]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		html := `
		<html>
		<fieldset>
			<legend>IGMP Enable Setting</legend>
			<table>
				<tr>
					<th>Router Port</th>
					<td>1</td>
					<td>2</td>
					<td>3</td>
					<td>4</td>
					<td>Trunk2</td>
				</tr>
				<tr>
					<th>static</th>`
		// Dynamically generate the IGMP state table based on enabledPorts.
		for i := 0; i <= 8; i++ {
			state := enabledPorts[i]
			if state == "on" {
				html += fmt.Sprintf(`<td><input type="checkbox" name="lPort_%d" checked></td>`, i)
			} else {
				html += fmt.Sprintf(`<td><input type="checkbox" name="lPort_%d"></td>`, i)
			}
		}
		html += `
				</tr>
			</table>
		</fieldset>
		</html>
		`
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}
}

// TestConfigurePortIGMPSnooping tests the ConfigurePortIGMPSnooping function.
func TestConfigurePortIGMPSnooping(t *testing.T) {
	tests := []struct {
		name          string
		portID        int
		enable        bool
		mockPorts     map[int]string
		expectedError bool
	}{
		{
			name:   "Enable Port 2",
			portID: 1,
			enable: true,
			mockPorts: map[int]string{
				0: "off", 1: "off", 2: "off", 3: "on", 8: "off",
			},
			expectedError: false,
		},
		{
			name:   "Disable Port 8",
			portID: 8,
			enable: false,
			mockPorts: map[int]string{
				0: "on", 1: "on", 2: "off", 3: "on", 8: "on",
			},
			expectedError: false,
		},
		{
			name:   "Port not in configuration (treated as off, can be enabled)",
			portID: 99,
			enable: true,
			mockPorts: map[int]string{
				0: "on", 1: "on", 2: "off", 3: "on", 8: "on",
			},
			expectedError: false, // Ports not in config are treated as "off" and can be enabled
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(igmpHandlerMock(tc.mockPorts))
			defer server.Close()

			client := &HRUIClient{
				URL:        server.URL,
				HttpClient: server.Client(),
			}

			err := client.ConfigurePortIGMPSnooping(context.Background(), tc.portID, tc.enable)
			if tc.expectedError {
				assert.Error(t, err, tc.name)
			} else {
				assert.NoError(t, err, tc.name)
			}
		})
	}
}

// TestGetPortIGMPSnooping tests the GetPortIGMPSnooping function.
func TestGetPortIGMPSnooping(t *testing.T) {
	tests := []struct {
		name          string
		portID        int
		mockPorts     map[int]string
		expectedState bool
		expectedError bool
	}{
		{
			name:   "Port enabled (on)",
			portID: 3,
			mockPorts: map[int]string{
				0: "off", 1: "off", 2: "off", 3: "on", 8: "off",
			},
			expectedState: true,
			expectedError: false,
		},
		{
			name:   "Port disabled (off)",
			portID: 1,
			mockPorts: map[int]string{
				0: "off", 1: "off", 2: "off", 3: "on", 8: "off",
			},
			expectedState: false,
			expectedError: false,
		},
		{
			name:   "Port not in configuration (returns false, not error)",
			portID: 99,
			mockPorts: map[int]string{
				0: "on", 1: "on", 2: "off", 3: "on", 8: "on",
			},
			expectedState: false, // Ports not in config are treated as disabled
			expectedError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(igmpHandlerMock(tc.mockPorts))
			defer server.Close()

			client := &HRUIClient{
				URL:        server.URL,
				HttpClient: server.Client(),
			}

			enabled, err := client.GetPortIGMPSnooping(context.Background(), tc.portID)
			if tc.expectedError {
				require.Error(t, err, tc.name)
			} else {
				require.NoError(t, err, tc.name)
				assert.Equal(t, tc.expectedState, enabled, tc.name)
			}
		})
	}
}
