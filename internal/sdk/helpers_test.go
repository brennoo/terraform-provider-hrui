package sdk_test

import (
	"net/http"
	"net/http/httptest"
)

// mockServerMock is a utility function to create a mock server that returns specific responses.
func mockServerMock(response string, code int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		_, _ = w.Write([]byte(response))
	}))
}
