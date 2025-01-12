package sdk

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

// Global lock to serialize IGMP port updates.
var igmpUpdateLock sync.Mutex

// IGMPConfig represents the IGMP Snooping configuration.
type IGMPConfig struct {
	Enabled bool
	Ports   map[int]bool
}

// ConfigureIGMPSnooping updates the global IGMP snooping setting (enable or disable).
func (c *HRUIClient) ConfigureIGMPSnooping(enable bool) error {
	return c.updateGlobalIGMP(enable)
}

// updateGlobalIGMP is a reusable method for updating the global IGMP state.
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

// EnableIGMPSnooping enables IGMP snooping globally.
func (c *HRUIClient) EnableIGMPSnooping() error {
	return c.updateGlobalIGMP(true)
}

// DisableIGMPSnooping turns off IGMP globally.
func (c *HRUIClient) DisableIGMPSnooping() error {
	return c.updateGlobalIGMP(false)
}

// ConfigurePortIGMPSnooping enables or disables IGMP snooping for a specific port.
func (c *HRUIClient) ConfigurePortIGMPSnooping(port int, enable bool) error {
	// Lock the operation to prevent race conditions
	igmpUpdateLock.Lock()
	defer igmpUpdateLock.Unlock()

	// Fetch current IGMP snooping states for all ports.
	currentState, err := c.GetAllPortsIGMPSnooping()
	if err != nil {
		return fmt.Errorf("failed to fetch IGMP configuration: %w", err)
	}

	// Check if the port ID exists in the current configuration.
	if _, exists := currentState[port]; !exists {
		return fmt.Errorf("invalid port: %d does not exist in configuration", port)
	}

	// Skip operation if the port is already in the desired state.
	if (enable && currentState[port] == "on") || (!enable && currentState[port] == "off") {
		log.Printf("[INFO] Port %d is already %s. Skipping operation.", port, map[bool]string{true: "enabled", false: "disabled"}[enable])
		return nil
	}

	// Modify the state for the target port.
	if enable {
		currentState[port] = "on"
	} else {
		delete(currentState, port) // Disabled ports are omitted.
	}

	// Construct the payload for all enabled ports.
	payload := url.Values{}
	for portID, state := range currentState {
		if state == "on" {
			payload.Add(fmt.Sprintf("lPort_%d", portID), "on")
		}
	}
	payload.Add("cmd", "set")
	log.Printf("[DEBUG] Constructed payload: %s", payload.Encode())

	// Send the configuration update to the IGMP settings endpoint.
	url := fmt.Sprintf("%s/igmp.cgi?page=igmp_static_router", c.URL)
	_, err = c.FormRequest(url, payload)
	if err != nil {
		return fmt.Errorf("failed to update IGMP snooping for port %d: %w", port, err)
	}

	log.Printf("[INFO] Port %d successfully updated to %s.", port, map[bool]string{true: "enabled", false: "disabled"}[enable])
	return nil
}

// EnablePortIGMPSnooping turns on IGMP snooping for a port.
func (c *HRUIClient) EnablePortIGMPSnooping(port int) error {
	return c.ConfigurePortIGMPSnooping(port, true)
}

// DisablePortIGMPSnooping turns off IGMP snooping for a port.
func (c *HRUIClient) DisablePortIGMPSnooping(port int) error {
	return c.ConfigurePortIGMPSnooping(port, false)
}

