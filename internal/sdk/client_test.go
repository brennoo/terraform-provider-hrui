package sdk_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/stretchr/testify/require"
)

func TestNewClient_SuccessfulAuthentication(t *testing.T) {
	// Simulating a successful 200 response on index.cgi (authentication check success)
	server := mockServerMock("", http.StatusOK)
	defer server.Close()

	// Create a new client
	clientObj, err := sdk.NewClient(server.URL, "testuser", "testpass", false)

	// Validate that no error occurred.
	require.NoError(t, err)
	require.NotNil(t, clientObj)
}

func TestNewClient_FailedAuthentication(t *testing.T) {
	// Simulate a response that redirects to a login page (authentication failure).
	redirectResponse := `<script>window.top.location.replace("/login.cgi")</script>`
	server := mockServerMock(redirectResponse, http.StatusOK)
	defer server.Close()

	// Attempt to create a new client when authentication should fail.
	clientObj, err := sdk.NewClient(server.URL, "invaliduser", "invalidpass", false)

	// Authentication should fail, so we expect an error.
	require.Error(t, err)
	require.Nil(t, clientObj) // Ensure the client was not created.
}

func TestClient_MakeRequest_Success(t *testing.T) {
	// Simulate a response for a generic GET request.
	responseBody := "Here is some content from a GET request!"
	server := mockServerMock(responseBody, http.StatusOK)
	defer server.Close()

	// Create an authenticated client.
	clientObj, _ := sdk.NewClient(server.URL, "testuser", "testpass", false)

	// Test the `MakeRequest` method.
	resp, err := clientObj.MakeRequest(server.URL + "/somepage.cgi")
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_SaveConfiguration_Success(t *testing.T) {
	// Simulate a successful configuration save (POST to /save.cgi).
	server := mockServerMock("", http.StatusOK)
	defer server.Close()

	// Create an authenticated client.
	clientObj, _ := sdk.NewClient(server.URL, "testuser", "testpass", false)

	// Test the `SaveConfiguration` method.
	err := clientObj.SaveConfiguration()
	require.NoError(t, err)
}

func TestClient_SaveConfiguration_Failure(t *testing.T) {
	// Set up a mock server that mimics /index.cgi authentication success and /save.cgi failure.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a proper response for authentication check (e.g., /index.cgi).
		if r.URL.Path == "/index.cgi" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Authenticated"))
			return
		}

		// Simulate a failure when hitting /save.cgi.
		if r.URL.Path == "/save.cgi" {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Error saving configuration"))
			return
		}

		// Generic 404 for any other unexpected paths.
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Create an authenticated client (simulated /index.cgi returns 200 OK).
	clientObj, err := sdk.NewClient(server.URL, "testuser", "testpass", false)
	require.NoError(t, err)
	require.NotNil(t, clientObj)

	// Test `SaveConfiguration` where the save fails (simulated /save.cgi returns 500).
	err = clientObj.SaveConfiguration()
	require.Error(t, err)

	// Update the error check to match the actual message format.
	require.Contains(t, err.Error(), "status code: 500")
}
