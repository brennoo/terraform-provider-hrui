package sdk

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// LoopFunctionType maps human-readable loop function values (like Off, Loop Detection, etc.)
// to their corresponding integer values needed by the internal system.
var LoopFunctionType = map[string]int{
	"Off":             0,
	"Loop Detection":  1,
	"Loop Prevention": 2,
	"Spanning Tree":   3,
}

// LoopProtocol represents loop protocol settings.
type LoopProtocol struct {
	LoopFunction string       // "Off", "Loop Detection", "Loop Prevention", "Spanning Tree"
	IntervalTime int          // Interval time (relevant for Loop Prevention)
	RecoverTime  int          // Recovery time (relevant for Loop Prevention)
	PortStatuses []PortStatus // Per-port Loop Prevention statuses
}

// PortStatus represents the status of a port under Loop Protocol control.
type PortStatus struct {
	Port       string // Port name
	Enable     bool   // Whether Loop Prevention is enabled on this port
	LoopState  string // Loop state ("Enable", "Disable")
	LoopStatus string // Loop operation status ("Forwarding", "Blocked", etc.)
}

// STPGlobalSettings holds the STP global settings.
type STPGlobalSettings struct {
	STPStatus        string // Overall STP status ("Enable", "Disable")
	ForceVersion     string // STP version ("STP", "RSTP")
	Priority         int    // Priority for the STP instance (values like 4096, 8192, 32768, etc.)
	MaxAge           int    // Maximum Age (seconds)
	HelloTime        int    // Hello Time (seconds)
	ForwardDelay     int    // Forwarding Delay (seconds)
	RootPriority     int    // Root bridge priority
	RootMAC          string // Root bridge MAC address
	RootPathCost     int    // Root path cost
	RootPort         string // Root port (number or identifier)
	RootMaxAge       int    // Root Maximum Age (seconds)
	RootHelloTime    int    // Root Hello Time (seconds)
	RootForwardDelay int    // Root Forward Delay (seconds)
}

// STPPort represents a switch port's STP settings.
type STPPort struct {
	Port           string // Port name
	State          string // Port operational state (e.g., Disabled, Forwarding)
	Role           string // Port role in STP (e.g., Designated, Alternate)
	PathCostConfig int    // Configured Path Cost
	PathCostActual int    // Actual Path Cost
	Priority       int    // Port Priority
	P2PConfig      string // Configured P2P setting (True, False, Auto)
	P2PActual      string // Actual P2P state
	EdgeConfig     string // Configured Edge setting (True, False)
	EdgeActual     string // Actual Edge state
}

// GetLoopProtocol fetches the loop protocol settings.
func (c *HRUIClient) GetLoopProtocol() (*LoopProtocol, error) {
	loopURL := c.URL + "/loop.cgi"

	respBody, err := c.Request("GET", loopURL, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch loop protocol page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse loop protocol HTML: %w", err)
	}

	protocol := &LoopProtocol{}

	// Extract loop function type
	doc.Find(`select[name="func_type"] option[selected]`).Each(func(i int, s *goquery.Selection) {
		protocol.LoopFunction = strings.TrimSpace(s.Text())
	})

	// Parse interval and recovery time if "Loop Prevention" is enabled
	if protocol.LoopFunction == "Loop Prevention" {
		protocol.IntervalTime, err = extractInt(doc, `input[name="interval_time"]`, "attribute", "value")
		if err != nil {
			return nil, fmt.Errorf("failed to extract interval time: %w", err)
		}

		protocol.RecoverTime, err = extractInt(doc, `input[name="recover_time"]`, "attribute", "value")
		if err != nil {
			return nil, fmt.Errorf("failed to extract recovery time: %w", err)
		}
	}

	// Parse port statuses
	protocol.PortStatuses = parsePortStatuses(doc)

	return protocol, nil
}

// ConfigureLoopProtocol updates the loop function and associated settings.
func (c *HRUIClient) ConfigureLoopProtocol(loopFunction string, intervalTime, recoverTime int, portStatuses []PortStatus) error {
	loopURL := c.URL + "/loop.cgi"
	funcType, valid := LoopFunctionType[loopFunction]
	if !valid {
		return fmt.Errorf("invalid loop function type: %s", loopFunction)
	}

	// Prepare form data for POST request
	formData := url.Values{
		"cmd":           {"loop"},
		"func_type":     {strconv.Itoa(funcType)},
		"interval_time": {strconv.Itoa(intervalTime)},
		"recover_time":  {strconv.Itoa(recoverTime)},
	}

	_, err := c.FormRequest(loopURL, formData)
	if err != nil {
		return fmt.Errorf("failed to update loop protocol: %w", err)
	}

	return nil
}

// GetSTPSettings fetches and parses the STP Global Settings page.
func (c *HRUIClient) GetSTPSettings() (*STPGlobalSettings, error) {
	stpURL := c.URL + "/loop.cgi?page=stp_global"

	respBody, err := c.Request("GET", stpURL, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch STP Global Settings page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse STP Global Settings HTML: %w", err)
	}

	return parseSTPGlobalSettings(doc)
}

