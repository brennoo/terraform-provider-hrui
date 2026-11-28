package sdk

import (
	"bytes"
	"context"
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

type PortIsolation struct {
	Port          string
	IsolationList []string
}

func (c *HRUIClient) GetPort(ctx context.Context, portID string) (*Port, error) {
	ports, err := c.ListPorts(ctx)
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
func (c *HRUIClient) GetPortByName(ctx context.Context, portName string) (int, error) {
	respBody, err := c.Request(ctx, "GET", fmt.Sprintf("%s/port.cgi", c.URL), nil, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch port.cgi: %w", err)
	}

	// Load the HTML body into goquery
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return 0, fmt.Errorf("failed to parse port.cgi HTML: %w", err)
	}

	// Iterate over all <select> elements with name="portid" and extract <option> values
	var portIDStr string
	doc.Find(`select[name="portid"] option`).EachWithBreak(func(i int, s *goquery.Selection) bool {
		// Extract and sanitize the port name
		text := strings.TrimSpace(s.Text())
		if text == portName {
			// Extract the numeric ID from the `value` attribute
			id, exists := s.Attr("value")
			if exists {
				portIDStr = id
				return false
			}
		}
		return true
	})

	// If portIDStr is still empty, fall back to v2.x checkbox layout (input[name='port0'], etc.)
	if portIDStr == "" {
		if idx, convErr := resolveZeroBasedPortIndex(portName); convErr == nil {
			if doc.Find(fmt.Sprintf("input[name='port%d']", idx)).Length() > 0 {
				return idx, nil
			}
		}
		return 0, fmt.Errorf("port name '%s' not found in port.cgi", portName)
	}

	// Convert the portIDStr to an int
	portID, err := strconv.Atoi(portIDStr)
	if err != nil {
		return 0, fmt.Errorf("invalid port ID '%s': %w", portIDStr, err)
	}

	return portID, nil
}

// ListPorts retrieves information about all switch ports.
func (c *HRUIClient) ListPorts(ctx context.Context) ([]*Port, error) {
	portURL := fmt.Sprintf("%s/port.cgi", c.URL)

	respBody, err := c.Request(ctx, "GET", portURL, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Port settings from HRUI: %w", err)
	}

	// Parse the HTML response using goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML output: %w", err)
	}

	var ports []*Port
	statusTable := doc.Find("body center fieldset table").Eq(2)
	if statusTable.Length() == 0 {
		statusTable = doc.Find("table").Last()
	}
	statusTable.Find("tr").Each(func(i int, tr *goquery.Selection) {
		if i < 2 { // Skip header rows
			return
		}

		port := &Port{}
		tds := tr.Find("td")
		if tds.Length() >= 8 {
			portIDText := strings.TrimSpace(tds.Eq(0).Text())
			switch {
			case strings.HasPrefix(portIDText, "Port"):
				port.ID = portIDText
			case strings.HasPrefix(portIDText, "Trunk"):
				port.ID = portIDText
				port.IsTrunk = true
			default:
				return
			}
			if strings.TrimSpace(tds.Eq(1).Text()) == "Enable" {
				port.State = 1
			}
			duplexCfg := strings.TrimSpace(tds.Eq(2).Text())
			duplexAct := strings.TrimSpace(tds.Eq(3).Text())
			speedCfg := strings.TrimSpace(tds.Eq(4).Text())
			speedAct := strings.TrimSpace(tds.Eq(5).Text())
			if speedCfg == "Auto" || duplexCfg == "Auto" {
				port.SpeedDuplexConfig = "Auto"
			} else {
				port.SpeedDuplexConfig = fmt.Sprintf("%s/%s", speedCfg, duplexCfg)
			}
			if speedAct == "Link Down" || duplexAct == "Link Down" {
				port.SpeedDuplexActual = "Link Down"
			} else {
				port.SpeedDuplexActual = speedAct + duplexAct
			}
			port.FlowControlConfig = strings.TrimSpace(tds.Eq(6).Text())
			port.FlowControlActual = strings.TrimSpace(tds.Eq(7).Text())
			ports = append(ports, port)
			return
		}
		tds.Each(func(k int, td *goquery.Selection) {
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
func (c *HRUIClient) ConfigurePort(ctx context.Context, port *Port) (*Port, error) {
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

	portID, err := c.GetPortByName(ctx, port.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve port name '%s': %w", port.ID, err)
	}

	form := url.Values{}
	form.Set("cmd", "port")
	form.Set("portid", strconv.Itoa(portID))
	form.Set(fmt.Sprintf("port%d", portID), "on")
	form.Set("state", strconv.Itoa(port.State))
	form.Set("speed_duplex", speedDuplexNumeric)
	form.Set("flow", flowControlNumeric)
	form.Set("flowctrl", flowControlNumeric)
	speedV2, duplexV2 := mapSpeedDuplexV2(port.SpeedDuplexConfig)
	form.Set("speed", speedV2)
	form.Set("duplex", duplexV2)

	portsURL := fmt.Sprintf("%s/port.cgi", c.URL)
	_, err = c.FormRequest(ctx, portsURL, form)
	if err != nil {
		return nil, fmt.Errorf("failed to update port settings: %w", err)
	}

	updatedPort, err := c.GetPort(ctx, port.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated port state: %w", err)
	}

	return updatedPort, nil
}

func mapSpeedDuplexV2(speedDuplex string) (speed, duplex string) {
	switch speedDuplex {
	case "10M/Half":
		return "0", "0"
	case "10M/Full":
		return "0", "1"
	case "100M/Half":
		return "1", "0"
	case "100M/Full":
		return "1", "1"
	case "1000M/Full":
		return "2", "1"
	case "2500M/Full":
		return "5", "1"
	case "10G/Full":
		return "4", "1"
	default:
		return "10", "2" // Auto
	}
}

// GetValidPorts fetches and returns the list of IDs of all ports available on the system.
func (c *HRUIClient) GetValidPorts(ctx context.Context) ([]int, error) {
	if c == nil {
		return nil, fmt.Errorf("HRUIClient is nil")
	}

	// Call ListPorts to fetch all ports.
	ports, err := c.ListPorts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list ports: %w", err)
	}

	// Collect all valid Port IDs.
	validPorts := []int{}
	for _, port := range ports {
		// Use GetPortByName to resolve the numeric ID for each port name.
		portID, err := c.GetPortByName(ctx, port.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid port name '%s': %w", port.ID, err)
		}

		// Append the valid port ID to the list.
		validPorts = append(validPorts, portID)
	}

	return validPorts, nil
}

// GetTotalPorts returns the current number of ports.
func (c *HRUIClient) GetTotalPorts(ctx context.Context) (int, error) {
	if c == nil {
		return 0, fmt.Errorf("HRUIClient is nil")
	}

	// Request the trunk group page
	url := fmt.Sprintf("%s/trunk.cgi?page=group", c.URL)
	body, err := c.Request(ctx, "GET", url, nil, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch trunk group page: %w", err)
	}

	// Parse the HTML document
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return 0, fmt.Errorf("failed to parse trunk group HTML output: %w", err)
	}

	// Find the <select> element with ID "portsel" (or name="ports" / input[name^='port'] on v2.x)
	totalPorts := doc.Find("select#portsel option").Length()
	if totalPorts == 0 {
		totalPorts = doc.Find("select[name='ports'] option").Length()
	}
	if totalPorts == 0 {
		totalPorts = doc.Find("form[action='/trunk.cgi?page=group'] input[type='checkbox'][name^='port']").Length()
	}

	if totalPorts == 0 {
		return 0, fmt.Errorf("no ports found in trunk group page")
	}

	return totalPorts, nil
}

// GetPortStatistics retrieves port statistics from the switch.
func (c *HRUIClient) GetPortStatistics(ctx context.Context) ([]*PortStatistics, error) {
	statsURL := fmt.Sprintf("%s/port.cgi?page=stats", c.URL)

	respBody, err := c.Request(ctx, "GET", statsURL, nil, nil)
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
func (c *HRUIClient) GetPortMirror(ctx context.Context) (*PortMirror, error) {
	// Fetch the mirroring configuration page
	urlMirror := fmt.Sprintf("%s/port.cgi?page=mirroring", c.URL)
	respBody, err := c.Request(ctx, "GET", urlMirror, nil, nil)
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

		tds := s.Find("td")
		if tds.Length() < 3 {
			return true
		}

		var direction, rawMirroringPort, rawMirroredPort string
		if tds.Length() >= 4 {
			// v2.x layout: td[0]=Mirror Group, td[1]=Source (Mirrored) Port, td[2]=Mirror Direction, td[3]=Destination (Mirroring) Port
			rawMirroredPort = strings.TrimSpace(tds.Eq(1).Text())
			direction = strings.TrimSpace(tds.Eq(2).Text())
			if strings.EqualFold(direction, "Both") {
				direction = "BOTH"
			}
			rawMirroringPort = strings.TrimSpace(tds.Eq(3).Text())
		} else {
			// v1.x layout: td[0]=Direction, td[1]=Mirroring Port, td[2]=Mirrored Port
			direction = strings.TrimSpace(tds.Eq(0).Text())
			rawMirroringPort = strings.TrimSpace(tds.Eq(1).Text())
			rawMirroredPort = strings.TrimSpace(tds.Eq(2).Text())
		}
		if direction == "" || direction == "-" || rawMirroringPort == "" || rawMirroringPort == "-" {
			return true
		}

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
func (c *HRUIClient) ConfigurePortMirror(ctx context.Context, p *PortMirror) error {
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
	mirroringPortID, err := c.GetPortByName(ctx, p.MirroringPort)
	if err != nil {
		return fmt.Errorf("failed to resolve mirroring port '%s': %w", p.MirroringPort, err)
	}
	form.Set("mirroring_port", strconv.Itoa(mirroringPortID))

	// Ensure the mirrored port is provided
	if p.MirroredPort == "" {
		return fmt.Errorf("mirrored port cannot be empty")
	}
	form.Set("mirror_group", "1")
	form.Set("mirrored_port", p.MirroredPort)
	if idx, convErr := resolveZeroBasedPortIndex(p.MirroredPort); convErr == nil {
		form.Set(fmt.Sprintf("port%d", idx), "on")
	}

	// Make the request to update port mirroring configuration
	_, err = c.FormRequest(ctx, urlMirror, form)
	if err != nil {
		return fmt.Errorf("failed to update port mirror settings: %w", err)
	}

	return nil
}

// DeletePortMirror removes the current port mirroring configuration.
func (c *HRUIClient) DeletePortMirror(ctx context.Context) error {
	// Construct the URL for deleting port mirroring configuration
	urlMirror := fmt.Sprintf("%s/port.cgi?page=delete_mirror", c.URL)
	form := url.Values{}
	form.Set("cmd", "del_mirror")

	// Send the form request to delete the configuration
	_, err := c.FormRequest(ctx, urlMirror, form)
	if err != nil {
		return fmt.Errorf("failed to delete port mirror settings: %w", err)
	}

	return nil
}

// GetPortIsolation fetches the current port isolation configuration.
func (c *HRUIClient) GetPortIsolation(ctx context.Context) ([]PortIsolation, error) {
	// Fetch the port isolation page from the device
	url := fmt.Sprintf("%s/port.cgi?page=isolation", c.URL)
	respBody, err := c.Request(ctx, "GET", url, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch port isolation page: %w", err)
	}

	// Parse the HTML response
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return nil, fmt.Errorf("failed to parse port isolation HTML: %w", err)
	}

	var isolations []PortIsolation

	// Traverse each row in the isolation table (supports both <fieldset> in v1.x and <div class="showdiv"> in v2.x)
	doc.Find("table").Last().Find("tr").Each(func(i int, row *goquery.Selection) {
		if i == 0 { // Skip header row
			return
		}

		// Extract columns in the current row
		cols := row.Find("td")
		if cols.Length() < 2 {
			return // Skip rows with invalid structure
		}

		portName := strings.TrimSpace(cols.Eq(0).Text())         // First column: Port name
		rawIsolationList := strings.TrimSpace(cols.Eq(1).Text()) // Second column: Isolation list

		isolationList := parseIsolationList(rawIsolationList) // Parse the isolation list

		// Append the parsed isolation to the result set
		isolations = append(isolations, PortIsolation{
			Port:          portName,
			IsolationList: isolationList,
		})
	})

	return isolations, nil
}

// parseIsolationList takes a raw isolation list string and parses it into normalized values.
func parseIsolationList(raw string) []string {
	if raw == "" {
		return []string{} // Return an empty slice for empty isolation lists
	}

	// Split the raw string on commas
	tokens := strings.Split(raw, ",")
	var isolationList []string

	for _, token := range tokens {
		token = strings.TrimSpace(token) // Trim whitespace
		if token == "" {
			continue
		}

		// Handle port ranges (e.g., "2-3" or "Port 2-Port 3")
		if strings.Contains(token, "-") {
			rangeParts := strings.Split(token, "-")
			if len(rangeParts) == 2 {
				startStr := strings.TrimSpace(rangeParts[0])
				endStr := strings.TrimSpace(rangeParts[1])

				// Extract numeric values, handling both "2" and "Port 2" formats
				var startPort, endPort int
				var err error

				if strings.HasPrefix(startStr, "Port ") {
					startPort, err = strconv.Atoi(strings.TrimPrefix(startStr, "Port "))
				} else {
					startPort, err = strconv.Atoi(startStr)
				}

				if err == nil {
					if strings.HasPrefix(endStr, "Port ") {
						endPort, err = strconv.Atoi(strings.TrimPrefix(endStr, "Port "))
					} else {
						endPort, err = strconv.Atoi(endStr)
					}

					if err == nil && startPort <= endPort {
						// Expand the range to individual ports
						for i := startPort; i <= endPort; i++ {
							isolationList = append(isolationList, fmt.Sprintf("Port %d", i))
						}
						continue
					}
				}
			}
			// If range parsing failed, treat as-is (could be "Trunk1-Trunk2" or similar)
			isolationList = append(isolationList, token)
		} else {
			// Handle individual ports
			switch {
			case isNumeric(token): // Numeric port (e.g., "1")
				isolationList = append(isolationList, fmt.Sprintf("Port %s", token))
			case strings.HasPrefix(token, "Port "): // Already normalized (e.g., "Port 2")
				isolationList = append(isolationList, token)
			default: // Non-numeric, e.g., "Trunk2"
				isolationList = append(isolationList, token)
			}
		}
	}

	return isolationList
}

// isNumeric checks if a string represents a valid numeric value.
func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func (c *HRUIClient) ConfigurePortIsolation(ctx context.Context, port string, isolationList []string) error {
	endpoint := fmt.Sprintf("%s/port.cgi?page=isolation", c.URL)

	formData := url.Values{}
	formData.Set("cmd", "portisolation")
	formData.Set("port", port)
	for _, isolation := range isolationList {
		formData.Add("isolationlist", isolation)
	}

	_, err := c.FormRequest(ctx, endpoint, formData)
	if err != nil {
		return fmt.Errorf("failed to configure port isolation: %w", err)
	}

	return nil
}

// DeletePortIsolation resets the isolation list for a specific port.
func (c *HRUIClient) DeletePortIsolation(ctx context.Context, port string) error {
	ports, err := c.ListPorts(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve ports to delete port isolation: %w", err)
	}

	var allPorts []string
	for _, p := range ports {
		allPorts = append(allPorts, p.ID)
	}

	endpoint := fmt.Sprintf("%s/port.cgi?page=isolation", c.URL)

	formData := url.Values{}
	formData.Set("cmd", "portisolation")
	formData.Set("port", port)
	for _, p := range allPorts {
		formData.Add("isolationlist", p)
	}

	_, err = c.FormRequest(ctx, endpoint, formData)
	if err != nil {
		return fmt.Errorf("failed to delete port isolation for port '%s': %w", port, err)
	}

	return nil
}
