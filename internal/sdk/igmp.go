package sdk

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

// IGMPConfig represents the IGMP Snooping configuration.
type IGMPConfig struct {
	Enabled bool
	Ports   map[int]bool
}

// Global lock to serialize IGMP port updates.
var igmpUpdateLock sync.Mutex

// ConfigureIGMPSnooping enables or disables IGMP snooping globally.
func (c *HRUIClient) ConfigureIGMPSnooping(enable bool) error {
	return c.updateGlobalIGMP(enable)
}

// EnableIGMPSnooping enables IGMP snooping globally.
func (c *HRUIClient) EnableIGMPSnooping() error {
	return c.updateGlobalIGMP(true)
}

// DisableIGMPSnooping disables IGMP snooping globally.
func (c *HRUIClient) DisableIGMPSnooping() error {
	return c.updateGlobalIGMP(false)
}

// updateGlobalIGMP handles the global IGMP snooping configuration change request.
func (c *HRUIClient) updateGlobalIGMP(enable bool) error {
	formData := url.Values{}
	if enable {
		formData.Set("enable_igmp", "on")
	}
	url := fmt.Sprintf("%s/igmp.cgi?page=enable_igmp", c.URL)

	if _, err := c.FormRequest(url, formData); err != nil {
		return fmt.Errorf("failed to update global IGMP snooping: %w", err)
	}
	return nil
}

// ConfigurePortIGMPSnooping enables or disables IGMP snooping for a specific port.
func (c *HRUIClient) ConfigurePortIGMPSnooping(portID int, enable bool) error {
	igmpUpdateLock.Lock()
	defer igmpUpdateLock.Unlock()

	currentState, err := c.GetAllPortsIGMPSnooping()
	if err != nil {
		return fmt.Errorf("failed to fetch IGMP configuration: %w", err)
	}

	// Check if the port ID exists in the current configuration.
	if _, exists := currentState[portID]; !exists {
		return fmt.Errorf("invalid port: %d does not exist in configuration", portID)
	}

	// Skip operation if the port is already in the desired state.
	isStateMatch := (enable && currentState[portID] == "on") || (!enable && currentState[portID] == "off")
	if isStateMatch {
		return nil
	}

	// Modify the state for the target port.
	if enable {
		currentState[portID] = "on"
	} else {
		delete(currentState, portID)
	}

	// Construct the payload for all enabled ports.
	payload := url.Values{}
	for id, state := range currentState {
		if state == "on" {
			payload.Add(fmt.Sprintf("lPort_%d", id), "on")
		}
	}
	payload.Add("cmd", "set")

	// Send the configuration update to the IGMP settings endpoint.
	url := fmt.Sprintf("%s/igmp.cgi?page=igmp_static_router", c.URL)
	if _, err := c.FormRequest(url, payload); err != nil {
		return fmt.Errorf("failed to update IGMP snooping for port %d: %w", portID, err)
	}

	return nil
}

// GetAllPortsIGMPSnooping retrieves the current IGMP snooping configuration for all ports.
func (c *HRUIClient) GetAllPortsIGMPSnooping() (map[int]string, error) {
	respBody, err := c.Request("GET", c.URL+"/igmp.cgi?page=dump", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch IGMP port statuses: %w", err)
	}

	portStates, err := parseAllPortsIGMPStatus(respBody)
	if err != nil {
		return nil, fmt.Errorf("failed to parse IGMP port statuses: %w", err)
	}

	return portStates, nil
}

// GetPortIGMPSnooping retrieves the IGMP snooping state for a specific port by its ID.
func (c *HRUIClient) GetPortIGMPSnooping(portID int) (bool, error) {
	allPorts, err := c.GetAllPortsIGMPSnooping()
	if err != nil {
		return false, fmt.Errorf("failed to fetch IGMP snooping states: %w", err)
	}

	status, exists := allPorts[portID]
	if !exists {
		return false, fmt.Errorf("port %d is not found in the IGMP configuration", portID)
	}

	return status == "on", nil
}

// UpdatePortIGMPSnoopingByName enables or disables IGMP snooping for a port by its name.
func (c *HRUIClient) UpdatePortIGMPSnoopingByName(portName string, enable bool) error {
	portID, err := c.GetPortByName(portName)
	if err != nil {
		return fmt.Errorf("failed to resolve port name %q to ID: %w", portName, err)
	}

	if err := c.ConfigurePortIGMPSnooping(portID, enable); err != nil {
		return fmt.Errorf("failed to update IGMP snooping for port %q (ID: %d): %w", portName, portID, err)
	}

	return nil
}

// GetPortIGMPSnoopingByName retrieves the IGMP snooping state for a port by its name.
func (c *HRUIClient) GetPortIGMPSnoopingByName(portName string) (bool, error) {
	portID, err := c.GetPortByName(portName)
	if err != nil {
		return false, fmt.Errorf("failed to resolve port name %q to ID: %w", portName, err)
	}

	return c.GetPortIGMPSnooping(portID)
}

// FetchIGMPConfig fetches and parses the complete IGMP configuration.
func (c *HRUIClient) FetchIGMPConfig() (*IGMPConfig, error) {
	url := fmt.Sprintf("%s/igmp.cgi?page=dump", c.URL)

	respBody, err := c.Request("GET", url, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch IGMP configuration: %w", err)
	}

	// Parse global IGMP state and port statuses.
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse IGMP configuration HTML: %w", err)
	}

	// Parse global IGMP state.
	var globalEnabled bool
	doc.Find("input[name=enable_igmp]").Each(func(_ int, s *goquery.Selection) {
		if _, exists := s.Attr("checked"); exists {
			globalEnabled = true
		}
	})

	// Parse port statuses.
	portStatus := make(map[int]bool)
	doc.Find("td input").Each(func(_ int, input *goquery.Selection) {
		nameAttr, exists := input.Attr("name")
		if !exists || !strings.HasPrefix(nameAttr, "lPort_") {
			return
		}

		var portID int
		if _, err := fmt.Sscanf(nameAttr, "lPort_%d", &portID); err == nil {
			_, isChecked := input.Attr("checked")
			portStatus[portID] = isChecked
		}
	})

	return &IGMPConfig{
		Enabled: globalEnabled,
		Ports:   portStatus,
	}, nil
}

// parseAllPortsIGMPStatus parses the IGMP snooping states for all ports from the response body.
func parseAllPortsIGMPStatus(respBody []byte) (map[int]string, error) {
	// Parse the HTML using goquery.
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse IGMP status response: %w", err)
	}

	// Port states to return.
	portStates := make(map[int]string)

	// Find relevant table rows.
	doc.Find("fieldset").Each(func(_ int, fieldset *goquery.Selection) {
		// Locate the specific table for static IGMP ports.
		fieldset.Find("tr").Each(func(_ int, row *goquery.Selection) {
			// Identify rows labeled "static."
			rowHeader := strings.TrimSpace(row.Find("th").Text())
			if rowHeader != "static" {
				return // Skip irrelevant rows.
			}

			// Traverse relevant columns.
			row.Find("td input[type='checkbox']").Each(func(_ int, input *goquery.Selection) {
				// Port ID is encoded in the `name` attribute (e.g., lPort_0).
				nameAttr, exists := input.Attr("name")
				if !exists || !strings.HasPrefix(nameAttr, "lPort_") {
					return
				}

				var portID int
				if _, err := fmt.Sscanf(nameAttr, "lPort_%d", &portID); err == nil {
					_, isChecked := input.Attr("checked")
					if isChecked {
						portStates[portID] = "on"
					} else {
						portStates[portID] = "off"
					}
				}
			})
		})
	})

	return portStates, nil
}