// SetSTPSettings updates the STP global settings.
func (c *HRUIClient) SetSTPSettings(stp *STPGlobalSettings) error {
	stpURL := c.URL + "/loop.cgi?page=stp_global"
	formData := url.Values{
		"cmd":      {"stp"},
		"version":  {stp.GetVersionValue()},
		"priority": {strconv.Itoa(stp.Priority)},
		"maxage":   {strconv.Itoa(stp.MaxAge)},
		"hello":    {strconv.Itoa(stp.HelloTime)},
		"delay":    {strconv.Itoa(stp.ForwardDelay)},
	}

	_, err := c.FormRequest(stpURL, formData)
	if err != nil {
		return fmt.Errorf("failed to update STP global settings: %w", err)
	}

	return nil
}

// SetSTPSettingsAsync performs a fire-and-forget POST request to update the STP Global Settings.
// needed due to a bug in the cgi for updating stp global settings that never returns.
func (c *HRUIClient) SetSTPSettingsAsync(stp *STPGlobalSettings) error {
	stpURL := c.URL + "/loop.cgi?page=stp_global"

	// Prepare form data for POST request
	data := url.Values{
		"version":  {stp.GetVersionValue()},
		"priority": {strconv.Itoa(stp.Priority)},
		"maxage":   {strconv.Itoa(stp.MaxAge)},
		"hello":    {strconv.Itoa(stp.HelloTime)},
		"delay":    {strconv.Itoa(stp.ForwardDelay)},
		"cmd":      {"stp"},
	}

	c.HttpClient.Timeout = 2 * time.Second
	_, err := c.FormRequest(stpURL, data)
	if err != nil {
		log.Printf("[WARN] POST request timed out or failed: %v", err)
		return nil
	}

	return nil
}

// GetSTPPortSettings fetches the STP port settings.
func (c *HRUIClient) GetSTPPortSettings() ([]STPPort, error) {
	stpURL := c.URL + "/loop.cgi?page=stp_port"

	respBody, err := c.Request("GET", stpURL, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch STP port settings page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Parsing options for STP integers
	parseSTPIntOptions := []ParseOption{
		WithDefaultValue(0),
		WithSpecialCases("Auto", "-"),
		WithLogging(),
	}

	// Target the last table (STP settings table)
	table := doc.Find("table").Last()

	// Parse rows within the table
	rows := table.Find("tr")
	var stpPorts []STPPort
	rows.Each(func(i int, row *goquery.Selection) {
		// Dynamically detect and skip header rows
		if row.Find("th").Length() > 0 {
			log.Printf("[DEBUG] Skipping header row %d", i)
			return
		}

		// Extract all <td> columns
		tds := row.Find("td")
		if tds.Length() != 10 {
			log.Printf("[DEBUG] Skipping row %d: column count = %d", i, tds.Length())
			return
		}

		// Parse the STP Port entry
		port := strings.TrimSpace(tds.Eq(0).Text())
		stpPort := STPPort{
			Port:           port,
			State:          strings.TrimSpace(tds.Eq(1).Text()),
			Role:           strings.TrimSpace(tds.Eq(2).Text()),
			PathCostConfig: *parseInt(tds.Eq(3).Text(), parseSTPIntOptions...),
			PathCostActual: *parseInt(tds.Eq(4).Text(), parseSTPIntOptions...),
			Priority:       *parseInt(tds.Eq(5).Text(), parseSTPIntOptions...),
			P2PConfig:      normalizeBoolString(tds.Eq(6).Text()),
			P2PActual:      normalizeBoolString(tds.Eq(7).Text()),
			EdgeConfig:     normalizeBoolString(tds.Eq(8).Text()),
			EdgeActual:     normalizeBoolString(tds.Eq(9).Text()),
		}

		log.Printf("[DEBUG] Parsed Port: %+v", stpPort)
		stpPorts = append(stpPorts, stpPort)
	})

	if len(stpPorts) == 0 {
		return nil, fmt.Errorf("no STP ports found in settings table")
	}

	return stpPorts, nil
}

// SetSTPPortSettings updates the STP settings for a specific port by its name.
func (c *HRUIClient) SetSTPPortSettings(portName string, pathCost, priority int, p2p, edge string) error {
	portID, err := c.GetPortByName(portName)
	if err != nil {
		return fmt.Errorf("failed to resolve port ID for port '%s': %w", portName, err)
	}

	// Construct the STP settings URL
	stpURL := c.URL + "/loop.cgi?page=stp_port"

	// Prepare the form data
	formData := url.Values{
		"cmd":      {"stp_port"},
		"portid":   {strconv.Itoa(portID)},
		"cost":     {strconv.Itoa(pathCost)},
		"priority": {strconv.Itoa(priority)},
		"p2p":      {strings.ToLower(p2p)},
		"edge":     {strings.ToLower(edge)},
		"submit":   {"+++Apply+++"},
	}

	// Send the request to update STP settings
	_, err = c.FormRequest(stpURL, formData)
	if err != nil {
		return fmt.Errorf("failed to update STP port settings for port '%s': %w", portName, err)
	}

	return nil
}

// GetSTPPort fetches a single STP port by its ID from the backend.
func (c *HRUIClient) GetSTPPort(portName string) (*STPPort, error) {
	ports, err := c.GetSTPPortSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch STP port settings: %w", err)
	}

	for _, port := range ports {
		if port.Port == portName {
			return &port, nil
		}
	}

	return nil, fmt.Errorf("port with name '%s' not found", portName)
}

