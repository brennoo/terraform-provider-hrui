package sdk

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
