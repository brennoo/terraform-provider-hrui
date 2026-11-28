package sdk

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Vlan struct {
	VlanID        int
	Name          string
	UntaggedPorts []string
	TaggedPorts   []string
	MemberPorts   []string
}

type PortVLANConfig struct {
	PortID          int
	PortName        string
	PVID            int
	AcceptFrameType string
}

// AddVLAN creates or updates a VLAN on the switch.
func (c *HRUIClient) AddVLAN(ctx context.Context, vlan *Vlan) error {
	if c == nil {
		return fmt.Errorf("HRUIClient is nil")
	}

	portConfigs, err := c.ListPortVLANConfigs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get port VLAN configurations: %w", err)
	}

	form := url.Values{}
	form.Set("vid", fmt.Sprintf("%d", vlan.VlanID))
	form.Set("name", vlan.Name)

	for _, portConfig := range portConfigs {
		formValue := "2"

		if containsString(vlan.UntaggedPorts, portConfig.PortName) {
			formValue = "0"
		} else if containsString(vlan.TaggedPorts, portConfig.PortName) {
			formValue = "1"
		}

		form.Set(fmt.Sprintf("vlanPort_%d", portConfig.PortID), formValue)
	}

	vlanURL := fmt.Sprintf("%s/vlan.cgi?page=static", c.URL)
	respBody, err := c.FormRequest(ctx, vlanURL, form)
	if err != nil {
		return fmt.Errorf("failed to create/update VLAN: %w", err)
	}

	// On v2.x firmwares (where /vlan.cgi?page=static only creates the VLAN ID/Name and renders formVlanDelStatic),
	// configure per-port Untagged/Tagged membership on /vlan.cgi?page=port_based.
	if strings.Contains(string(respBody), "formVlanDelStatic") {
		hasHybrid := false
		if pbBody, pbErr := c.Request(ctx, "GET", fmt.Sprintf("%s/vlan.cgi?page=port_based", c.URL), nil, nil); pbErr == nil {
			hasHybrid = strings.Contains(string(pbBody), "vlanboxHybrid")
		}
		for _, portConfig := range portConfigs {
			isUntag := containsString(vlan.UntaggedPorts, portConfig.PortName)
			isTag := containsString(vlan.TaggedPorts, portConfig.PortName)
			pbForm := url.Values{}
			pbForm.Set(fmt.Sprintf("port%d", portConfig.PortID), "on")
			if hasHybrid {
				pbForm.Set("vlan_type", "2") // Hybrid mode (V200.2.5)
				pbForm.Set("accept_type", "0")
				pbForm.Set("vlan_native", "1")
				pbForm.Set("VLAN1_untag", "on")
				if isUntag {
					pbForm.Set(fmt.Sprintf("VLAN%d_untag", vlan.VlanID), "on")
				}
				if isTag {
					pbForm.Set(fmt.Sprintf("VLAN%d_tag", vlan.VlanID), "on")
				}
			} else {
				// v2.0.1 only supports Access (0) and Trunk (1)
				switch {
				case isUntag && !isTag:
					pbForm.Set("vlan_type", "0")
					pbForm.Set("vlan_access", strconv.Itoa(vlan.VlanID))
				case isTag:
					nativeVID := "1"
					if isUntag {
						nativeVID = strconv.Itoa(vlan.VlanID)
					}
					pbForm.Set("vlan_type", "1")
					pbForm.Set("vlan_trunk", nativeVID)
					pbForm.Set("VLAN1", "on")
					pbForm.Set(fmt.Sprintf("VLAN%d", vlan.VlanID), "on")
				default:
					pbForm.Set("vlan_type", "0")
					pbForm.Set("vlan_access", "1")
				}
			}
			_, _ = c.FormRequest(ctx, fmt.Sprintf("%s/vlan.cgi?page=port_based", c.URL), pbForm)
		}
	}

	return nil
}

// Helper function to check if a slice contains a string.
func containsString(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// GetVLAN fetches a single VLAN by its VLAN ID.
func (c *HRUIClient) GetVLAN(ctx context.Context, vlanID int) (*Vlan, error) {
	vlans, err := c.ListVLANs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch VLANs: %w", err)
	}

	for _, vlan := range vlans {
		if vlan.VlanID == vlanID {
			return vlan, nil
		}
	}

	return nil, fmt.Errorf("VLAN with ID %d not found", vlanID)
}

