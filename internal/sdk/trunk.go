package sdk

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type TrunkConfig struct {
	ID    int
	Type  string
	Ports []int
}

// ListAvailableTrunks fetches available Trunks on the device.
func (c *HRUIClient) ListAvailableTrunks() ([]TrunkConfig, error) {
	// Fetch the HTML page
	endpoint := c.URL + "/trunk.cgi?page=group"
	respBody, err := c.Request("GET", endpoint, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch trunk page: %w", err)
	}

	// Parse the HTML response
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML response: %w", err)
	}

	// Find the `<select name="id">` dropdown and extract trunk IDs
	var trunkConfigs []TrunkConfig
	doc.Find("select[name='id'] option").Each(func(i int, s *goquery.Selection) {
		// Get the trunk ID value
		id, _ := strconv.Atoi(s.AttrOr("value", "0"))

		// Add the trunk ID as a TrunkConfig
		trunkConfigs = append(trunkConfigs, TrunkConfig{
			ID: id,
		})
	})

	// Return the available trunk configurations
	return trunkConfigs, nil
}

// ConfigureTrunk sends configuration for a Trunk.
func (c *HRUIClient) ConfigureTrunk(config *TrunkConfig) error {
	form := url.Values{}

	// Set trunk group ID and type
	form.Set("id", strconv.Itoa(config.ID))
	form.Set("trunk_type", strconv.Itoa(parseTrunkType(config.Type)))

	// Convert ports to 0-indexed before sending them to the API
	for _, port := range config.Ports {
		form.Add("ports", strconv.Itoa(port-1))
	}

	form.Set("cmd", "trunk")

	// Endpoint for creating/modifying trunk groups
	endpoint := c.URL + "/trunk.cgi?page=group"

	// Use FormRequest to send the form and handle errors
	_, err := c.FormRequest(endpoint, form)
	if err != nil {
		return fmt.Errorf("failed to configure trunk: %w", err)
	}

	return nil
}

// Helper function to map trunk type strings to integers.
func parseTrunkType(trunkType string) int {
	switch trunkType {
	case "static":
		return 0
	case "LACP":
		return 1
	default:
		return 0
	}
}

// DeleteTrunk sends a delete request for a Trunk.
func (c *HRUIClient) DeleteTrunk(id int) error {
	form := url.Values{}
	form.Set("id", strconv.Itoa(id))
	form.Set("cmd", "group_remove")

	endpoint := c.URL + "/trunk.cgi?page=group_remove"
	_, err := c.FormRequest(endpoint, form)

	return err
}

// GetTrunk fetches details of a configured Trunk by its ID.
func (c *HRUIClient) GetTrunk(id int) (*TrunkConfig, error) {
	endpoint := c.URL + "/trunk.cgi?page=group"
	respBody, err := c.Request("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, err
	}

	// Find the row containing the trunk information
	var trunk *TrunkConfig
	doc.Find("form[action='/trunk.cgi?page=group_remove'] table tr").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			return
		}

		trunkIDText := strings.TrimSpace(s.Find("td").Eq(0).Text())
		if trunkIDText != fmt.Sprintf("Trunk%d", id) {
			return
		}

		trunkType := strings.TrimSpace(s.Find("td").Eq(1).Text())
		portsText := strings.TrimSpace(s.Find("td").Eq(2).Text())

		var ports []int
		for _, p := range strings.Split(portsText, ",") {
			port := parseInt(p, WithDefaultValue(0))
			ports = append(ports, *port)
		}

		trunk = &TrunkConfig{
			ID:    id,
			Type:  trunkType,
			Ports: ports,
		}
	})

	if trunk == nil {
		return nil, errors.New("trunk not found")
	}

	return trunk, nil
}

// ListConfiguredTrunks fetches configured Trunks from the device.
func (c *HRUIClient) ListConfiguredTrunks() ([]TrunkConfig, error) {
	endpoint := c.URL + "/trunk.cgi?page=group"
	respBody, err := c.Request("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, err
	}

	var trunkConfigs []TrunkConfig
	doc.Find("form[action='/trunk.cgi?page=group_remove'] table tr").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			return
		}

		idText := strings.TrimSpace(s.Find("td").Eq(0).Text())
		id := parseInt(idText, WithTrimPrefix("Trunk"), WithDefaultValue(0))

		trunkType := strings.TrimSpace(s.Find("td").Eq(1).Text())
		portsText := strings.TrimSpace(s.Find("td").Eq(2).Text())

		var ports []int
		for _, p := range strings.Split(portsText, ",") {
			port := parseInt(p, WithDefaultValue(0))
			ports = append(ports, *port)
		}

		trunkConfigs = append(trunkConfigs, TrunkConfig{
			ID:    *id,
			Type:  trunkType,
			Ports: ports,
		})
	})

	return trunkConfigs, nil
}
