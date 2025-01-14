package sdk

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetStormControlStatus(t *testing.T) {
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
				<td align="center">Port 4</td><td align="center">2490000</td><td align="center">Off</td><td align="center">Off</td><td align="center">Off</td>
			</tr>
		</table>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(mockHTML)); err != nil {
			t.Fatalf("failed to write mock HTML response: %v", err)
		}
	}))

	defer server.Close()

	client := &HRUIClient{HttpClient: server.Client(), URL: server.URL}
	config, err := client.GetStormControlStatus()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := StormControlConfig{
		Entries: []StormControlEntry{
			{
				Port:                     1,
				BroadcastRateKbps:        nil,
				KnownMulticastRateKbps:   intPointer(25000),
				UnknownUnicastRateKbps:   intPointer(25000),
				UnknownMulticastRateKbps: nil,
			},
			{
				Port:                     4,
				BroadcastRateKbps:        intPointer(2490000),
				KnownMulticastRateKbps:   nil,
				UnknownUnicastRateKbps:   nil,
				UnknownMulticastRateKbps: nil,
			},
		},
	}

	// Fixing comparison using helper to dereference pointers
	for i, entry := range config.Entries {
		if !compareStormControlEntry(entry, expected.Entries[i]) {
			t.Errorf("entry %d: expected %+v, got %+v", i, expected.Entries[i], entry)
		}
	}
}

func TestGetPortMaxRate(t *testing.T) {
	mockHTML := `
		<table border="1">
			<tr>
				<td align="center">Port 1</td>
				<td align="center">(1-2500000)(kbps)</td>
			</tr>
		</table>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(mockHTML)); err != nil {
			t.Fatalf("failed to write mock HTML response: %v", err)
		}
	}))

	defer server.Close()

	client := &HRUIClient{HttpClient: server.Client(), URL: server.URL}
	maxRate, err := client.GetPortMaxRate(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedRate := int64(2500000)
	if maxRate != expectedRate {
		t.Errorf("expected max rate %d, got %d", expectedRate, maxRate)
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
	jumboFrame, err := client.GetJumboFrame()
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

	// Mock server for handling POST request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %v", r.Method)
		}

		// Parse the POST form data
		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse form data: %v", err)
		}

		cmd := r.FormValue("cmd")
		jumboframe := r.FormValue("jumboframe")

		// Assert the form data values
		if cmd != "jumboframe" {
			t.Errorf("expected cmd to be 'jumboframe', got %v", cmd)
		}
		if jumboframe != "3" { // "3" corresponds to 9216 in dropdown mapping
			t.Errorf("expected jumboframe to be '3', got %v", jumboframe)
		}

		// Respond with a success page (acknowledgment)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<title>Jumbo Frame Setting</title>`))
	}))
	defer server.Close()

	client := &HRUIClient{HttpClient: server.Client(), URL: server.URL}
	err := client.SetJumboFrame(expectedFrameSize)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetJumboFrameInvalidSize(t *testing.T) {
	client := &HRUIClient{HttpClient: http.DefaultClient, URL: "http://example.com"}

	invalidFrameSize := 1500 // This size is not supported
	err := client.SetJumboFrame(invalidFrameSize)
	if err == nil {
		t.Fatalf("expected error for invalid frame size, got nil")
	}

	expectedError := "invalid Jumbo Frame size '1500': supported sizes are 1522, 1536, 1552, 9216, 16383"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error message to contain '%s', got '%v'", expectedError, err.Error())
	}
}
