package sdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetStormControlStatus(t *testing.T) {
	// Mock storm control table response
	mockHTML := `
		<table border="1">
			<tr>
				<th>Port</th>
				<th>Broadcast (kbps)</th>
				<th>Known Multicast (kbps)</th>
				<th>Unknown Unicast (kbps)</th>
				<th>Unknown Multicast (kbps)</th>
			</tr>
			<tr>
				<td align="center">Port 1</td><td align="center">Off</td><td align="center">25000</td><td align="center">25000</td><td align="center">Off</td>
			</tr>
			<tr>
				<td align="center">Trunk1</td><td align="center">2490000</td><td align="center">Off</td><td align="center">Off</td><td align="center">Off</td>
			</tr>
		</table>`

	// Mock port.cgi HTML response
	mockPortHTMLResponse := `
		<html>
		<body>
			<form action="/port.cgi" method="get">
				<select name="portid">
					<option value="1">Port 1</option>
					<option value="2">Port 2</option>
					<option value="3">Trunk1</option>
				</select>
			</form>
		</body>
		</html>`

	// Mock server with conditional response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/port.cgi" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(mockPortHTMLResponse))
			return
		}
		if _, err := w.Write([]byte(mockHTML)); err != nil {
			t.Fatalf("failed to write mock HTML response: %v", err)
		}
	}))
	defer server.Close()

	// Create the test client
	client := &HRUIClient{HttpClient: server.Client(), URL: server.URL}

	// Test GetStormControlStatus
	config, err := client.GetStormControlStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected values
	expected := StormControlConfig{
		Entries: []StormControlEntry{
			{
				Port:                     "Port 1", // Using name as string
				BroadcastRateKbps:        nil,
				KnownMulticastRateKbps:   intPointer(25000),
				UnknownUnicastRateKbps:   intPointer(25000),
				UnknownMulticastRateKbps: nil,
			},
			{
				Port:                     "Trunk1", // Using name as string
				BroadcastRateKbps:        intPointer(2490000),
				KnownMulticastRateKbps:   nil,
				UnknownUnicastRateKbps:   nil,
				UnknownMulticastRateKbps: nil,
			},
		},
	}

	// Ensure entries match expected values
	for i, entry := range config.Entries {
		if !compareStormControlEntry(entry, expected.Entries[i]) {
			t.Errorf("entry %d: expected %+v, got %+v", i, expected.Entries[i], entry)
		}
	}
}

func TestGetPortMaxRate(t *testing.T) {
	mockHTML := `
    <html>
    <body>
        <table border="1">
            <tr>
                <th align="center">Port</th>
                <th align="center">Broadcast (kbps)</th>
                <th align="center">Known Multicast (kbps)</th>
                <th align="center">Unknown Unicast (kbps)</th>
                <th align="center">Unknown Multicast (kbps)</th>
            </tr>
            <tr>
                <td align="center">Port 1</td>
                <td align="center">(1-2500000)(kbps)</td>
                <td align="center">(1-25000)(kbps)</td>
                <td align="center">(1-50000)(kbps)</td>
                <td align="center">(1-100000)(kbps)</td>
            </tr>
            <tr>
                <td align="center">Port 2</td>
                <td align="center">(1-2490000)(kbps)</td>
                <td align="center">Off</td>
                <td align="center">Off</td>
                <td align="center">(1-50000)(kbps)</td>
            </tr>
        </table>
    </body>
    </html>`

	// Mock HTTP server to serve the HTML
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(mockHTML))
		if err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := &HRUIClient{HttpClient: server.Client(), URL: server.URL}

	// Test case: Valid port (Port 1)
	rate, err := client.GetPortMaxRate(context.Background(), "Port 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rate != 2500000 {
		t.Errorf("expected rate 2500000, got %d", rate)
	}

	// Test case: Valid port (Port 2)
	rate, err = client.GetPortMaxRate(context.Background(), "Port 2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rate != 2490000 {
		t.Errorf("expected rate 2490000, got %d", rate)
	}

	// Test case: Non-existent port (Port 3)
	_, err = client.GetPortMaxRate(context.Background(), "Port 3")
	if err == nil {
		t.Fatal("expected an error, received none")
	}
	if !strings.Contains(err.Error(), "rate information not found for port 'Port 3'") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// Helper function to create int pointer.
