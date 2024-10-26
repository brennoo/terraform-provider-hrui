package sdk

import (
	"fmt"
	"net/http"
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

// Create VLAN creates or updates a VLAN on the switch, computing NotMemberPorts if needed.
func (c *HRUIClient) CreateVLAN(vlan *Vlan, totalPorts int) error {
	// Compute NotMemberPorts: ports that are neither tagged nor untagged
	allAssignablePorts := generatePortRange(1, totalPorts)
	notMemberPorts := computeNotMemberPorts(vlan.UntaggedPorts, vlan.TaggedPorts, allAssignablePorts)

	// Build the form data
	form := url.Values{}
	form.Set("vid", fmt.Sprintf("%d", vlan.VlanID))
	form.Set("name", vlan.Name)

	// Set untagged and tagged ports
	for _, port := range vlan.UntaggedPorts {
		form.Set(fmt.Sprintf("vlanPort_%d", port), "0")
	}

	for _, port := range vlan.TaggedPorts {
		form.Set(fmt.Sprintf("vlanPort_%d", port), "1")
	}

	// Add NotMemberPorts (members explicitly marked as not part of the VLAN)
	for _, port := range notMemberPorts {
		form.Set(fmt.Sprintf("vlanPort_%d", port), "2")
	}

	// Submit the form
	vlanURL := fmt.Sprintf("%s/vlan.cgi?page=static", c.URL)
	httpReq, err := http.NewRequest(http.MethodPost, vlanURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create/update VLAN request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Perform the request
	httpResp, err := c.HttpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to submit VLAN form: %w", err)
	}
	defer httpResp.Body.Close()

	// Check for successful submission
	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create/update VLAN, received status code: %d", httpResp.StatusCode)
	}

	if c.Autosave {
		return c.SaveConfiguration()
	}

	return nil
}

// Helper function to compute not member ports based on total ports available on the switch
func computeNotMemberPorts(untagged, tagged, allPorts []int) []int {
	// Create a map to track all ports that are tagged or untagged
	memberPortsMap := make(map[int]struct{})
	for _, port := range untagged {
		memberPortsMap[port] = struct{}{}
	}
	for _, port := range tagged {
		memberPortsMap[port] = struct{}{}
	}

	// Find ports that are not tagged or untagged
	var notMemberPorts []int
	for _, port := range allPorts {
		if _, ok := memberPortsMap[port]; !ok {
			notMemberPorts = append(notMemberPorts, port)
		}
	}

	return notMemberPorts
}

// Helper function to generate a range of port numbers (1 - totalports)
func generatePortRange(start, end int) []int {
	var ports []int
	for i := start; i <= end; i++ {
		ports = append(ports, i)
	}
	return ports
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
	doc.Find("form[name=formVlanStatus] table tr").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			// Skip header row
			return
		}

		vlan := &Vlan{}
		s.Find("td").Each(func(j int, td *goquery.Selection) {
			text := strings.TrimSpace(td.Text())
			switch j {
			case 0:
				fmt.Sscanf(text, "%d", &vlan.VlanID)
			case 1:
				vlan.Name = text
			case 2:
				vlan.MemberPorts = parsePortRangeSafe(text)
			case 3:
				vlan.TaggedPorts = parsePortRangeSafe(text)
			case 4:
				vlan.UntaggedPorts = parsePortRangeSafe(text)
			}
		})
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
	httpReq, err := http.NewRequest(http.MethodPost, deleteURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create delete VLAN request: %w", err)
	}

	// Set the appropriate content type
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpResp, err := c.HttpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to delete VLAN: %w", err)
	}
	defer httpResp.Body.Close()

	// Check for the correct status code
	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete VLAN, status code: %d", httpResp.StatusCode)
	}

	if c.Autosave {
		return c.SaveConfiguration()
	}

	return nil
}

// parsePortRangeSafe ensures that empty or invalid ports are handled, returning an empty slice if no ports exist
func parsePortRangeSafe(portRange string) []int {
	// Handle empty `tagged_ports` or invalid entries like `"-"` resolving to an empty list (not `[0]`)
	if portRange == "-" || portRange == "" {
		return []int{}
	}

	// Otherwise, parse the port range into a list of ints
	parts := strings.Split(portRange, "-")
	if len(parts) == 2 {
		start, _ := strconv.Atoi(parts[0])
		end, _ := strconv.Atoi(parts[1])
		var result []int
		for i := start; i <= end; i++ {
			result = append(result, i)
		}
		return result
	} else if len(parts) == 1 {
		port, _ := strconv.Atoi(parts[0])
		return []int{port}
	}
	return []int{}
}
