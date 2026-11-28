package sdk

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// BandwidthControl holds the ingress and egress rate configuration for a given port.
type BandwidthControl struct {
	Port        string // The port identifier
	IngressRate string // The ingress bandwidth rate
	EgressRate  string // The egress bandwidth rate
}

// GetBandwidthControl retrieves the bandwidth control configuration for each port.
func (c *HRUIClient) GetBandwidthControl(ctx context.Context) ([]BandwidthControl, error) {
	// URL to the bandwidth control page
	urlBw := fmt.Sprintf("%s/port.cgi?page=bw_ctrl", c.URL)

	// Perform HTTP GET request
	respBody, err := c.Request(ctx, "GET", urlBw, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bandwidth control page: %w", err)
	}

	// Parse the HTML response using goquery
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return nil, fmt.Errorf("failed to parse bandwidth control HTML: %w", err)
	}

	// Slice to hold bandwidth control data
	var controls []BandwidthControl

	// Locate the bandwidth control table and process rows
	doc.Find("table").Last().Find("tr").Each(func(i int, row *goquery.Selection) {
		// Skip the header row
		if i == 0 {
			return
		}

		// Extract columns from the row
		cols := row.Find("td")
		if cols.Length() != 3 {
			// Skip malformed rows that don't have exactly 3 columns
			return
		}

		// Parse and normalize the values
		port := strings.TrimSpace(cols.Eq(0).Text())        // First column: Port
		ingressRate := strings.TrimSpace(cols.Eq(1).Text()) // Second column: Ingress Rate
		egressRate := strings.TrimSpace(cols.Eq(2).Text())  // Third column: Egress Rate

		// Append the parsed bandwidth control values for this port
		controls = append(controls, BandwidthControl{
			Port:        port,
			IngressRate: ingressRate,
			EgressRate:  egressRate,
		})
	})

	return controls, nil
}

// ConfigureBandwidthControl configures ingress or egress bandwidth control for a specific port.
func (c *HRUIClient) ConfigureBandwidthControl(ctx context.Context, portName string, isIngress, enable bool, rate string) error {
	// Resolve the numeric port ID from the port name
	portID, err := c.GetPortByName(ctx, portName)
	if err != nil {
		return fmt.Errorf("failed to resolve port '%s': %w", portName, err)
	}

	return c.configureBandwidthByID(ctx, portID, isIngress, enable, rate)
}

// configureBandwidthByID sets bandwidth control for a port using its numeric ID.
func (c *HRUIClient) configureBandwidthByID(ctx context.Context, portID int, isIngress, enable bool, rate string) error {
	return c.withOpLock(func() error {
		bandwidthType := "1"
		if isIngress {
			bandwidthType = "0"
		}

		state := "0"
		if enable {
			state = "1"
		}

		rateValue := "Unlimited"
		if enable && strings.ToLower(rate) != "unlimited" {
			rateValue = rate
		}

		// On v2.x firmwares, /port.cgi?page=bwctrl updates both in_rate and out_rate at once using port%d checkboxes.
		controls, err := c.GetBandwidthControl(ctx)
		if err != nil {
			return fmt.Errorf("failed to read current bandwidth control state for port %d: %w", portID, err)
		}

		inRateV2 := ""
		outRateV2 := ""
		if portID >= 0 && portID < len(controls) {
			if !strings.EqualFold(controls[portID].IngressRate, "Unlimited") {
				inRateV2 = controls[portID].IngressRate
			}
			if !strings.EqualFold(controls[portID].EgressRate, "Unlimited") {
				outRateV2 = controls[portID].EgressRate
			}
		}
		if isIngress {
			if rateValue == "Unlimited" {
				inRateV2 = ""
			} else {
				inRateV2 = rateValue
			}
		} else {
			if rateValue == "Unlimited" {
				outRateV2 = ""
			} else {
				outRateV2 = rateValue
			}
		}

		form := url.Values{}
		form.Set("cmd", "bandwidthcontrol")
		form.Set("portid", strconv.Itoa(portID))
		form.Set(fmt.Sprintf("port%d", portID), "on")
		form.Set("type", bandwidthType)
		form.Set("state", state)
		form.Set("rate", rateValue)
		form.Set("in_rate", inRateV2)
		form.Set("out_rate", outRateV2)
		form.Set("submit", "+++Apply+++")

		endpoint := fmt.Sprintf("%s/port.cgi?page=bwctrl", c.URL)

		if _, err := c.FormRequest(ctx, endpoint, form); err != nil {
			return fmt.Errorf("failed to configure bandwidth control for port ID '%d': %w", portID, err)
		}

		return nil
	})
}
