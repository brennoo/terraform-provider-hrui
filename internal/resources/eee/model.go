package eee

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// eeeModel represents the state model for the EEE Terraform resource.
type eeeModel struct {
	ID      types.String `tfsdk:"id"`
	Enabled types.Bool   `tfsdk:"enabled"`
}
