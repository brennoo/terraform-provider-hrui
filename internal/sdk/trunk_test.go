package sdk

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mocked HTML Response.
const mockHTML = `
<form action='/trunk.cgi?page=group_remove' method='post'>
	<table align='center' class='mid' width='100'>
                <tr>
                    <td>Header1</td>
                    <td>Header2</td>
                    <td>Header3</td>
                </tr>
		<tr bgcolor='#d4d0c8'>
			<td>Trunk1</td>
			<td>static</td>
			<td>3-4</td>
		</tr>
		<tr bgcolor='#d4d0c8'>
			<td>Trunk2</td>
			<td>LACP</td>
			<td>5-6</td>
		</tr>
	</table>
</form>`

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
	trunk, err := client.GetTrunk(1)
	assert.NoError(t, err)

	// Expected trunk
	expected := &TrunkConfig{
		ID:    1,
		Type:  "static",
		Ports: []int{3, 4},
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
	trunks, err := client.ListConfiguredTrunks()
	assert.NoError(t, err)

	// Expected trunks.
	expected := []TrunkConfig{
		{
			ID:    1,
			Type:  "static",
			Ports: []int{3, 4},
		},
		{
			ID:    2,
			Type:  "LACP",
			Ports: []int{5, 6},
		},
	}

	// Assert that the actual parsed data matches the expected data
	assert.Equal(t, expected, trunks)
}
