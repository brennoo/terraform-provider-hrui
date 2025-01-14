package jumbo_frame

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// jumboFrameModel represents the Terraform resource data model for Jumbo Frame settings.
type jumboFrameModel struct {
	Size types.Int64 `tfsdk:"size"`
}
