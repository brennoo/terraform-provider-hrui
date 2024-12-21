package sdk

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var speedDuplexMapping = map[int]string{
	0: "Auto",
	1: "10M/Half",
	2: "10M/Full",
	3: "100M/Half",
	4: "100M/Full",
	5: "1000M/Full",
	6: "2500M/Full",
	8: "10G/Full",
}

var flowControlMapping = map[int]string{
	0: "Off",
	1: "On",
}

var stateMapping = map[string]int{
	"Enable":  1,
	"Disable": 0,
}

type Port struct {
	ID          int
	State       int
	SpeedDuplex string
	FlowControl string
}

type PortStatistics struct {
	Port       int
	State      int
	LinkStatus string
	TxGoodPkt  int64
	TxBadPkt   int64
	RxGoodPkt  int64
	RxBadPkt   int64
}

func (c *HRUIClient) GetPort(portID int) (*Port, error) {
	ports, err := c.GetAllPorts()
	if err != nil {
		log.Print("failed to fetch Ports:", err)
		return nil, err
	}

	// Search for the Port with the specified ID in the list of Ports retrieved
	for _, port := range ports {
		if port.ID == portID {
			return port, nil
		}
	}
	return nil, fmt.Errorf("port with ID %d not found", portID)
}

func (c *HRUIClient) GetAllPorts() ([]*Port, error) {
	portURL := fmt.Sprintf("%s/port.cgi?page=static", c.URL)

	resp, err := c.MakeRequest(portURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Port settings from HRUI: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML output: %w", err)
	}

	var ports []*Port
	doc.Find("body center fieldset table").Each(func(i int, s *goquery.Selection) {
		if i != 2 { // 3rd table
			return
		}
		s.Find("tr").Each(func(j int, tr *goquery.Selection) {
			if j < 2 {
				return
			}
			port := &Port{}
			tr.Find("td").Each(func(k int, td *goquery.Selection) {
				text := strings.TrimSpace(td.Text())
				switch k {
				case 0:
					text = strings.TrimPrefix(text, "Port ")
					fmt.Sscanf(text, "%d", &port.ID)
				case 1:
					if text == "Enable" {
						port.State = 1
					} else {
						port.State = 0
					}
				case 2:
					port.SpeedDuplex = text
				case 4:
					port.FlowControl = text
				}
			})

			ports = append(ports, port)
		})
	})
	return ports, nil
}

func (c *HRUIClient) UpdatePortSettings(port *Port) error {
	form := url.Values{}
	form.Set("cmd", "port")
	form.Set("portid", strconv.Itoa(port.ID))
	form.Set("state", strconv.Itoa(port.State))
	form.Set("speed_duplex", port.SpeedDuplex)
	form.Set("flow", port.FlowControl)

	// submit the form
	portsURL := fmt.Sprintf("%s/port.cgi", c.URL)
	_, err := c.PostForm(portsURL, form)
	if err != nil {
		return fmt.Errorf("failed to update port settings: %w", err)
	}

	if c.Autosave {
		return c.SaveConfiguration()
	}

	return nil
}

func (c *HRUIClient) GetTotalPorts() (int, error) {
	ports, err := c.GetAllPorts()
	if err != nil {
		return 0, fmt.Errorf("failed to get total ports: %w", err)
	}
	return len(ports), nil
}

func (c *HRUIClient) GetPortStatistics() ([]*PortStatistics, error) {
	statsURL := fmt.Sprintf("%s/port.cgi?page=stats", c.URL)

	// Send GET request to the port statistics page
	resp, err := c.HttpClient.Get(statsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch port statistics page: %v", err)
	}
	defer resp.Body.Close()

	// Load the HTML into goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML document: %v", err)
	}

	var stats []*PortStatistics

	// Select rows in the table, skipping the header rows
	doc.Find("table tr").Each(func(i int, sel *goquery.Selection) {
		// Skip header rows
		if i == 0 {
			return
		}

		// Extract columns (td elements)
		columns := sel.Find("td")
		if columns.Length() < 7 {
			// Not enough columns for the statistics table row
			return
		}

		// Extract each field from the table row
		portText := strings.TrimSpace(columns.Eq(0).Text())
		portNumber, err := strconv.Atoi(strings.TrimPrefix(portText, "Port "))
		if err != nil {
			log.Printf("Skipping invalid port number '%s': %v", portText, err)
			return
		}

		stateStr := strings.TrimSpace(columns.Eq(1).Text())
		state, ok := stateMapping[stateStr]
		if !ok {
			log.Printf("Unknown state '%s' for Port %d, defaulting to 'Disable'", stateStr, portNumber)
			state = 0 // Default to "Disable" if state is not recognized
		}

		linkStatus := strings.TrimSpace(columns.Eq(2).Text())

		txGoodPkt, err := strconv.ParseInt(strings.TrimSpace(columns.Eq(3).Text()), 10, 64)
		if err != nil {
			txGoodPkt = 0
			log.Printf("Error parsing TxGoodPkt for Port %d: %v", portNumber, err)
		}

		txBadPkt, err := strconv.ParseInt(strings.TrimSpace(columns.Eq(4).Text()), 10, 64)
		if err != nil {
			txBadPkt = 0
			log.Printf("Error parsing TxBadPkt for Port %d: %v", portNumber, err)
		}

		rxGoodPkt, err := strconv.ParseInt(strings.TrimSpace(columns.Eq(5).Text()), 10, 64)
		if err != nil {
			rxGoodPkt = 0
			log.Printf("Error parsing RxGoodPkt for Port %d: %v", portNumber, err)
		}

		rxBadPkt, err := strconv.ParseInt(strings.TrimSpace(columns.Eq(6).Text()), 10, 64)
		if err != nil {
			rxBadPkt = 0
			log.Printf("Error parsing RxBadPkt for Port %d: %v", portNumber, err)
		}

		// Append the parsed data as a PortStatistics entry
		stats = append(stats, &PortStatistics{
			Port:       portNumber,
			State:      state,
			LinkStatus: linkStatus,
			TxGoodPkt:  txGoodPkt,
			TxBadPkt:   txBadPkt,
			RxGoodPkt:  rxGoodPkt,
			RxBadPkt:   rxBadPkt,
		})
	})

	// Return the list of port statistics
	return stats, nil
}