// FetchIGMPConfig retrieves and parses the full IGMP configuration.
func (c *HRUIClient) FetchIGMPConfig() (*IGMPConfig, error) {
	url := fmt.Sprintf("%s/igmp.cgi?page=dump", c.URL)

	// Fetch the HTML page containing the IGMP configuration.
	respBody, err := c.Request("GET", url, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch IGMP configuration: %w", err)
	}

	// Parse the HTML page using goquery.
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse IGMP configuration HTML: %w", err)
	}

	// Parse global IGMP state.
	var globalEnabled bool
	doc.Find("input[name=enable_igmp]").Each(func(_ int, s *goquery.Selection) {
		_, globalEnabled = s.Attr("checked")
	})

	// Parse port statuses.
	portStatus := make(map[int]bool)
	doc.Find("td input").Each(func(_ int, input *goquery.Selection) {
		nameAttr, exists := input.Attr("name")
		if !exists || !strings.HasPrefix(nameAttr, "lPort_") {
			return
		}

		// Extract the port ID from the checkbox name (e.g., `lPort_0`).
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

// ========================= new functions here

func (c *HRUIClient) UpdatePortIGMPSnoopingByName(portName string, enable bool) error {
	// Resolve the port name to a logical port ID.
	portID, err := c.ResolvePortNameToID(portName)
	if err != nil {
		return fmt.Errorf("failed to resolve port name %s: %w", portName, err)
	}

	// Use the standard `ConfigurePortIGMPSnooping` function to handle the operation.
	err = c.ConfigurePortIGMPSnooping(portID, enable)
	if err != nil {
		return fmt.Errorf("failed to update IGMP snooping state for port %s: %w", portName, err)
	}

	log.Printf("[INFO] Port %s successfully updated to %s.", portName, map[bool]string{true: "enabled", false: "disabled"}[enable])
	return nil
}

// ListIGMPPorts fetches a list of valid ports with their logical IDs from the backend.
func (c *HRUIClient) ListIGMPPorts() (map[int]string, error) {
	// Fetch the IGMP configuration page.
	respBody, err := c.Request("GET", c.URL+"/igmp.cgi?page=dump", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch IGMP configuration page: %w", err)
	}

	// Parse the HTML response using goquery.
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse IGMP configuration HTML: %w", err)
	}

	// Build a map of logical IDs to port names.
	portMap := make(map[int]string)

	// Use goquery to find and parse the relevant table rows.
	doc.Find("form[action='/igmp.cgi?page=igmp_static_router'] table tr").First().Find("td").Each(func(i int, s *goquery.Selection) {
		portName := strings.TrimSpace(s.Text())
		// Ensure the port name is valid and neither empty nor undesired.
		if portName != "" && !strings.Contains(portName, "static") && !strings.Contains(portName, "dynamic") && portName != "Router Port" {
			portMap[i] = portName // Use the index `i` as the logical ID.
		}
	})

	// Check if any ports were found.
	if len(portMap) == 0 {
		return nil, fmt.Errorf("no valid ports found on the backend")
	}

	return portMap, nil
}

func (c *HRUIClient) GetPortIGMPSnoopingByName(portName string) (bool, error) {
	portID, err := c.ResolvePortNameToID(portName)
	if err != nil {
		return false, fmt.Errorf("failed to resolve port name %s: %w", portName, err)
	}
	return c.GetPortIGMPSnooping(portID)
}

// ResolvePortNameToID maps a port name to its logical ID.
func (c *HRUIClient) ResolvePortNameToID(portName string) (int, error) {
	portList, err := c.ListIGMPPorts()
	if err != nil {
		return -1, fmt.Errorf("failed to fetch port list: %w", err)
	}

	portMap := make(map[string]int)
	for logicalID, name := range portList {
		portMap[name] = logicalID
	}

	portID, exists := portMap[portName]
	if !exists {
		return -1, fmt.Errorf("unknown port name %s", portName)
	}

	return portID, nil
}

func (c *HRUIClient) GetPortIGMPSnooping(portID int) (bool, error) {
	// Fetch all port statuses.
	allPorts, err := c.GetAllPortsIGMPSnooping()
	if err != nil {
		return false, fmt.Errorf("failed to fetch IGMP snooping states: %w", err)
	}

	// Check if the port exists in the state map.
	status, exists := allPorts[portID]
	if !exists {
		return false, fmt.Errorf("port %d is not found in the IGMP configuration", portID)
	}

	// Return the port status.
	return status == "on", nil
}

func parseAllPortsIGMPStatus(respBody []byte) map[int]string {
	// Parse the HTML response using goquery.
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		log.Fatalf("[ERROR] Failed to parse IGMP status response: %v", err)
	}

	// Initialize the map to store port states.
	portStates := make(map[int]string)

	// Traverse the relevant table rows and extract data.
	doc.Find("table").Find("tr").Each(func(rowIndex int, row *goquery.Selection) {
		// Extract cells (`<td>`) from the current row.
		cols := row.Find("td")

		// Validate rows: Skip rows with less than 2 columns (irrelevant rows like "Router Port").
		if cols.Length() < 2 {
			return
		}

		// Handle rows for "static" IGMP states (port checkboxes).
		portType := strings.TrimSpace(row.Find("th").Text())
		if portType == "static" {
			cols.Each(func(colIndex int, col *goquery.Selection) {
				// Get the checkbox's `name` attribute (e.g., `lPort_0`).
				nameAttr, exists := col.Find("input[type='checkbox']").Attr("name")
				if !exists {
					return // Skip if no checkbox exists.
				}

				// Extract the port ID based on "lPort_<portID>".
				var portID int
				_, err := fmt.Sscanf(nameAttr, "lPort_%d", &portID)
				if err != nil {
					log.Printf("[DEBUG] Failed to parse port ID from name attribute: %v\n", nameAttr)
					return
				}

				// Check if the checkbox is checked (enable IGMP snooping).
				checkedAttr, checked := col.Find("input[type='checkbox']").Attr("checked")
				if checked && checkedAttr != "false" {
					portStates[portID] = "on"
				} else {
					portStates[portID] = "off"
				}
			})
		}
	})

	// Debug output for the parsed state.
	log.Printf("[DEBUG] Parsed port states: %v\n", portStates)
	return portStates
}

func (c *HRUIClient) GetAllPortsIGMPSnooping() (map[int]string, error) {
	// Call the backend to fetch the IGMP states.
	respBody, err := c.Request("GET", c.URL+"/igmp.cgi?page=dump", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch IGMP port statuses: %w", err)
	}

	// Parse the port statuses from the response.
	portStates := parseAllPortsIGMPStatus(respBody)
	log.Printf("[DEBUG] Fetched current IGMP snooping state: %v", portStates)

	return portStates, nil
}
