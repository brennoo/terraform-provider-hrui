package sdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Common HTML Responses.
const (
	sampleVLANHTMLResponse = `  
	<html>  
	<body>  
	<form name='formVlanStatus'>  
	<table>  
	<tr>  
	<th>ID</th> <th>Name</th> <th>Members</th> <th>Tagged</th> <th>Untagged</th>  
	</tr>  
	<tr>  
	<td width="60">  
	<a method="post" href="/vlan.cgi?page=getVlanEntry&pickVlanId=10">10</a>  
	</td>  
	<td>myvlan-lala</td>  
	<td nowrap>2-3,Trunk2</td>  
	<td nowrap>Trunk2</td>  
	<td nowrap>2-3</td>  
	<td width="50">  
	<input type="checkbox" name="remove_10" id=vlan_3>  
	</td>  
	</tr>  
	</table>  
	</form>  
	</body>  
	</html>`

	sampleVLANPVIDHTMLResponse = `  
	<html>  
	<body>  
	<table>  
	<tr> <th>Port</th> <th>PVID</th> <th>Accepted Frame Type</th> </tr>  
	<tr> <td>Port 1</td> <td>1</td> <td>All</td> </tr>  
	<tr> <td>Port 2</td> <td>20</td> <td>Tagged Only</td> </tr>  
	<tr> <td>Port 10</td> <td>100</td> <td>All</td> </tr>  
	</table>  
	</body>  
	</html>`

	samplePortHTMLResponse = `  
	<html>  
	<body>  
	<form action="/port.cgi" method="get">  
	<select name="portid">  
	<option value="0">Port 1</option>  
	<option value="1">Port 2</option>  
	<option value="2">Port 3</option>  
	<option value="8">Trunk2</option>   
	<option value="10">Port 10</option>  
	</select>  
	</form>  
	</body>  
	</html>`
)

// Helper for Mock Server.
func setupMockServer() *httptest.Server {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/port.cgi" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(samplePortHTMLResponse))
			return
		}
		if strings.Contains(r.URL.Path, "/vlan.cgi") {
			if strings.Contains(r.URL.RawQuery, "page=port_based") {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(sampleVLANPVIDHTMLResponse))
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(sampleVLANHTMLResponse))
			return
		}
		http.NotFound(w, r)
	}))
	return mock
}

// Mock Client Creator.
func setupClient(mockURL string) *HRUIClient {
	return &HRUIClient{
		URL:        mockURL,
		HttpClient: &http.Client{},
	}
}

// Test GetVLAN.
func TestGetVLAN(t *testing.T) {
	mock := setupMockServer()
	defer mock.Close()

	client := setupClient(mock.URL)

	vlan, err := client.GetVLAN(context.Background(), 10)
	require.NoError(t, err)
	require.NotNil(t, vlan)

	assert.Equal(t, 10, vlan.VlanID)
	assert.Equal(t, "myvlan-lala", vlan.Name)
	assert.ElementsMatch(t, []string{"Port 2", "Port 3", "Trunk2"}, vlan.MemberPorts)
}

// Test AddVLAN.
func TestAddVLAN(t *testing.T) {
	mock := setupMockServer()
	defer mock.Close()

	client := setupClient(mock.URL)

	newVLAN := &Vlan{
		VlanID:        10,
		Name:          "TestVLAN",
		UntaggedPorts: []string{"eth1"},
		TaggedPorts:   []string{"eth2"},
	}

	err := client.AddVLAN(context.Background(), newVLAN)
	require.NoError(t, err)
}

// Test RemoveVLAN.
func TestRemoveVLAN(t *testing.T) {
	mock := setupMockServer()
	defer mock.Close()

	client := setupClient(mock.URL)

	require.NoError(t, client.RemoveVLAN(context.Background(), 10))
}

// Test ListVLANs.
func TestListVLANs(t *testing.T) {
	mock := setupMockServer()
	defer mock.Close()

	client := setupClient(mock.URL)

	vlans, err := client.ListVLANs(context.Background())
	assert.NoError(t, err)
	assert.NotEmpty(t, vlans)
}

// Test ListPortVLANConfigs.
func TestListPortVLANConfigs(t *testing.T) {
	mock := setupMockServer()
	defer mock.Close()

	client := setupClient(mock.URL)

	configs, err := client.ListPortVLANConfigs(context.Background())
	assert.NoError(t, err)
	assert.NotEmpty(t, configs)

	expected := []*PortVLANConfig{
		{PortName: "Port 1", PortID: 0, PVID: 1, AcceptFrameType: "All"},
		{PortName: "Port 2", PortID: 1, PVID: 20, AcceptFrameType: "Tagged Only"},
		{PortName: "Port 10", PortID: 10, PVID: 100, AcceptFrameType: "All"},
	}

	assert.Equal(t, expected, configs)
}

// Test ListVLANs With Trunk Members.
func TestListVLANsWithTrunkMembers(t *testing.T) {
	mock := setupMockServer()
	defer mock.Close()

	client := setupClient(mock.URL)

	configs, err := client.ListPortVLANConfigs(context.Background())
	assert.NoError(t, err)
	assert.NotEmpty(t, configs)
}
