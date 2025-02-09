package sdk

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Vlan struct {
	VlanID        int
	Name          string
	UntaggedPorts []string
	TaggedPorts   []string
	MemberPorts   []string
}

type PortVLANConfig struct {
	PortID          int
	PortName        string
	PVID            int
	AcceptFrameType string
}

// AddVLAN creates or updates a VLAN on the switch.
func (c *HRUIClient) AddVLAN(vlan *Vlan) error {
	if c == nil {
		return fmt.Errorf("HRUIClient is nil")
	}

	portConfigs, err := c.ListPortVLANConfigs()
	if err != nil {
		return fmt.Errorf("failed to get port VLAN configurations: %w", err)
	}

	form := url.Values{}
	form.Set("vid", fmt.Sprintf("%d", vlan.VlanID))
	form.Set("name", vlan.Name)

	for _, portConfig := range portConfigs {
		formValue := "2"

		if containsString(vlan.UntaggedPorts, portConfig.PortName) {
			formValue = "0"
		} else if containsString(vlan.TaggedPorts, portConfig.PortName) {
			formValue = "1"
		}

		form.Set(fmt.Sprintf("vlanPort_%d", portConfig.PortID), formValue)
	}

	vlanURL := fmt.Sprintf("%s/vlan.cgi?page=static", c.URL)
	_, err = c.FormRequest(vlanURL, form)
	if err != nil {
		return fmt.Errorf("failed to create/update VLAN: %w", err)
	}

	return nil
}

// Helper function to check if a slice contains a string.
func containsString(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// GetVLAN fetches a single VLAN by its VLAN ID.
func (c *HRUIClient) GetVLAN(vlanID int) (*Vlan, error) {
	vlans, err := c.ListVLANs()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch VLANs: %w", err)
	}

	for _, vlan := range vlans {
		if vlan.VlanID == vlanID {
			return vlan, nil
		}
	}

	return nil, fmt.Errorf("VLAN with ID %d not found", vlanID)
}

// ListVLANs fetches the list of VLANs using port names.
func (c *HRUIClient) ListVLANs() ([]*Vlan, error) {
	if c == nil {
		return nil, fmt.Errorf("HRUIClient is nil")
	}

	vlanURL := fmt.Sprintf("%s/vlan.cgi?page=static", c.URL)
	respBody, err := c.Request("GET", vlanURL, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch VLAN configuration from HRUI: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse VLAN HTML output: %w", err)
	}

	var vlans []*Vlan

	doc.Find("form[name='formVlanStatus'] table tr").Each(func(i int, s *goquery.Selection) {
		if i == 0 { // Skip table header row
			return
		}

		vlan := &Vlan{}

		// Extract VLAN ID from the <a> tag in the first table cell
		vlanIDText := strings.TrimSpace(s.Find("td:nth-child(1) a").Text())
		if vlanIDText == "" {
			fmt.Println("Skipping row: VLAN ID not found")
			return
		}

		vlanID, err := strconv.Atoi(vlanIDText)
		if err != nil {
			fmt.Printf("Error parsing VLAN ID (%s): %v\n", vlanIDText, err)
			return
		}
		vlan.VlanID = vlanID

		// Extract VLAN Name
		vlan.Name = strings.TrimSpace(s.Find("td:nth-child(2)").Text())

		// Extract port configurations
		vlan.MemberPorts = c.ParsePortRange(strings.TrimSpace(s.Find("td:nth-child(3)").Text()))
		vlan.TaggedPorts = c.ParsePortRange(strings.TrimSpace(s.Find("td:nth-child(4)").Text()))
		vlan.UntaggedPorts = c.ParsePortRange(strings.TrimSpace(s.Find("td:nth-child(5)").Text()))

		vlans = append(vlans, vlan)
	})

	return vlans, nil
}

// RemoveVLAN deletes a VLAN by its VLAN ID from the switch.
func (c *HRUIClient) RemoveVLAN(vlanID int) error {
	if c == nil {
		return fmt.Errorf("HRUIClient is nil")
	}

	form := url.Values{}
	form.Set(fmt.Sprintf("remove_%d", vlanID), "on")

	deleteURL := fmt.Sprintf("%s/vlan.cgi?page=getRmvVlanEntry", c.URL)

	_, err := c.FormRequest(deleteURL, form)
	if err != nil {
		return fmt.Errorf("failed to delete VLAN: %w", err)
	}

	return nil
}

