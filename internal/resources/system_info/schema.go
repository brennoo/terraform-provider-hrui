package system_info

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func (d *systemInfoDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{ // Added the 'id' attribute
				Computed:    true,
				Description: "The ID of the system info resource.",
			},
			"device_model": schema.StringAttribute{
				Computed:    true,
				Description: "The device model of the HRUI switch.",
			},
			"mac_address": schema.StringAttribute{
				Computed:    true,
				Description: "The MAC address of the HRUI switch.",
			},
			"ip_address": schema.StringAttribute{
				Computed:    true,
				Description: "The IP address of the HRUI switch.",
			},
			"netmask": schema.StringAttribute{
				Computed:    true,
				Description: "The netmask of the HRUI switch.",
			},
			"gateway": schema.StringAttribute{
				Computed:    true,
				Description: "The gateway of the HRUI switch.",
			},
			"firmware_version": schema.StringAttribute{
				Computed:    true,
				Description: "The firmware version of the HRUI switch.",
			},
			"firmware_date": schema.StringAttribute{
				Computed:    true,
				Description: "The firmware date of the HRUI switch.",
			},
			"hardware_version": schema.StringAttribute{
				Computed:    true,
				Description: "The hardware version of the HRUI switch.",
			},
		},
	}
}
