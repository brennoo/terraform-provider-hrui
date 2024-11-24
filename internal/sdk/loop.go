package sdk

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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
	Port       int    // Port number
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
	Port       int    // Port number (e.g., "Port 1")
	State      string // Current state (e.g., "Forwarding", "Blocked", "Listening")
	Role       string // Port role (e.g., "Designated", "Root", "Disabled")
	PathCost   int    // Path cost value for the port
	Priority   int    // Port priority
	P2P        string // Point-to-point link status configuration (e.g., "Auto", "True", "False")
	P2PActual  string // Actual P2P operational status (e.g., "Auto", "True", "False")
	Edge       string // Edge port configuration (e.g., "True", "False")
	EdgeActual string // Actual edge port operational status (e.g., "True", "False")
}

// GetLoopProtocol fetches the loop protocol settings.
func (client *HRUIClient) GetLoopProtocol() (*LoopProtocol, error) {
	loopURL := client.URL + "/loop.cgi"
	resp, err := client.HttpClient.Get(loopURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch loop protocol page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
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
		protocol.IntervalTime, _ = extractIntAttribute(doc, `input[name="interval_time"]`, "value")
		protocol.RecoverTime, _ = extractIntAttribute(doc, `input[name="recover_time"]`, "value")
	}

	// Parse port statuses
	protocol.PortStatuses = parsePortStatuses(doc)

	return protocol, nil
}

