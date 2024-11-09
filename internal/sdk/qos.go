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

// QoSQueueWeight represents the "Queue Weight" for a queue.
type QoSQueueWeight struct {
	Queue  int
	Weight string
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
			Queue:  queueID,
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
	// Prepare form values for the HTTP request
	data := url.Values{}
	data.Set("cmd", "portprio")                      // API command for modifying port priority
	data.Set("portid", strconv.Itoa(portID-1))       // Port ID (0-based)
	data.Set("port_priority", strconv.Itoa(queue-1)) // The new QoS queue value to set (0-based)

	// Prepare the endpoint to send the update request to
	updateURL := client.URL + "/qos.cgi?page=port_pri"

	// Send the POST request to update the QoS Port Queue
	resp, err := client.HttpClient.PostForm(updateURL, data)
	if err != nil {
		return fmt.Errorf("failed to update QoS Port Queue: %w", err)
	}
	defer resp.Body.Close()

	// Check for non-OK responses and return meaningful errors
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusNotFound:
			return fmt.Errorf("QoS update failed: API endpoint not found")
		case http.StatusInternalServerError:
			return fmt.Errorf("QoS update failed: Internal server error")
		default:
			return fmt.Errorf("error: received unexpected status code %d from QoS update", resp.StatusCode)
		}
	}

	// If there is no error, return nil (the operation was successful)
	return nil
}

// GetAllQOSQueueWeights fetches the current queues and weights from the HTML page.
func (client *HRUIClient) GetAllQOSQueueWeights() ([]QoSQueueWeight, error) {
	resp, err := client.HttpClient.Get(client.URL + "/qos.cgi?page=pkt_sch")
	if err != nil {
		return nil, fmt.Errorf("failed to request queue weights page: %w", err)
	}
	defer resp.Body.Close()

	// Parse the HTML document using goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML for queue weights: %w", err)
	}

	var queueWeights []QoSQueueWeight

	// Iterate over table rows to extract queue and weight information
	doc.Find("table").Last().Find("tr").Each(func(i int, row *goquery.Selection) {
		queueText := row.Find("td:first-child").Text()
		weightText := row.Find("td:nth-child(2)").Text()

		if queueText == "" || weightText == "" || queueText == "Queue" {
			return
		}

		queueID, err := strconv.Atoi(queueText)
		if err != nil {
			// If the queueText is not a valid integer, skip it.
			return
		}

		// Add the queue and its corresponding weight
		queueWeights = append(queueWeights, QoSQueueWeight{
			Queue:  queueID,
			Weight: weightText,
		})
	})

	return queueWeights, nil
}

// UpdateQOSQueueWeight updates the weight for a given queue.
func (client *HRUIClient) UpdateQOSQueueWeight(queue, weight int) error {
	data := url.Values{}
	data.Set("cmd", "qweight")                 // Command for setting queue weight
	data.Set("queueid", strconv.Itoa(queue-1)) // Queue (0-based for the backend)
	data.Set("weight", strconv.Itoa(weight))   // Weight (already 0-based as expected)

	updateURL := client.URL + "/qos.cgi?page=que_weight"

	// Sending POST request to update the queue weight
	resp, err := client.HttpClient.PostForm(updateURL, data)
	if err != nil {
		return fmt.Errorf("failed to update QoS Queue Weight: %w", err)
	}
	defer resp.Body.Close()

	// Check for a successful response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d when updating QoS Queue Weight", resp.StatusCode)
	}

	return nil
}
