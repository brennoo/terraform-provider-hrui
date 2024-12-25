package sdk

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

// Global lock to serialize IGMP port updates.
var igmpUpdateLock sync.Mutex

// IGMPConfig represents the IGMP Snooping configuration.
type IGMPConfig struct {
	Enabled bool
	Ports   []bool
}

// UpdateIGMPSnooping updates the global IGMP snooping setting (enable or disable).
func (c *HRUIClient) UpdateIGMPSnooping(enable bool) error {
	formData := url.Values{}
	if enable {
		formData.Set("enable_igmp", "on")
	}

	url := fmt.Sprintf("%s/igmp.cgi?page=enable_igmp", c.URL)

	_, err := c.ExecuteFormRequest(url, formData)
	if err != nil {
		return fmt.Errorf("failed to update global IGMP snooping: %w", err)
	}

	if c.Autosave {
		return c.SaveConfiguration()
	}
	return nil
}

// EnableIGMPSnooping enables IGMP snooping globally.
func (c *HRUIClient) EnableIGMPSnooping() error {
	return c.UpdateIGMPSnooping(true)
}

// DisableIGMPSnooping disables IGMP snooping globally.
func (c *HRUIClient) DisableIGMPSnooping() error {
	return c.UpdateIGMPSnooping(false)
}

// UpdatePortIGMPSnooping enables or disables IGMP snooping for a specific port.
func (c *HRUIClient) UpdatePortIGMPSnooping(port int, enable bool) error {
	// Lock the operation to prevent race conditions
	igmpUpdateLock.Lock()
	defer igmpUpdateLock.Unlock()

	totalPorts, err := c.GetTotalPorts()
	if err != nil {
		return fmt.Errorf("failed to retrieve total number of ports: %w", err)
	}

	if port < 1 || port > totalPorts {
		return fmt.Errorf("port %d is out of range, valid ports are between 1 and %d", port, totalPorts)
	}

	// Fetch current configuration of all ports
	config, err := c.GetIGMPConfig()
	if err != nil {
		return fmt.Errorf("failed to retrieve current IGMP configuration: %w", err)
	}

	// Update the specified port's state
	portIndex := port - 1
	config.Ports[portIndex] = enable

	// Construct the complete payload for all ports
	formData := url.Values{}
	formData.Set("cmd", "set")
	for i, enabled := range config.Ports {
		if enabled {
			formData.Set(fmt.Sprintf("lPort_%d", i), "on")
		}
	}

	url := fmt.Sprintf("%s/igmp.cgi?page=igmp_static_router", c.URL)

	_, err = c.ExecuteFormRequest(url, formData)
	if err != nil {
		return fmt.Errorf("failed to update IGMP snooping for port %d: %w", port, err)
	}

	if c.Autosave {
		return c.SaveConfiguration()
	}
	return nil
}

// EnablePortIGMPSnooping enables IGMP snooping for a specific port.
func (c *HRUIClient) EnablePortIGMPSnooping(port int) error {
	return c.UpdatePortIGMPSnooping(port, true)
}

// DisablePortIGMPSnooping disables IGMP snooping for a specific port.
func (c *HRUIClient) DisablePortIGMPSnooping(port int) error {
	return c.UpdatePortIGMPSnooping(port, false)
}

// GetPortIGMPSnooping retrieves the IGMP snooping status for a specific port.
func (c *HRUIClient) GetPortIGMPSnooping(port int) (bool, error) {
	totalPorts, err := c.GetTotalPorts()
	if err != nil {
		return false, fmt.Errorf("failed to retrieve total ports: %w", err)
	}

	if port < 1 || port > totalPorts {
		return false, fmt.Errorf("port %d is out of range, valid ports are between 1 and %d", port, totalPorts)
	}

	url := fmt.Sprintf("%s/igmp.cgi?page=dump", c.URL)

	// Use ExecuteRequest to fetch the IGMP configuration
	respBody, err := c.ExecuteRequest("GET", url, nil, nil)
	if err != nil {
		return false, fmt.Errorf("failed to fetch IGMP configuration for port %d: %w", port, err)
	}

	// Parse the HTML document using goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return false, fmt.Errorf("failed to parse IGMP configuration HTML: %w", err)
	}

	var portStatus bool
	doc.Find("tr").Each(func(_ int, tr *goquery.Selection) {
		tr.Find("th").Each(func(_ int, th *goquery.Selection) {
			if strings.TrimSpace(th.Text()) == "static" {
				tr.Find("td input").Each(func(portIndex int, input *goquery.Selection) {
					if portIndex == port-1 { // Match the zero-indexed port number
						_, portStatus = input.Attr("checked") // 'checked' indicates enabled
					}
				})
			}
		})
	})

	return portStatus, nil
}

// GetIGMPConfig retrieves the full IGMP snooping configuration for the device.
func (c *HRUIClient) GetIGMPConfig() (*IGMPConfig, error) {
	url := fmt.Sprintf("%s/igmp.cgi?page=dump", c.URL)

	// Use ExecuteRequest to fetch the IGMP configuration
	respBody, err := c.ExecuteRequest("GET", url, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch IGMP configuration: %w", err)
	}

	// Parse the HTML document using goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse IGMP configuration HTML: %w", err)
	}

	var globalEnabled bool
	doc.Find("input[name=enable_igmp]").Each(func(_ int, s *goquery.Selection) {
		_, globalEnabled = s.Attr("checked")
	})

	totalPorts, err := c.GetTotalPorts()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve total ports: %w", err)
	}

	portStatus := make([]bool, totalPorts)
	doc.Find("tr").Each(func(_ int, tr *goquery.Selection) {
		tr.Find("th").Each(func(_ int, th *goquery.Selection) {
			if strings.TrimSpace(th.Text()) == "static" {
				tr.Find("td input").Each(func(portIndex int, input *goquery.Selection) {
					if portIndex < totalPorts {
						_, isChecked := input.Attr("checked")
						portStatus[portIndex] = isChecked
					}
				})
			}
		})
	})

	return &IGMPConfig{
		Enabled: globalEnabled,
		Ports:   portStatus,
	}, nil
}
