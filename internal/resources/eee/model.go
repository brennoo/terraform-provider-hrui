package eee

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// eeeModel represents the state model for the EEE Terraform resource.
type eeeModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}