func intPointer(v int) *int {
	return &v
}

// Helper function to compare StormControlEntry.
func compareStormControlEntry(a, b StormControlEntry) bool {
	return a.Port == b.Port &&
		comparePointers(a.BroadcastRateKbps, b.BroadcastRateKbps) &&
		comparePointers(a.KnownMulticastRateKbps, b.KnownMulticastRateKbps) &&
		comparePointers(a.UnknownUnicastRateKbps, b.UnknownUnicastRateKbps) &&
		comparePointers(a.UnknownMulticastRateKbps, b.UnknownMulticastRateKbps)
}

// Helper function to compare int pointers.
func comparePointers(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func TestGetJumboFrame(t *testing.T) {
	mockHTML := `
		<html>
		<body>
			<form method="post" name="jumboframe" action="/fwd.cgi?page=jumboframe">
				<select name="jumboframe">
					<option value="0">1522</option>
					<option value="1">1536</option>
					<option value="2">1552</option>
					<option value="3">9216</option>
					<option value="4" selected>16383</option>
				</select>
			</form>
		</body>
		</html>`

	// Mock server to simulate Jumbo Frame settings page
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET request, got %v", r.Method)
		}
		if _, err := w.Write([]byte(mockHTML)); err != nil {
			t.Fatalf("failed to write mock HTML response: %v", err)
		}
	}))
	defer server.Close()

	client := &HRUIClient{HttpClient: server.Client(), URL: server.URL}
	jumboFrame, err := client.GetJumboFrame(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected frame size is 16383 (selected entry in mock HTML).
	expectedFrameSize := 16383
	if jumboFrame.FrameSize != expectedFrameSize {
		t.Errorf("expected frame size %d, got %d", expectedFrameSize, jumboFrame.FrameSize)
	}
}

func TestSetJumboFrame(t *testing.T) {
	expectedFrameSize := 9216 // The frame size we are setting (mapped to "3" in the form)

	responseHTML := `
<html>
<head><title>Jumbo Frame Setting</title></head>
<body>
<center>
<fieldset>
<legend>Jumbo Frame Setting</legend>
<form method="post" name="jumboframe" action="/fwd.cgi?page=jumboframe">
<select name="jumboframe">
  <option value="0">1522</option>
  <option value="1">1536</option>
  <option value="2">1552</option>
  <option value="3" selected>9216</option>
  <option value="4">16383</option>
</select>
</form>
</fieldset>
</center>
</body>
</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				t.Fatalf("failed to parse form data: %v", err)
			}
			cmd := r.FormValue("cmd")
			jumboframe := r.FormValue("jumboframe")

			if cmd != "jumboframe" {
				t.Errorf("expected cmd to be 'jumboframe', got %v", cmd)
			}
			if jumboframe != "3" { // "3" corresponds to 9216 in dropdown mapping
				t.Errorf("expected jumboframe to be '3', got %v", jumboframe)
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(responseHTML))
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(responseHTML))
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}))
	defer server.Close()

	client := &HRUIClient{HttpClient: server.Client(), URL: server.URL}
	appliedSize, err := client.SetJumboFrame(context.Background(), expectedFrameSize)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if appliedSize != expectedFrameSize {
		t.Fatalf("expected applied size %d, got %d", expectedFrameSize, appliedSize)
	}
}

func TestSetJumboFrameInvalidSize(t *testing.T) {
	client := &HRUIClient{HttpClient: http.DefaultClient, URL: "http://example.com"}

	invalidFrameSize := 1500 // This size is not supported
	_, err := client.SetJumboFrame(context.Background(), invalidFrameSize)
	if err == nil {
		t.Fatalf("expected error for invalid frame size, got nil")
	}

	expectedError := "invalid Jumbo Frame size '1500': supported sizes are 1522, 1536, 1552, 9216, 16383"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error message to contain '%s', got '%v'", expectedError, err.Error())
	}
}
