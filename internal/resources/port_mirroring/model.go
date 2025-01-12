package port_mirroring

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type portMirroringModel struct {
	MirrorDirection types.String `tfsdk:"mirror_direction"`
	MirroringPort   types.String `tfsdk:"mirroring_port"`
	MirroredPort    types.String `tfsdk:"mirrored_port"`
}
