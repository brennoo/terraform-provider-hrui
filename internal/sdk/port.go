package sdk

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var speedDuplexMapping = map[int]string{
	0: "Auto",
	1: "10M/Half",
	2: "10M/Full",
	3: "100M/Half",
	4: "100M/Full",
	5: "1000M/Full",
	6: "2500M/Full",
	8: "10G/Full",
}

var flowControlMapping = map[int]string{
	0: "Off",
	1: "On",
}

var stateMapping = map[string]int{
	"Enable":  1,
	"Disable": 0,
}

type Port struct {
	ID                string
	IsTrunk           bool
	State             int
	SpeedDuplexConfig string
	SpeedDuplexActual string
	FlowControlConfig string
	FlowControlActual string
}

type PortStatistics struct {
	Port       string
	State      int
	LinkStatus string
	TxGoodPkt  int64
	TxBadPkt   int64
	RxGoodPkt  int64
	RxBadPkt   int64
}

type PortMirror struct {
	MirrorDirection string
	MirroringPort   string
	MirroredPort    string
}

func (c *HRUIClient) GetPort(portID string) (*Port, error) {
	ports, err := c.ListPorts()
	if err != nil {
		log.Printf("failed to fetch Ports: %v", err)
		return nil, err
	}

	// Search for the Port with the specified ID in the list of Ports retrieved
	for _, port := range ports {
		if port.ID == portID {
			return port, nil
		}
	}
	return nil, fmt.Errorf("port with ID %s not found", portID)
}

// GetPortByName fetches port.cgi, parses it, and resolves the numeric port ID for a given port name.
func (c *HRUIClient) GetPortByName(portName string) (string, error) {
	respBody, err := c.Request("GET", fmt.Sprintf("%s/port.cgi", c.URL), nil, nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch port.cgi: %w", err)
	}

	// Load the HTML body into goquery
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return "", fmt.Errorf("failed to parse port.cgi HTML: %w", err)
	}

	// Iterate over all <select> elements with name="portid" and extract <option> values
	var portID string
	doc.Find(`select[name="portid"] option`).EachWithBreak(func(i int, s *goquery.Selection) bool {
		// Extract and sanitize the port name
		text := strings.TrimSpace(s.Text())
		if text == portName {
			// Extract the numeric ID from the `value` attribute
			id, exists := s.Attr("value")
			if exists {
				portID = id
				return false
			}
		}
		return true
	})

	// If portID is still empty, the portName was not found
	if portID == "" {
		return "", fmt.Errorf("port name '%s' not found in port.cgi", portName)
	}

	return portID, nil
}

// ListPorts retrieves information about all switch ports.
func (c *HRUIClient) ListPorts() ([]*Port, error) {
	portURL := fmt.Sprintf("%s/port.cgi", c.URL)

	respBody, err := c.Request("GET", portURL, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Port settings from HRUI: %w", err)
	}

	// Parse the HTML response using goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML output: %w", err)
	}

	var ports []*Port
	doc.Find("body center fieldset table").Eq(2).Find("tr").Each(func(i int, tr *goquery.Selection) {
		if i < 2 { // Skip header rows
			return
		}

		port := &Port{}
		tr.Find("td").Each(func(k int, td *goquery.Selection) {
			text := strings.TrimSpace(td.Text())
			switch k {
			case 0: // Port ID
				if strings.HasPrefix(text, "Port") {
					port.ID = text
					port.IsTrunk = false
				} else if strings.HasPrefix(text, "Trunk") {
					port.ID = text
					port.IsTrunk = true
				} else {
					log.Printf("Unexpected port ID format: %s", text)
					return
				}
			case 1: // State
				if text == "Enable" {
					port.State = 1
				} else {
					port.State = 0
				}
			case 2: // Configured Speed/Duplex
				port.SpeedDuplexConfig = text
			case 3: // Actual Speed/Duplex
				port.SpeedDuplexActual = text
			case 4: // Configured Flow Control
				port.FlowControlConfig = text
			case 5: // Actual Flow Control
				port.FlowControlActual = text
			}
		})
		ports = append(ports, port)
	})

	return ports, nil
}

