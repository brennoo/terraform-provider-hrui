package sdk

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test GetVLAN function.
func TestGetVLAN(t *testing.T) {
	htmlResponse := `
        <html>
        <body>
        <center>
        <fieldset>
        <legend>802.1Q VLAN</legend>
        <form method="post" action="/vlan.cgi?page=getRmvVlanEntry" name=formVlanStatus>
         <table border="1">
          <tr>
           <th nowrap width="60">VLAN</th>
           <th nowrap>VLAN Name</th>
           <th nowrap>Member Ports</th>
           <th nowrap>Tagged Ports</th>
           <th nowrap>Untagged Ports</th>
           <th nowrap width="50">Delete</th>
          </tr>
          <tr>
           <td width="60"><a method="post" href="/vlan.cgi?page=getVlanEntry&pickVlanId=1">1</a></td>
           <td></td>
           <td nowrap>1-6</td>
           <td nowrap>-</td>
           <td nowrap>1-6</td>
           <td width="50"><input type="checkbox" name="remove_1" id=vlan_0 disabled></td>
          </tr>
          <tr>
           <td width="60"><a method="post" href="/vlan.cgi?page=getVlanEntry&pickVlanId=4">4</a></td>
           <td>sechs</td>
           <td nowrap>6</td>
           <td nowrap>-</td>
           <td nowrap>6</td>
           <td width="50"><input type="checkbox" name="remove_4" id=vlan_1></td>
          </tr>
          <tr>
           <td width="60"><a method="post" href="/vlan.cgi?page=getVlanEntry&pickVlanId=5">5</a></td>
           <td>sechs</td>
           <td nowrap>5</td>
           <td nowrap>-</td>
           <td nowrap>5</td>
           <td width="50"><input type="checkbox" name="remove_5" id=vlan_2></td>
          </tr>
          <tr>
           <td width="60"><a method="post" href="/vlan.cgi?page=getVlanEntry&pickVlanId=10">10</a></td>
           <td>myvlan1</td>
           <td nowrap>1,3-5</td>
           <td nowrap>1,4-5</td>
           <td nowrap>3</td>
           <td width="50"><input type="checkbox" name="remove_10" id=vlan_3></td>
          </tr>
         </table>
         <br style="line-height:50%">
         <input type="submit" name="Delete" value="    Delete    ">
         <input type="button" value=" Select All " onclick="vlanstatic_selAll(4)">
        </form>
        </fieldset>
        </center>
        </body>
        </html>
`

	server := mockServerMock(htmlResponse, http.StatusOK)
	defer server.Close()

	client, err := NewClient(server.URL, "testuser", "testpass", false)
	require.NoError(t, err)
	require.NotNil(t, client, "Client should not be nil")

	// Test fetching an existing VLAN
	vlan, err := client.GetVLAN(10)
	assert.NoError(t, err)
	assert.Equal(t, 10, vlan.VlanID)
	assert.Equal(t, "myvlan1", vlan.Name)
	assert.Equal(t, []int{1, 3, 4, 5}, vlan.MemberPorts)
	assert.Equal(t, []int{1, 4, 5}, vlan.TaggedPorts)
	assert.Equal(t, []int{3}, vlan.UntaggedPorts)

	// Test fetching a non-existent VLAN
	vlan, err = client.GetVLAN(20)
	assert.Error(t, err)
	assert.Nil(t, vlan)
}

// Test CreateVLAN function - test NotMemberPorts calculation.
func TestCreateVLAN_NotMemberPorts(t *testing.T) {
	// Create a mock server to test VLAN creation
	server := mockServerMock("", http.StatusOK)
	defer server.Close()

	client, _ := NewClient(server.URL, "testuser", "testpass", false)
	totalPorts := 6

	vlan := &Vlan{
		VlanID:        10,
		Name:          "VLAN-Test",
		UntaggedPorts: []int{1, 2}, // Members
		TaggedPorts:   []int{3},    // Members
	}

	err := client.CreateVLAN(vlan, totalPorts)
	require.NoError(t, err)
}

// Test DeleteVLAN is still unchanged.
func TestDeleteVLAN(t *testing.T) {
	// Simulate deleting a VLAN
	server := mockServerMock("", http.StatusOK)
	defer server.Close()

	// Create an HRUIClient and point it to the mock server URL
	client, _ := NewClient(server.URL, "testuser", "testpass", false)

	// Test that deleting the VLAN sends the correct form submission
	err := client.DeleteVLAN(10)
	require.NoError(t, err)
}

