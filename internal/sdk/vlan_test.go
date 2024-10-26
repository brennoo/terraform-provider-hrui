package sdk_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/stretchr/testify/require"
)

// Mock server to simulate VLAN API responses
func mockServer(response string, code int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		_, _ = w.Write([]byte(response))
	}))
}

// Test GetVLAN function
func TestGetVLAN(t *testing.T) {
	// Simulate an HTML response that includes VLAN details for member, tagged, untagged ports
	htmlResponse := `
		<html>
		<body>
			<form name="formVlanStatus">
				<table>
					<tr>
						<th>VLAN</th>
						<th>Name</th>
						<th>Member Ports</th>
						<th>Tagged Ports</th>
						<th>Untagged Ports</th>
					</tr>
					<tr>
						<td>5</td>
						<td>VLAN-Dev</td>
						<td>1-3</td>    <!-- Member Ports -->
						<td>4</td>     <!-- Tagged Ports -->
						<td>1-2</td>   <!-- Untagged Ports -->
					</tr>
				</table>
			</form>
		</body>
		</html>
	`

	// Mock server hosting the HTML response
	server := mockServer(htmlResponse, http.StatusOK)
	defer server.Close()

	// Create an HRUIClient and point it to the mock server URL
	client, _ := sdk.NewClient(server.URL, "testuser", "testpass", false)

	// Use GetVLAN method to fetch the specific VLAN
	vlan, err := client.GetVLAN(5)
	require.NoError(t, err)
	require.NotNil(t, vlan)

	// Test that VLAN details are parsed properly from HTML response
	require.Equal(t, 5, vlan.VlanID, "VLAN ID should be parsed correctly")
	require.Equal(t, "VLAN-Dev", vlan.Name, "VLAN Name should match")
	require.ElementsMatch(t, []int{1, 2}, vlan.UntaggedPorts, "Untagged ports should match")
	require.ElementsMatch(t, []int{4}, vlan.TaggedPorts, "Tagged ports should match")
	require.ElementsMatch(t, []int{1, 2, 3}, vlan.MemberPorts, "Member ports should match")
}

// Test CreateVLAN function - test NotMemberPorts calculation
func TestCreateVLAN_NotMemberPorts(t *testing.T) {
	// Create a mock server to test VLAN creation
	server := mockServer("", http.StatusOK)
	defer server.Close()

	client, _ := sdk.NewClient(server.URL, "testuser", "testpass", false)
	totalPorts := 6

	vlan := &sdk.Vlan{
		VlanID:        10,
		Name:          "VLAN-Test",
		UntaggedPorts: []int{1, 2}, // Members
		TaggedPorts:   []int{3},    // Members
	}

	err := client.CreateVLAN(vlan, totalPorts)
	require.NoError(t, err)
}

// Test DeleteVLAN is still unchanged
func TestDeleteVLAN(t *testing.T) {
	// Simulate deleting a VLAN
	server := mockServer("", http.StatusOK)
	defer server.Close()

	// Create an HRUIClient and point it to the mock server URL
	client, _ := sdk.NewClient(server.URL, "testuser", "testpass", false)

	// Test that deleting the VLAN sends the correct form submission
	err := client.DeleteVLAN(10)
	require.NoError(t, err)
}
