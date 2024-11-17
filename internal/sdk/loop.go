package sdk

import (
	"fmt"
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

// LoopProtocol holds loop protocol settings.
type LoopProtocol struct {
	LoopFunction string       // "Off", "Loop Detection", "Loop Prevention", "Spanning Tree"
	IntervalTime int          // Interval time (only relevant for Loop Detection/Prevention)
	RecoverTime  int          // Recover time (only relevant for Loop Detection/Prevention)
	PortStatuses []PortStatus // Per-port Loop Prevention statuses
}

// PortStatus represents the status of a port under Loop Protocol control.
type PortStatus struct {
	Port       int    // Port number (e.g., 1, 2, 3...)
	Enable     bool   // Whether Loop Prevention is enabled on this port
	LoopState  string // Loop state ("Enable", "Disable")
	LoopStatus string // Loop status ("Forwarding", "Blocked", etc.)
}

// STPGlobalSettings holds the parsed Spanning Tree Protocol (STP) global settings.
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

// UpdateLoopProtocol updates the loop function and associated port settings.
func (client *HRUIClient) UpdateLoopProtocol(loopFunction string, intervalTime, recoverTime int, portStatuses []PortStatus) error {
	funcType, valid := LoopFunctionType[loopFunction]
	if !valid {
		return fmt.Errorf("invalid loop_function: %s", loopFunction)
	}

	// Prepare form data to update the main loop settings.
	data := url.Values{}
	data.Set("cmd", "loop")
	data.Set("func_type", strconv.Itoa(funcType))

	// Only apply interval/recovery times if using Loop Prevention.
	if funcType == 2 { // Loop Prevention
		data.Set("interval_time", strconv.Itoa(intervalTime))
		data.Set("recover_time", strconv.Itoa(recoverTime))
	}

	// Send the main loop protocol update.
	loopURL := client.URL + "/loop.cgi"
	resp, err := client.HttpClient.PostForm(loopURL, data)
	if err != nil {
		return fmt.Errorf("failed to update loop protocol: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Update port-specific loop settings.
	for _, port := range portStatuses {
		portData := url.Values{}
		portData.Set("cmd", "rlp")
		portData.Set("portid", strconv.Itoa(port.Port))
		portData.Set("portEnable", strconv.Itoa(boolToInt(port.Enable)))

		resp, err := client.HttpClient.PostForm(client.URL+"/loop_port.cgi", portData)
		if err != nil {
			return fmt.Errorf("failed to update port loop prevention: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to update port loop prevention: unexpected status %d", resp.StatusCode)
		}
	}

	return nil
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

	// Extract loop function setting.
	doc.Find(`select[name="func_type"] option[selected]`).Each(func(i int, s *goquery.Selection) {
		protocol.LoopFunction = strings.TrimSpace(s.Text())
	})

	// Parse interval and recovery times if Loop Prevention is enabled.
	if protocol.LoopFunction == "Loop Prevention" {
		doc.Find(`input[name="interval_time"]`).Each(func(i int, s *goquery.Selection) {
			protocol.IntervalTime = atoiSafe(s.AttrOr("value", "0"))
		})

		doc.Find(`input[name="recover_time"]`).Each(func(i int, s *goquery.Selection) {
			protocol.RecoverTime = atoiSafe(s.AttrOr("value", "0"))
		})

		// Parse port statuses.
		protocol.PortStatuses = parsePortStatuses(doc)
	}

	return protocol, nil
}

// parsePortStatuses parses the port table and returns a list of port statuses.
func parsePortStatuses(doc *goquery.Document) []PortStatus {
	var portStatuses []PortStatus

	// Assuming the first table contains the port data.
	doc.Find("table").First().Find("tr").Each(func(i int, tr *goquery.Selection) {
		if i == 0 { // Skip header
			return
		}

		tds := tr.Find("td")
		if tds.Length() < 3 {
			return
		}

		port := atoiSafe(strings.TrimSpace(tds.Eq(0).Text()))
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

// GetSTPPortSettings fetches and parses STP port settings.
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

	return parseSTPPorts(doc), nil
}

// parseSTPPorts parses STP port settings from the document.
func parseSTPPorts(doc *goquery.Document) []STPPort {
	var stpPorts []STPPort

	doc.Find("table").First().Find("tr").Each(func(i int, tr *goquery.Selection) {
		if i <= 1 { // Skip header rows
			return
		}

		tds := tr.Find("td")
		if tds.Length() < 10 {
			return
		}

		port := atoiSafe(strings.Split(strings.TrimSpace(tds.Eq(0).Text()), " ")[1])
		state := strings.TrimSpace(tds.Eq(1).Text())
		role := strings.TrimSpace(tds.Eq(2).Text())
		pathCost := atoiSafe(strings.TrimSpace(tds.Eq(4).Text()))
		priority := atoiSafe(strings.TrimSpace(tds.Eq(5).Text()))
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

	return stpPorts
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

// GetSTPSettings fetches and parses the STP Global Settings page.
func (client *HRUIClient) GetSTPSettings() (*STPGlobalSettings, error) {
	stpURL := client.URL + "/loop.cgi?page=stp_global"
	resp, err := client.HttpClient.Get(stpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch STP Global Settings page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse STP Global Settings HTML: %w", err)
	}

	return parseSTPGlobalSettings(doc), nil
}

// parseSTPGlobalSettings converts the document into an STPGlobalSettings structure.
func parseSTPGlobalSettings(doc *goquery.Document) *STPGlobalSettings {
	stpSettings := &STPGlobalSettings{}

	doc.Find("th:contains('Spanning Tree Status')").Next().Each(func(i int, s *goquery.Selection) {
		stpSettings.STPStatus = strings.TrimSpace(s.Text())
	})

	doc.Find("select[name='version'] option[selected]").Each(func(i int, s *goquery.Selection) {
		stpSettings.ForceVersion = strings.TrimSpace(s.Text())
	})

	doc.Find("select[name='priority'] option[selected]").Each(func(i int, s *goquery.Selection) {
		stpSettings.Priority = atoiSafe(strings.TrimSpace(s.Text()))
	})

	doc.Find("input[name='maxage']").Each(func(i int, s *goquery.Selection) {
		stpSettings.MaxAge = atoiSafe(s.AttrOr("value", "0"))
	})

	doc.Find("input[name='hello']").Each(func(i int, s *goquery.Selection) {
		stpSettings.HelloTime = atoiSafe(s.AttrOr("value", "0"))
	})

	doc.Find("input[name='delay']").Each(func(i int, s *goquery.Selection) {
		stpSettings.ForwardDelay = atoiSafe(s.AttrOr("value", "0"))
	})

	doc.Find("th:contains('Root Priority')").Next().Each(func(i int, s *goquery.Selection) {
		stpSettings.RootPriority = atoiSafe(strings.TrimSpace(s.Text()))
	})

	doc.Find("th:contains('Root MAC Address')").Next().Each(func(i int, s *goquery.Selection) {
		stpSettings.RootMAC = strings.TrimSpace(s.Text())
	})

	doc.Find("th:contains('Root Path Cost')").Next().Each(func(i int, s *goquery.Selection) {
		stpSettings.RootPathCost = atoiSafe(strings.TrimSpace(s.Text()))
	})

	doc.Find("th:contains('Root Port')").Next().Each(func(i int, s *goquery.Selection) {
		stpSettings.RootPort = strings.TrimSpace(s.Text())
	})

	doc.Find("th:contains('Root Maximum Age')").Next().Each(func(i int, s *goquery.Selection) {
		stpSettings.RootMaxAge = atoiSafe(strings.Split(strings.TrimSpace(s.Text()), " ")[0])
	})

	doc.Find("th:contains('Root Hello Time')").Next().Each(func(i int, s *goquery.Selection) {
		stpSettings.RootHelloTime = atoiSafe(strings.Split(strings.TrimSpace(s.Text()), " ")[0])
	})

	doc.Find("th:contains('Root Forward Delay')").Next().Each(func(i int, s *goquery.Selection) {
		stpSettings.RootForwardDelay = atoiSafe(strings.Split(strings.TrimSpace(s.Text()), " ")[0])
	})

	return stpSettings
}

// UpdateSTPSettings sends a POST request to update the STP Global Settings.
func (client *HRUIClient) UpdateSTPSettings(stp *STPGlobalSettings) error {
	stpURL := client.URL + "/loop.cgi?page=stp_global"

	data := url.Values{}
	data.Set("version", stp.GetVersionValue())
	data.Set("priority", strconv.Itoa(stp.Priority))
	data.Set("maxage", strconv.Itoa(stp.MaxAge))
	data.Set("hello", strconv.Itoa(stp.HelloTime))
	data.Set("delay", strconv.Itoa(stp.ForwardDelay))
	data.Set("cmd", "stp")

	req, err := http.NewRequest("POST", stpURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create POST request: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error updating STP Global Settings, status code: %d", resp.StatusCode)
	}

	return nil
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

// Helper to safely convert a string to an int.
func atoiSafe(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// Helper to convert a boolean to an integer (like a ternary operator).
func boolToInt(condition bool) int {
	if condition {
		return 1
	}
	return 0
}
