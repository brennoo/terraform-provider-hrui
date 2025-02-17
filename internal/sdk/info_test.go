package sdk

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSystemInfo(t *testing.T) {
	tests := []struct {
		name       string
		html       string
		expected   map[string]string
		shouldFail bool
		httpStatus int
	}{
		{
			"Valid system info",
			`<table>
				<tr><th>OS</th><td>Linux</td></tr>
				<tr><th>Version</th><td>1.2.3</td></tr>
			</table>`,
			map[string]string{
				"OS":      "Linux",
				"Version": "1.2.3",
			},
			false,
			http.StatusOK,
		},
		{
			"Empty table",
			`<table></table>`,
			map[string]string{},
			false,
			http.StatusOK,
		},
		{
			"Request error",
			"",
			nil,
			true,
			http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.httpStatus)
				if _, err := w.Write([]byte(tt.html)); err != nil {
					t.Errorf("failed to write response: %v", err)
				}
			}))
			defer server.Close()

			client := &HRUIClient{URL: server.URL, HttpClient: server.Client()}
			info, err := client.GetSystemInfo()

			if tt.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, info)
			}
		})
	}
}
