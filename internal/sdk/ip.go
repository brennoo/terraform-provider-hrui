package sdk

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

type IPAddressSettings struct {
	DHCPEnabled bool
	IPAddress   string
	Netmask     string
	Gateway     string
}

func (c *HRUIClient) GetIPAddressSettings() (*IPAddressSettings, error) {
	ipSettingsURL := fmt.Sprintf("%s/ip.cgi", c.URL)

	resp, err := c.MakeRequest(ipSettingsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch IP Settings configuration: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse IP Settings HTML output: %w", err)
	}

	settings := &IPAddressSettings{}

	settings.DHCPEnabled = false
	doc.Find("select[name='dhcp_state'] option").Each(func(i int, s *goquery.Selection) {
		if _, ok := s.Attr("selected"); ok {
			dhcpEnabledStr, _ := s.Attr("value")
			settings.DHCPEnabled, _ = strconv.ParseBool(dhcpEnabledStr)
		}
	})
	doc.Find("input[name=ip]").Each(func(i int, s *goquery.Selection) {
		ipValue, _ := s.Attr("value")
		settings.IPAddress = ipValue
	})
	doc.Find("input[name=netmask]").Each(func(i int, s *goquery.Selection) {
		netmaskValue, _ := s.Attr("value")
		settings.Netmask = netmaskValue
	})
	doc.Find("input[name=gateway]").Each(func(i int, s *goquery.Selection) {
		gatewayValue, _ := s.Attr("value")
		settings.Gateway = gatewayValue
	})

	return settings, nil
}

func (c *HRUIClient) UpdateIPAddressSettings(settings *IPAddressSettings) error {
	form := url.Values{}
	form.Set("dhcp_state", "0")
	if settings.DHCPEnabled {
		form.Set("dhcp_state", "1")
	}
	form.Set("ip", settings.IPAddress)
	form.Set("netmask", settings.Netmask)
	form.Set("gateway", settings.Gateway)

	ipSettingsURL := fmt.Sprintf("%s/ip.cgi", c.URL)
	_, err := c.PostForm(ipSettingsURL, form)
	if err != nil {
		return fmt.Errorf("failed to update IP Settings: %w", err)
	}

	if c.Autosave {
		return c.SaveConfiguration()
	}

	return nil
}
