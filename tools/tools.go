package tools

import (
	// This package is used for the go generate directives,
	// but is not included in the build itself.
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"

	// Strict go fmt
	_ "mvdan.cc/gofumpt"
)
