package sdk

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mocked HTML Response.
const mockHTML = `
<form method="post" action="/trunk.cgi?page=group_remove">
<table>
	<tr>
		<th style="width:80px;">Group ID</th>
		<th style="width:80px;">Type</th>
		<th style="width:150px;">Member port</th>
		<th style="width:150px;">Aggregated Port</th>
		<th style="width:80px;">Select</th>
	</tr>
	<tr>
		<td >Trunk2</td>
		<td >static</td>
		<td>2,6</td>
		<td>2,6</td>
		<td><input type="checkbox" name="remove_1" id=trunk_0></td>
	</tr>
  </table>
<br style="line-height:50%">
<input type="submit" name="Remove" value="    Delete    ">
<input type="button" value=" Select All " onclick="trunk_selAll(1)">
<input type="hidden" name="cmd" value="group_remove">
</form>
`

func TestGetTrunk(t *testing.T) {
	// Mock the HTTP server response with the above HTML
	mockServer := mockServerMock(mockHTML, http.StatusOK)
	defer mockServer.Close()

	// Create a client instance
	client := &HRUIClient{
		URL:        mockServer.URL,
		HttpClient: &http.Client{},
	}

	// Call GetTrunk
	trunk, err := client.GetTrunk(context.Background(), 2)
	assert.NoError(t, err)

	// Expected trunk
	expected := &TrunkConfig{
		ID:    2,
		Type:  "static",
		Ports: []int{2, 6},
	}

	// Assert that the actual parsed data matches the expected data
	assert.Equal(t, expected, trunk)
}

func TestListConfiguredTrunks(t *testing.T) {
	// Mock the HTTP server response with the above HTML
	mockServer := mockServerMock(mockHTML, http.StatusOK)
	defer mockServer.Close()

	// Create a client instance
	client := &HRUIClient{
		URL:        mockServer.URL,
		HttpClient: &http.Client{},
	}

	// Call ListConfiguredTrunks
	trunks, err := client.ListConfiguredTrunks(context.Background())
	assert.NoError(t, err)

	// Expected trunks.
	expected := []TrunkConfig{
		{
			ID:    2,
			Type:  "static",
			Ports: []int{2, 6},
		},
	}

	// Assert that the actual parsed data matches the expected data
	assert.Equal(t, expected, trunks)
}
