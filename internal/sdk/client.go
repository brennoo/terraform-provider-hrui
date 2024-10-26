package sdk

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

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
			Jar: jar,
		},
	}

	// Authenticate the client
	err = client.Authenticate()
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate HRUIClient: %w", err)
	}

	return client, nil
}

// Authenticate sets the authentication cookie for the HRUI system.
func (c *HRUIClient) Authenticate() error {
	// Generate MD5 hash of the username + password (as per HRUI spec).
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

	resp, err := c.MakeRequest(authURL)
	if err != nil {
		return fmt.Errorf("authentication validation request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check if a redirection to the login page (indicating failed authentication) occurs
	if resp.StatusCode == http.StatusOK {
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return fmt.Errorf("error parsing response body with goquery: %w", err)
		}

		// Check for any sign of being redirected to a login page via a script tag
		scriptContent := ""
		doc.Find("script").Each(func(i int, s *goquery.Selection) {
			scriptContent = s.Text()
		})

		if strings.Contains(scriptContent, `window.top.location.replace("/login.cgi")`) {
			return fmt.Errorf("authentication failed: redirected to login page")
		}
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// MakeRequest performs a simple GET request and returns the response.
func (c *HRUIClient) MakeRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating GET request: %w", err)
	}

	// Perform the GET request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making GET request: %w", err)
	}

	return resp, nil
}

// SaveConfiguration saves the configuration by making a POST request to `/save.cgi`.
func (c *HRUIClient) SaveConfiguration() error {
	url := fmt.Sprintf("%s/save.cgi", c.URL)
	httpReq, err := http.NewRequest(http.MethodPost, url, strings.NewReader("cmd=save"))
	if err != nil {
		return fmt.Errorf("error creating save request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpResp, err := c.HttpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to save HRUI configuration: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to save HRUI configuration, status code: %d", httpResp.StatusCode)
	}

	return nil
}