func TestGetAllVLANs(t *testing.T) {
	htmlResponse := `
        <html>
        <body>
        <center>
        <fieldset>
        <legend>802.1Q VLAN</legend>
        <form method="post" action="/vlan.cgi?page=getRmvVlanEntry" name=formVlanStatus>
         <table border="1">
          <tr>
           <th nowrap width="60">VLAN</th>
           <th nowrap>VLAN Name</th>
           <th nowrap>Member Ports</th>
           <th nowrap>Tagged Ports</th>
           <th nowrap>Untagged Ports</th>
           <th nowrap width="50">Delete</th>
          </tr>
          <tr>
           <td width="60"><a method="post" href="/vlan.cgi?page=getVlanEntry&pickVlanId=1">1</a></td>
           <td></td>
           <td nowrap>1-6</td>
           <td nowrap>-</td>
           <td nowrap>1-6</td>
           <td width="50"><input type="checkbox" name="remove_1" id=vlan_0 disabled></td>
          </tr>
          <tr>
           <td width="60"><a method="post" href="/vlan.cgi?page=getVlanEntry&pickVlanId=4">4</a></td>
           <td>vier</td>
           <td nowrap>6</td>
           <td nowrap>-</td>
           <td nowrap>6</td>
           <td width="50"><input type="checkbox" name="remove_4" id=vlan_1></td>
          </tr>
          <tr>
           <td width="60"><a method="post" href="/vlan.cgi?page=getVlanEntry&pickVlanId=5">5</a></td>
           <td>funf</td>
           <td nowrap>5</td>
           <td nowrap>-</td>
           <td nowrap>5</td>
           <td width="50"><input type="checkbox" name="remove_5" id=vlan_2></td>
          </tr>
          <tr>
           <td width="60"><a method="post" href="/vlan.cgi?page=getVlanEntry&pickVlanId=10">10</a></td>
           <td>myvlan1</td>
           <td nowrap>1,3-5</td>
           <td nowrap>1,4-5</td>
           <td nowrap>3</td>
           <td width="50"><input type="checkbox" name="remove_10" id=vlan_3></td>
          </tr>
         </table>
         <br style="line-height:50%">
         <input type="submit" name="Delete" value="    Delete    ">
         <input type="button" value=" Select All " onclick="vlanstatic_selAll(4)">
        </form>
        </fieldset>
        </center>
        </body>
        </html>
`

	server := mockServerMock(htmlResponse, http.StatusOK)
	defer server.Close()

	client, _ := NewClient(server.URL, "testuser", "testpass", false)

	vlans, err := client.GetAllVLANs()
	assert.NoError(t, err)

	// Print the number of VLANs found
	fmt.Printf("Number of VLANs found: %d\n", len(vlans))

	// Print the extracted VLANs
	for _, vlan := range vlans {
		fmt.Printf("VLAN: %+v\n", vlan)
	}

	assert.Equal(t, 4, len(vlans))

	expectedVLANs := []*Vlan{
		{VlanID: 1, Name: "", MemberPorts: []int{1, 2, 3, 4, 5, 6}, TaggedPorts: []int{}, UntaggedPorts: []int{1, 2, 3, 4, 5, 6}},
		{VlanID: 4, Name: "vier", MemberPorts: []int{6}, TaggedPorts: []int{}, UntaggedPorts: []int{6}},
		{VlanID: 5, Name: "funf", MemberPorts: []int{5}, TaggedPorts: []int{}, UntaggedPorts: []int{5}},
		{VlanID: 10, Name: "myvlan1", MemberPorts: []int{1, 3, 4, 5}, TaggedPorts: []int{1, 4, 5}, UntaggedPorts: []int{3}},
	}
	assert.Equal(t, expectedVLANs, vlans)
}

func TestGetAllPortVLANConfigs(t *testing.T) {
	htmlResponse := `
        <html><head></head>
        <body>
            <table border="1">
                <tbody>
                    <tr>
                        <th width="90">Port</th>
                        <th width="90">PVID</th>
                        <th width="160">Accepted Frame Type</th>
                    </tr>
                    <tr>
                        <td align="center">Port 1</td>
                        <td align="center">1</td>
                        <td align="center" style="width:180px;">All</td>
                    </tr>
                    <tr>
                        <td align="center">Port 2</td>
                        <td align="center">1</td>
                        <td align="center" style="width:180px;">All</td>
                    </tr>
                    <tr>
                        <td align="center">Port 3</td>
                        <td align="center">1</td>
                        <td align="center" style="width:180px;">All</td>
                    </tr>
                    <tr>
                        <td align="center">Port 4</td>
                        <td align="center">1</td>
                        <td align="center" style="width:180px;">All</td>
                    </tr>
                    <tr>
                        <td align="center">Port 5</td>
                        <td align="center">1</td>
                        <td align="center" style="width:180px;">All</td>
                    </tr>
                    <tr>
                        <td align="center">Port 6</td>
                        <td align="center">10</td>
                        <td align="center" style="width:180px;">All</td>
                    </tr>
                </tbody>
            </table>
        </body>
        </html>
    `

	server := mockServerMock(htmlResponse, http.StatusOK)
	defer server.Close()

	client, _ := NewClient(server.URL, "testuser", "testpass", false)

	configs, err := client.GetAllPortVLANConfigs()
	assert.NoError(t, err)

	// Print the number of configs found
	fmt.Printf("Number of configs found: %d\n", len(configs))

	// Print the extracted configs
	for _, config := range configs {
		fmt.Printf("Config: %+v\n", config)
	}

	assert.Equal(t, 6, len(configs))

	// Assert the values for each port
	expectedConfigs := []*PortVLANConfig{
		{Port: 1, PVID: 1, AcceptFrameType: "All"},
		{Port: 2, PVID: 1, AcceptFrameType: "All"},
		{Port: 3, PVID: 1, AcceptFrameType: "All"},
		{Port: 4, PVID: 1, AcceptFrameType: "All"},
		{Port: 5, PVID: 1, AcceptFrameType: "All"},
		{Port: 6, PVID: 10, AcceptFrameType: "All"},
	}
	assert.Equal(t, expectedConfigs, configs)
}
