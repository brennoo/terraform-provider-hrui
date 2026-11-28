package sdk

import (
	"context"
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
func (c *HRUIClient) ListAvailableTrunks(ctx context.Context) ([]TrunkConfig, error) {
	endpoint := c.URL + "/trunk.cgi?page=group"
	respBody, err := c.Request(ctx, "GET", endpoint, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch trunk page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML response: %w", err)
	}

	var trunkConfigs []TrunkConfig
	doc.Find("select[name='id'] option").Each(func(i int, s *goquery.Selection) {
		id, _ := strconv.Atoi(s.AttrOr("value", "0"))
		trunkConfigs = append(trunkConfigs, TrunkConfig{
			ID: id,
		})
	})

	return trunkConfigs, nil
}

// ConfigureTrunk sends configuration for a Trunk.
func (c *HRUIClient) ConfigureTrunk(ctx context.Context, config *TrunkConfig) error {
	form := url.Values{}

	form.Set("id", strconv.Itoa(config.ID))
	form.Set("trunk_type", strconv.Itoa(parseTrunkType(config.Type)))

	// Convert ports to 0-indexed before sending them to the API (supports both v1.x 'ports' and v2.x 'port%d')
	for _, port := range config.Ports {
		idx := port - 1
		form.Add("ports", strconv.Itoa(idx))
		form.Set(fmt.Sprintf("port%d", idx), "on")
	}

	form.Set("cmd", "trunk")

	endpoint := c.URL + "/trunk.cgi?page=group"
	_, err := c.FormRequest(ctx, endpoint, form)
	if err != nil {
		return fmt.Errorf("failed to configure trunk: %w", err)
	}

	return nil
}

// Helper function to map trunk type strings to integers.
func parseTrunkType(trunkType string) int {
	switch strings.ToUpper(strings.TrimSpace(trunkType)) {
	case "LACP":
		return 1
	default:
		return 0
	}
}

// DeleteTrunk sends a delete request for a Trunk.
func (c *HRUIClient) DeleteTrunk(ctx context.Context, id int) error {
	removeKey := fmt.Sprintf("remove_%d", id)

	// Inspect the group table to find the exact remove checkbox name (e.g. remove_0 vs remove_1) when available.
	if respBody, err := c.Request(ctx, "GET", c.URL+"/trunk.cgi?page=group", nil, nil); err == nil {
		if doc, parseErr := goquery.NewDocumentFromReader(strings.NewReader(string(respBody))); parseErr == nil {
			doc.Find("form[action='/trunk.cgi?page=group_remove'] table tr").Each(func(i int, s *goquery.Selection) {
				if i == 0 {
					return
				}
				cfg, chkName, ok := parseTrunkTableRow(s)
				if ok && cfg.ID == id && chkName != "" {
					removeKey = chkName
				}
			})
		}
	}

	form := url.Values{}
	form.Set("id", strconv.Itoa(id))
	form.Set(removeKey, "on")
	form.Set("cmd", "group_remove")

	endpoint := c.URL + "/trunk.cgi?page=group_remove"
	_, err := c.FormRequest(ctx, endpoint, form)
	return err
}

// GetTrunk fetches details of a configured Trunk by its ID.
func (c *HRUIClient) GetTrunk(ctx context.Context, id int) (*TrunkConfig, error) {
	trunks, err := c.ListConfiguredTrunks(ctx)
	if err != nil {
		return nil, err
	}

	for i := range trunks {
		if trunks[i].ID == id {
			return &trunks[i], nil
		}
	}

	return nil, errors.New("trunk not found")
}

// ListConfiguredTrunks fetches configured Trunks from the device.
func (c *HRUIClient) ListConfiguredTrunks(ctx context.Context) ([]TrunkConfig, error) {
	endpoint := c.URL + "/trunk.cgi?page=group"
	respBody, err := c.Request(ctx, "GET", endpoint, nil, nil)
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
		if cfg, _, ok := parseTrunkTableRow(s); ok {
			trunkConfigs = append(trunkConfigs, cfg)
		}
	})

	return trunkConfigs, nil
}

func parseTrunkTableRow(s *goquery.Selection) (TrunkConfig, string, bool) {
	tds := s.Find("td")
	if tds.Length() < 3 {
		return TrunkConfig{}, "", false
	}

	// In v1.x: td[0]=TrunkID, td[1]=Type, td[2]=Member Ports, td[last]=<input type="checkbox" name="remove_X">
	// In v2.x: td[0]=<input type="checkbox" name="remove_X">, td[1]=TrunkID, td[2]=Type, td[3]=Member Ports
	colOffset := 0
	if tds.Eq(0).Find("input[type='checkbox']").Length() > 0 && tds.Length() >= 4 {
		colOffset = 1
	}

	idText := strings.TrimSpace(tds.Eq(colOffset).Text())
	if !strings.HasPrefix(idText, "Trunk") {
		return TrunkConfig{}, "", false
	}
	id := parseInt(idText, WithTrimPrefix("Trunk"), WithDefaultValue(0))
	if id == nil {
		return TrunkConfig{}, "", false
	}

	trunkType := strings.TrimSpace(tds.Eq(colOffset + 1).Text())
	portsText := strings.TrimSpace(tds.Eq(colOffset + 2).Text())
	ports := parseTrunkPortsText(portsText)

	chkName := s.Find("input[type='checkbox']").AttrOr("name", "")
	return TrunkConfig{
		ID:    *id,
		Type:  trunkType,
		Ports: ports,
	}, chkName, true
}

func parseTrunkPortsText(portsText string) []int {
	var ports []int
	if portsText == "" || portsText == "-" {
		return ports
	}
	for part := range strings.SplitSeq(portsText, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if startStr, endStr, ok := strings.Cut(part, "-"); ok {
			startPort := parseInt(strings.TrimSpace(startStr), WithDefaultValue(0))
			endPort := parseInt(strings.TrimSpace(endStr), WithDefaultValue(0))
			if startPort != nil && endPort != nil && *startPort > 0 && *startPort <= *endPort {
				for i := *startPort; i <= *endPort; i++ {
					ports = append(ports, i)
				}
			}
			continue
		}
		port := parseInt(part, WithDefaultValue(0))
		if port != nil && *port > 0 {
			ports = append(ports, *port)
		}
	}
	return ports
}
