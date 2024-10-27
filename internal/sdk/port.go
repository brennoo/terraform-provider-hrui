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

type Port struct {
	ID          int
	State       int
	SpeedDuplex string
	FlowControl string
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
