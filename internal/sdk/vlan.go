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
	Port            int
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
			// Handle error, e.g., log a warning or skip the row
			fmt.Printf("Error parsing VLAN ID: %v\n", err)
			return
		}
		vlan.VlanID = vlanID

		// Extract VLAN name
		vlan.Name = strings.TrimSpace(s.Find("td:nth-child(2)").Text())

		// Extract port ranges
		vlan.MemberPorts = parsePortRangeSafe(strings.TrimSpace(s.Find("td:nth-child(3)").Text()))
		vlan.TaggedPorts = parsePortRangeSafe(strings.TrimSpace(s.Find("td:nth-child(4)").Text()))
		vlan.UntaggedPorts = parsePortRangeSafe(strings.TrimSpace(s.Find("td:nth-child(5)").Text()))

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

		// Extract Port number
		portText := strings.TrimSpace(s.Find("td:nth-child(1)").Text())
		var port int
		_, err := fmt.Sscanf(portText, "Port %d", &port)
		if err != nil {
			fmt.Println("Failed to parse port number:", err)
			return
		}
		config.Port = port

		// Extract PVID
		pvidText := strings.TrimSpace(s.Find("td:nth-child(2)").Text())
		var pvid int
		_, err = fmt.Sscanf(pvidText, "%d", &pvid)
		if err != nil {
			fmt.Println("Failed to parse PVID:", err)
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
		if config.Port == port {
			return config, nil
		}
	}

	return nil, fmt.Errorf("port %d not found in %d ports", port, len(configs))
}

// SetPortVLANConfig sets the VLAN configuration for a specific port on the switch.
func (c *HRUIClient) SetPortVLANConfig(config *PortVLANConfig) error {
	form := url.Values{}
	form.Set("ports", fmt.Sprintf("%d", config.Port-1))
	form.Set("pvid", fmt.Sprintf("%d", config.PVID))

	acceptFrameTypeMap := map[string]string{
		"All":        "0",
		"Tag-only":   "1",
		"Untag-only": "2",
	}

	form.Set("vlan_accept_frame_type", acceptFrameTypeMap[config.AcceptFrameType])

	// Submit the form
	portVLANURL := fmt.Sprintf("%s/vlan.cgi?page=port_based", c.URL)
	_, err := c.FormRequest(portVLANURL, form)
	if err != nil {
		return fmt.Errorf("failed to set port VLAN config: %w", err)
	}

	return nil
}

// parsePortRangeSafe ensures that empty or invalid ports are handled, returning an empty slice if no ports exist.
func parsePortRangeSafe(portRange string) []int {
	if portRange == "-" || portRange == "" {
		return []int{}
	}

	var result []int
	for _, part := range strings.Split(portRange, ",") { // Split by comma
		subParts := strings.Split(part, "-")
		if len(subParts) == 2 {
			start, _ := strconv.Atoi(subParts[0])
			end, _ := strconv.Atoi(subParts[1])
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
		} else if len(subParts) == 1 {
			port, _ := strconv.Atoi(subParts[0])
			result = append(result, port)
		}
	}
	return result
}