// ConfigurePort updates the configuration for a single port.
func (c *HRUIClient) ConfigurePort(port *Port) (*Port, error) {
	var speedDuplexNumeric string
	for k, v := range speedDuplexMapping {
		if v == port.SpeedDuplexConfig {
			speedDuplexNumeric = strconv.Itoa(k)
			break
		}
	}
	if speedDuplexNumeric == "" {
		return nil, fmt.Errorf("invalid SpeedDuplex value: %s", port.SpeedDuplexConfig)
	}

	var flowControlNumeric string
	for k, v := range flowControlMapping {
		if v == port.FlowControlConfig {
			flowControlNumeric = strconv.Itoa(k)
			break
		}
	}
	if flowControlNumeric == "" {
		return nil, fmt.Errorf("invalid FlowControl value: %s", port.FlowControlConfig)
	}

	portID, err := c.GetPortByName(port.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve port name '%s': %w", port.ID, err)
	}

	form := url.Values{}
	form.Set("cmd", "port")
	form.Set("portid", portID)
	form.Set("state", strconv.Itoa(port.State))
	form.Set("speed_duplex", speedDuplexNumeric)
	form.Set("flow", flowControlNumeric)

	portsURL := fmt.Sprintf("%s/port.cgi", c.URL)
	_, err = c.FormRequest(portsURL, form)
	if err != nil {
		return nil, fmt.Errorf("failed to update port settings: %w", err)
	}

	updatedPort, err := c.GetPort(port.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated port state: %w", err)
	}

	return updatedPort, nil
}

// GetValidPorts fetches and returns the list of IDs of all ports available on the system.
func (c *HRUIClient) GetValidPorts() ([]int, error) {
	if c == nil {
		return nil, fmt.Errorf("HRUIClient is nil")
	}

	// Call ListPorts to fetch all ports.
	ports, err := c.ListPorts()
	if err != nil {
		return nil, fmt.Errorf("failed to list ports: %w", err)
	}

	// Collect all valid Port IDs.
	validPorts := []int{}
	for _, port := range ports {
		// Use GetPortByName to resolve the numeric ID for each port name.
		portID, err := c.GetPortByName(port.ID)
		portIDNum, err := strconv.Atoi(portID)
		if err != nil {
			return nil, fmt.Errorf("invalid port name '%s': %w", port.ID, err)
		}

		// Append the valid port ID to the list.
		validPorts = append(validPorts, portIDNum)
	}

	return validPorts, nil
}

// GetTotalPorts returns the current number of ports.
func (c *HRUIClient) GetTotalPorts() (int, error) {
	if c == nil {
		return 0, fmt.Errorf("HRUIClient is nil")
	}

	// Request the trunk group page
	url := fmt.Sprintf("%s/trunk.cgi?page=group", c.URL)
	body, err := c.Request("GET", url, nil, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch trunk group page: %w", err)
	}

	// Parse the HTML document
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return 0, fmt.Errorf("failed to parse trunk group HTML output: %w", err)
	}

	// Find the <select> element with ID "portsel" and count the number of <option> elements
	totalPorts := 0
	doc.Find("select#portsel option").Each(func(i int, s *goquery.Selection) {
		totalPorts++
	})

	if totalPorts == 0 {
		return 0, fmt.Errorf("no ports found in trunk group page")
	}

	return totalPorts, nil
}

// GetPortStatistics retrieves port statistics from the switch.
func (c *HRUIClient) GetPortStatistics() ([]*PortStatistics, error) {
	statsURL := fmt.Sprintf("%s/port.cgi?page=stats", c.URL)

	respBody, err := c.Request("GET", statsURL, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch port statistics page: %w", err)
	}

	// Parse the HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML document: %w", err)
	}

	var stats []*PortStatistics

	// Select rows in the table, skipping the header rows
	doc.Find("table tr").Each(func(i int, sel *goquery.Selection) {
		if i == 0 {
			return // Skip header
		}

		columns := sel.Find("td")
		port := &PortStatistics{}
		portID := strings.TrimSpace(columns.Eq(0).Text())
		port.Port = portID

		stateStr := strings.TrimSpace(columns.Eq(1).Text())
		port.State, _ = stateMapping[stateStr]
		port.LinkStatus = strings.TrimSpace(columns.Eq(2).Text())

		port.TxGoodPkt, _ = strconv.ParseInt(strings.TrimSpace(columns.Eq(3).Text()), 10, 64)
		port.TxBadPkt, _ = strconv.ParseInt(strings.TrimSpace(columns.Eq(4).Text()), 10, 64)
		port.RxGoodPkt, _ = strconv.ParseInt(strings.TrimSpace(columns.Eq(5).Text()), 10, 64)
		port.RxBadPkt, _ = strconv.ParseInt(strings.TrimSpace(columns.Eq(6).Text()), 10, 64)

		stats = append(stats, port)
	})

	return stats, nil
}

