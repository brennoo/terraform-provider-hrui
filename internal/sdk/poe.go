package sdk

import (
	"bytes"
	"cmp"
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// PoESystem represents the global PoE system settings and telemetry.
type PoESystem struct {
	PowerLimitWatts       int
	PowerConsumptionWatts float64
	PowerUsagePercent     int
}

// PoEPortConfig represents the administrative PoE configuration for a port on /pse_system.cgi.
type PoEPortConfig struct {
	Port          string
	Enabled       bool
	PSEType       string
	WatchdogState bool
}

// PoEPortStatus represents the full configuration and live telemetry for a PoE port from /pse_port.cgi.
type PoEPortStatus struct {
	Port          string
	Enabled       bool
	PSEType       string
	WatchdogState bool
	Operation     string
	Class         string
	PowerWatts    float64
	VoltageVolts  float64
	CurrentMA     int
}

var pseTypeToValue = map[string]string{
	"BT":    "0",
	"HiPoE": "1",
	"AT":    "2",
	"AF":    "3",
}

// GetPoESystem fetches the global PoE power budget from /pse_system.cgi and live consumption from /pse_port.cgi.
func (c *HRUIClient) GetPoESystem(ctx context.Context) (*PoESystem, error) {
	sysURL := fmt.Sprintf("%s/pse_system.cgi", c.URL)
	sysBody, err := c.Request(ctx, "GET", sysURL, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PoE system page: %w", err)
	}

	sysDoc, err := goquery.NewDocumentFromReader(bytes.NewReader(sysBody))
	if err != nil {
		return nil, fmt.Errorf("failed to parse PoE system HTML: %w", err)
	}

	pwrVal, exists := sysDoc.Find("input[name='pse_sys_pwr']").Attr("value")
	if !exists {
		return nil, fmt.Errorf("pse_sys_pwr input not found on /pse_system.cgi")
	}
	limitWatts, err := strconv.Atoi(strings.TrimSpace(pwrVal))
	if err != nil {
		return nil, fmt.Errorf("invalid pse_sys_pwr value %q: %w", pwrVal, err)
	}

	status := &PoESystem{
		PowerLimitWatts: limitWatts,
	}

	portURL := fmt.Sprintf("%s/pse_port.cgi", c.URL)
	portBody, err := c.Request(ctx, "GET", portURL, nil, nil)
	if err == nil {
		if portDoc, parseErr := goquery.NewDocumentFromReader(bytes.NewReader(portBody)); parseErr == nil {
			if conVal, ok := portDoc.Find("input[name='pse_con_pwr']").Attr("value"); ok {
				if v, convErr := strconv.ParseFloat(strings.TrimSpace(conVal), 64); convErr == nil {
					status.PowerConsumptionWatts = v
				}
			}
			if pctVal, ok := portDoc.Find("input[name='pse_con_pwr_percent']").Attr("value"); ok {
				if v, convErr := strconv.Atoi(strings.TrimSpace(pctVal)); convErr == nil {
					status.PowerUsagePercent = v
				}
			}
		}
	}

	return status, nil
}

// SetPoESystemPowerLimit updates the total PoE power budget (in Watts) on /pse_system.cgi.
func (c *HRUIClient) SetPoESystemPowerLimit(ctx context.Context, powerLimitWatts int) error {
	if powerLimitWatts <= 0 {
		return fmt.Errorf("powerLimitWatts must be positive, got %d", powerLimitWatts)
	}

	form := url.Values{}
	form.Set("pse_sys_pwr", strconv.Itoa(powerLimitWatts))
	form.Set("state", "1")
	form.Set("pse_type", "2")
	form.Set("watchdog_state", "0")
	form.Set("cmd", "poe")
	form.Set("submit", "Apply")

	endpoint := fmt.Sprintf("%s/pse_system.cgi", c.URL)
	if _, err := c.FormRequest(ctx, endpoint, form); err != nil {
		return fmt.Errorf("failed to set PoE system power limit: %w", err)
	}
	return nil
}

// ListPoEPortConfigs parses the per-port PoE administrative configurations from /pse_system.cgi.
func (c *HRUIClient) ListPoEPortConfigs(ctx context.Context) ([]PoEPortConfig, error) {
	endpoint := fmt.Sprintf("%s/pse_system.cgi", c.URL)
	body, err := c.Request(ctx, "GET", endpoint, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PoE system page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse PoE system HTML: %w", err)
	}

	var configs []PoEPortConfig
	doc.Find("table").Last().Find("tr").Each(func(i int, row *goquery.Selection) {
		if i == 0 {
			return
		}
		cols := row.Find("td")
		if cols.Length() < 4 {
			return
		}
		portName := strings.TrimSpace(cols.Eq(0).Text())
		if portName == "" {
			return
		}
		adminStr := strings.TrimSpace(cols.Eq(1).Text())
		pseType := strings.TrimSpace(cols.Eq(2).Text())
		dogStr := strings.TrimSpace(cols.Eq(3).Text())

		configs = append(configs, PoEPortConfig{
			Port:          portName,
			Enabled:       strings.EqualFold(adminStr, "Enable") || adminStr == "1",
			PSEType:       pseType,
			WatchdogState: strings.EqualFold(dogStr, "Enable") || dogStr == "1",
		})
	})

	return configs, nil
}

// GetPoEPortConfig fetches the PoE configuration for a single port from /pse_system.cgi.
func (c *HRUIClient) GetPoEPortConfig(ctx context.Context, port string) (*PoEPortConfig, error) {
	configs, err := c.ListPoEPortConfigs(ctx)
	if err != nil {
		return nil, err
	}
	target := strings.TrimSpace(port)
	for i := range configs {
		if strings.EqualFold(configs[i].Port, target) {
			return &configs[i], nil
		}
	}
	return nil, fmt.Errorf("PoE port %q not found", port)
}

// ConfigurePoEPort updates the PoE configuration for a port via /pse_system.cgi.
func (c *HRUIClient) ConfigurePoEPort(ctx context.Context, cfg PoEPortConfig) error {
	endpoint := fmt.Sprintf("%s/pse_system.cgi", c.URL)
	body, err := c.Request(ctx, "GET", endpoint, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch PoE system page before update: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to parse PoE system HTML: %w", err)
	}

	sysPwr := strings.TrimSpace(doc.Find("input[name='pse_sys_pwr']").AttrOr("value", "100"))
	sysPwr = cmp.Or(sysPwr, "100")

	portValue := ""
	targetPort := strings.TrimSpace(cfg.Port)
	doc.Find("select[name='portid'] option").EachWithBreak(func(_ int, opt *goquery.Selection) bool {
		if strings.EqualFold(strings.TrimSpace(opt.Text()), targetPort) {
			portValue = strings.TrimSpace(opt.AttrOr("value", ""))
			return false
		}
		return true
	})

	if portValue == "" {
		portNum, convErr := resolveZeroBasedPortIndex(targetPort)
		if convErr != nil {
			return fmt.Errorf("PoE port %q not found on /pse_system.cgi: %w", cfg.Port, convErr)
		}
		portValue = strconv.Itoa(portNum)
	}

	pseTypeName := cmp.Or(strings.TrimSpace(cfg.PSEType), "AT")
	pseTypeVal, ok := pseTypeToValue[pseTypeName]
	if !ok {
		return fmt.Errorf("invalid PoE PSE type %q (expected BT, HiPoE, AT, or AF)", cfg.PSEType)
	}

	stateVal := "0"
	if cfg.Enabled {
		stateVal = "1"
	}
	dogVal := "0"
	if cfg.WatchdogState {
		dogVal = "1"
	}

	form := url.Values{}
	form.Set("pse_sys_pwr", sysPwr)
	form.Set("portid", portValue)
	form.Set("state", stateVal)
	form.Set("pse_type", pseTypeVal)
	form.Set("watchdog_state", dogVal)
	form.Set("cmd", "poe")
	form.Set("submit", "Apply")

	if _, err := c.FormRequest(ctx, endpoint, form); err != nil {
		return fmt.Errorf("failed to configure PoE port %q: %w", cfg.Port, err)
	}
	return nil
}

// ListPoEPortStatuses parses the per-port PoE configuration and live telemetry from /pse_port.cgi.
func (c *HRUIClient) ListPoEPortStatuses(ctx context.Context) ([]PoEPortStatus, error) {
	endpoint := fmt.Sprintf("%s/pse_port.cgi", c.URL)
	body, err := c.Request(ctx, "GET", endpoint, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PoE port status page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse PoE port status HTML: %w", err)
	}

	var statuses []PoEPortStatus
	doc.Find("table").Last().Find("tr").Each(func(i int, row *goquery.Selection) {
		if i == 0 {
			return
		}
		cols := row.Find("td")
		if cols.Length() < 9 {
			return
		}
		portName := strings.TrimSpace(cols.Eq(0).Text())
		if portName == "" {
			return
		}
		adminStr := strings.TrimSpace(cols.Eq(1).Text())
		pseType := strings.TrimSpace(cols.Eq(2).Text())
		dogStr := strings.TrimSpace(cols.Eq(3).Text())
		opStr := strings.TrimSpace(cols.Eq(4).Text())
		classStr := strings.TrimSpace(cols.Eq(5).Text())
		pwrStr := strings.TrimSpace(cols.Eq(6).Text())
		voltStr := strings.TrimSpace(cols.Eq(7).Text())
		currStr := strings.TrimSpace(cols.Eq(8).Text())

		var pwrWatts, voltVolts float64
		var currMA int
		if pwrStr != "" && pwrStr != "-" {
			pwrWatts, _ = strconv.ParseFloat(pwrStr, 64)
		}
		if voltStr != "" && voltStr != "-" {
			voltVolts, _ = strconv.ParseFloat(voltStr, 64)
		}
		if currStr != "" && currStr != "-" {
			currMA, _ = strconv.Atoi(currStr)
		}

		statuses = append(statuses, PoEPortStatus{
			Port:          portName,
			Enabled:       strings.EqualFold(adminStr, "Enable") || adminStr == "1",
			PSEType:       pseType,
			WatchdogState: strings.EqualFold(dogStr, "Enable") || dogStr == "1",
			Operation:     opStr,
			Class:         classStr,
			PowerWatts:    pwrWatts,
			VoltageVolts:  voltVolts,
			CurrentMA:     currMA,
		})
	})

	return statuses, nil
}

// GetPoEPortStatus fetches the live PoE status and configuration for a single port from /pse_port.cgi.
func (c *HRUIClient) GetPoEPortStatus(ctx context.Context, port string) (*PoEPortStatus, error) {
	statuses, err := c.ListPoEPortStatuses(ctx)
	if err != nil {
		return nil, err
	}
	target := strings.TrimSpace(port)
	for i := range statuses {
		if strings.EqualFold(statuses[i].Port, target) {
			return &statuses[i], nil
		}
	}
	return nil, fmt.Errorf("PoE port status for %q not found", port)
}

// ResetPoEPort triggers a power cycle on a PoE port via POST /pse_port.cgi.
func (c *HRUIClient) ResetPoEPort(ctx context.Context, port string) error {
	idx, err := resolveZeroBasedPortIndex(port)
	if err != nil {
		return err
	}
	form := url.Values{}
	form.Set("port_id", strconv.Itoa(idx))

	endpoint := fmt.Sprintf("%s/pse_port.cgi", c.URL)
	if _, err := c.FormRequest(ctx, endpoint, form); err != nil {
		return fmt.Errorf("failed to reset PoE port %q: %w", port, err)
	}
	return nil
}

func resolveZeroBasedPortIndex(port string) (int, error) {
	trimmed := strings.TrimSpace(port)
	if _, after, ok := strings.Cut(strings.ToLower(trimmed), "port"); ok {
		trimmed = strings.TrimSpace(after)
	}
	n, err := strconv.Atoi(trimmed)
	if err != nil || n < 1 {
		return 0, fmt.Errorf("invalid port identifier %q", port)
	}
	return n - 1, nil
}
