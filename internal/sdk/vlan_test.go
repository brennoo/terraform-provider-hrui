package sdk

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetVLAN(t *testing.T) {
	htmlResponse := `
<form name='formVlanStatus'>
  <table>
    <tr>
      <th>ID</th>
      <th>Name</th>
      <th>Members</th>
      <th>Tagged</th>
      <th>Untagged</th>
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
</form>`
	server := mockServerMock(htmlResponse, http.StatusOK)
	defer server.Close()

	client, err := NewClient(server.URL, "testuser", "testpass", false)
	require.NoError(t, err)

	// Test fetching an existing VLAN
	vlan, err := client.GetVLAN(10)
	require.NoError(t, err, "Expected to find VLAN 10, but got error")

	// Ensure VLAN fields are correctly parsed
	require.NotNil(t, vlan, "Expected a VLAN object, but got nil")
	assert.Equal(t, 10, vlan.VlanID)
	assert.Equal(t, "myvlan-lala", vlan.Name)

	assert.ElementsMatch(t, []string{"Port 2", "Port 3", "Trunk2"}, vlan.MemberPorts)
	assert.ElementsMatch(t, []string{"Trunk2"}, vlan.TaggedPorts)
	assert.ElementsMatch(t, []string{"Port 2", "Port 3"}, vlan.UntaggedPorts)

	// Test fetching a non-existent VLAN (should return an error)
	vlan, err = client.GetVLAN(20)
	assert.Error(t, err, "Expected error for VLAN 20, but got none")
	assert.Nil(t, vlan, "Expected nil VLAN for non-existent VLAN ID 20")
}

func TestAddVLAN(t *testing.T) {
	vlanHTMLResponse := `
        <html>
        <body>
        <center>
        <fieldset>
        <legend>802.1Q VLAN</legend>
        <form method="post" action="/vlan.cgi?page=getRmvVlanEntry">
         <table border="1">
          <tr>
           <td><a href="/vlan.cgi?page=getVlanEntry&pickVlanId=10">10</a></td>
           <td>TestVLAN</td>
           <td>1,2</td>
           <td>2</td>
           <td>1</td>
           <td><input type="checkbox"></td>
          </tr>
         </table>
        </form>
        </fieldset>
        </center>
        </body>
        </html> `

	portVLANHTMLResponse := `
	<table>
		<tr>
			<th>Port ID</th><th>Port Name</th><th>PVID</th><th>Accept Frame Type</th>
		</tr>
		<tr>
			<td>1</td><td>eth1</td><td>10</td><td>All</td>
		</tr>
		<tr>
			<td>2</td><td>eth2</td><td>10</td><td>Tagged Only</td>
		</tr>
	</table>
	`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "vlan.cgi") && strings.Contains(r.URL.RawQuery, "page=port_based"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(portVLANHTMLResponse))
			return

		case strings.Contains(r.URL.Path, "vlan.cgi") && strings.Contains(r.URL.RawQuery, "page=add"):
			w.WriteHeader(http.StatusOK)
			return

		case strings.Contains(r.URL.Path, "vlan.cgi"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(vlanHTMLResponse))
			return

		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := &HRUIClient{
		URL:        server.URL,
		HttpClient: server.Client(),
	}

	newVLAN := &Vlan{
		VlanID:        10,
		Name:          "TestVLAN",
		UntaggedPorts: []string{"eth1"},
		TaggedPorts:   []string{"eth2"},
	}

	err := client.AddVLAN(newVLAN)

	require.NoError(t, err, "AddVLAN should not return an error")
}

func TestRemoveVLAN(t *testing.T) {
	server := mockServerMock("", http.StatusOK)
	defer server.Close()

	client, _ := NewClient(server.URL, "testuser", "testpass", false)

	require.NoError(t, client.RemoveVLAN(10))
}

