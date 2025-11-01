package sdk

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GetEEE fetches the current EEE (Energy Efficient Ethernet) status from the device.
// Returns `true` if EEE is enabled, `false` if disabled.
func (c *HRUIClient) GetEEE(ctx context.Context) (bool, error) {
	// Issue a GET request to `/eee.cgi`
	endpoint := fmt.Sprintf("%s/eee.cgi", c.URL)
	responseBody, err := c.Request(ctx, "GET", endpoint, nil, nil)
	if err != nil {
		return false, fmt.Errorf("failed to fetch EEE status: %w", err)
	}

	// Parse the HTML response to determine the current EEE status using goquery
	return parseEEEHTMLWithGoQuery(responseBody)
}

// SetEEE updates the EEE (Energy Efficient Ethernet) status on the device.
// Pass `true` to enable EEE or `false` to disable it.
func (c *HRUIClient) SetEEE(ctx context.Context, enabled bool) error {
	// Construct the POST form data
	funcType := "0" // Default to Disable
	if enabled {
		funcType = "1" // Enable
	}
	formData := url.Values{}
	formData.Set("func_type", funcType)
	formData.Set("cmd", "loop") // Required field per the HTML form

	// Issue a POST request to `/eee.cgi`
	endpoint := fmt.Sprintf("%s/eee.cgi", c.URL)
	_, err := c.FormRequest(ctx, endpoint, formData)
	if err != nil {
		return fmt.Errorf("failed to update EEE status: %w", err)
	}

	return nil
}

// parseEEEHTMLWithGoQuery parses the HTML response from `/eee.cgi` to determine the current EEE status.
func parseEEEHTMLWithGoQuery(body []byte) (bool, error) {
	// Load the HTML document using GoQuery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return false, fmt.Errorf("failed to parse HTML document: %w", err)
	}

	// Find the <select name="func_type"> element to determine the selected value
	selectedValue := ""
	doc.Find("select[name='func_type'] option").Each(func(i int, s *goquery.Selection) {
		// Check if this option is "selected"
		_, exists := s.Attr("selected")
		if exists {
			selectedValue = strings.TrimSpace(s.Text())
		}
	})

	// Validate that a selected value was found
	if selectedValue == "" {
		return false, fmt.Errorf("failed to find selected EEE status in HTML")
	}

	// Interpret the selected value ("Disable" = false, "Enable" = true)
	switch strings.ToLower(selectedValue) {
	case "disable":
		return false, nil
	case "enable":
		return true, nil
	default:
		return false, fmt.Errorf("unexpected value for EEE status: %s", selectedValue)
	}
}
