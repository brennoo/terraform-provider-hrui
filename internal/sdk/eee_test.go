package sdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetEEE(t *testing.T) {
	// Mock HTML response for EEE page
	mockHTML := `
		<html>
		<body>
			<form method="post" name="eee" action="/eee.cgi">
				<select name="func_type">
					<option value="0">Disable</option>
					<option value="1" selected>Enable</option>
				</select>
			</form>
		</body>
		</html>`

	// Mock server to simulate EEE settings page
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
	enabled, err := client.GetEEE(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected state is "Enable" (true) from HTML
	expected := true
	if enabled != expected {
		t.Errorf("expected EEE to be %v, got %v", expected, enabled)
	}
}

func TestSetEEE(t *testing.T) {
	// Mock server to capture and verify POST requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %v", r.Method)
		}

		// Parse the POST form data
		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse form data: %v", err)
		}

		funcType := r.FormValue("func_type")
		cmd := r.FormValue("cmd")

		// Assert the form data values
		if cmd != "loop" {
			t.Errorf("expected cmd to be 'loop', got %v", cmd)
		}

		// Respond with success HTML
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<title>EEE Setting</title>`))

		// Check if EEE is being enabled or disabled
		expectedFuncType := "1" // Enable
		if funcType != expectedFuncType {
			t.Errorf("expected func_type to be '%s', got '%s'", expectedFuncType, funcType)
		}
	}))
	defer server.Close()

	client := &HRUIClient{HttpClient: server.Client(), URL: server.URL}

	// Enable EEE
	err := client.SetEEE(context.Background(), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetEEEDisable(t *testing.T) {
	// Mock server to capture and verify POST requests for disabling EEE
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %v", r.Method)
		}

		// Parse the POST form data
		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse form data: %v", err)
		}

		funcType := r.FormValue("func_type")
		cmd := r.FormValue("cmd")

		// Assert the form data values
		if cmd != "loop" {
			t.Errorf("expected cmd to be 'loop', got %v", cmd)
		}

		// Respond with success HTML
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<title>EEE Setting</title>`))

		// Check if EEE is being disabled
		expectedFuncType := "0" // Disable
		if funcType != expectedFuncType {
			t.Errorf("expected func_type to be '%s', got '%s'", expectedFuncType, funcType)
		}
	}))
	defer server.Close()

	client := &HRUIClient{HttpClient: server.Client(), URL: server.URL}

	// Disable EEE
	err := client.SetEEE(context.Background(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetEEEParseError(t *testing.T) {
	// Mock faulty HTML response for EEE page
	mockHTML := `
		<html>
		<body>
			<form method="post" name="eee" action="/eee.cgi">
				<select name="func_type">
					<!-- Missing options or selected attribute -->
				</select>
			</form>
		</body>
		</html>`

	// Mock server
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
	_, err := client.GetEEE(context.Background())
	if err == nil {
		t.Fatalf("expected error due to missing selected option, got nil")
	}

	// Assert error message
	expectedError := "failed to find selected EEE status in HTML"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error message to contain '%s', got '%v'", expectedError, err.Error())
	}
}
