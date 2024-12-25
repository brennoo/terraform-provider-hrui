package sdk_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"

	"github.com/stretchr/testify/require"
)

func createAuthenticatedClient(t *testing.T, server *httptest.Server) *sdk.HRUIClient {
	clientObj, err := sdk.NewClient(server.URL, "testuser", "testpass", false)
	require.NoError(t, err)
	require.NotNil(t, clientObj)
	require.NotNil(t, clientObj.HttpClient)
	return clientObj
}

func TestNewClient_SuccessfulAuthentication(t *testing.T) {
	// Simulating a successful 200 response on index.cgi (authentication check success)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/index.cgi" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(""))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create a new client
	clientObj, err := sdk.NewClient(server.URL, "testuser", "testpass", false)

	// Validate that no error occurred and the client is not nil
	require.NoError(t, err)
	require.NotNil(t, clientObj)

	// Ensure HttpClient is not nil before using it
	require.NotNil(t, clientObj.HttpClient)
}

func TestNewClient_FailedAuthentication(t *testing.T) {
	// Simulate a response that redirects to a login page (authentication failure).
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/index.cgi" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<script>window.top.location.replace("/login.cgi")</script>`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Attempt to create a new client when authentication should fail.
	clientObj, err := sdk.NewClient(server.URL, "invaliduser", "invalidpass", false)

	// Authentication should fail, so we expect an error.
	require.Error(t, err)
	require.Nil(t, clientObj) // Ensure the client was not created.
}

func TestClient_SaveConfiguration_Success(t *testing.T) {
	// Simulate a successful configuration save (POST to /save.cgi).
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/index.cgi":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(""))
		case "/save.cgi":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(""))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create an authenticated client.
	clientObj := createAuthenticatedClient(t, server)

	// Test the `SaveConfiguration` method.
	err := clientObj.SaveConfiguration()
	require.NoError(t, err)
}

func TestClient_SaveConfiguration_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/index.cgi":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(""))
		case "/save.cgi":
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Error saving configuration"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create an authenticated client.
	clientObj := createAuthenticatedClient(t, server)

	// Test `SaveConfiguration` where the save fails (simulated /save.cgi returns 500).
	err := clientObj.SaveConfiguration()
	require.Error(t, err)

	// Update the error check to match the actual message format.
	require.Contains(t, err.Error(), "returned status 500: Error saving configuration")
}

func TestClient_ExecuteRequest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/index.cgi":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(""))
		case "/test":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Success"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	clientObj := createAuthenticatedClient(t, server)

	// Execute a GET request
	respBody, err := clientObj.ExecuteRequest("GET", fmt.Sprintf("%s/test", server.URL), nil, nil)
	require.NoError(t, err)
	require.Equal(t, "Success", string(respBody))
}

func TestClient_ExecuteRequest_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/index.cgi" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(""))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	clientObj := createAuthenticatedClient(t, server)

	// Execute a GET request
	_, err := clientObj.ExecuteRequest("GET", fmt.Sprintf("%s/test", server.URL), nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "/test returned status 404")
}

func TestClient_ExecuteFormRequest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/index.cgi":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(""))
		case "/test":
			if r.Method == "POST" && r.URL.Path == "/test" {
				err := r.ParseForm()
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				if r.Form.Get("param1") == "value1" && r.Form.Get("param2") == "value2" {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("Success"))
				} else {
					w.WriteHeader(http.StatusBadRequest)
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	clientObj := createAuthenticatedClient(t, server)

	formData := url.Values{}
	formData.Set("param1", "value1")
	formData.Set("param2", "value2")

	// Execute a POST request with form data
	respBody, err := clientObj.ExecuteFormRequest(fmt.Sprintf("%s/test", server.URL), formData)
	require.NoError(t, err)
	require.Equal(t, "Success", string(respBody))
}

func TestClient_ExecuteFormRequest_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/index.cgi" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(""))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	clientObj := createAuthenticatedClient(t, server)

	formData := url.Values{}
	formData.Set("param1", "value1")

	// Execute a POST request with form data
	_, err := clientObj.ExecuteFormRequest(fmt.Sprintf("%s/test", server.URL), formData)
	require.Error(t, err)
	require.Contains(t, err.Error(), "/test returned status 404")
}
