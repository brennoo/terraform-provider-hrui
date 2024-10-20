package client

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

// Simple client for interacting with the HRUI web UI
type Client struct {
	URL        string
	Username   string
	Password   string
	Autosave   bool
	HttpClient *http.Client
}

// NewClient creates a new authenticated client.
func NewClient(url, username, password string, autosave bool) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	c := &Client{
		URL:      url,
		Username: username,
		Password: password,
		Autosave: autosave,
		HttpClient: &http.Client{
			Jar: jar,
		},
	}

	err = c.Authenticate()
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate client: %w", err)
	}

	return c, nil
}

// Authenticate authenticates the client by setting the authentication cookie
func (c *Client) Authenticate() error {
	// Generate MD5 hash of username+password
	hash := md5.Sum([]byte(c.Username + c.Password))
	cookieValue := hex.EncodeToString(hash[:])

	// Create authentication cookie
	authCookie := &http.Cookie{
		Name:  "admin",
		Value: cookieValue,
		Path:  "/",
	}

	// Set the cookie directly using the URL
	u, err := url.Parse(c.URL)
	if err != nil {
		return fmt.Errorf("error parsing URL: %w", err)
	}
	c.HttpClient.Jar.SetCookies(u, []*http.Cookie{authCookie})

	// Validate the authentication by making a request to index.cgi or /
	return c.validateAuth()
}

// validateAuth uses MakeRequest to check if the authentication was successful
func (c *Client) validateAuth() error {
	authURL := c.URL + "/index.cgi" // You can change this to just "/" if needed

	resp, err := c.MakeRequest(authURL)
	if err != nil {
		return fmt.Errorf("authentication validation request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP 200 but potential redirection to login
	if resp.StatusCode == http.StatusOK {
		// Use goquery to parse the response body
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return fmt.Errorf("error parsing response body with goquery: %w", err)
		}

		// Look for a script tag that contains the redirection to login page
		scriptContent := ""
		doc.Find("script").Each(func(i int, s *goquery.Selection) {
			scriptContent = s.Text()
		})

		// Check if the script tag contains a redirection to login
		if strings.Contains(scriptContent, `window.top.location.replace("/login.cgi")`) {
			return fmt.Errorf("authentication failed: redirected to login page")
		}
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// MakeRequest sends a GET request and returns the response
func (c *Client) MakeRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	return resp, nil
}

// SaveConfiguration makes a POST request to /save.cgi to save the configuration
func (c *Client) SaveConfiguration() error {
	url := fmt.Sprintf("%s/save.cgi", c.URL)
	httpReq, err := http.NewRequest(http.MethodPost, url, strings.NewReader("cmd=save"))
	if err != nil {
		return fmt.Errorf("unable to create save request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpResp, err := c.HttpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("unable to save HRUI configuration, got error: %w", err)
	}

	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("unable to save HRUI configuration, got HTTP status code: %d", httpResp.StatusCode)
	}

	return nil
}
