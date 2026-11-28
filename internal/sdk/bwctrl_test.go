package sdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConfigureBandwidthControl(t *testing.T) {
	// Mock HTML response for `/port.cgi`
	mockPortHTML := `
    <html>
        <head>
            <title>Port Configuration</title>
        </head>
        <body>
            <form method="get" action="/port.cgi">
                <select name="portid" multiple size="6">
                    <option value="0">Port 1</option>
                    <option value="1">Port 2</option>
                    <option value="2">Port 3</option>
                    <option value="3">Port 4</option>
                </select>
            </form>
        </body>
    </html>`

	// Mock HTML response for `/port.cgi?page=bwctrl`
	mockBandwidthHTML := `
    <html>
        <head>
            <title>Bandwidth Control Setting</title>
        </head>
        <body>
            <form method="post" action="/port.cgi?page=bwctrl">
                <table>
                    <tr>
                        <td><input type="text" name="rate" value="Unlimited"></td>
                    </tr>
                </table>
            </form>
        </body>
    </html>`

	// Mock HTML response for `/port.cgi?page=bw_ctrl` (v2.x current-state probe)
	mockBwCtrlHTML := `
    <html>
        <body>
            <table>
                <tr><td>Port</td><td>Ingress Rate</td><td>Egress Rate</td></tr>
                <tr><td>Port 1</td><td>Unlimited</td><td>Unlimited</td></tr>
                <tr><td>Port 2</td><td>Unlimited</td><td>Unlimited</td></tr>
                <tr><td>Port 3</td><td>Unlimited</td><td>Unlimited</td></tr>
                <tr><td>Port 4</td><td>Unlimited</td><td>Unlimited</td></tr>
            </table>
        </body>
    </html>`

	// Mock HTML response for `/fwd.cgi?page=storm_ctrl`
	mockStormCtrlHTML := `
                <table border="1">
                        <tr>
                                <td align="center">Port 1</td>
                                <td align="center">(1-2500000)(kbps)</td>
                        </tr>
                </table>`

	// Set up HTTP test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/port.cgi" && r.URL.Query().Get("page") == "" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(mockPortHTML))
		} else if r.URL.Path == "/port.cgi" && r.URL.Query().Get("page") == "bwctrl" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(mockBandwidthHTML))
		} else if r.URL.Path == "/port.cgi" && r.URL.Query().Get("page") == "bw_ctrl" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(mockBwCtrlHTML))
		} else if r.URL.Path == "/fwd.cgi" && r.URL.Query().Get("page") == "storm_ctrl" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(mockStormCtrlHTML))
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Inject the mock server URL into HRUIClient
	client := &HRUIClient{
		URL:        server.URL,
		HttpClient: server.Client(),
	}

	// Define test cases
	testCases := []struct {
		name      string
		port      string
		isIngress bool
		enable    bool
		rate      string
		expectErr bool
	}{
		{"Enable ingress with specific rate", "Port 1", true, true, "1000", false},
		{"Enable egress with Unlimited rate", "Port 2", false, true, "Unlimited", false},
		{"Disable ingress control (rate ignored)", "Port 3", true, false, "0", false},
		{"Set egress rate to literal 0", "Port 4", false, true, "0", false},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.ConfigureBandwidthControl(context.Background(), tc.port, tc.isIngress, tc.enable, tc.rate)
			if (err != nil) != tc.expectErr {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestConfigureBandwidthControl_ProbeFails ensures that when the current-state
// probe (GetBandwidthControl) fails, the function returns an error instead of
// silently posting Unlimited for the untouched direction.
func TestConfigureBandwidthControl_ProbeFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer server.Close()

	client := &HRUIClient{
		URL:        server.URL,
		HttpClient: server.Client(),
	}

	err := client.configureBandwidthByID(context.Background(), 0, true, true, "1000")
	if err == nil {
		t.Fatal("expected an error when the bandwidth control probe fails, but got nil")
	}
}
