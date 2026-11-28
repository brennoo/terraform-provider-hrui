package sdk

import (
	"context"
	"fmt"
	"net/url"
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
	Port   string // Port name associated with the MAC address (e.g., "Port 1" or "Trunk2")
}

// StaticMACEntry represents a single entry in the static MAC address table.
type StaticMACEntry struct {
	ID         int
	MACAddress string
	VLANID     int
	Port       string
}

// Utility function to format port string properly.
func formatPort(rawPort string) string {
	rawPort = strings.TrimSpace(rawPort)
	if _, err := strconv.Atoi(rawPort); err == nil {
		return "Port " + rawPort
	}
	return rawPort
}

// GetMACAddressTable fetches and parses the MAC table from the switch.
func (c *HRUIClient) GetMACAddressTable(ctx context.Context) ([]MACAddressEntry, error) {
	url := c.URL + "/mac.cgi?page=fwd_tbl"

	respBody, err := c.Request(ctx, "GET", url, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("error fetching MAC address table: %w", err)
	}

	// Load the HTML response into goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse MAC table HTML: %w", err)
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
		port := formatPort(strings.TrimSpace(tds.Eq(4).Text()))

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

// GetStaticMACAddressTable retrieves the static MAC address table.
func (c *HRUIClient) GetStaticMACAddressTable(ctx context.Context) ([]StaticMACEntry, error) {
	url := c.URL + "/mac.cgi?page=static"

	respBody, err := c.Request(ctx, "GET", url, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("error fetching static MAC address table: %w", err)
	}

	// Parse the HTML document using goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %w", err)
	}

	// Slice to hold the parsed MAC entries
	var entries []StaticMACEntry

	// Find all rows in the static MAC table
	doc.Find("form[action='/mac.cgi?page=staticdel'] table tr").Each(func(i int, row *goquery.Selection) {
		// Skip the header (first row)
		if i == 0 {
			return
		}

		// Extract content from each column (v1.x has 5 columns; v2.x has 4 columns with checkbox in td[0])
		columns := row.Find("td")
		switch columns.Length() {
		case 5:
			id, _ := strconv.Atoi(strings.TrimSpace(columns.Eq(0).Text()))
			mac := strings.TrimSpace(columns.Eq(1).Text())
			vlan, _ := strconv.Atoi(strings.TrimSpace(columns.Eq(2).Text()))
			port := formatPort(strings.TrimSpace(columns.Eq(3).Text()))
			entries = append(entries, StaticMACEntry{
				ID:         id,
				MACAddress: mac,
				VLANID:     vlan,
				Port:       port,
			})
		case 4:
			delVal := strings.TrimSpace(columns.Eq(0).Find("input[name='del']").AttrOr("value", "0"))
			id, _ := strconv.Atoi(delVal)
			mac := strings.TrimSpace(columns.Eq(1).Text())
			vlan, _ := strconv.Atoi(strings.TrimSpace(columns.Eq(2).Text()))
			port := formatPort(strings.TrimSpace(columns.Eq(3).Text()))
			entries = append(entries, StaticMACEntry{
				ID:         id,
				MACAddress: mac,
				VLANID:     vlan,
				Port:       port,
			})
		}
	})

	return entries, nil
}

// AddStaticMACEntry adds a new static MAC address entry by sending a POST request.
func (c *HRUIClient) AddStaticMACEntry(ctx context.Context, mac string, vlanID int, portName string) error {
	portID, err := c.GetPortByName(ctx, portName)
	if err != nil {
		return fmt.Errorf("failed to resolve port name '%s': %w", portName, err)
	}

	formData := url.Values{
		"mac":  []string{mac},
		"vlan": []string{strconv.Itoa(vlanID)},
		"src":  []string{strconv.Itoa(portID)},
		"cmd":  []string{"macstatic"},
	}

	_, err = c.FormRequest(ctx, c.URL+"/mac.cgi?page=static", formData)
	if err != nil {
		return fmt.Errorf("error adding static MAC address: %w", err)
	}

	return nil
}

// RemoveStaticMACEntries deletes one or more static MAC address entries.
func (c *HRUIClient) RemoveStaticMACEntries(ctx context.Context, macEntries []StaticMACEntry) error {
	// Prepare the form data
	formData := url.Values{
		"cmd": []string{"macstatictbl"},
	}

	for _, entry := range macEntries {
		checkboxValue := entry.MACAddress + "_" + strconv.Itoa(entry.VLANID)
		if respBody, getErr := c.Request(ctx, "GET", c.URL+"/mac.cgi?page=static", nil, nil); getErr == nil {
			if doc, parseErr := goquery.NewDocumentFromReader(strings.NewReader(string(respBody))); parseErr == nil {
				doc.Find("form[action='/mac.cgi?page=staticdel'] table tr").Each(func(i int, row *goquery.Selection) {
					if i == 0 {
						return
					}
					cols := row.Find("td")
					if cols.Length() == 4 {
						mac := strings.TrimSpace(cols.Eq(1).Text())
						vlan, _ := strconv.Atoi(strings.TrimSpace(cols.Eq(2).Text()))
						if strings.EqualFold(mac, entry.MACAddress) && vlan == entry.VLANID {
							if v := strings.TrimSpace(cols.Eq(0).Find("input[name='del']").AttrOr("value", "")); v != "" {
								checkboxValue = v
							}
						}
					}
				})
			}
		}
		formData.Add("del", checkboxValue)
	}

	_, err := c.FormRequest(ctx, c.URL+"/mac.cgi?page=staticdel", formData)
	if err != nil {
		return fmt.Errorf("error deleting static MAC addresses: %w", err)
	}

	return nil
}
