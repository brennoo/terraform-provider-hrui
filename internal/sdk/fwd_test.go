package sdk

import (
	"net/http"
	"net/http/httptest"
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