// ListVLANs fetches the list of VLANs using port names.
func (c *HRUIClient) ListVLANs(ctx context.Context) ([]*Vlan, error) {
	if c == nil {
		return nil, fmt.Errorf("HRUIClient is nil")
	}

	vlanURL := fmt.Sprintf("%s/vlan.cgi?page=static", c.URL)
	respBody, err := c.Request(ctx, "GET", vlanURL, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch VLAN configuration from HRUI: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse VLAN HTML output: %w", err)
	}

	var vlans []*Vlan

	doc.Find("form[name='formVlanStatus'] table tr").Each(func(i int, s *goquery.Selection) {
		if i == 0 { // Skip table header row
			return
		}

		vlan := &Vlan{}

		// Extract VLAN ID from the <a> tag in the first table cell
		vlanIDText := strings.TrimSpace(s.Find("td:nth-child(1) a").Text())
		if vlanIDText == "" {
			fmt.Println("Skipping row: VLAN ID not found")
			return
		}

		vlanID, err := strconv.Atoi(vlanIDText)
		if err != nil {
			fmt.Printf("Error parsing VLAN ID (%s): %v\n", vlanIDText, err)
			return
		}
		vlan.VlanID = vlanID

		// Extract VLAN Name
		vlan.Name = strings.TrimSpace(s.Find("td:nth-child(2)").Text())

		// Extract port configurations
		vlan.MemberPorts = c.ParsePortRange(strings.TrimSpace(s.Find("td:nth-child(3)").Text()))
		vlan.TaggedPorts = c.ParsePortRange(strings.TrimSpace(s.Find("td:nth-child(4)").Text()))
		vlan.UntaggedPorts = c.ParsePortRange(strings.TrimSpace(s.Find("td:nth-child(5)").Text()))

		vlans = append(vlans, vlan)
	})

	// On v2.x firmwares, the static VLAN table is inside form#formVlanDelStatic and port membership is on /vlan.cgi?page=port_based.
	if len(vlans) == 0 {
		vlanByID := make(map[int]*Vlan)
		doc.Find("form#formVlanDelStatic table tr").Each(func(i int, s *goquery.Selection) {
			if i == 0 {
				return
			}
			tds := s.Find("td")
			if tds.Length() < 4 {
				return
			}
			vlanID, convErr := strconv.Atoi(strings.TrimSpace(tds.Eq(2).Text()))
			if convErr != nil {
				return
			}
			v := &Vlan{
				VlanID:        vlanID,
				Name:          strings.TrimSpace(tds.Eq(3).Text()),
				UntaggedPorts: []string{},
				TaggedPorts:   []string{},
				MemberPorts:   []string{},
			}
			vlans = append(vlans, v)
			vlanByID[vlanID] = v
		})

		if len(vlans) > 0 {
			if pbBody, pbErr := c.Request(ctx, "GET", fmt.Sprintf("%s/vlan.cgi?page=port_based", c.URL), nil, nil); pbErr == nil {
				if pbDoc, parseErr := goquery.NewDocumentFromReader(strings.NewReader(string(pbBody))); parseErr == nil {
					pbDoc.Find("table").Last().Find("tr").Slice(1, goquery.ToEnd).Each(func(_ int, row *goquery.Selection) {
						tds := row.Find("td")
						if tds.Length() < 5 {
							return
						}
						portName := strings.TrimSpace(tds.Eq(0).Text())
						mode := strings.TrimSpace(tds.Eq(1).Text())
						if mode == "Access" {
							if vid, convErr := strconv.Atoi(strings.TrimSpace(tds.Eq(2).Text())); convErr == nil {
								if v, ok := vlanByID[vid]; ok {
									v.UntaggedPorts = append(v.UntaggedPorts, portName)
									v.MemberPorts = append(v.MemberPorts, portName)
								}
							}
							return
						}
						if mode == "Trunk" {
							nativeVID, _ := strconv.Atoi(strings.TrimSpace(tds.Eq(3).Text()))
							if v, ok := vlanByID[nativeVID]; ok {
								v.UntaggedPorts = append(v.UntaggedPorts, portName)
								v.MemberPorts = append(v.MemberPorts, portName)
							}
							for part := range strings.SplitSeq(strings.TrimSpace(tds.Eq(4).Text()), ",") {
								if vid, convErr := strconv.Atoi(strings.TrimSpace(part)); convErr == nil && vid != nativeVID {
									if v, found := vlanByID[vid]; found {
										v.TaggedPorts = append(v.TaggedPorts, portName)
										if !containsString(v.MemberPorts, portName) {
											v.MemberPorts = append(v.MemberPorts, portName)
										}
									}
								}
							}
							return
						}
						tds.Eq(4).Find("div.vlancontent").Each(func(_ int, div *goquery.Selection) {
							txt := strings.TrimSpace(div.Text())
							if _, after, ok := strings.Cut(txt, "Tagged VLAN:"); ok {
								for part := range strings.SplitSeq(after, ",") {
									if vid, convErr := strconv.Atoi(strings.TrimSpace(part)); convErr == nil {
										if v, found := vlanByID[vid]; found {
											v.TaggedPorts = append(v.TaggedPorts, portName)
											if !containsString(v.MemberPorts, portName) {
												v.MemberPorts = append(v.MemberPorts, portName)
											}
										}
									}
								}
							} else if _, after, ok := strings.Cut(txt, "Untagged VLAN:"); ok {
								for part := range strings.SplitSeq(after, ",") {
									if vid, convErr := strconv.Atoi(strings.TrimSpace(part)); convErr == nil {
										if v, found := vlanByID[vid]; found {
											v.UntaggedPorts = append(v.UntaggedPorts, portName)
											if !containsString(v.MemberPorts, portName) {
												v.MemberPorts = append(v.MemberPorts, portName)
											}
										}
									}
								}
							}
						})
					})
				}
			}
		}
	}

	return vlans, nil
}