// UpdateLoopProtocol updates the loop function and associated settings.
func (client *HRUIClient) UpdateLoopProtocol(loopFunction string, intervalTime, recoverTime int, portStatuses []PortStatus) error {
	funcType, valid := LoopFunctionType[loopFunction]
	if !valid {
		return fmt.Errorf("invalid loop function type: %s", loopFunction)
	}

	// Prepare form data for POST request
	data := url.Values{
		"cmd":           {"loop"},
		"func_type":     {strconv.Itoa(funcType)},
		"interval_time": {strconv.Itoa(intervalTime)},
		"recover_time":  {strconv.Itoa(recoverTime)},
	}

	// Send update to backend
	resp, err := client.HttpClient.PostForm(client.URL+"/loop.cgi", data)
	if err != nil {
		return fmt.Errorf("failed to update loop protocol: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// GetSTPSettings fetches and parses the STP Global Settings page.
func (client *HRUIClient) GetSTPSettings() (*STPGlobalSettings, error) {
	stpURL := client.URL + "/loop.cgi?page=stp_global"
	resp, err := client.HttpClient.Get(stpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch STP Global Settings page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse STP Global Settings HTML: %w", err)
	}

	return parseSTPGlobalSettings(doc)
}

// UpdateSTPSettings updates the STP global settings.
func (client *HRUIClient) UpdateSTPSettings(stp *STPGlobalSettings) error {
	data := url.Values{
		"cmd":      {"stp"},
		"version":  {stp.ForceVersion},
		"priority": {strconv.Itoa(stp.Priority)},
		"maxage":   {strconv.Itoa(stp.MaxAge)},
		"hello":    {strconv.Itoa(stp.HelloTime)},
		"delay":    {strconv.Itoa(stp.ForwardDelay)},
	}

	resp, err := client.HttpClient.PostForm(client.URL+"/loop.cgi?page=stp_global", data)
	if err != nil {
		return fmt.Errorf("failed to update STP global settings: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (client *HRUIClient) GetSTPPortSettings() ([]STPPort, error) {
	stpURL := client.URL + "/loop.cgi?page=stp_port"
	resp, err := client.HttpClient.Get(stpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch STP port settings page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var stpPorts []STPPort
	doc.Find("table").First().Find("tr").Each(func(i int, tr *goquery.Selection) {
		if i <= 1 { // Skip header rows
			return
		}

		tds := tr.Find("td")
		if tds.Length() < 10 {
			return
		}

		port := parseInt(strings.Split(strings.TrimSpace(tds.Eq(0).Text()), " ")[1])
		state := strings.TrimSpace(tds.Eq(1).Text())
		role := strings.TrimSpace(tds.Eq(2).Text())
		pathCost := parseInt(strings.TrimSpace(tds.Eq(4).Text()))
		priority := parseInt(strings.TrimSpace(tds.Eq(5).Text()))
		p2pConfig := strings.TrimSpace(tds.Eq(6).Text())
		p2pActual := strings.TrimSpace(tds.Eq(7).Text())
		edgeConfig := strings.TrimSpace(tds.Eq(8).Text())
		edgeActual := strings.TrimSpace(tds.Eq(9).Text())

		stpPorts = append(stpPorts, STPPort{
			Port:       port,
			State:      state,
			Role:       role,
			PathCost:   pathCost,
			Priority:   priority,
			P2P:        p2pConfig,
			P2PActual:  p2pActual,
			Edge:       edgeConfig,
			EdgeActual: edgeActual,
		})
	})

	return stpPorts, nil
}

// UpdateSTPPortSettings updates the STP settings for a specific port.
func (client *HRUIClient) UpdateSTPPortSettings(portID, pathCost, priority int, p2p, edge string) error {
	postData := url.Values{}
	postData.Set("cmd", "stp_port")
	postData.Set("portid", strconv.Itoa(portID))
	postData.Set("cost", strconv.Itoa(pathCost))
	postData.Set("priority", strconv.Itoa(priority))
	postData.Set("p2p", p2p)
	postData.Set("edge", edge)

	stpURL := client.URL + "/loop.cgi?page=stp_port"
	resp, err := client.HttpClient.PostForm(stpURL, postData)
	if err != nil {
		return fmt.Errorf("failed to update STP port settings: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update STP port settings: unexpected status %d", resp.StatusCode)
	}

	return nil
}

// parsePortStatuses parses the port table and returns a list of port statuses.
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

		port := parseInt(strings.TrimSpace(tds.Eq(0).Text()))
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
	if settings.Priority, err = extractIntAttribute(doc, "select[name='priority'] option[selected]", "value"); err != nil {
		return nil, err
	}

	// Parse MaxAge, HelloTime, ForwardDelay (uses attribute value)
	if settings.MaxAge, err = extractIntAttribute(doc, "input[name='maxage']", "value"); err != nil {
		return nil, err
	}
	if settings.HelloTime, err = extractIntAttribute(doc, "input[name='hello']", "value"); err != nil {
		return nil, err
	}
	if settings.ForwardDelay, err = extractIntAttribute(doc, "input[name='delay']", "value"); err != nil {
		return nil, err
	}

	// Parse Root Priority
	if settings.RootPriority, err = extractInt(doc, "th:contains('Root Priority') + td"); err != nil {
		return nil, err
	}

	// Parse Root MAC Address
	if settings.RootMAC, err = extractText(doc, "th:contains('Root MAC Address') + td"); err != nil {
		return nil, err
	}

	// Parse Root Path Cost
	if settings.RootPathCost, err = extractInt(doc, "th:contains('Root Path Cost') + td"); err != nil {
		return nil, err
	}

	// Parse Root Port
	if settings.RootPort, err = extractText(doc, "th:contains('Root Port') + td"); err != nil {
		return nil, err
	}

	// Parse RootMaxAge, removing units like "Sec"
	if maxAgeRaw, err := extractText(doc, "th:contains('Root Maximum Age') + td"); err != nil {
		return nil, err
	} else {
		settings.RootMaxAge = parseInt(strings.Fields(maxAgeRaw)[0]) // Split on space and take the first part
	}

	// Parse RootHelloTime, removing units like "Sec"
	if helloTimeRaw, err := extractText(doc, "th:contains('Root Hello Time') + td"); err != nil {
		return nil, err
	} else {
		settings.RootHelloTime = parseInt(strings.Fields(helloTimeRaw)[0])
	}

	// Parse RootForwardDelay, removing units like "Sec"
	if forwardDelayRaw, err := extractText(doc, "th:contains('Root Forward Delay') + td"); err != nil {
		return nil, err
	} else {
		settings.RootForwardDelay = parseInt(strings.Fields(forwardDelayRaw)[0])
	}

	log.Printf("[DEBUG] Parsed STPGlobalSettings: %+v", settings)
	return settings, nil
}

func parseInt(value string) int {
	if i, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
		return i
	}
	return 0
}

// parseBool interprets "enabled" or "true" as true, and anything else as false.
func parseBool(value string) bool {
	trimmedValue := strings.TrimSpace(strings.ToLower(value))
	return trimmedValue == "enabled" || trimmedValue == "true"
}

// extractInt extracts an integer from the text content of a given selector.
func extractInt(doc *goquery.Document, selector string) (int, error) {
	text, err := extractText(doc, selector)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(text))
}

// extractIntAttribute extracts an integer from the value of an attribute (e.g., `value`).
func extractIntAttribute(doc *goquery.Document, selector, attr string) (int, error) {
	value := doc.Find(selector).AttrOr(attr, "")
	return strconv.Atoi(strings.TrimSpace(value))
}
