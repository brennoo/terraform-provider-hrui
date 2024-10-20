package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func (p *hruiProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "hrui"
}
