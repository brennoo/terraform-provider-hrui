package sdk

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TODO: this needs some love :-)

// FlattenInt64List normalizes and converts []int to []types.Int64 for use with Terraform
func FlattenInt64List(intList []int) []types.Int64 {
	var result []types.Int64
	for _, v := range intList {
		result = append(result, types.Int64Value(int64(v)))
	}
	if len(result) == 0 {
		return []types.Int64{}
	}
	return result
}

// ConvertTerraformIntList converts a Terraform []types.Int64 to a native Go int slice
func ConvertTerraformIntList(intList []types.Int64) []int {
	var result []int
	for _, v := range intList {
		result = append(result, int(v.ValueInt64()))
	}
	return result
}

// ConvertToNativeIntList converts a Terraform []types.Int64 (used in provider) to a native Go []int (used in SDK)
func ConvertToNativeIntList(intList []types.Int64) []int {
	var result []int
	for _, v := range intList {
		if !v.IsNull() {
			result = append(result, int(v.ValueInt64()))
		}
	}
	return result
}

// ConvertToTFInt64List converts a native []int to []types.Int64 for Terraform usage
func ConvertToTFInt64List(intList []int) []types.Int64 {
	var result []types.Int64
	for _, port := range intList {
		result = append(result, types.Int64Value(int64(port)))
	}
	return result
}