func TestListVLANs(t *testing.T) {
	htmlResponse := `
<html>
<body>
<center>
<fieldset>
<legend>802.1Q VLAN</legend>
<form name="formVlanStatus" action="/vlan.cgi?page=getRmvVlanEntry">
 <table border="1">
  <tr>
   <th>VLAN</th>
   <th>VLAN Name</th>
   <th>Member Ports</th>
   <th>Tagged Ports</th>
   <th>Untagged Ports</th>
  </tr>
  <tr>
   <td><a href="/vlan.cgi?page=getVlanEntry&pickVlanId=10">10</a></td>
   <td>myvlan1</td>
   <td>1,3-5</td>
   <td>1,5</td>
   <td>3-4</td>
  </tr>
 </table>
</form>
</fieldset>
</center>
</body>
</html>`

	server := mockServerMock(htmlResponse, http.StatusOK)
	defer server.Close()

	client, _ := NewClient(server.URL, "testuser", "testpass", false)

	vlans, err := client.ListVLANs()
	assert.NoError(t, err)
	assert.NotEmpty(t, vlans)
}

func TestListPortVLANConfigs(t *testing.T) {
	portHTMLResponse := `
    <html>
    <body>
        <form action="/port.cgi" method="get">
            <select name="portid">
                <option value="1">Port 1</option>
                <option value="2">Port 2</option>
            </select>
        </form>
    </body>
    </html>
    `

	vlanHTMLResponse := `
    <html>
    <body>
    <table>
    <tr> <th>Port</th> <th>PVID</th> <th>Accepted Frame Type</th> </tr>
    <tr> <td>Port 1</td> <td>1</td> <td>All</td> </tr>
    <tr> <td>Port 2</td> <td>20</td> <td>Tagged Only</td> </tr>
    </table>
    </body>
    </html>
    `

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/port.cgi" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(portHTMLResponse))
			return
		}

		if r.URL.Path == "/vlan.cgi" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(vlanHTMLResponse))
			return
		}

		http.NotFound(w, r)
	}))
	defer mock.Close()

	client := &HRUIClient{
		URL:        mock.URL,
		HttpClient: &http.Client{},
	}

	// Run function
	configs, err := client.ListPortVLANConfigs()
	assert.NoError(t, err, "ListPortVLANConfigs should succeed")
	assert.NotEmpty(t, configs, "VLAN configs should not be empty")

	expected := []*PortVLANConfig{
		{PortName: "Port 1", PortID: 1, PVID: 1, AcceptFrameType: "All"},
		{PortName: "Port 2", PortID: 2, PVID: 20, AcceptFrameType: "Tagged Only"},
	}

	assert.Equal(t, expected, configs, "VLAN output mismatch")
}

func TestListVLANsWithTrunkMembers(t *testing.T) {
	portHTMLResponse := `
    <html>
    <body>
        <form action="/port.cgi" method="get">
            <select name="portid">
                <option value="1">Port 1</option>
                <option value="2">Port 2</option>
            </select>
        </form>
    </body>
    </html>
    `

	vlanHTMLResponse := `
    <html>
    <body>
    <table>
    <tr> <th>Port</th> <th>PVID</th> <th>Accepted Frame Type</th> </tr>
    <tr> <td>Port 1</td> <td>1</td> <td>All</td> </tr>
    <tr> <td>Port 2</td> <td>20</td> <td>Tagged Only</td> </tr>
    </table>
    </body>
    </html>
    `

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/port.cgi" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(portHTMLResponse))
			return
		}

		if r.URL.Path == "/vlan.cgi" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(vlanHTMLResponse))
			return
		}

		http.NotFound(w, r)
	}))

	defer mock.Close()

	client := &HRUIClient{
		URL:        mock.URL,
		HttpClient: &http.Client{},
	}

	// Execute function
	configs, err := client.ListPortVLANConfigs()
	assert.NoError(t, err, "ListPortVLANConfigs should succeed")
	assert.NotEmpty(t, configs, "Expected VLAN configurations, but got empty list")

	expected := []*PortVLANConfig{
		{PortName: "Port 1", PortID: 1, PVID: 1, AcceptFrameType: "All"},
		{PortName: "Port 2", PortID: 2, PVID: 20, AcceptFrameType: "Tagged Only"},
	}

	assert.Equal(t, expected, configs, "VLAN output mismatch")
}
