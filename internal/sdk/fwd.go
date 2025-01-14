package sdk

import (
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
	Port                     int  `json:"port"`                        // Port number as an integer
	BroadcastRateKbps        *int `json:"broadcast_rate_kbps"`         // Broadcast rate in kbps, nil if "Off"
	KnownMulticastRateKbps   *int `json:"known_multicast_rate_kbps"`   // Known Multicast, nil if "Off"
	UnknownUnicastRateKbps   *int `json:"unknown_unicast_rate_kbps"`   // Unknown Unicast, nil if "Off"
	UnknownMulticastRateKbps *int `json:"unknown_multicast_rate_kbps"` // Unknown Multicast, nil if "Off"
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
func (c *HRUIClient) GetStormControlStatus() (*StormControlConfig, error) {
	respBody, err := c.Request("GET", c.URL+"/fwd.cgi?page=storm_ctrl", nil, nil)
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

		portString := strings.TrimSpace(cols.Eq(0).Text())
		port := parseInt(portString, WithTrimPrefix("Port "))
		if port == nil {
			return // Skip invalid port entries
		}

		broadcast := parseInt(strings.TrimSpace(cols.Eq(1).Text()), parseRateOptions()...)
		knownMulticast := parseInt(strings.TrimSpace(cols.Eq(2).Text()), parseRateOptions()...)
		unknownUnicast := parseInt(strings.TrimSpace(cols.Eq(3).Text()), parseRateOptions()...)
		unknownMulticast := parseInt(strings.TrimSpace(cols.Eq(4).Text()), parseRateOptions()...)

		entry := StormControlEntry{
			Port:                     *port,
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
	stormType string, // Type of storm control: "Broadcast", "Known Multicast", etc.
	ports []int, // Ports to apply settings to, as integers.
	state bool, // Whether to enable or disable storm control.
	rate *int64, // Rate in kbps, or nil if disabling.
) error {
	// Convert integer ports to backend's string format
	var backendPorts []string
	for _, port := range ports {
		backendPorts = append(backendPorts, intToBackendPort(port))
	}

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

	// Set rate only if the state is enabled
	if state && rate != nil {
		formData.Set("rate", strconv.FormatInt(*rate, 10))
	}

	// Add ports
	formData.Set("portid", strings.Join(backendPorts, ","))

	respBody, err := c.FormRequest(c.URL+"/fwd.cgi?page=storm_ctrl", formData)
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
// from the "Rate (kbps)" column in the storm control HTML page.
func (c *HRUIClient) GetPortMaxRate(port int) (int64, error) {
	portString := intToBackendPort(port)

	respBody, err := c.Request("GET", c.URL+"/fwd.cgi?page=storm_ctrl", nil, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch storm control page: %w", err)
	}

	// Parse the HTML content using goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return 0, fmt.Errorf("error parsing storm control page: %w", err)
	}

	// Look for the row corresponding to the given port
	var rateText string
	doc.Find("tr").Each(func(i int, s *goquery.Selection) {
		// Check if this row contains the specified port string
		if strings.Contains(s.Text(), portString) {
			// Look for a cell with the Rate (kbps) info (e.g., "(1-2500000)(kbps)")
			s.Find("td").Each(func(j int, td *goquery.Selection) {
				if strings.Contains(td.Text(), "(kbps)") {
					rateText = td.Text()
				}
			})
		}
	})

	// Handle cases where the port or rate was not found
	if rateText == "" {
		return 0, fmt.Errorf("could not find rate information for port '%s'", portString)
	}

	// Extract the max rate from the text using regex
	re := regexp.MustCompile(`1-(\d+).*kbps`) // Match the "1-XXXXXX" structure and extract XXXXXX
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

func intToBackendPort(port int) string {
	return fmt.Sprintf("Port %d", port)
}

// GetJumboFrame retrieves the current Jumbo Frame configuration from the HTML page.
func (c *HRUIClient) GetJumboFrame() (*JumboFrame, error) {
	// URL to the Jumbo Frame page
	urlJumbo := fmt.Sprintf("%s/fwd.cgi?page=jumboframe", c.URL)

	// Perform HTTP GET request
	respBody, err := c.Request("GET", urlJumbo, nil, nil)
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
func (c *HRUIClient) SetJumboFrame(frameSize int) error {
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
	respBody, err := c.FormRequest(endpoint, formData)
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
