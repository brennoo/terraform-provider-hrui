package sdk

import (
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
		if r.URL.Path == "/index.cgi" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(""))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	clientObj, err := NewClient(server.URL, "testuser", "testpass", false)
	require.NoError(t, err)
	require.NotNil(t, clientObj)
	require.NotNil(t, clientObj.HttpClient)
}

func TestNewClient_FailedAuthentication(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/index.cgi" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<script>window.top.location.replace("/login.cgi")</script>`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	clientObj, err := NewClient(server.URL, "invaliduser", "invalidpass", false)
	require.Error(t, err)
	require.Nil(t, clientObj)
}

func TestClient_CommitChanges_Success(t *testing.T) {
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

	clientObj := createAuthenticatedClient(t, server)
	err := clientObj.CommitChanges()
	require.NoError(t, err)
}

func TestClient_CommitChanges_Failure(t *testing.T) {
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

	clientObj := createAuthenticatedClient(t, server)
	err := clientObj.CommitChanges()
	require.Error(t, err)
	require.Contains(t, err.Error(), "returned status 500: Error saving configuration")
}

func TestClient_Request_Success(t *testing.T) {
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
	respBody, err := clientObj.Request("GET", fmt.Sprintf("%s/test", server.URL), nil, nil)
	require.NoError(t, err)
	require.Equal(t, "Success", string(respBody))
}

func TestClient_Request_Failure(t *testing.T) {
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
	_, err := clientObj.Request("GET", fmt.Sprintf("%s/test", server.URL), nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "/test returned status 404")
}

func TestClient_FormRequest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/index.cgi":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(""))
		case "/test":
			if r.Method == "POST" {
				err := r.ParseForm()
				require.NoError(t, err)

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

	respBody, err := clientObj.FormRequest(fmt.Sprintf("%s/test", server.URL), formData)
	require.NoError(t, err)
	require.Equal(t, "Success", string(respBody))
}

func TestClient_FormRequest_Failure(t *testing.T) {
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

	_, err := clientObj.FormRequest(fmt.Sprintf("%s/test", server.URL), formData)
	require.Error(t, err)
	require.Contains(t, err.Error(), "/test returned status 404")
}
