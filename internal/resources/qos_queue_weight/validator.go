package qos_queue_weight

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// WeightValidator ensures that weight is either "Strict priority" or a valid number between 1 and 15.
type WeightValidator struct{}

// Description explains what the validator does in documentation.
func (v WeightValidator) Description(context.Context) string {
	return "Validates that the weight is either \"Strict priority\" or a number between 1 and 15."
}

// MarkdownDescription provides the description in markdown style.
func (v WeightValidator) MarkdownDescription(context.Context) string {
	return "Validates that the weight is either **`Strict priority`** or a number between **1** and **15**."
}

// ValidateString checks if the provided value is either "Strict priority" or between 1 and 15.
func (v WeightValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	weightStr := req.ConfigValue.ValueString()

	// If the value is "Strict priority", it's valid.
	if weightStr == "Strict priority" {
		return
	}

	// Try to convert value to an integer
	weightInt, err := strconv.Atoi(weightStr)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Weight",
			fmt.Sprintf("Value '%s' is invalid for weight. It must be either 'Strict priority' or an integer between 1 and 15.", weightStr))
		return
	}

	// If the value is 0, instruct the user to use "Strict priority" instead
	if weightInt == 0 {
		resp.Diagnostics.AddError(
			"Using 0 as Weight",
			"'0' is equivalent to 'Strict priority'. Please use the term 'Strict priority' for better readability.",
		)
		return
	}

	// Make sure the value is between 1 and 15
	if weightInt < 1 || weightInt > 15 {
		resp.Diagnostics.AddError("Weight Out of Range",
			"Value must be between 1 and 15, or 'Strict priority' for priority queues.")
	}
}

// NewWeightValidator returns an instance of the WeightValidator.
func NewWeightValidator() validator.String {
	return WeightValidator{}
}
