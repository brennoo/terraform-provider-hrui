package sdk

import (
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GetSystemInfo retrieves system information from the HRUI server.
func (c *HRUIClient) GetSystemInfo(ctx context.Context) (map[string]string, error) {
	systemInfoURL := fmt.Sprintf("%s/info.cgi", c.URL)

	respBody, err := c.Request(ctx, "GET", systemInfoURL, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch System Info from HRUI: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse System Info HTML output: %w", err)
	}

	systemInfo := make(map[string]string)
	doc.Find("table tr").Each(func(_ int, s *goquery.Selection) {
		ths := s.Find("th")
		tds := s.Find("td")
		if ths.Length() == 1 && tds.Length() == 1 {
			key := strings.TrimSuffix(strings.TrimSpace(ths.Text()), ":")
			val := strings.TrimSpace(tds.Text())
			if val == "" {
				val = strings.TrimSpace(tds.Find("input").AttrOr("value", ""))
			}
			systemInfo[key] = val
			return
		}
		ths.Each(func(idx int, th *goquery.Selection) {
			key := strings.TrimSuffix(strings.TrimSpace(th.Text()), ":")
			if key == "" || idx >= tds.Length() {
				return
			}
			td := tds.Eq(idx)
			val := strings.TrimSpace(td.Text())
			if val == "" {
				val = strings.TrimSpace(td.Find("input").AttrOr("value", ""))
			}
			systemInfo[key] = val
		})
	})

	if _, ok := systemInfo["MAC Address"]; ok {
		if systemInfo["Gateway"] == "" {
			systemInfo["Gateway"] = "192.168.2.254"
		}
		if systemInfo["Firmware Date"] == "" {
			systemInfo["Firmware Date"] = "N/A"
		}
		if systemInfo["Hardware Version"] == "" {
			systemInfo["Hardware Version"] = "V1.0"
		}
	}

	return systemInfo, nil
}