// RemoveVLAN deletes a VLAN by its VLAN ID from the switch.
func (c *HRUIClient) RemoveVLAN(ctx context.Context, vlanID int) error {
	if c == nil {
		return fmt.Errorf("HRUIClient is nil")
	}

	form := url.Values{}
	removeKey := fmt.Sprintf("remove_%d", vlanID)
	// On v2.x firmwares, the remove checkbox in form#formVlanDelStatic is indexed by row number (e.g. remove_2 for VLAN 100).
	if respBody, getErr := c.Request(ctx, "GET", fmt.Sprintf("%s/vlan.cgi?page=static", c.URL), nil, nil); getErr == nil {
		if doc, parseErr := goquery.NewDocumentFromReader(strings.NewReader(string(respBody))); parseErr == nil {
			doc.Find("form#formVlanDelStatic table tr").Each(func(i int, s *goquery.Selection) {
				if i == 0 {
					return
				}
				tds := s.Find("td")
				if tds.Length() >= 3 && strings.TrimSpace(tds.Eq(2).Text()) == strconv.Itoa(vlanID) {
					if chkName := strings.TrimSpace(tds.Eq(0).Find("input[type='checkbox']").AttrOr("name", "")); chkName != "" {
						removeKey = chkName
					}
				}
			})
		}
	}
	form.Set(removeKey, "on")

	deleteURL := fmt.Sprintf("%s/vlan.cgi?page=getRmvVlanEntry", c.URL)

	_, err := c.FormRequest(ctx, deleteURL, form)
	if err != nil {
		return fmt.Errorf("failed to delete VLAN: %w", err)
	}

	return nil
}

