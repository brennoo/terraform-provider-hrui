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
	UntaggedPorts []int
	TaggedPorts   []int
	MemberPorts   []int
}

type PortVLANConfig struct {
	PortID          int
	Name            string
	PVID            int
	AcceptFrameType string
}

// AddVLAN creates or updates a VLAN on the switch, computing NotMemberPorts if needed.
func (c *HRUIClient) AddVLAN(vlan *Vlan, totalPorts int) error {
	if c == nil {
		return fmt.Errorf("HRUIClient is nil")
	}

	// Build the form data
	form := url.Values{}
	form.Set("vid", fmt.Sprintf("%d", vlan.VlanID))
	form.Set("name", vlan.Name)

	// Set untagged, tagged
	for port := 1; port <= totalPorts; port++ {
		formValue := "2" // Default: Not Member
		if contains(vlan.UntaggedPorts, port) {
			formValue = "0" // Untagged
		} else if contains(vlan.TaggedPorts, port) {
			formValue = "1" // Tagged
		}

		form.Set(fmt.Sprintf("vlanPort_%d", port-1), formValue)
	}

	for _, trunkPort := range vlan.MemberPorts {
		if trunkPort > totalPorts {
			form.Set(fmt.Sprintf("vlanPort_%d", trunkPort), "1")
		}
	}

	vlanURL := fmt.Sprintf("%s/vlan.cgi?page=static", c.URL)

	_, err := c.FormRequest(vlanURL, form)
	if err != nil {
		return fmt.Errorf("failed to create/update VLAN: %w", err)
	}

	return nil
}

// Helper function to check if a slice contains an integer.
func contains(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// GetVLAN fetches a single VLAN by its VLAN ID by filtering results from ListVLANs.
func (c *HRUIClient) GetVLAN(vlanID int) (*Vlan, error) {
	vlans, err := c.ListVLANs()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch VLANs: %w", err)
	}

	// Search for the VLAN with the specified ID in the list of VLANs retrieved
	for _, vlan := range vlans {
		if vlan.VlanID == vlanID {
			return vlan, nil
		}
	}

	return nil, fmt.Errorf("VLAN with ID %d not found", vlanID)
}

// ListVLANs fetches the list of VLANs, setting member ports directly instead of notmemberports.
func (c *HRUIClient) ListVLANs() ([]*Vlan, error) {
	if c == nil {
		return nil, fmt.Errorf("HRUIClient is nil")
	}

	// URL for VLAN page
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

	// Select the table within the form with name "formVlanStatus"
	doc.Find("form[name='formVlanStatus'] table tr").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			// Skip header row
			return
		}

		vlan := &Vlan{}

		// Extract VLAN ID
		vlanIDText := strings.TrimSpace(s.Find("td:nth-child(1)").Text())
		vlanID, err := strconv.Atoi(vlanIDText)
		if err != nil {
			fmt.Printf("Error parsing VLAN ID: %v\n", err)
			return
		}
		vlan.VlanID = vlanID

		// Extract VLAN name
		vlan.Name = strings.TrimSpace(s.Find("td:nth-child(2)").Text())
		// Extract port ranges
		vlan.MemberPorts = c.parsePortRange(strings.TrimSpace(s.Find("td:nth-child(3)").Text()))
		vlan.TaggedPorts = c.parsePortRange(strings.TrimSpace(s.Find("td:nth-child(4)").Text()))
		vlan.UntaggedPorts = c.parsePortRange(strings.TrimSpace(s.Find("td:nth-child(5)").Text()))

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

// ListPortVLANConfigs fetches the VLAN configuration for all ports on the switch.
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

	// Iterate over each row and extract the VLAN configuration
	rows.Each(func(i int, s *goquery.Selection) {
		config := &PortVLANConfig{}

		// Extract the port name (e.g., "Port 1", "Trunk2").
		portName := strings.TrimSpace(s.Find("td:nth-child(1)").Text())
		if portName == "" {
			fmt.Printf("Skipping row %d due to missing port name\n", i)
			return
		}
		config.Name = portName

		// Use GetPortByName to fetch the numeric PortID for the resolved port name.
		portID, err := c.GetPortByName(portName)
		if err != nil {
			fmt.Printf("Skipping row %d due to PortID resolution error for '%s': %v\n", i, portName, err)
			return
		}

		parsedPortID, err := strconv.Atoi(portID)
		if err != nil {
			fmt.Printf("Skipping row %d due to PortID parsing error for '%s': %v\n", i, portName, err)
			return
		}
		config.PortID = parsedPortID

		// Extract PVID
		pvidText := strings.TrimSpace(s.Find("td:nth-child(2)").Text())
		var pvid int
		_, err = fmt.Sscanf(pvidText, "%d", &pvid)
		if err != nil {
			fmt.Printf("Skipping row %d due to PVID parsing error: %v\n", i, err)
			return
		}
		config.PVID = pvid

		// Extract Accepted Frame Type
		config.AcceptFrameType = strings.TrimSpace(s.Find("td:nth-child(3)").Text())

		// Append the parsed configuration to the slice
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

	// Resolve PortID using Name if PortID is invalid
	if config.PortID < 0 {
		if config.Name == "" {
			return fmt.Errorf("invalid configuration: both PortID and Name are missing")
		}

		portID, err := c.GetPortByName(config.Name)
		if err != nil {
			return fmt.Errorf("failed to resolve PortID for Name '%s': %w", config.Name, err)
		}

		portIDInt, err := strconv.Atoi(portID)
		if err != nil {
			return fmt.Errorf("invalid PortID '%s' for port '%s': %w", portID, config.Name, err)
		}

		config.PortID = portIDInt
	}

	// Ensure PortID is valid
	if config.PortID <= 0 {
		return fmt.Errorf("invalid PortID after resolution: %d", config.PortID)
	}

	form := url.Values{}
	form.Set("ports", fmt.Sprintf("%d", config.PortID))
	form.Set("pvid", fmt.Sprintf("%d", config.PVID))

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

func (c *HRUIClient) parsePortRange(portRange string) []int {
	if portRange == "-" || portRange == "" {
		return []int{}
	}

	var result []int
	for _, part := range strings.Split(portRange, ",") {
		subParts := strings.TrimSpace(part)

		// Check if the part is a trunk port (e.g., "Trunk1")
		if strings.HasPrefix(subParts, "Trunk") {
			// Retrieve trunk port ID dynamically
			trunkPort, err := c.GetPortByName(subParts)
			if err != nil {
				// Skip this trunk port if resolution fails (log optional)
				fmt.Printf("Error resolving trunk port %s: %v\n", subParts, err)
				continue
			}
			// Convert the resolved port ID to an integer
			if resolvedID, err := strconv.Atoi(trunkPort); err == nil {
				result = append(result, resolvedID)
			}
			continue
		}

		// Parse ranges or single numbers
		rangeParts := strings.Split(subParts, "-")
		if len(rangeParts) == 2 { // It's a range, e.g., "1-3"
			start, err1 := strconv.Atoi(rangeParts[0])
			end, err2 := strconv.Atoi(rangeParts[1])
			if err1 == nil && err2 == nil {
				for i := start; i <= end; i++ {
					result = append(result, i)
				}
			}
		} else if len(rangeParts) == 1 { // Single number
			port, err := strconv.Atoi(rangeParts[0])
			if err == nil {
				result = append(result, port)
			}
		}
	}
	return result
}
