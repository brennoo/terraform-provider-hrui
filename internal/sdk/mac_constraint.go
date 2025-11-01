package sdk

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// MACLimit represents the MAC entry limit for a specific port.
type MACLimit struct {
	Port    string
	Enabled bool
	Limit   *int
}

// GetMACLimits fetches the current MAC limits configuration for all ports.
func (c *HRUIClient) GetMACLimits(ctx context.Context) ([]MACLimit, error) {
	// Execute a GET request to retrieve the MAC constraints HTML page.
	respBody, err := c.Request(ctx, "GET", fmt.Sprintf("%s/mac_constraint.cgi", c.URL), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch MAC constraints page: %w", err)
	}

	// Parse the HTML using goquery.
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse MAC constraints HTML: %w", err)
	}

	// Extract MAC limits from the table at the bottom of the page.
	var macLimits []MACLimit
	doc.Find("table").Last().Find("tr").Each(func(i int, selection *goquery.Selection) {
		// Skip the header row.
		if i == 0 {
			return
		}

		columns := selection.Find("td")
		port := strings.TrimSpace(columns.Eq(0).Text())
		limitStr := strings.TrimSpace(columns.Eq(1).Text())

		// Determine if the limit is "Unlimited" or a numeric value.
		var limit *int
		enabled := limitStr != "Unlimited"
		if enabled {
			limitValue, err := strconv.Atoi(limitStr)
			if err == nil {
				limit = &limitValue
			}
		}

		// Append the data to the macLimits slice.
		macLimits = append(macLimits, MACLimit{
			Port:    port,
			Enabled: enabled,
			Limit:   limit,
		})
	})

	return macLimits, nil
}

// SetMACLimit sets the MAC limit for a specific port.
func (c *HRUIClient) SetMACLimit(ctx context.Context, portID int, enabled bool, limit *int) error {
	// Build the form data.
	formData := url.Values{}
	formData.Set("cmd", "mac_constraint")
	formData.Set("portid", strconv.Itoa(portID))
	if enabled {
		formData.Set("state", "1") // Enable the limit.
		if limit != nil {
			formData.Set("limit", strconv.Itoa(*limit)) // Set the limit value.
		} else {
			formData.Set("limit", "Unlimited") // No numeric limit.
		}
	} else {
		formData.Set("state", "0")         // Disable the limit.
		formData.Set("limit", "Unlimited") // Set it back to "Unlimited".
	}

	// Send the POST request to apply changes.
	respBody, err := c.FormRequest(ctx, fmt.Sprintf("%s/mac_constraint.cgi", c.URL), formData)
	if err != nil {
		return fmt.Errorf("failed to update MAC constraints: %w", err)
	}

	// Optionally, handle any errors reported in the response.
	if strings.Contains(string(respBody), "Error") {
		return fmt.Errorf("device reported an error when setting MAC constraints: %s", string(respBody))
	}

	return nil
}
