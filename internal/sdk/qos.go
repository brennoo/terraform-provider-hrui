package sdk

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

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

// ListQoSPortQueues fetches and parses QoS port queues from the HTML page
// and resolves their Port IDs using GetPortByName.
func (c *HRUIClient) ListQoSPortQueues(ctx context.Context) ([]QoSPortQueue, error) {
	// Perform the HTTP request to fetch QoS table.
	respBody, err := c.Request(ctx, "GET", c.URL+"/qos.cgi?page=port_pri", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to request QoS Port Queues: %w", err)
	}

	// Parse the HTML document using goquery.
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse QoS Port Queues HTML: %w", err)
	}

	var portQueues []QoSPortQueue

	// Iterate over rows in the last table.
	doc.Find("table").Last().Find("tr").Each(func(i int, row *goquery.Selection) {
		portText := strings.TrimSpace(row.Find("td:first-child").Text())
		queueText := strings.TrimSpace(row.Find("td:nth-child(2)").Text())

		// Skip invalid rows or header rows.
		if portText == "" || queueText == "" || portText == "Port" {
			return
		}

		// Resolve the port ID using the port name through GetPortByName.
		portID, err := c.GetPortByName(ctx, portText)
		if err != nil {
			// Log or handle the invalid port name scenario.
			fmt.Printf("Warning: Unable to resolve port '%s'. Skipping this row.\n", portText)
			return
		}

		// Parse the Queue ID value.
		queueIDOptions := []ParseOption{
			WithDefaultValue(0),     // Default to 0 if no valid value is found.
			WithSpecialCases("N/A"), // Handle "N/A" or other special cases.
		}
		queueID := parseInt(queueText, queueIDOptions...)
		if queueID == nil {
			fmt.Printf("Warning: Invalid queue value '%s' for port '%s'. Skipping this row.\n", queueText, portText)
			return
		}

		// Append the resolved port name and queue ID to the list.
		portQueues = append(portQueues, QoSPortQueue{
			PortID: portID,
			Queue:  *queueID,
		})
	})

	return portQueues, nil
}

// GetQoSPortQueue fetches the QoS port queue by portID (0-based input).
func (c *HRUIClient) GetQoSPortQueue(ctx context.Context, portID int) (*QoSPortQueue, error) {
	portQueues, err := c.ListQoSPortQueues(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve list of QoS port queues: %w", err)
	}

	for _, portQueue := range portQueues {
		if portQueue.PortID == (portID) {
			return &portQueue, nil
		}
	}

	return nil, fmt.Errorf("QoS Port Queue not found for Port ID %d", portID)
}

// SetQoSPortQueue updates the QoS port queue for the given port.
func (c *HRUIClient) SetQoSPortQueue(ctx context.Context, portID, queue int) error {
	// Prepare form values for the HTTP request
	data := url.Values{}
	data.Set("cmd", "portprio")                      // API command for modifying port priority
	data.Set("portid", strconv.Itoa(portID))         // Port ID
	data.Set("port_priority", strconv.Itoa(queue-1)) // The new QoS queue value to set (0-based)

	// Prepare the endpoint to send the update request to
	updateURL := c.URL + "/qos.cgi?page=port_pri"

	// Send the POST request to update the QoS Port Queue
	_, err := c.FormRequest(ctx, updateURL, data)
	if err != nil {
		return fmt.Errorf("failed to update QoS Port Queue: %w", err)
	}

	return nil
}

// ListQoSQueueWeights fetches the current queues and weights from the HTML page.
func (c *HRUIClient) ListQoSQueueWeights(ctx context.Context) ([]QoSQueueWeight, error) {
	// Use Request to fetch the HTML page with QoS queue weights
	respBody, err := c.Request(ctx, "GET", c.URL+"/qos.cgi?page=pkt_sch", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to request queue weights page: %w", err)
	}

	// Parse the HTML document using goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML for queue weights: %w", err)
	}

	var queueWeights []QoSQueueWeight

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

// SetQoSQueueWeight updates the weight for a given queue.
func (c *HRUIClient) SetQoSQueueWeight(ctx context.Context, queue, weight int) error {
	data := url.Values{}
	data.Set("cmd", "qweight")                 // Command for setting queue weight
	data.Set("queueid", strconv.Itoa(queue-1)) // Queue (0-based for the backend)
	data.Set("weight", strconv.Itoa(weight))   // Weight (already 0-based as expected)

	updateURL := c.URL + "/qos.cgi?page=que_weight"

	// Send the POST request to update the queue weight using FormRequest
	_, err := c.FormRequest(ctx, updateURL, data)
	if err != nil {
		return fmt.Errorf("failed to update QoS Queue Weight: %w", err)
	}

	return nil
}
