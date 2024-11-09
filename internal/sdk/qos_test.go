package sdk

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetAllQOSPortQueues(t *testing.T) {
	// Sample HTML output from the device.
	html := `
<html>
<head> 
  <title>Port-based Priority</title> 
  <link rel="stylesheet" type="text/css" href="/style.css"> 
  <script type="text/javascript"> 
  </script> 
</head> 

<body> 
  <center> 
    <fieldset> 
      <legend>Port to Queue Setting</legend> 
      <form method="post" action="/qos.cgi?page=port_pri"> 
        <table border="1"> 
          <tr> 
            <th width="100">Port</th> 
            <th width="130">Queue</th> 
          </tr> 
          <tr> 
            <td align="center"> 
              <select name="portid" multiple size="6"> 
                <option value="0">Port 1 
                <option value="1">Port 2 
                <option value="2">Port 3 
                <option value="3">Port 4 
                <option value="4">Port 5 
                <option value="5">Port 6 
              </select> 
            </td> 
            <td align="center"> 
              <select name="port_priority"> 
                <option value="0">1 
                <option value="1">2 
                <option value="2">3 
                <option value="3">4 
                <option value="4">5 
                <option value="5">6 
                <option value="6">7 
                <option value="7">8 
              </select> 
            </td> 
          </tr> 
        </table> 
        <br style="line-height:50%"> 
        <input type="submit" value="      Apply      "> 
        <input type="hidden" name="cmd" value="portprio"> 
      </form> 
      <hr> 
      <table border="1"> 
        <tr> 
          <th width="100">Port</th> 
          <th width="130">Queue</th> 
        </tr> 
        <tr> 
          <td align="center">Port 1</td> 
          <td align="center">1</td> 
        </tr> 
        <tr> 
          <td align="center">Port 2</td> 
          <td align="center">8</td> 
        </tr> 
        <tr> 
          <td align="center">Port 3</td> 
          <td align="center">1</td> 
        </tr> 
        <tr> 
          <td align="center">Port 4</td> 
          <td align="center">1</td> 
        </tr> 
        <tr> 
          <td align="center">Port 5</td> 
          <td align="center">1</td> 
        </tr> 
        <tr> 
          <td align="center">Port 6</td> 
          <td align="center">3</td> 
        </tr> 
      </table> 
      <input type="hidden" name="cmd" value="portprios"> 
      <br> 
    </fieldset> 
    <p> 
  </center> 
</body>
</html>
`

	// Create a mock HTTP server to return the sample HTML.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, html)
	}))
	defer server.Close()

	// Create an HRUIClient with the mock server URL.
	client := &HRUIClient{
		URL:        server.URL,
		HttpClient: http.DefaultClient,
	}

	// Call GetAllQOSPortQueues to parse the HTML.
	portQueues, err := client.GetAllQOSPortQueues()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Define the expected parsed output.
	expectedQueues := []QoSPortQueue{
		{PortID: 1, Queue: 1},
		{PortID: 2, Queue: 8},
		{PortID: 3, Queue: 1},
		{PortID: 4, Queue: 1},
		{PortID: 5, Queue: 1},
		{PortID: 6, Queue: 3},
	}

	// Check if the number of parsed queues matches the expected count.
	if len(portQueues) != len(expectedQueues) {
		t.Errorf("Expected %d port queues, got %d", len(expectedQueues), len(portQueues))
	}

	// Compare each parsed queue with the expected values.
	for i, queue := range portQueues {
		if queue.PortID != expectedQueues[i].PortID || queue.Queue != expectedQueues[i].Queue {
			t.Errorf("Mismatch at index %d. Expected: %v, Got: %v", i, expectedQueues[i], queue)
		}
	}
}