// GetPortMirror fetches the current port mirroring configuration (if any).
func (c *HRUIClient) GetPortMirror() (*PortMirror, error) {
	// Fetch the mirroring configuration page
	urlMirror := fmt.Sprintf("%s/port.cgi?page=mirroring", c.URL)
	respBody, err := c.Request("GET", urlMirror, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Port Mirror settings: %w", err)
	}

	// Parse the HTML response using goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML response: %w", err)
	}

	// Define a PortMirror struct to store the configuration
	mirror := &PortMirror{}
	found := false

	// Locate the table in the `<form>` with action `/port.cgi?page=delete_mirror`.
	doc.Find("form[action='/port.cgi?page=delete_mirror'] table tr").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if i == 0 { // Skip the header row
			return true
		}

		// Extract and clean the table row values
		direction := strings.TrimSpace(s.Find("td").Eq(0).Text())
		rawMirroringPort := strings.TrimSpace(s.Find("td").Eq(1).Text())
		rawMirroredPort := strings.TrimSpace(s.Find("td").Eq(2).Text())

		// Format the values as needed
		mirroringPort := formatPortValue(rawMirroringPort)
		mirroredPort := formatPortValue(rawMirroredPort)

		// Populate the PortMirror struct
		mirror.MirrorDirection = direction
		mirror.MirroringPort = mirroringPort
		mirror.MirroredPort = mirroredPort
		found = true
		return false // Exit after reading the first match
	})

	// If no configuration is found, return nil without error
	if !found {
		return nil, nil
	}

	return mirror, nil
}

// formatPortValue converts raw port values (IDs or Trunk names) to the expected format.
func formatPortValue(raw string) string {
	// Check if raw value is a numeric port ID (e.g., "1", "2") and format it
	if _, err := strconv.Atoi(raw); err == nil {
		return fmt.Sprintf("Port %s", raw)
	}

	// If not numeric, assume it's a Trunk name and return as-is
	return raw
}

// ConfigurePortMirror sets up or updates port mirroring with the given configuration.
func (c *HRUIClient) ConfigurePortMirror(p *PortMirror) error {
	// Construct the URL for configuring port mirroring
	urlMirror := fmt.Sprintf("%s/port.cgi?page=mirroring", c.URL)
	form := url.Values{}
	form.Set("cmd", "mirror")

	// Map the mirror direction to its backend representation
	mirrorDirs := map[string]string{
		"Rx":   "1", // RX only
		"Tx":   "2", // TX only
		"BOTH": "3", // Both RX and TX
	}

	// Validate and set the mirror direction
	if dirValue, exists := mirrorDirs[p.MirrorDirection]; exists {
		form.Set("mirror_direction", dirValue)
	} else {
		return fmt.Errorf("invalid mirror direction '%s'. Valid values: Rx, Tx, BOTH", p.MirrorDirection)
	}

	// Resolve the mirroring port ID
	mirroringPortID, err := c.GetPortByName(p.MirroringPort)
	if err != nil {
		return fmt.Errorf("failed to resolve mirroring port '%s': %w", p.MirroringPort, err)
	}
	form.Set("mirroring_port", mirroringPortID)

	// Ensure the mirrored port is provided
	if p.MirroredPort == "" {
		return fmt.Errorf("mirrored port cannot be empty")
	}
	form.Set("mirrored_port", p.MirroredPort)

	// Make the request to update port mirroring configuration
	_, err = c.FormRequest(urlMirror, form)
	if err != nil {
		return fmt.Errorf("failed to update port mirror settings: %w", err)
	}

	return nil
}

// DeletePortMirror removes the current port mirroring configuration.
func (c *HRUIClient) DeletePortMirror() error {
	// Construct the URL for deleting port mirroring configuration
	urlMirror := fmt.Sprintf("%s/port.cgi?page=delete_mirror", c.URL)
	form := url.Values{}
	form.Set("cmd", "del_mirror")

	// Send the form request to delete the configuration
	_, err := c.FormRequest(urlMirror, form)
	if err != nil {
		return fmt.Errorf("failed to delete port mirror settings: %w", err)
	}

	return nil
}
