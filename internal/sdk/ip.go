package sdk

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// IPAddressSettings represents the IP configuration.
type IPAddressSettings struct {
	DHCPEnabled bool
	IPAddress   string
	Netmask     string
	Gateway     string
}

// GetIPAddressSettings retrieves the IP address settings from the HRUI server.
func (c *HRUIClient) GetIPAddressSettings(ctx context.Context) (*IPAddressSettings, error) {
	// Construct the IP settings URL
	ipSettingsURL := fmt.Sprintf("%s/ip.cgi", c.URL)

	// Execute GET request to fetch IP settings
	respBody, err := c.Request(ctx, "GET", ipSettingsURL, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch IP Settings configuration: %w", err)
	}

	// Parse the response body as HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse IP Settings HTML output: %w", err)
	}

	// Extract IP settings from HTML
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

	// On v2.x firmwares, the ip/netmask/gateway inputs are injected via JS innerHTML inside <script>.
	rawHTML := string(respBody)
	if settings.IPAddress == "" {
		if m := regexp.MustCompile(`name=['"]ip['"][^>]*value=['"]([^'"]+)['"]`).FindStringSubmatch(rawHTML); len(m) > 1 {
			settings.IPAddress = m[1]
		}
	}
	if settings.Netmask == "" {
		if m := regexp.MustCompile(`name=['"]netmask['"][^>]*value=['"]([^'"]+)['"]`).FindStringSubmatch(rawHTML); len(m) > 1 {
			settings.Netmask = m[1]
		}
	}
	if settings.Gateway == "" {
		if m := regexp.MustCompile(`name=['"]gateway['"][^>]*value=['"]([^'"]+)['"]`).FindStringSubmatch(rawHTML); len(m) > 1 {
			settings.Gateway = m[1]
		}
	}

	return settings, nil
}

// SetIPAddressSettings updates the IP address settings on the HRUI server.
func (c *HRUIClient) SetIPAddressSettings(ctx context.Context, settings *IPAddressSettings) error {
	form := url.Values{}
	form.Set("cmd", "ip")
	form.Set("dhcp_state", "0")
	if settings.DHCPEnabled {
		form.Set("dhcp_state", "1")
	}
	form.Set("ip", settings.IPAddress)
	form.Set("netmask", settings.Netmask)
	form.Set("gateway", settings.Gateway)

	ipSettingsURL := fmt.Sprintf("%s/ip.cgi", c.URL)
	_, err := c.FormRequest(ctx, ipSettingsURL, form)
	if err != nil {
		return fmt.Errorf("failed to update IP Settings: %w", err)
	}

	return nil
}
