package port_settings

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

const (
	speedDuplexField = "speed_duplex"
	flowControlField = "flow_control"
)

var portSettingSchema = map[string]rschema.Attribute{
	"id": dschema.StringAttribute{
		Computed:    true,
		Description: "The ID of the port setting data source.",
	},
	"port_id": rschema.Int64Attribute{
		Required:    true,
		Description: "The ID of the port to configure.",
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.UseStateForUnknown(),
		},
	},
	"enabled": rschema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Whether the port is enabled.",
	},
	speedDuplexField: rschema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "The speed and duplex mode of the port.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	flowControlField: rschema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "The flow control setting of the port.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
}

// Removed duplicate definitions of portSettingSpeed and portSettingFlowControl

func (r *portSettingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Attributes: portSettingSchema,
	}
}

func (d *portSettingDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the port setting data source.",
			},
			"port_id": dschema.Int64Attribute{
				Required:    true,
				Description: "The ID of the port to read.",
			},
			"enabled": dschema.BoolAttribute{
				Computed:    true,
				Description: "Whether the port is enabled.",
			},
			"speed": dschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dschema.Attribute{
					"config": dschema.StringAttribute{
						Computed:    true,
						Description: "Configured speed and duplex mode.",
					},
					"actual": dschema.StringAttribute{
						Computed:    true,
						Description: "Actual speed and duplex mode.",
					},
				},
				Description: "Speed and duplex settings of the port.",
			},
			"flow_control": dschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dschema.Attribute{
					"config": dschema.StringAttribute{
						Computed:    true,
						Description: "Configured flow control setting.",
					},
					"actual": dschema.StringAttribute{
						Computed:    true,
						Description: "Actual flow control setting.",
					},
				},
				Description: "Flow control settings of the port.",
			},
		},
	}
}
