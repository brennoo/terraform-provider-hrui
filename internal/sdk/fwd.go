package sdk

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// StormControlEntry represents the configuration for a specific port.
type StormControlEntry struct {
	Port                     string `json:"port"`                        // Port name as a string
	BroadcastRateKbps        *int   `json:"broadcast_rate_kbps"`         // Broadcast rate in kbps, nil if "Off"
	KnownMulticastRateKbps   *int   `json:"known_multicast_rate_kbps"`   // Known Multicast, nil if "Off"
	UnknownUnicastRateKbps   *int   `json:"unknown_unicast_rate_kbps"`   // Unknown Unicast, nil if "Off"
	UnknownMulticastRateKbps *int   `json:"unknown_multicast_rate_kbps"` // Unknown Multicast, nil if "Off"
}

// StormControlConfig represents all the storm control entries in the table.
type StormControlConfig struct {
	Entries []StormControlEntry `json:"entries"`
}

// JumboFrame represents the current selected Jumbo Frame size.
type JumboFrame struct {
	FrameSize int
}

// GetStormControlStatus fetches the current storm control status from the HTML page.
func (c *HRUIClient) GetStormControlStatus(ctx context.Context) (*StormControlConfig, error) {
	respBody, err := c.Request(ctx, "GET", c.URL+"/fwd.cgi?page=storm_ctrl", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch storm control page: %w", err)
	}

	// Parse the HTML page using goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, errors.New("failed to parse HTML response")
	}

	// Define parse options for rates
	parseRateOptions := func() []ParseOption {
		return []ParseOption{
			WithSpecialCases("Auto", "Off"),
			WithReturnNilOnSpecialCases(),
		}
	}

	// Find the last table row entries in the Storm Control status table
	var entries []StormControlEntry

	doc.Find("table").Last().Find("tr").Each(func(i int, row *goquery.Selection) {
		// Skip the header row
		if i == 0 {
			return
		}

		// Extract port status from row
		cols := row.Find("td")
		if cols.Length() != 5 {
			return // Skip rows that don't have all 5 fields
		}

		port := strings.TrimSpace(cols.Eq(0).Text())
		if port == "" {
			return // Skip invalid port entries
		}

		broadcast := parseInt(strings.TrimSpace(cols.Eq(1).Text()), parseRateOptions()...)
		knownMulticast := parseInt(strings.TrimSpace(cols.Eq(2).Text()), parseRateOptions()...)
		unknownUnicast := parseInt(strings.TrimSpace(cols.Eq(3).Text()), parseRateOptions()...)
		unknownMulticast := parseInt(strings.TrimSpace(cols.Eq(4).Text()), parseRateOptions()...)

		entry := StormControlEntry{
			Port:                     port,
			BroadcastRateKbps:        broadcast,
			KnownMulticastRateKbps:   knownMulticast,
			UnknownUnicastRateKbps:   unknownUnicast,
			UnknownMulticastRateKbps: unknownMulticast,
		}

		entries = append(entries, entry)
	})

	// Return the parsed results
	return &StormControlConfig{Entries: entries}, nil
}

// SetStormControlConfig updates the storm control settings for specific ports.
func (c *HRUIClient) SetStormControlConfig(
	ctx context.Context,
	stormType string, // Type of storm control: "Broadcast", "Known Multicast", etc.
	ports []string, // Ports to apply settings to, as strings
	state bool, // Whether to enable or disable storm control.
	rate *int64, // Rate in kbps, or nil if disabling.
) error {
	// Determine state value ("On" -> 1, "Off" -> 0)
	stateValue := "0"
	if state {
		stateValue = "1"
	}

	// Build POST data
	formData := url.Values{}
	formData.Set("storm_filter", stormTypeToID(stormType))
	formData.Set("action", stateValue)
	formData.Set("cmd", "storm")

	// Add each port as a separate "portid" key in the form data.
	for _, portName := range ports {
		// Replace spaces with "+" to make it URL-safe.
		encodedPortName := strings.ReplaceAll(portName, " ", "+")
		formData.Add("portid", encodedPortName)
	}

	// Set rate only if the state is enabled
	if state && rate != nil {
		formData.Set("rate", strconv.FormatInt(*rate, 10))
	}

	respBody, err := c.FormRequest(ctx, c.URL+"/fwd.cgi?page=storm_ctrl", formData)
	if err != nil {
		return fmt.Errorf("failed to update storm control settings: %w", err)
	}

	// Parse the response body to check for errors
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return errors.New("failed to parse HTML response")
	}

	// Look for the error script indicating an invalid control rate
	errorScript := doc.Find("script").First().Text()
	if strings.Contains(errorScript, "alert.cgi?alertmsg=Invalid Control rate !!") {
		return errors.New("invalid control rate: the rate provided is outside the allowed range")
	}

	// Check for unexpected response content
	if !strings.Contains(doc.Find("title").First().Text(), "Storm Control") {
		return errors.New("unexpected response while updating storm control settings")
	}

	return nil
}

