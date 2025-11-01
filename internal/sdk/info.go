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
	doc.Find("table tr").Each(func(i int, s *goquery.Selection) {
		key := s.Find("th").Text()
		value := s.Find("td").Text()
		systemInfo[key] = value
	})

	return systemInfo, nil
}
