package port_isolation

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type portIsolationModel struct {
	Port          types.String `tfsdk:"port"`
	IsolationList types.List   `tfsdk:"isolation_list"`
}