// GetPortMaxRate retrieves the maximum allowed traffic rate (kbps) for a specific port
// using the provided human-readable port name (e.g., "Port 1") rather than port ID.
func (c *HRUIClient) GetPortMaxRate(ctx context.Context, portName string) (int64, error) {
	// Fetch the storm control HTML page.
	respBody, err := c.Request(ctx, "GET", c.URL+"/fwd.cgi?page=storm_ctrl", nil, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch storm control page: %w", err)
	}

	// Parse the HTML content using goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return 0, fmt.Errorf("error parsing storm control page: %w", err)
	}

	// Locate the table row for the specified port name and extract the rate column's text.
	rateText, err := findRateText(doc, portName)
	if err != nil {
		return 0, fmt.Errorf("failed to get rate text for port '%s': %w", portName, err)
	}

	// Extract the maximum rate from the rate text.
	maxRate, err := extractMaxRate(rateText)
	if err != nil {
		return 0, fmt.Errorf("failed to extract max rate for port '%s': %w", portName, err)
	}

	return maxRate, nil
}

// findRateText locates the rate text (e.g., "1-10000000(kbps)") for the specified port name.
func findRateText(doc *goquery.Document, portName string) (string, error) {
	var rateText string
	found := false

	// Iterate over table rows to find the port's row.
	doc.Find("tr").EachWithBreak(func(i int, row *goquery.Selection) bool {
		// Check if this row contains the specified port name.
		if strings.Contains(row.Text(), portName) {
			// Locate the cell containing "(kbps)" text.
			row.Find("td").EachWithBreak(func(j int, cell *goquery.Selection) bool {
				if strings.Contains(cell.Text(), "(kbps)") {
					rateText = strings.TrimSpace(cell.Text())
					found = true
					return false // Stop iterating over cells.
				}
				return true // Continue searching cells.
			})
			return false // Stop iterating over rows.
		}
		return true // Continue searching rows.
	})

	// If the rate text was not found, return an error.
	if !found {
		return "", fmt.Errorf("rate information not found for port '%s'", portName)
	}

	return rateText, nil
}

// extractMaxRate parses and extracts the maximum rate value (e.g., 2500000)
// from a rate string like "(1-2500000)(kbps)".
func extractMaxRate(rateText string) (int64, error) {
	// Normalize the input to remove unexpected spaces.
	rateText = strings.TrimSpace(rateText)

	// Update the regex to match "(1-XXXXXX)(kbps)"
	re := regexp.MustCompile(`\(?1-(\d+)\)?\(kbps\)`)

	matches := re.FindStringSubmatch(rateText)

	if len(matches) < 2 {
		return 0, fmt.Errorf("failed to extract rate information from text: %s", rateText)
	}

	// Convert the extracted max rate to int64
	maxRate, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("error converting max rate to integer: %w", err)
	}

	return maxRate, nil
}

// Helper function to convert storm type to its corresponding ID.
func stormTypeToID(stormType string) string {
	switch strings.ToLower(stormType) {
	case "broadcast":
		return "3"
	case "known multicast":
		return "2"
	case "unknown unicast":
		return "0"
	case "unknown multicast":
		return "1"
	default:
		return "3" // Default to "Broadcast"
	}
}

// GetJumboFrame retrieves the current Jumbo Frame configuration from the HTML page.
func (c *HRUIClient) GetJumboFrame(ctx context.Context) (*JumboFrame, error) {
	// URL to the Jumbo Frame page
	urlJumbo := fmt.Sprintf("%s/fwd.cgi?page=jumboframe", c.URL)

	// Perform HTTP GET request
	respBody, err := c.Request(ctx, "GET", urlJumbo, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Jumbo Frame page: %w", err)
	}

	// Parse the HTML response using goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse Jumbo Frame HTML: %w", err)
	}

	// Look for the currently selected option in the <select> element
	selectedOption := doc.Find("select[name='jumboframe'] option[selected]").Text()
	selectedOption = strings.TrimSpace(selectedOption)

	// Convert the selected option to an integer (frame size)
	frameSize, err := strconv.Atoi(selectedOption)
	if err != nil {
		return nil, fmt.Errorf("failed to parse selected Jumbo Frame size ('%s'): %w", selectedOption, err)
	}

	// Return the current Jumbo Frame
	return &JumboFrame{FrameSize: frameSize}, nil
}

// SetJumboFrame sets the Jumbo Frame size on the device.
func (c *HRUIClient) SetJumboFrame(ctx context.Context, frameSize int) error {
	// Map the FrameSize to its corresponding dropdown value
	frameSizeValue := map[int]string{
		1522:  "0",
		1536:  "1",
		1552:  "2",
		9216:  "3",
		16383: "4",
	}[frameSize]

	// Ensure the frameSize is valid, and the value exists in the mapping
	if frameSizeValue == "" {
		return fmt.Errorf("invalid Jumbo Frame size '%d': supported sizes are 1522, 1536, 1552, 9216, 16383", frameSize)
	}

	// Construct the POST form data
	formData := url.Values{}
	formData.Set("cmd", "jumboframe")
	formData.Set("jumboframe", frameSizeValue) // Selected value from the mapping

	// URL of the Jumbo Frame page
	endpoint := fmt.Sprintf("%s/fwd.cgi?page=jumboframe", c.URL)

	// Send the POST request
	respBody, err := c.FormRequest(ctx, endpoint, formData)
	if err != nil {
		return fmt.Errorf("failed to set Jumbo Frame size '%d': %w", frameSize, err)
	}

	// Parse the response to check for issues
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return fmt.Errorf("failed to parse response after setting Jumbo Frame: %w", err)
	}

	// Check for an error indication in the response
	if !strings.Contains(doc.Find("title").Text(), "Jumbo Frame Setting") {
		return errors.New("unexpected response after setting Jumbo Frame")
	}

	return nil
}
