package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
)

// Schema defines the provider-level schema for configuration data.
func (p *hruiProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The HRUI Terraform Provider allows you to manage network switches made by HRUI (Shenzhen HongRui Optical Technology Co., Ltd) that have a Web UI.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "URL of the HRUI switch web interface. Can also be set using the `HRUI_URL` environment variable.",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for authentication. Can also be set using the `HRUI_USERNAME` environment variable.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for authentication. Can also be set using the `HRUI_PASSWORD` environment variable.",
				Optional:            true,
			},
			"autosave": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Enable automatic saving of configuration changes after resource creation or updates. Can also be set using the `HRUI_AUTOSAVE` environment variable.",
			},
		},
	}
}