// TestGetAllQOSQueueWeights_Success tests successfully parsing queue weights from an HTML page.
func TestGetAllQOSQueueWeights_Success(t *testing.T) {
	// Simulate HTML response
	htmlResponse := `
		<html>
		<body>
		<table>
			<tr>
				<td>Queue</td>
				<td>Weight</td>
			</tr>
			<tr>
				<td>1</td>
				<td>Strict priority</td>
			</tr>
			<tr>
				<td>2</td>
				<td>15</td>
			</tr>
		</table>
		</body>
		</html>
	`

	// Create a test server that returns the above HTML
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(htmlResponse))
	}))
	defer server.Close()

	// Create an HRUIClient pointing to the test server
	client := &HRUIClient{
		HttpClient: server.Client(),
		URL:        server.URL,
	}

	// Call the actual method to test
	queueWeights, err := client.GetAllQOSQueueWeights()
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}

	// Check the returned queue weights
	expectedQueues := []QoSQueueWeight{
		{Queue: 1, Weight: "Strict priority"},
		{Queue: 2, Weight: "15"},
	}

	if len(queueWeights) != len(expectedQueues) {
		t.Fatalf("expected %d queue weights, but got %d", len(expectedQueues), len(queueWeights))
	}

	for i, expected := range expectedQueues {
		if queueWeights[i] != expected {
			t.Errorf("expected queue weight %v, but got %v", expected, queueWeights[i])
		}
	}
}

// TestGetAllQOSQueueWeights_EmptyTable tests the case where the HTML contains no queue data.
func TestGetAllQOSQueueWeights_EmptyTable(t *testing.T) {
	htmlResponse := `
		<html>
		<body>
		<table>
			<tr>
				<td>Queue</td>
				<td>Weight</td>
			</tr>
		</table>
		</body>
		</html>
	`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(htmlResponse))
	}))
	defer server.Close()

	client := &HRUIClient{
		HttpClient: server.Client(),
		URL:        server.URL,
	}

	queueWeights, err := client.GetAllQOSQueueWeights()
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}

	if len(queueWeights) != 0 {
		t.Fatalf("expected no queue weights, but got %d", len(queueWeights))
	}
}

// TestGetAllQOSQueueWeights_MalformedHTML tests handling of malformed HTML.
func TestGetAllQOSQueueWeights_MalformedHTML(t *testing.T) {
	// Simulate malformed HTML
	htmlResponse := `
		<html>
		<body>
		<table>
			some malformed html content here
		</table>
		</body>
		</html>
	`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(htmlResponse))
	}))
	defer server.Close()

	client := &HRUIClient{
		HttpClient: server.Client(),
		URL:        server.URL,
	}

	// Call the method to test; it should not return an error.
	queueWeights, err := client.GetAllQOSQueueWeights()
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}

	// We expect no valid queues from this malformed HTML
	if len(queueWeights) != 0 {
		t.Fatalf("expected no queue weights from malformed HTML, but got %d", len(queueWeights))
	}
}

// TestUpdateQOSQueueWeight_Success tests a successful POST request for updating queue weights.
func TestUpdateQOSQueueWeight_Success(t *testing.T) {
	// Simulate the backend responding to the POST request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			t.Fatalf("expected to parse form data, but got error: %v", err)
		}

		// Check that the correct parameters are sent
		queueID := r.PostFormValue("queueid")
		weight := r.PostFormValue("weight")
		if queueID != "0" || weight != "15" {
			t.Fatalf("expected queueid=0 and weight=15, but got queueid=%s and weight=%s", queueID, weight)
		}

		// Respond with OK to simulate a successful operation
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a client pointing to the test server
	client := &HRUIClient{
		HttpClient: server.Client(),
		URL:        server.URL,
	}

	// Call the actual method to test
	err := client.UpdateQOSQueueWeight(1, 15) // Set weight 15 for queue 1 (which is 0-based for the backend)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
}

// TestUpdateQOSQueueWeight_Failure tests failure scenarios for updating queue weights.
func TestUpdateQOSQueueWeight_Failure(t *testing.T) {
	// Simulate the backend returning 500 Internal Server Error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create a client pointing to the test server
	client := &HRUIClient{
		HttpClient: server.Client(),
		URL:        server.URL,
	}

	// Call the method to test
	err := client.UpdateQOSQueueWeight(1, 10)
	if err == nil || !strings.Contains(err.Error(), "unexpected status code 500") {
		t.Fatalf("expected an error with status code 500, but got %v", err)
	}
}