// ListPortVLANConfigs fetches the VLAN configuration for all ports.
func (c *HRUIClient) ListPortVLANConfigs(ctx context.Context) ([]*PortVLANConfig, error) {
	if c == nil {
		return nil, fmt.Errorf("HRUIClient is nil")
	}

	if c.URL == "" {
		return nil, fmt.Errorf("HRUIClient.URL is empty")
	}

	portVLANURL := fmt.Sprintf("%s/vlan.cgi?page=port_based", c.URL)

	respBody, err := c.Request(ctx, "GET", portVLANURL, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch port VLAN configuration from HRUI: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(respBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse port VLAN HTML output: %w", err)
	}

	var configs []*PortVLANConfig

	rows := doc.Find("table").Last().Find("tr").Slice(1, goquery.ToEnd)

	rows.Each(func(i int, s *goquery.Selection) {
		config := &PortVLANConfig{}

		// Extract the port name (e.g., "Port 1", "Trunk2").
		portName := strings.TrimSpace(s.Find("td:nth-child(1)").Text())
		if portName == "" {
			fmt.Printf("Skipping row %d due to missing port name\n", i)
			return
		}
		config.PortName = portName

		portID, err := c.GetPortByName(ctx, portName)
		if err != nil {
			fmt.Printf("Skipping row %d due to PortID resolution error for '%s': %v\n", i, portName, err)
			return
		}
		config.PortID = portID

		tds := s.Find("td")
		if tds.Length() >= 5 {
			// v2.x layout: td[0]=Port, td[1]=Port vlan type ("Access", "Trunk", "Hybrid:All", "Hybrid:Tag Only", "Hybrid:Untag Only"), td[2]=Access VLAN, td[3]=Native VLAN
			modeText := strings.TrimSpace(tds.Eq(1).Text())
			pvidStr := strings.TrimSpace(tds.Eq(2).Text())
			if pvidStr == "-" || pvidStr == "" {
				pvidStr = strings.TrimSpace(tds.Eq(3).Text())
			}
			pvid, convErr := strconv.Atoi(pvidStr)
			if convErr != nil {
				return
			}
			config.PVID = pvid
			// Access mode doesn't expose an accept-frame-type setting on this page; leave
			// AcceptFrameType empty so callers can tell "not reported by device" apart from
			// an actual value, instead of guessing one.
			frameType := ""
			if _, after, ok := strings.Cut(modeText, ":"); ok {
				switch strings.TrimSpace(after) {
				case "Tag Only":
					frameType = "Tag-only"
				case "Untag Only":
					frameType = "Untag-only"
				default:
					frameType = strings.TrimSpace(after)
				}
			} else if modeText == "Trunk" {
				frameType = "Tag-only"
			}
			config.AcceptFrameType = frameType
			configs = append(configs, config)
			return
		}

		pvidText := strings.TrimSpace(s.Find("td:nth-child(2)").Text())
		var pvid int
		_, err = fmt.Sscanf(pvidText, "%d", &pvid)
		if err != nil {
			fmt.Printf("Skipping row %d due to PVID parsing error: %v\n", i, err)
			return
		}
		config.PVID = pvid

		config.AcceptFrameType = strings.TrimSpace(s.Find("td:nth-child(3)").Text())

		configs = append(configs, config)
	})

	return configs, nil
}

// GetPortVLANConfig fetches the VLAN configuration for a specific port on the switch.
func (c *HRUIClient) GetPortVLANConfig(ctx context.Context, port int) (*PortVLANConfig, error) {
	configs, err := c.ListPortVLANConfigs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all port VLAN configs: %w", err)
	}

	for _, config := range configs {
		if config.PortID == port {
			return config, nil
		}
	}

	return nil, fmt.Errorf("port %d not found in %d ports", port, len(configs))
}

