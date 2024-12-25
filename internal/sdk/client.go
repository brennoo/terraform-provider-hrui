package sdk

import (
	"bufio"
	"bytes"
	"crypto/md5" //#nosec G501 -- HRUI switch auth requires it
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Client handles communication with the HRUI device, managing VLANs and other networking functionality.
type HRUIClient struct {
	URL        string
	Username   string
	Password   string
	Autosave   bool
	HttpClient *http.Client
}

// NewClient initializes and authenticates a new HRUIClient.
func NewClient(url, username, password string, autosave bool) (*HRUIClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	// Initialize the client
	client := &HRUIClient{
		URL:      url,
		Username: username,
		Password: password,
		Autosave: autosave,
		HttpClient: &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
		},
	}

	// Authenticate the client
	err = client.Authenticate()
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate HRUIClient: %w", err)
	}

	if client.HttpClient == nil {
		return nil, fmt.Errorf("HttpClient was not initialized")
	}

	return client, nil
}

// Authenticate sets the authentication cookie for the HRUI system.
func (c *HRUIClient) Authenticate() error {
	// Generate MD5 hash of the username + password (as per HRUI spec).
	//#nosec G401
	hash := md5.Sum([]byte(c.Username + c.Password))
	cookieValue := hex.EncodeToString(hash[:])

	// Create authentication cookie
	authCookie := &http.Cookie{
		Name:  "admin",
		Value: cookieValue,
		Path:  "/",
	}

	// Set the cookie directly using the parsed URL
	u, err := url.Parse(c.URL)
	if err != nil {
		return fmt.Errorf("error parsing URL: %w", err)
	}
	c.HttpClient.Jar.SetCookies(u, []*http.Cookie{authCookie})

	// Validate the authentication
	return c.validateAuth()
}

// validateAuth checks whether the authentication was successful.
func (c *HRUIClient) validateAuth() error {
	authURL := fmt.Sprintf("%s/index.cgi", c.URL)

	// Execute the GET request using ExecuteRequest
	responseBody, err := c.ExecuteRequest("GET", authURL, nil, nil)
	if err != nil {
		return fmt.Errorf("authentication validation request failed: %w", err)
	}

	// Parse the response body for validation
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(responseBody)))
	if err != nil {
		return fmt.Errorf("error parsing response body with goquery: %w", err)
	}

	// Check for redirection to the login page in a <script> tag
	scriptContent := ""
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		scriptContent = s.Text()
	})

	if strings.Contains(scriptContent, `window.top.location.replace("/login.cgi")`) {
		return fmt.Errorf("authentication failed: redirected to login page")
	}

	return nil
}

// ExecuteRequest handles all HTTP methods and returns the response body as a byte slice.
func (client *HRUIClient) ExecuteRequest(method, endpoint string, body io.Reader, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers, if provided
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute %s request to %s: %w", method, endpoint, err)
	}

	// Ensure Body.Close() is called and its error is checked
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log or return the close error if necessary
			fmt.Printf("warning: error closing response body: %v\n", closeErr)
		}
	}()

	// Use a scanner to read the response body in chunks
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 1024), 1024*1024) // Adjust buffer size as needed

	var respBody bytes.Buffer
	for scanner.Scan() {
		respBody.Write(scanner.Bytes())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP request to %s returned status %d: %s", endpoint, resp.StatusCode, string(respBody.Bytes()))
	}

	return respBody.Bytes(), nil
}

// ExecuteFormRequest simplifies form submissions via POST and returns the response body as a byte slice.
func (client *HRUIClient) ExecuteFormRequest(endpoint string, formData url.Values) ([]byte, error) {
	formEncoded := formData.Encode()
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	return client.ExecuteRequest("POST", endpoint, strings.NewReader(formEncoded), headers)
}

// SaveConfiguration saves the configuration by making a POST request to `/save.cgi`.
func (c *HRUIClient) SaveConfiguration() error {
	if c == nil {
		return fmt.Errorf("HRUIClient is nil")
	}

	if c.HttpClient == nil {
		return fmt.Errorf("HttpClient is nil in HRUIClient")
	}

	url := fmt.Sprintf("%s/save.cgi", c.URL)
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	// Execute the POST request using ExecuteRequest
	respBody, err := c.ExecuteRequest("POST", url, strings.NewReader("cmd=save"), headers)
	if err != nil {
		return fmt.Errorf("failed to save HRUI configuration: %w", err)
	}

	// Check if the body contains an error message (e.g., for a 500 status code)
	if strings.Contains(string(respBody), "Error saving configuration") {
		return fmt.Errorf("failed to save HRUI configuration: %s", string(respBody))
	}

	return nil
}
