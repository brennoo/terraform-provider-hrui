package sdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetMACLimits tests the GetMACLimits function for parsing the MAC limits table.
func TestGetMACLimits(t *testing.T) {
	// Mock HTML for the MAC limits table
	mockHTML := `
		<table border="1">
			<tr>
				<th width="90">Port</th>
				<th width="120">Entry Limits</th>
			</tr>
			<tr>
				<td>Port 1</td>
				<td>Unlimited</td>
			</tr>
			<tr>
				<td>Port 2</td>
				<td>100</td>
			</tr>
			<tr>
				<td>Port 3</td>
				<td>Unlimited</td>
			</tr>
			<tr>
				<td>Port 4</td>
				<td>200</td>
			</tr>
		</table>`

	// Mock server to simulate the /mac_constraint.cgi endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Respond with the mock HTML
		if _, err := w.Write([]byte(mockHTML)); err != nil {
			t.Fatalf("failed to write mock HTML response: %v", err)
		}
	}))
	defer server.Close()

	// Initialize the client with the mock server URL
	client := &HRUIClient{
		HttpClient: server.Client(),
		URL:        server.URL,
	}

	// Call the GetMACLimits function
	macLimits, err := client.GetMACLimits(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Define the expected results
	expected := []MACLimit{
		{Port: "Port 1", Enabled: false, Limit: nil},
		{Port: "Port 2", Enabled: true, Limit: intPointer(100)},
		{Port: "Port 3", Enabled: false, Limit: nil},
		{Port: "Port 4", Enabled: true, Limit: intPointer(200)},
	}

	// Compare the results with the expected output
	if len(macLimits) != len(expected) {
		t.Fatalf("expected %d MAC limits, got %d", len(expected), len(macLimits))
	}

	for i, limit := range macLimits {
		expectedLimit := expected[i]
		if limit.Port != expectedLimit.Port || limit.Enabled != expectedLimit.Enabled || !comparePointers(limit.Limit, expectedLimit.Limit) {
			t.Errorf("entry %d: expected %+v, got %+v", i, expectedLimit, limit)
		}
	}
}

// TestSetMACLimit tests the SetMACLimit function for enabling, disabling, and setting specific limits.
func TestSetMACLimit(t *testing.T) {
	// Mock server for handling POST requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %v", r.Method)
		}

		// Parse the form data
		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse form data: %v", err)
		}

		// Get form values for validation
		portID := r.FormValue("portid")
		state := r.FormValue("state")
		limit := r.FormValue("limit")

		// Validate form data based on the test case
		testID := r.Header.Get("Test-ID")
		switch testID {
		case "EnableWithLimit":
			if portID != "1" || state != "1" || limit != "100" {
				t.Errorf("unexpected form values for EnableWithLimit: portid=%v, state=%v, limit=%v", portID, state, limit)
			}
		case "Disable":
			if portID != "2" || state != "0" || limit != "Unlimited" {
				t.Errorf("unexpected form values for Disable: portid=%v, state=%v, limit=%v", portID, state, limit)
			}
		case "EnableUnlimited":
			if portID != "3" || state != "1" || limit != "Unlimited" {
				t.Errorf("unexpected form values for EnableUnlimited: portid=%v, state=%v, limit=%v", portID, state, limit)
			}
		default:
			t.Errorf("invalid Test-ID header: %v", testID)
		}

		// Respond with a success message
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Initialize the client with the mock server URL
	client := &HRUIClient{
		HttpClient: server.Client(),
		URL:        server.URL,
	}

	// Test case 1: Enable with specific limit
	t.Run("EnableWithLimit", func(t *testing.T) {
		reqHeader := http.Header{}
		reqHeader.Set("Test-ID", "EnableWithLimit")
		client.HttpClient.Transport = addTestIDHeader(reqHeader, server.Client().Transport)

		err := client.SetMACLimit(context.Background(), 1, true, intPointer(100))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// Test case 2: Disable MAC limit
	t.Run("Disable", func(t *testing.T) {
		reqHeader := http.Header{}
		reqHeader.Set("Test-ID", "Disable")
		client.HttpClient.Transport = addTestIDHeader(reqHeader, server.Client().Transport)

		err := client.SetMACLimit(context.Background(), 2, false, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// Test case 3: Enable with unlimited limit
	t.Run("EnableUnlimited", func(t *testing.T) {
		reqHeader := http.Header{}
		reqHeader.Set("Test-ID", "EnableUnlimited")
		client.HttpClient.Transport = addTestIDHeader(reqHeader, server.Client().Transport)

		err := client.SetMACLimit(context.Background(), 3, true, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// Helper function to add custom headers in the test transport layer.
func addTestIDHeader(headers http.Header, baseTransport http.RoundTripper) http.RoundTripper {
	return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		for key, values := range headers {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
		return baseTransport.RoundTrip(req)
	})
}

// Helper RoundTripperFunc definition.
type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
