package sdk

import (
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
	Port   int    // Port number associated with the MAC address
}

// StaticMACEntry represents a single entry in the static MAC address table.
type StaticMACEntry struct {
	ID         int
	MACAddress string
	VLANID     int
	Port       int
}

// GetMACAddressTable fetches and parses the MAC table from the switch.
func (c *HRUIClient) GetMACAddressTable() ([]MACAddressEntry, error) {
	url := c.URL + "/mac.cgi?page=fwd_tbl"

	respBody, err := c.ExecuteRequest("GET", url, nil, nil)
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

// GetStaticMACAddressTable retrieves the static MAC address table.
func (c *HRUIClient) GetStaticMACAddressTable() ([]StaticMACEntry, error) {
	url := c.URL + "/mac.cgi?page=static"

	respBody, err := c.ExecuteRequest("GET", url, nil, nil)
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

		// Extract content from each column
		columns := row.Find("td")
		if columns.Length() == 5 { // Ensure the row has 5 columns
			id, _ := strconv.Atoi(strings.TrimSpace(columns.Eq(0).Text()))
			mac := strings.TrimSpace(columns.Eq(1).Text())
			vlan, _ := strconv.Atoi(strings.TrimSpace(columns.Eq(2).Text()))
			port, _ := strconv.Atoi(strings.TrimSpace(columns.Eq(3).Text()))

			// Append to the entries slice
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

// AddStaticMACAddress adds a new static MAC address entry by sending a POST request.
func (c *HRUIClient) AddStaticMACAddress(mac string, vlanID int, port int) error {
	formData := url.Values{
		"mac":  {mac},
		"vlan": {strconv.Itoa(vlanID)},
		"src":  {strconv.Itoa(port)},
		"cmd":  {"macstatic"},
	}

	_, err := c.ExecuteFormRequest(c.URL+"/mac.cgi?page=static", formData)
	if err != nil {
		return fmt.Errorf("error adding static MAC address: %w", err)
	}

	return nil
}

// DeleteStaticMACAddress deletes one or more static MAC address entries.
func (c *HRUIClient) DeleteStaticMACAddress(macEntries []StaticMACEntry) error {
	// Prepare the form data
	formData := url.Values{
		"cmd": {"macstatictbl"},
	}

	for _, entry := range macEntries {
		checkboxValue := entry.MACAddress + "_" + strconv.Itoa(entry.VLANID)
		formData.Add("del", checkboxValue)
	}

	_, err := c.ExecuteFormRequest(c.URL+"/mac.cgi?page=staticdel", formData)
	if err != nil {
		return fmt.Errorf("error deleting static MAC addresses: %w", err)
	}

	return nil
}
