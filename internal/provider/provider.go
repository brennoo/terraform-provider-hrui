package provider

import (
	"context"

	"github.com/brennoo/terraform-provider-hrui/internal/resources/bandwidth_control"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/eee"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/igmp_snooping"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/igmp_snooping_static"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/ip_address_settings"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/jumbo_frame"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/loop_protocol"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/mac_limit"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/mac_static"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/mac_table"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/port_isolation"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/port_mirroring"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/port_settings"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/port_statistics"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/qos_port_queue"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/qos_queue_weight"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/storm_control"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/stp_global"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/stp_port"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/system_info"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/trunk_group"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/vlan_8021q"
	"github.com/brennoo/terraform-provider-hrui/internal/resources/vlan_vid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ provider.Provider = &hruiProvider{}

// hruiProvider defines the provider implementation.
type hruiProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally.
	version string
}

// New is a helper function to simplify provider server and testing logic.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &hruiProvider{
			version: version,
		}
	}
}

// DataSources - Defines the provider's data sources, i.e. the things that
// can be queried.
func (p *hruiProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		system_info.NewDataSource,
		port_settings.NewDataSource,
		vlan_8021q.NewDataSource,
		vlan_vid.NewDataSource,
		qos_port_queue.NewDataSource,
		qos_queue_weight.NewDataSource,
		mac_table.NewDataSource,
		mac_static.NewDataSource,
		port_statistics.NewDataSource,
	}
}

// Resources - Defines the provider's resources, i.e. the things that can
// be created, updated, and deleted.
func (p *hruiProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		ip_address_settings.NewResource,
		port_settings.NewResource,
		vlan_8021q.NewResource,
		vlan_vid.NewResource,
		qos_port_queue.NewResource,
		qos_queue_weight.NewResource,
		loop_protocol.NewResource,
		stp_global.NewResource,
		stp_port.NewResource,
		mac_static.NewResource,
		storm_control.NewResource,
		igmp_snooping.NewResource,
		igmp_snooping_static.NewResource,
		trunk_group.NewResource,
		port_mirroring.NewResource,
		port_isolation.NewResource,
		bandwidth_control.NewResource,
		jumbo_frame.NewResource,
		eee.NewResource,
		mac_limit.NewResource,
	}
}
