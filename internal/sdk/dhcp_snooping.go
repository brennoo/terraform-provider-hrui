package sdk

import (
	"bytes"
	"cmp"
	"context"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// DHCPSnoopingConfig represents the global DHCP snooping state and per-port trust configuration.
type DHCPSnoopingConfig struct {
	Enabled      bool
	TrustedPorts []string
}

type dhcpSnoopingFormInfo struct {
	globalInputName string
	globalEnabled   bool
	isV2Checkboxes  bool
	radioGroupNames []string // ordered by 1-based port index for v1.x (e.g. vlanPort_0, vlanPort_1, ...)
	checkboxNames   []string // ordered by 1-based port index for v2.x (e.g. port1, port2, ...)
	trustedPorts    []string
}

func (c *HRUIClient) fetchDHCPSnoopingForm(ctx context.Context) (*dhcpSnoopingFormInfo, error) {
	endpoint := fmt.Sprintf("%s/dhcp_snooping.cgi?page=dump", c.URL)
	body, err := c.Request(ctx, "GET", endpoint, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch DHCP snooping page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse DHCP snooping HTML: %w", err)
	}

	info := &dhcpSnoopingFormInfo{
		globalInputName: "enable_dhcpsnooping",
		trustedPorts:    []string{},
	}

	globalChk := doc.Find("form[action='/dhcp_snooping.cgi?page=enable_dhcpsnooping'] input[type='checkbox']").First()
	if globalChk.Length() == 0 {
		globalChk = doc.Find("input[name='enable_dhcpsnooping'], input[name='enable_dhcpsnp']").First()
	}
	if globalChk.Length() > 0 {
		if name, ok := globalChk.Attr("name"); ok && strings.TrimSpace(name) != "" {
			info.globalInputName = strings.TrimSpace(name)
		}
		_, info.globalEnabled = globalChk.Attr("checked")
	}

	staticForm := doc.Find("form[action='/dhcp_snooping.cgi?page=static']")
	v2Chks := staticForm.Find("input[type='checkbox'][name^='port']")
	if v2Chks.Length() > 0 {
		info.isV2Checkboxes = true
		v2Chks.Each(func(i int, s *goquery.Selection) {
			name := strings.TrimSpace(s.AttrOr("name", ""))
			if name == "" {
				return
			}
			info.checkboxNames = append(info.checkboxNames, name)
			portNum := len(info.checkboxNames)
			portLabel := fmt.Sprintf("Port %d", portNum)
			if _, checked := s.Attr("checked"); checked {
				info.trustedPorts = append(info.trustedPorts, portLabel)
			}
		})
		return info, nil
	}

	// v1.x layout uses radio buttons <input type="radio" name="vlanPort_%d" value="1|0">
	seenGroups := make(map[string]int)
	staticForm.Find("input[type='radio'][name^='vlanPort_']").Each(func(_ int, s *goquery.Selection) {
		name := strings.TrimSpace(s.AttrOr("name", ""))
		if name == "" {
			return
		}
		portIdx, exists := seenGroups[name]
		if !exists {
			info.radioGroupNames = append(info.radioGroupNames, name)
			portIdx = len(info.radioGroupNames)
			seenGroups[name] = portIdx
		}
		val := strings.TrimSpace(s.AttrOr("value", ""))
		_, checked := s.Attr("checked")
		if checked && val == "1" {
			portLabel := fmt.Sprintf("Port %d", portIdx)
			if !slices.Contains(info.trustedPorts, portLabel) {
				info.trustedPorts = append(info.trustedPorts, portLabel)
			}
		}
	})

	return info, nil
}

// GetDHCPSnooping retrieves the current DHCP snooping global state and trusted ports from /dhcp_snooping.cgi?page=dump.
func (c *HRUIClient) GetDHCPSnooping(ctx context.Context) (*DHCPSnoopingConfig, error) {
	info, err := c.fetchDHCPSnoopingForm(ctx)
	if err != nil {
		return nil, err
	}
	return &DHCPSnoopingConfig{
		Enabled:      info.globalEnabled,
		TrustedPorts: info.trustedPorts,
	}, nil
}

// SetDHCPSnoopingGlobal enables or disables global DHCP snooping via POST /dhcp_snooping.cgi?page=enable_dhcpsnooping.
func (c *HRUIClient) SetDHCPSnoopingGlobal(ctx context.Context, enabled bool) error {
	info, err := c.fetchDHCPSnoopingForm(ctx)
	if err != nil {
		return err
	}

	form := url.Values{}
	if enabled {
		form.Set(cmp.Or(info.globalInputName, "enable_dhcpsnooping"), "on")
	}

	endpoint := fmt.Sprintf("%s/dhcp_snooping.cgi?page=enable_dhcpsnooping", c.URL)
	if _, err := c.FormRequest(ctx, endpoint, form); err != nil {
		return fmt.Errorf("failed to set DHCP snooping global state: %w", err)
	}
	return nil
}

// SetDHCPSnoopingPorts updates the trusted ports list via POST /dhcp_snooping.cgi?page=static.
func (c *HRUIClient) SetDHCPSnoopingPorts(ctx context.Context, trustedPorts []string) error {
	info, err := c.fetchDHCPSnoopingForm(ctx)
	if err != nil {
		return err
	}

	trustedSet := make(map[int]bool, len(trustedPorts))
	for _, p := range trustedPorts {
		idx0, convErr := resolveZeroBasedPortIndex(p)
		if convErr != nil {
			return convErr
		}
		trustedSet[idx0+1] = true
	}

	form := url.Values{}
	if info.isV2Checkboxes {
		form.Set("cmd", "dhcpsnpfrm")
		for i, chkName := range info.checkboxNames {
			portNum := i + 1
			if trustedSet[portNum] {
				form.Set(chkName, "on")
			}
		}
	} else {
		for i, radioName := range info.radioGroupNames {
			portNum := i + 1
			if trustedSet[portNum] {
				form.Set(radioName, "1")
			} else {
				form.Set(radioName, "0")
			}
		}
	}

	endpoint := fmt.Sprintf("%s/dhcp_snooping.cgi?page=static", c.URL)
	if _, err := c.FormRequest(ctx, endpoint, form); err != nil {
		return fmt.Errorf("failed to configure DHCP snooping trusted ports: %w", err)
	}
	return nil
}

// ConfigureDHCPSnooping updates both the global DHCP snooping state and the trusted ports list.
// On v2.x firmwares, the per-port trust form is only rendered when global DHCP snooping is enabled,
// so when enabling we enable global state first, and when disabling we clear/update ports before disabling.
func (c *HRUIClient) ConfigureDHCPSnooping(ctx context.Context, cfg DHCPSnoopingConfig) error {
	if cfg.Enabled {
		if err := c.SetDHCPSnoopingGlobal(ctx, true); err != nil {
			return err
		}
		return c.SetDHCPSnoopingPorts(ctx, cfg.TrustedPorts)
	}
	if err := c.SetDHCPSnoopingPorts(ctx, cfg.TrustedPorts); err != nil {
		return err
	}
	return c.SetDHCPSnoopingGlobal(ctx, false)
}
