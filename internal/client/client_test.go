package client

import (
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
)

func TestClient_Authenticate(t *testing.T) {
	// Mock server to simulate authentication validation
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/index.cgi" {
			// Simulate redirection to login in the body despite a 200 OK status
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`
				<html>
				<head><title>Redirect</title></head>
				<body>
				<script type="text/javascript">
				window.top.location.replace("/login.cgi");
				</script>
				</body>
				</html>`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	// Create a test HTTP client with a cookie jar
	jar, _ := cookiejar.New(nil)
	client := &Client{
		URL:        mockServer.URL,
		Username:   "admin",
		Password:   "password123",
		HttpClient: &http.Client{Jar: jar},
	}

	// Call the Authenticate method, expecting failure due to redirection to login
	err := client.Authenticate()
	if err == nil || err.Error() != "authentication failed: redirected to login page" {
		t.Fatalf("expected authentication failure due to login redirection, got %v", err)
	}
}

func TestClient_MakeRequest(t *testing.T) {
	// Mock server to simulate a simple HTTP request
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	}))
	defer mockServer.Close()

	// Create a test HTTP client with a cookie jar
	jar, _ := cookiejar.New(nil)
	client := &Client{
		URL:        mockServer.URL,
		Username:   "admin",
		Password:   "password123",
		HttpClient: &http.Client{Jar: jar},
	}

	// Test MakeRequest method
	resp, err := client.MakeRequest(mockServer.URL + "/index.cgi")
	if err != nil {
		t.Fatalf("MakeRequest failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status OK, got %v", resp.StatusCode)
	}
}
