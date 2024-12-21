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

// StormControlConfig represents all the storm control entries in the table
type StormControlConfig struct {
	Entries []StormControlEntry `json:"entries"`
}

// GetStormControlStatus fetches the current storm control status from the HTML page
func (c *HRUIClient) GetStormControlStatus() (*StormControlConfig, error) {
	// Perform the GET request to fetch the page content
	resp, err := c.HttpClient.Get(c.URL + "/fwd.cgi?page=storm_ctrl")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse the HTML page using goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, errors.New("failed to parse HTML response")
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
		port, err := backendPortToInt(portString)
		if err != nil {
			return // Skip invalid port entries
		}

		broadcast := parseRate(strings.TrimSpace(cols.Eq(1).Text()))
		knownMulticast := parseRate(strings.TrimSpace(cols.Eq(2).Text()))
		unknownUnicast := parseRate(strings.TrimSpace(cols.Eq(3).Text()))
		unknownMulticast := parseRate(strings.TrimSpace(cols.Eq(4).Text()))

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

// Helper function to parse rates from text (converts "Off" to nil)
func parseRate(rate string) *int {
	if rate == "Off" {
		return nil
	}

	// Convert from string to integer
	value, err := strconv.Atoi(rate)
	if err != nil {
		return nil
	}

	return &value
}

// UpdateStormControl updates the storm control settings for specific ports
func (c *HRUIClient) UpdateStormControl(
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

	// Perform the POST request
	resp, err := c.HttpClient.PostForm(c.URL+"/fwd.cgi?page=storm_ctrl", formData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Parse the response body to check for errors
	doc, err := goquery.NewDocumentFromReader(resp.Body)
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
	// Convert the provided integer port to the backend's string format
	portString := intToBackendPort(port)

	// Step 1: Fetch the storm control page
	resp, err := c.HttpClient.Get(c.URL + "/fwd.cgi?page=storm_ctrl")
	if err != nil {
		return 0, fmt.Errorf("failed to fetch storm control page: %v", err)
	}
	defer resp.Body.Close()

	// Step 2: Parse the HTML content using goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error parsing storm control page: %v", err)
	}

	// Step 3: Look for the row corresponding to the given port
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

	// Step 4: Handle cases where the port or rate was not found
	if rateText == "" {
		return 0, fmt.Errorf("could not find rate information for port '%s'", portString)
	}

	// Step 5: Extract the max rate from the text using regex
	re := regexp.MustCompile(`1-(\d+).*kbps`) // Match the "1-XXXXXX" structure and extract XXXXXX
	matches := re.FindStringSubmatch(rateText)
	if len(matches) < 2 {
		return 0, fmt.Errorf("failed to extract rate information from text: %s", rateText)
	}

	// Convert the extracted max rate to int64
	maxRate, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("error converting max rate to integer: %v", err)
	}

	return maxRate, nil
}

// Helper function to convert storm type to its corresponding ID
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

func backendPortToInt(port string) (int, error) {
	parts := strings.Split(port, " ")
	if len(parts) != 2 || parts[0] != "Port" {
		return 0, fmt.Errorf("invalid port format: %s", port)
	}
	return strconv.Atoi(parts[1])
}
