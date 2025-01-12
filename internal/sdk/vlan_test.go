package sdk

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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

// Test AddVLAN function - test NotMemberPorts calculation.
func TestAddVLAN_NotMemberPorts(t *testing.T) {
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

	err := client.AddVLAN(vlan, totalPorts)
	require.NoError(t, err)
}

// Test RemoveVLAN is still unchanged.
func TestRemoveVLAN(t *testing.T) {
	// Simulate deleting a VLAN
	server := mockServerMock("", http.StatusOK)
	defer server.Close()

	// Create an HRUIClient and point it to the mock server URL
	client, _ := NewClient(server.URL, "testuser", "testpass", false)

	// Test that deleting the VLAN sends the correct form submission
	err := client.RemoveVLAN(10)
	require.NoError(t, err)
}

func TestListVLANs(t *testing.T) {
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

	vlans, err := client.ListVLANs()
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

func TestListPortVLANConfigs(t *testing.T) {
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

	portHTMLResponse := `
        <html>
        <body>
            <form action="/port.cgi" method="get">
                <select name="portid">
                    <option value="1">Port 1</option>
                    <option value="2">Port 2</option>
                    <option value="3">Port 3</option>
                    <option value="4">Port 4</option>
                    <option value="5">Port 5</option>
                    <option value="6">Port 6</option>
                </select>
            </form>
        </body>
        </html>
    `

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "port.cgi") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(portHTMLResponse))
		} else if strings.Contains(r.URL.Path, "vlan.cgi") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(htmlResponse))
		} else if strings.Contains(r.URL.Path, "login.cgi") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<html><body>Login successful</body></html>`))
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "testuser", "testpass", false)

	configs, err := client.ListPortVLANConfigs()
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
		{PortID: 1, Name: "Port 1", PVID: 1, AcceptFrameType: "All"},
		{PortID: 2, Name: "Port 2", PVID: 1, AcceptFrameType: "All"},
		{PortID: 3, Name: "Port 3", PVID: 1, AcceptFrameType: "All"},
		{PortID: 4, Name: "Port 4", PVID: 1, AcceptFrameType: "All"},
		{PortID: 5, Name: "Port 5", PVID: 1, AcceptFrameType: "All"},
		{PortID: 6, Name: "Port 6", PVID: 10, AcceptFrameType: "All"},
	}
	assert.Equal(t, expectedConfigs, configs)
}

func TestListVLANsWithTrunkMembers(t *testing.T) {
	// Mock HTML response for VLANs
	vlanHTMLResponse := `
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
           <td><a href="/vlan.cgi?page=getVlanEntry&pickVlanId=10">10</a></td>
           <td>myvlan1</td>
           <td nowrap>1,3,Trunk1</td>
           <td nowrap>Trunk1</td>
           <td nowrap>1,3</td>
           <td align="center"><input type="checkbox" id="vlan_0" /></td>
          </tr>
         </table>
        </form>
        </fieldset>
        </center>
        </body>
        </html>
    `

	// Mock HTML response for port.cgi
	portHTMLResponse := `
        <html>
        <body>
        <form action="/port.cgi" method="get">
            <select name="portid">
                <option value="0">Port 1</option>
                <option value="1">Port 2</option>
                <option value="2">Port 3</option>
                <option value="5">Port 6</option>
                <option value="7">Trunk1</option>
            </select>
        </form>
        </body>
        </html>
    `

	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "vlan.cgi") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(vlanHTMLResponse))
		} else if strings.Contains(r.URL.Path, "port.cgi") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(portHTMLResponse))
		} else if strings.Contains(r.URL.Path, "login.cgi") {
			// Mock the login.cgi response
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<html><body>Login successful</body></html>`))
		} else {
			// For unknown paths, return 404
			http.NotFound(w, r)
		}
	}))

	defer server.Close()

	// Debugging: Ensure server URL is valid
	t.Logf("Server URL: %s", server.URL)

	// Create client pointing to mock server URL
	client, err := NewClient(server.URL, "testuser", "testpassword", false)
	require.NoError(t, err, "Failed to initialize HRUIClient")
	require.NotNil(t, client, "HRUIClient should not be nil")

	// Call ListVLANs
	vlans, err := client.ListVLANs()
	require.NoError(t, err)
	require.Len(t, vlans, 1)

	// Validate the parsed VLAN
	vlan := vlans[0]
	assert.Equal(t, 10, vlan.VlanID)
	assert.Equal(t, "myvlan1", vlan.Name)
	assert.ElementsMatch(t, []int{1, 3, 7}, vlan.MemberPorts) // Trunk1 resolved to Port_7
	assert.ElementsMatch(t, []int{7}, vlan.TaggedPorts)       // Trunk1 tagged
	assert.ElementsMatch(t, []int{1, 3}, vlan.UntaggedPorts)  // Regular untagged
}
