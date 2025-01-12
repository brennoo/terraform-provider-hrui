package sdk

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock Generator for Dynamic Ports.
func mockIGMPHTMLDynamicPorts(portNames []string) string {
	html := `
<html>
<head>
<title>IGMP</title>
</head>
<body>
<form action="/igmp.cgi?page=igmp_static_router">
  <table>
    <tr>
      <th>Router Port</th>
`
	// Add port names dynamically
	for _, port := range portNames {
		html += fmt.Sprintf("<td>%s</td>", port)
	}

	html += `
    </tr>
    <tr>
      <th>static</th>
`

	// Simulate checkboxes for each port
	for i := range portNames {
		html += fmt.Sprintf(`<td><input type="checkbox" name="lPort_%d"></td>`, i)
	}

	html += `
    </tr>
  </table>
</form>
</body>
</html>
`
	return html
}

// Mock Generator for Static Ports (with Enabled States).
func mockIGMPHTMLStaticPorts(enabledPorts []int) string {
	html := `
<html>
<head>
<title>IGMP</title>
</head>
<body>
<form action="/igmp.cgi?page=igmp_static_router">
  <table>
    <tr>
      <th>Router Port</th>
      <td>1</td>
      <td>2</td>
      <td>3</td>
      <td>4</td>
      <td>5</td>
      <td>6</td>
    </tr>
    <tr>
      <th>static</th>
`
	// Add port IGMP snooping states dynamically
	for i := 0; i < 6; i++ {
		if contains(enabledPorts, i+1) {
			html += fmt.Sprintf(`<td><input type="checkbox" name="lPort_%d" checked></td>`, i)
		} else {
			html += fmt.Sprintf(`<td><input type="checkbox" name="lPort_%d"></td>`, i)
		}
	}

	html += `
    </tr>
  </table>
</form>
</body>
</html>
`
	return html
}

// Test for dynamic ResolvePortNameToID function.
func TestResolvePortNameToID(t *testing.T) {
	tests := []struct {
		name          string
		portName      string
		mockHTML      string
		expectedID    int
		expectedError bool
	}{
		{"Valid Port 1", "1", mockIGMPHTMLDynamicPorts([]string{"1", "2", "3", "4", "Trunk2"}), 0, false},
		{"Valid Port 3", "3", mockIGMPHTMLDynamicPorts([]string{"1", "2", "3", "4", "Trunk2"}), 2, false},
		{"Valid Port Trunk2", "Trunk2", mockIGMPHTMLDynamicPorts([]string{"1", "2", "3", "4", "Trunk2"}), 4, false},
		{"Invalid Port Name", "7", mockIGMPHTMLDynamicPorts([]string{"1", "2", "3", "4", "Trunk2"}), -1, true},
		{"Empty Port Name", "", mockIGMPHTMLDynamicPorts([]string{"1", "2", "3", "4", "Trunk2"}), -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/igmp.cgi" && r.URL.Query().Get("page") == "dump" {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(tt.mockHTML))
					return
				}
				http.Error(w, "not found", http.StatusNotFound)
			}))
			defer server.Close()

			client := &HRUIClient{
				URL:        server.URL,
				HttpClient: server.Client(),
			}

			id, err := client.ResolvePortNameToID(tt.portName)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedID, id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}

// Test for GetPortIGMPSnooping function.
func TestGetPortIGMPSnooping(t *testing.T) {
	tests := []struct {
		name           string
		port           int
		mockHTML       string
		expectedStatus bool
		expectedError  bool
	}{
		{"Port 3 IGMP Enabled", 2, mockIGMPHTMLStaticPorts([]int{3}), true, false},
		{"Port 3 IGMP Disabled", 2, mockIGMPHTMLStaticPorts([]int{1, 2, 4, 5, 6}), false, false},
		{"Invalid Port Number", 7, mockIGMPHTMLStaticPorts([]int{1, 3}), false, true},
		{"Port 2 IGMP Enabled", 1, mockIGMPHTMLStaticPorts([]int{2}), true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/igmp.cgi" && r.URL.Query().Get("page") == "dump" {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(tt.mockHTML))
					return
				}
				http.Error(w, "not found", http.StatusNotFound)
			}))
			defer server.Close()

			client := &HRUIClient{
				URL:        server.URL,
				HttpClient: server.Client(),
			}

			status, err := client.GetPortIGMPSnooping(tt.port)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, status)
			}
		})
	}
}
