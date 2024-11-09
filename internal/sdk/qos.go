package sdk

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

// QoSPortQueue represents the QoS queue configuration for a port.
type QoSPortQueue struct {
	PortID int
	Queue  int
}

// GetAllQOSPortQueues fetches and parses QoS port queues from the HTML page.
func (client *HRUIClient) GetAllQOSPortQueues() ([]QoSPortQueue, error) {
	response, err := client.HttpClient.Get(client.URL + "/qos.cgi?page=port_pri")
	if err != nil {
		return nil, fmt.Errorf("failed to request QoS Port Queues: %w", err)
	}
	defer response.Body.Close()

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse QoS Port Queues HTML: %w", err)
	}

	var portQueues []QoSPortQueue

	doc.Find("table").Last().Find("tr").Each(func(i int, row *goquery.Selection) {
		portText := row.Find("td:first-child").Text()
		queueText := row.Find("td:nth-child(2)").Text()

		if portText == "" || queueText == "" || portText == "Port" {
			return
		}

		portID, err := parsePortID(portText)
		if err != nil {
			return
		}

		queueID, err := parseQueueID(queueText)
		if err != nil {
			return
		}

		portQueues = append(portQueues, QoSPortQueue{
			PortID: portID,
			Queue:  queueID, // No need to subtract 1 here
		})
	})

	return portQueues, nil
}

// parsePortID extracts the port number from the given text.
func parsePortID(portText string) (int, error) {
	re := regexp.MustCompile(`Port (\d+)`)
	match := re.FindStringSubmatch(portText)
	if len(match) != 2 {
		return 0, fmt.Errorf("invalid port text format: %s", portText)
	}
	return strconv.Atoi(match[1])
}

// parseQueueID extracts the queue ID from the given text.
func parseQueueID(queueText string) (int, error) {
	re := regexp.MustCompile(`\d+`)
	match := re.FindString(queueText)
	if match == "" {
		return 0, fmt.Errorf("invalid queue text format: %s", queueText)
	}
	return strconv.Atoi(match)
}

// GetQOSPortQueue fetches the QoS port queue by portID (0-based input).
func (client *HRUIClient) GetQOSPortQueue(portID int) (*QoSPortQueue, error) {
	portQueues, err := client.GetAllQOSPortQueues()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve list of QoS port queues: %w", err)
	}

	for _, portQueue := range portQueues {
		if portQueue.PortID == (portID) {
			return &portQueue, nil
		}
	}

	return nil, fmt.Errorf("QoS Port Queue not found for Port ID %d", portID+1)
}

// UpdateQOSPortQueue updates the QoS port queue for the given port.
func (client *HRUIClient) UpdateQOSPortQueue(portID, queue int) error {
	data := url.Values{}
	data.Set("cmd", "portprio")                      // API command for modifying port priority
	data.Set("portid", strconv.Itoa(portID-1))       // The port ID (0-based)
	data.Set("port_priority", strconv.Itoa(queue-1)) // The new QoS queue value to set (0-based)

	updateURL := client.URL + "/qos.cgi?page=port_pri"

	var resp *http.Response
	var err error
	for i := 0; i < 3; i++ { // Retry up to 3 times
		resp, err = client.HttpClient.PostForm(updateURL, data)
		if err != nil {
			if err.Error() == "EOF" {
				continue
			}
			return fmt.Errorf("failed to update QoS Port Queue: %w", err)
		}
		break
	}
	if err != nil {
		return fmt.Errorf("failed to update QoS Port Queue (after retries): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("QoS update failed: API endpoint not found")
		} else if resp.StatusCode == http.StatusInternalServerError {
			return fmt.Errorf("QoS update failed: Internal server error")
		}
		return fmt.Errorf("error: received unexpected status code %d from QoS update", resp.StatusCode)
	}

	return nil
}
