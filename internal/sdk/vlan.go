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

// Create VLAN creates or updates a VLAN on the switch, computing NotMemberPorts if needed.
func (c *HRUIClient) CreateVLAN(vlan *Vlan, totalPorts int) error {
	// Build the form data
	form := url.Values{}
	form.Set("vid", fmt.Sprintf("%d", vlan.VlanID))
	form.Set("name", vlan.Name)

	// Set untagged, tagged, and not member ports
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
	// Submit using PostForm method
	_, err := c.PostForm(vlanURL, form)
	if err != nil {
		return fmt.Errorf("failed to create/update VLAN: %w", err)
	}

	if c.Autosave {
		return c.SaveConfiguration()
	}

	return nil
}

// Helper function to check if a slice contains an integer
func contains(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// GetVLAN fetches a single VLAN by its VLAN ID by filtering results from GetAllVLANs.
func (c *HRUIClient) GetVLAN(vlanID int) (*Vlan, error) {
	vlans, err := c.GetAllVLANs()
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

// GetAllVLANs fetches the list of VLANs, setting member ports directly instead of notmemberports.
func (c *HRUIClient) GetAllVLANs() ([]*Vlan, error) {
	// URL for VLAN page
	vlanURL := fmt.Sprintf("%s/vlan.cgi?page=static", c.URL)

	resp, err := c.MakeRequest(vlanURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch VLAN configuration from HRUI: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
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

// DeleteVLAN deletes a VLAN by its VLAN ID from the switch.
func (c *HRUIClient) DeleteVLAN(vlanID int) error {
	form := url.Values{}
	form.Set(fmt.Sprintf("remove_%d", vlanID), "on")

	// Create the delete URL
	deleteURL := fmt.Sprintf("%s/vlan.cgi?page=getRmvVlanEntry", c.URL)

	_, err := c.PostForm(deleteURL, form)
	if err != nil {
		return fmt.Errorf("failed to delete VLAN: %w", err)
	}

	if c.Autosave {
		return c.SaveConfiguration()
	}

	return nil
}

// GetAllPortVLANConfigs fetches the VLAN configuration for all ports on the switch.
func (c *HRUIClient) GetAllPortVLANConfigs() ([]*PortVLANConfig, error) {
	portVLANURL := fmt.Sprintf("%s/vlan.cgi?page=port_based", c.URL)

	// Make the request to get the VLAN configuration page
	resp, err := c.MakeRequest(portVLANURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch port VLAN configuration from HRUI: %w", err)
	}
	defer resp.Body.Close()

	// Parse the HTML response
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse port VLAN HTML output: %w", err)
	}

	var configs []*PortVLANConfig

	// Select the rows from the table, skipping the header row
	rows := doc.Find("table").Last().Find("tr").Slice(1, goquery.ToEnd)

	// Iterate over each row and extract the VLAN configuration
	rows.Each(func(i int, s *goquery.Selection) {
		config := &PortVLANConfig{}

		// Extract Port number
		portText := strings.TrimSpace(s.Find("td:nth-child(1)").Text())
		fmt.Sscanf(portText, "Port %d", &config.Port)

		// Extract PVID
		pvidText := strings.TrimSpace(s.Find("td:nth-child(2)").Text())
		fmt.Sscanf(pvidText, "%d", &config.PVID)

		// Extract Accepted Frame Type
		config.AcceptFrameType = strings.TrimSpace(s.Find("td:nth-child(3)").Text())

		// Append the parsed configuration to the slice
		configs = append(configs, config)
	})

	return configs, nil
}

// GetPortVLANConfig fetches the VLAN configuration for a specific port on the switch.
func (c *HRUIClient) GetPortVLANConfig(port int) (*PortVLANConfig, error) {
	configs, err := c.GetAllPortVLANConfigs()
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
	_, err := c.PostForm(portVLANURL, form)
	if err != nil {
		return fmt.Errorf("failed to set port VLAN config: %w", err)
	}

	if c.Autosave {
		return c.SaveConfiguration()
	}

	return nil
}

// parsePortRangeSafe ensures that empty or invalid ports are handled, returning an empty slice if no ports exist
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