// ListPortVLANConfigs fetches the VLAN configuration for all ports.
func (c *HRUIClient) ListPortVLANConfigs() ([]*PortVLANConfig, error) {
	if c == nil {
		return nil, fmt.Errorf("HRUIClient is nil")
	}

	if c.URL == "" {
		return nil, fmt.Errorf("HRUIClient.URL is empty")
	}

	portVLANURL := fmt.Sprintf("%s/vlan.cgi?page=port_based", c.URL)

	respBody, err := c.Request("GET", portVLANURL, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch port VLAN configuration from HRUI: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse port VLAN HTML output: %w", err)
	}

	var configs []*PortVLANConfig

	rows := doc.Find("table").Last().Find("tr").Slice(1, goquery.ToEnd)

	rows.Each(func(i int, s *goquery.Selection) {
		config := &PortVLANConfig{}

		// Extract the port name (e.g., "Port 1", "Trunk2").
		portName := strings.TrimSpace(s.Find("td:nth-child(1)").Text())
		if portName == "" {
			fmt.Printf("Skipping row %d due to missing port name\n", i)
			return
		}
		config.PortName = portName

		portID, err := c.GetPortByName(portName)
		if err != nil {
			fmt.Printf("Skipping row %d due to PortID resolution error for '%s': %v\n", i, portName, err)
			return
		}
		config.PortID = portID

		pvidText := strings.TrimSpace(s.Find("td:nth-child(2)").Text())
		var pvid int
		_, err = fmt.Sscanf(pvidText, "%d", &pvid)
		if err != nil {
			fmt.Printf("Skipping row %d due to PVID parsing error: %v\n", i, err)
			return
		}
		config.PVID = pvid

		config.AcceptFrameType = strings.TrimSpace(s.Find("td:nth-child(3)").Text())

		configs = append(configs, config)
	})

	return configs, nil
}

// GetPortVLANConfig fetches the VLAN configuration for a specific port on the switch.
func (c *HRUIClient) GetPortVLANConfig(port int) (*PortVLANConfig, error) {
	configs, err := c.ListPortVLANConfigs()
	if err != nil {
		return nil, fmt.Errorf("failed to get all port VLAN configs: %w", err)
	}

	for _, config := range configs {
		if config.PortID == port {
			return config, nil
		}
	}

	return nil, fmt.Errorf("port %d not found in %d ports", port, len(configs))
}

// SetPortVLANConfig sets the VLAN configuration for a specific port on the switch.
func (c *HRUIClient) SetPortVLANConfig(config *PortVLANConfig) error {
	if c == nil {
		return fmt.Errorf("HRUIClient is nil")
	}

	if config == nil {
		return fmt.Errorf("PortVLANConfig is nil")
	}

	// Resolve PortID if it's invalid or missing
	if config.PortID <= 0 {
		if config.PortName == "" {
			return fmt.Errorf("invalid configuration: both PortID and Name are missing")
		}

		portID, err := c.GetPortByName(config.PortName)
		if err != nil {
			return fmt.Errorf("failed to resolve PortID for Name '%s': %w", config.PortName, err)
		}

		config.PortID = portID
	}

	// Validate that PortID is now valid
	if config.PortID <= 0 {
		return fmt.Errorf("invalid PortID after resolution: %d", config.PortID)
	}

	// Prepare form data
	form := url.Values{}
	form.Set("ports", strconv.Itoa(config.PortID))
	form.Set("pvid", strconv.Itoa(config.PVID))

	// Map Accepted Frame Types to their respective values
	acceptFrameTypeMap := map[string]string{
		"All":        "0",
		"Tag-only":   "1",
		"Untag-only": "2",
	}

	frameTypeValue, ok := acceptFrameTypeMap[config.AcceptFrameType]
	if !ok {
		return fmt.Errorf("invalid Accepted Frame Type: %s", config.AcceptFrameType)
	}
	form.Set("vlan_accept_frame_type", frameTypeValue)

	// Submit the form
	portVLANURL := fmt.Sprintf("%s/vlan.cgi?page=port_based", c.URL)
	_, err := c.FormRequest(portVLANURL, form)
	if err != nil {
		return fmt.Errorf("failed to set port VLAN config: %w", err)
	}

	return nil
}

func (c *HRUIClient) ParsePortRange(portStr string) []string {
	ports := []string{}
	entries := strings.Split(portStr, ",")

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)

		if entry == "-" || entry == "" {
			continue
		}

		if strings.Contains(entry, "-") {
			rangeParts := strings.Split(entry, "-")
			if len(rangeParts) == 2 {
				start, err1 := strconv.Atoi(rangeParts[0])
				end, err2 := strconv.Atoi(rangeParts[1])
				if err1 == nil && err2 == nil {
					for i := start; i <= end; i++ {
						ports = append(ports, fmt.Sprintf("Port %d", i))
					}
				}
			}
		} else if strings.HasPrefix(entry, "Trunk") {
			// Keep named ports like "Trunk2"
			ports = append(ports, entry)
		} else {
			// Handle individual ports (e.g., "2" → "Port 2")
			if num, err := strconv.Atoi(entry); err == nil {
				ports = append(ports, fmt.Sprintf("Port %d", num))
			}
		}
	}
	return ports
}
