package sdk

import (
	"crypto/md5" //#nosec G501 -- HRUI switch auth requires it
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func createAuthenticatedClient(t *testing.T, server *httptest.Server) *HRUIClient {
	clientObj, err := NewClient(server.URL, "testuser", "testpass", false)
	require.NoError(t, err)
	require.NotNil(t, clientObj)
	require.NotNil(t, clientObj.HttpClient)
	return clientObj
}

func TestNewClient_SuccessfulAuthentication(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login.cgi":
			if r.Method == "POST" {
				fmt.Println("Handling POST request to /login.cgi")
				err := r.ParseForm()
				require.NoError(t, err) // Ensure no form parsing errors

				username := r.FormValue("username")
				response := r.FormValue("Response")
				language := r.FormValue("language")
				//#nosec G401
				// nosemgrep use-of-md5
				expectedResponse := fmt.Sprintf("%x", md5.Sum([]byte("testuser"+"testpass")))

				if username == "testuser" && response == expectedResponse && language == "EN" {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("Login successful"))
				} else {
					w.WriteHeader(http.StatusUnauthorized)
					_, _ = w.Write([]byte("Unauthorized"))
				}
			} else if r.Method == "GET" {
				fmt.Println("Handling GET request to /login.cgi")

				// Simulate session validation
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`<html><body><h1>Session is valid</h1></body></html>`))
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}

		case "/index.cgi":
			fmt.Println("Handling /index.cgi")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Welcome"))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	clientObj, err := NewClient(server.URL, "testuser", "testpass", false)
	require.NoError(t, err, "Failed to authenticate the client")
	require.NotNil(t, clientObj)
	require.NotNil(t, clientObj.HttpClient)
}

func TestNewClient_FailedAuthentication(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login.cgi" {
			if r.Method == "POST" {
				// Simulate invalid credentials by returning 401 Unauthorized
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`<script>window.top.location.replace("/login.cgi")</script>`))
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed) // Handle non-POST methods correctly
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	clientObj, err := NewClient(server.URL, "invaliduser", "invalidpass", false)
	require.Error(t, err)
	require.Nil(t, clientObj)
	require.Contains(t, err.Error(), "failed to authenticate HRUIClient")
}

func TestClient_CommitChanges_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login.cgi":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Login successful")) // Simulate login success
		case "/save.cgi":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Configuration saved"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	clientObj := createAuthenticatedClient(t, server)
	err := clientObj.CommitChanges()
	require.NoError(t, err)
}

func TestClient_CommitChanges_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login.cgi":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Login successful"))
		case "/save.cgi":
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Error saving configuration"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	clientObj := createAuthenticatedClient(t, server)
	err := clientObj.CommitChanges()
	require.Error(t, err)
	require.Contains(t, err.Error(), "Error saving configuration")
}

func TestClient_Request_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login.cgi":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Login successful"))
		case "/test":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Success"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	clientObj := createAuthenticatedClient(t, server)
	respBody, err := clientObj.Request("GET", fmt.Sprintf("%s/test", server.URL), nil, nil)
	require.NoError(t, err)
	require.Equal(t, "Success", string(respBody))
}

func TestClient_Request_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login.cgi":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Login successful"))
		case "/test":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	clientObj := createAuthenticatedClient(t, server)
	_, err := clientObj.Request("GET", fmt.Sprintf("%s/test", server.URL), nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "/test returned status 404")
}

func TestClient_FormRequest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login.cgi":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Login successful"))
		case "/test":
			if r.Method == "POST" {
				_ = r.ParseForm()
				if r.PostFormValue("param1") == "value1" && r.PostFormValue("param2") == "value2" {
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

	respBody, err := clientObj.FormRequest(fmt.Sprintf("%s/test", server.URL), formData)
	require.NoError(t, err)
	require.Equal(t, "Success", string(respBody))
}

func TestClient_FormRequest_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login.cgi":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Login successful"))
		case "/test":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	clientObj := createAuthenticatedClient(t, server)
	formData := url.Values{}
	formData.Set("param1", "value1")

	_, err := clientObj.FormRequest(fmt.Sprintf("%s/test", server.URL), formData)
	require.Error(t, err)
	require.Contains(t, err.Error(), "/test returned status 404")
}