func parsePortStatuses(doc *goquery.Document) []PortStatus {
	var portStatuses []PortStatus

	doc.Find("table").First().Find("tr").Each(func(i int, tr *goquery.Selection) {
		if i == 0 { // Skip the header
			return
		}

		tds := tr.Find("td")
		if tds.Length() < 3 {
			return
		}

		port := strings.TrimSpace(tds.Eq(0).Text())

		// Extract other fields
		loopState := strings.TrimSpace(tds.Eq(1).Text())
		loopStatus := strings.TrimSpace(tds.Eq(2).Text())
		enable := loopState == "Enable"

		portStatuses = append(portStatuses, PortStatus{
			Port:       port,
			Enable:     enable,
			LoopState:  loopState,
			LoopStatus: loopStatus,
		})
	})

	return portStatuses
}

// Extracts text from the given CSS selector, returning an empty string on failure.
func extractText(doc *goquery.Document, selector string) (string, error) {
	selection := strings.TrimSpace(doc.Find(selector).Text())
	if selection == "" {
		return "", fmt.Errorf("missing value for selector: %s", selector)
	}
	return selection, nil
}

// parseSTPGlobalSettings extracts STP configuration data from the HTML using goquery.
func parseSTPGlobalSettings(doc *goquery.Document) (*STPGlobalSettings, error) {
	settings := &STPGlobalSettings{}
	var err error

	// Parse STP status
	if settings.STPStatus, err = extractText(doc, "th:contains('Spanning Tree Status') + td"); err != nil {
		return nil, err
	}

	// Parse ForceVersion
	if settings.ForceVersion, err = extractText(doc, "select[name='version'] option[selected]"); err != nil {
		return nil, err
	}

	// Parse Priority (uses attribute value)
	if settings.Priority, err = extractInt(doc, "select[name='priority'] option[selected]", "attribute", "value"); err != nil {
		return nil, err
	}

	// Parse MaxAge, HelloTime, ForwardDelay (uses attribute value)
	if settings.MaxAge, err = extractInt(doc, "input[name='maxage']", "attribute", "value"); err != nil {
		return nil, err
	}
	if settings.HelloTime, err = extractInt(doc, "input[name='hello']", "attribute", "value"); err != nil {
		return nil, err
	}
	if settings.ForwardDelay, err = extractInt(doc, "input[name='delay']", "attribute", "value"); err != nil {
		return nil, err
	}

	// Parse Root Priority
	if settings.RootPriority, err = extractInt(doc, "th:contains('Root Priority') + td", "text"); err != nil {
		return nil, err
	}

	// Parse Root MAC Address
	if settings.RootMAC, err = extractText(doc, "th:contains('Root MAC Address') + td"); err != nil {
		return nil, err
	}

	// Parse Root Path Cost
	if settings.RootPathCost, err = extractInt(doc, "th:contains('Root Path Cost') + td", "text"); err != nil {
		return nil, err
	}

	// Parse Root Port
	if settings.RootPort, err = extractText(doc, "th:contains('Root Port') + td"); err != nil {
		return nil, err
	}

	// Parse RootMaxAge, removing " Sec"
	if rawValue, err := extractText(doc, "th:contains('Root Maximum Age') + td"); err != nil {
		return nil, err
	} else {
		settings.RootMaxAge = *parseInt(rawValue, WithTrimSuffix(" Sec"))
	}

	// Parse RootHelloTime, removing " Sec"
	if rawValue, err := extractText(doc, "th:contains('Root Hello Time') + td"); err != nil {
		return nil, err
	} else {
		settings.RootHelloTime = *parseInt(rawValue, WithTrimSuffix(" Sec"))
	}

	// Parse RootForwardDelay, removing " Sec"
	if rawValue, err := extractText(doc, "th:contains('Root Forward Delay') + td"); err != nil {
		return nil, err
	} else {
		settings.RootForwardDelay = *parseInt(rawValue, WithTrimSuffix(" Sec"))
	}

	log.Printf("[DEBUG] Parsed STPGlobalSettings: %+v", settings)
	return settings, nil
}

func normalizeBoolString(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true":
		return "True"
	case "false":
		return "False"
	case "auto":
		return "Auto"
	default:
		return value
	}
}

// GetVersionValue returns a numeric value for the version (STP/RSTP).
func (stp *STPGlobalSettings) GetVersionValue() string {
	switch stp.ForceVersion {
	case "STP":
		return "0"
	case "RSTP":
		return "1"
	default:
		return "0"
	}
}
