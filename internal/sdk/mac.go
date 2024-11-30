package sdk

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// MACAddressEntry represents a single entry in the MAC address table.
type MACAddressEntry struct {
	ID     int    // Sequence number of the entry
	MAC    string // MAC address in the format xx:xx:xx:xx:xx:xx
	VLANID int    // VLAN ID associated with the MAC address
	Type   string // Type of the entry (e.g., "dynamic" or "static")
	Port   int    // Port number associated with the MAC address
}

// GetMACAddressTable fetches and parses the MAC table from the switch.
func (c *HRUIClient) GetMACAddressTable() ([]MACAddressEntry, error) {
	// Send an HTTP GET request to fetch the MAC table page.
	resp, err := c.HttpClient.Get(c.URL + "/mac.cgi?page=fwd_tbl")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch MAC address table: status " + strconv.Itoa(resp.StatusCode))
	}

	// Load the HTML response into goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, errors.New("failed to parse MAC table HTML: " + err.Error())
	}

	// Extract MAC address entries from the table
	var entries []MACAddressEntry
	doc.Find("table tr").Each(func(i int, row *goquery.Selection) {
		// Skip the header row
		if i == 0 {
			return
		}

		// Extract <td> elements
		tds := row.Find("td")
		if tds.Length() < 5 {
			return
		}

		// Parse the data
		id, _ := strconv.Atoi(strings.TrimSpace(tds.Eq(0).Text()))
		mac := strings.TrimSpace(tds.Eq(1).Text())
		vlanID, _ := strconv.Atoi(strings.TrimSpace(tds.Eq(2).Text()))
		entryType := strings.ToLower(strings.TrimSpace(tds.Eq(3).Text()))
		port, _ := strconv.Atoi(strings.TrimSpace(tds.Eq(4).Text()))

		// Append the entry
		entries = append(entries, MACAddressEntry{
			ID:     id,
			MAC:    mac,
			VLANID: vlanID,
			Type:   entryType,
			Port:   port,
		})
	})

	return entries, nil
}
