package sdk

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
)

func (c *HRUIClient) GetSystemInfo() (map[string]string, error) {
	systemInfoURL := fmt.Sprintf("%s/info.cgi", c.URL)

	resp, err := c.MakeRequest(systemInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch System Info from HRUI: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
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
