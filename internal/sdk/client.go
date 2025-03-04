package sdk

import (
	"crypto/md5" //#nosec G501 -- HRUI switch auth requires it
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
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
	err = client.Login()
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate HRUIClient: %w", err)
	}

	if client.HttpClient == nil {
		return nil, fmt.Errorf("HttpClient was not initialized")
	}

	return client, nil
}

// Login sets the authentication cookie for the HRUI system.
func (c *HRUIClient) Login() error {
	// Generate MD5 hash of the username + password (as per HRUI spec).
	//#nosec G401
	// nosemgrep use-of-md5
	hash := md5.Sum([]byte(c.Username + c.Password))
	cookieValue := hex.EncodeToString(hash[:])

	// Create authentication cookie
	// nosemgrep cookie-missing-secure
	authCookie := &http.Cookie{
		Name:     "admin",
		Value:    cookieValue,
		Path:     "/",
		HttpOnly: true,
	}

	// Parse the base URL
	u, err := url.Parse(c.URL)
	if err != nil {
		return fmt.Errorf("error parsing URL: %w", err)
	}

	// Set the cookie in the HTTP client jar
	c.HttpClient.Jar.SetCookies(u, []*http.Cookie{authCookie})

	// Construct the login.cgi URL
	loginURL := fmt.Sprintf("%s/login.cgi", strings.TrimSuffix(c.URL, "/"))

	// Prepare POST form data
	formData := url.Values{}
	formData.Set("username", c.Username)
	formData.Set("Response", cookieValue)
	formData.Set("language", "EN")

	// Create a POST request to login.cgi
	req, err := http.NewRequest("POST", loginURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("error creating POST request for login.cgi: %w", err)
	}

	// Add appropriate headers for the form submission
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Send the POST request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending POST request to login.cgi: %w", err)
	}
	// Ensure response body is closed properly
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			fmt.Printf("Warning: error closing response body: %v\n", cerr)
		}
	}()
	// Check for HTTP errors in the response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) // for debugging
		return fmt.Errorf("unexpected status code %d from login.cgi: %s", resp.StatusCode, string(body))
	}

	return c.ValidateAuthCookie()
}

// ValidateAuthCookie checks whether the authentication was successful.
func (c *HRUIClient) ValidateAuthCookie() error {
	authURL := fmt.Sprintf("%s/login.cgi", c.URL)

	// Execute the GET request using Request
	responseBody, err := c.Request("GET", authURL, nil, nil)
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
		return fmt.Errorf("authentication failed: redirected to login page\n\n%s", string(responseBody))
	}

	return nil
}

// Request handles all HTTP methods and returns the response body as a byte slice.
func (c *HRUIClient) Request(method, endpoint string, body io.Reader, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers, if provided
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute %s request to %s: %w", method, endpoint, err)
	}

	// Ensure Body.Close() is called and its error is checked
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("warning: error closing response body: %v\n", closeErr)
		}
	}()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP request to %s returned status %d: %s", endpoint, resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// FormRequest simplifies form submissions via POST and returns the response body as a byte slice.
func (c *HRUIClient) FormRequest(endpoint string, formData url.Values) ([]byte, error) {
	formEncoded := formData.Encode()
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	// Send the POST request using the Request function
	respBody, err := c.Request("POST", endpoint, strings.NewReader(formEncoded), headers)
	if err != nil {
		return nil, err
	}

	// Check for error dialogs in the HTML response
	responseStr := string(respBody)
	if strings.Contains(responseStr, "alert.cgi?alertmsg=") {
		re := regexp.MustCompile(`alertmsg=([^']+)`)
		match := re.FindStringSubmatch(responseStr)
		if len(match) > 1 {
			alertMsg, _ := url.QueryUnescape(match[1])
			return nil, fmt.Errorf("device reported an error: %s", alertMsg)
		}

		// Generic error fallback (in case the extraction fails)
		return nil, fmt.Errorf("device reported an unknown error")
	}

	// If Autosave is enabled, save the configuration after the form request
	if c.Autosave {
		if err := c.CommitChanges(); err != nil {
			return nil, fmt.Errorf("form request succeeded, but saving configuration failed: %w", err)
		}
	}

	return respBody, nil
}

// CommitChanges saves the configuration by making a POST request to `/save.cgi`.
func (c *HRUIClient) CommitChanges() error {
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

	const maxRetries = 3
	const retryDelay = 2 * time.Second

	var lastError error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Execute the POST request using Request
		respBody, err := c.Request("POST", url, strings.NewReader("cmd=save"), headers)
		if err == nil {
			// Check if the body contains an error message (e.g., for a 500 status code)
			if strings.Contains(string(respBody), "Error saving configuration") {
				lastError = fmt.Errorf("failed to save HRUI configuration: %s", string(respBody))
			} else {
				// Save succeeded
				return nil
			}
		} else {
			lastError = fmt.Errorf("failed to save HRUI configuration: %w", err)
		}

		// If not the last attempt, wait before retrying
		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	// Return the last error after exhausting all retries
	return fmt.Errorf("save configuration failed after %d retries: %w", maxRetries, lastError)
}
