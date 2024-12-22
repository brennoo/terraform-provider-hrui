package sdk

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// IGMPConfig represents the IGMP Snooping configuration.
type IGMPConfig struct {
	Enabled bool
	Ports   []bool
}

// UpdateIGMPSnooping updates the global IGMP snooping setting (enable or disable).
func (c *HRUIClient) UpdateIGMPSnooping(enable bool) error {
	form := "enable_igmp=on"
	if !enable {
		form = ""
	}

	url := fmt.Sprintf("%s/igmp.cgi?page=enable_igmp", c.URL)
	resp, err := c.HttpClient.Post(url, "application/x-www-form-urlencoded", bytes.NewBufferString(form))
	if err != nil {
		return fmt.Errorf("failed to update global IGMP snooping: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code %d while updating global IGMP snooping", resp.StatusCode)
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
	totalPorts, err := c.GetTotalPorts()
	if err != nil {
		return fmt.Errorf("failed to retrieve total number of ports: %w", err)
	}

	if port < 1 || port > totalPorts {
		return fmt.Errorf("port %d is out of range, valid ports are between 1 and %d", port, totalPorts)
	}

	form := []string{"cmd=set"}
	if enable {
		form = append(form, fmt.Sprintf("lPort_%d=on", port-1))
	}

	url := fmt.Sprintf("%s/igmp.cgi?page=igmp_static_router", c.URL)
	resp, err := c.HttpClient.Post(url, "application/x-www-form-urlencoded", bytes.NewBufferString(strings.Join(form, "&")))
	if err != nil {
		return fmt.Errorf("failed to update IGMP snooping for port %d: %w", port, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code %d while updating IGMP snooping for port %d", resp.StatusCode, port)
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
	resp, err := c.HttpClient.Get(url)
	if err != nil {
		return false, fmt.Errorf("failed to fetch IGMP configuration for port %d: %w", port, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, fmt.Errorf("unexpected status code %d while fetching IGMP configuration", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
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
	resp, err := c.HttpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch IGMP configuration: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code %d while fetching IGMP configuration", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse IGMP configuration HTML: %w", err)
	}

	globalEnabled := false
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