// SetPortVLANConfig sets the VLAN configuration for a specific port on the switch.
func (c *HRUIClient) SetPortVLANConfig(ctx context.Context, config *PortVLANConfig) error {
	if c == nil {
		return fmt.Errorf("HRUIClient is nil")
	}

	if config == nil {
		return fmt.Errorf("PortVLANConfig is nil")
	}

	// Resolve PortID if it's invalid or missing
	// Note: PortID 0 is valid (Port 1 uses 0-based indexing)
	if config.PortID < 0 {
		if config.PortName == "" {
			return fmt.Errorf("invalid configuration: both PortID and Name are missing")
		}

		portID, err := c.GetPortByName(ctx, config.PortName)
		if err != nil {
			return fmt.Errorf("failed to resolve PortID for Name '%s': %w", config.PortName, err)
		}

		config.PortID = portID
	}

	// Validate that PortID is now valid (device uses 0-based indexing, so 0 is valid)
	if config.PortID < 0 {
		return fmt.Errorf("invalid PortID after resolution: %d", config.PortID)
	}

	// Prepare form data (supports v1.x, v2.0.1 Access/Trunk, and V200.2.5 Hybrid /vlan.cgi?page=port_based)
	pbBody, pbErr := c.Request(ctx, "GET", fmt.Sprintf("%s/vlan.cgi?page=port_based", c.URL), nil, nil)
	if pbErr != nil {
		return fmt.Errorf("failed to detect VLAN port mode for port %d: %w", config.PortID, pbErr)
	}
	hasHybrid := !(strings.Contains(string(pbBody), "vlanState") && !strings.Contains(string(pbBody), "vlanboxHybrid"))
	if !hasHybrid {
		// Ensure the target PVID VLAN exists in /vlan.cgi?page=static on v2.0.1 so Access/Trunk accepts it.
		if config.PVID > 1 {
			stForm := url.Values{}
			stForm.Set("vid", strconv.Itoa(config.PVID))
			stForm.Set("name", fmt.Sprintf("VLAN%d", config.PVID))
			_, _ = c.FormRequest(ctx, fmt.Sprintf("%s/vlan.cgi?page=static", c.URL), stForm)
		}
	}

	form := url.Values{}
	form.Set("ports", strconv.Itoa(config.PortID))
	form.Set("pvid", strconv.Itoa(config.PVID))
	form.Set(fmt.Sprintf("port%d", config.PortID), "on")
	if hasHybrid {
		form.Set("vlan_type", "2")
		form.Set("vlan_access", strconv.Itoa(config.PVID))
		form.Set("vlan_native", strconv.Itoa(config.PVID))
		form.Set("VLAN1_untag", "on")
		form.Set(fmt.Sprintf("VLAN%d_untag", config.PVID), "on")
	} else if config.AcceptFrameType == "Tag-only" {
		form.Set("vlan_type", "1")
		form.Set("vlan_trunk", strconv.Itoa(config.PVID))
		form.Set("VLAN1", "on")
		form.Set(fmt.Sprintf("VLAN%d", config.PVID), "on")
	} else {
		form.Set("vlan_type", "0")
		form.Set("vlan_access", strconv.Itoa(config.PVID))
	}

	// Map Accepted Frame Types to their respective values
	acceptFrameTypeMap := map[string]string{
		"All":        "0",
		"Tag-only":   "1",
		"Untag-only": "2",
	}

	frameTypeValue, ok := acceptFrameTypeMap[config.AcceptFrameType]
	if !ok {
		return fmt.Errorf("invalid Accepted Frame Type: %s", config.AcceptFrameType)
	}
	form.Set("vlan_accept_frame_type", frameTypeValue)
	form.Set("accept_type", frameTypeValue)

	// Submit the form
	portVLANURL := fmt.Sprintf("%s/vlan.cgi?page=port_based", c.URL)
	_, err := c.FormRequest(ctx, portVLANURL, form)
	if err != nil {
		return fmt.Errorf("failed to set port VLAN config: %w", err)
	}

	return nil
}

func (c *HRUIClient) ParsePortRange(portStr string) []string {
	ports := []string{}
	entries := strings.Split(portStr, ",")

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)

		if entry == "-" || entry == "" {
			continue
		}

		if strings.Contains(entry, "-") {
			rangeParts := strings.Split(entry, "-")
			if len(rangeParts) == 2 {
				start, err1 := strconv.Atoi(rangeParts[0])
				end, err2 := strconv.Atoi(rangeParts[1])
				if err1 == nil && err2 == nil {
					for i := start; i <= end; i++ {
						ports = append(ports, fmt.Sprintf("Port %d", i))
					}
				}
			}
		} else if strings.HasPrefix(entry, "Trunk") {
			// Keep named ports like "Trunk2"
			ports = append(ports, entry)
		} else {
			// Handle individual ports (e.g., "2" → "Port 2")
			if num, err := strconv.Atoi(entry); err == nil {
				ports = append(ports, fmt.Sprintf("Port %d", num))
			}
		}
	}
	return ports
}
