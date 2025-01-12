package port_settings

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// portSettingModel maps the resource and data source schema data.
type portSettingModel struct {
	Port        types.String            `tfsdk:"port"`
	Enabled     types.Bool              `tfsdk:"enabled"`
	Speed       *portSettingSpeed       `tfsdk:"speed"`
	FlowControl *portSettingFlowControl `tfsdk:"flow_control"`
}

type portSettingSpeed struct {
	Config types.String `tfsdk:"config"`
	Actual types.String `tfsdk:"actual"`
}

type portSettingFlowControl struct {
	Config types.String `tfsdk:"config"`
	Actual types.String `tfsdk:"actual"`
}
