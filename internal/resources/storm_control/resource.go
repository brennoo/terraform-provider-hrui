package storm_control

import (
	"context"
	"fmt"
	"strings"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure implementation satisfies the resource.Resource interface.
var _ resource.Resource = &stormControlResource{}

type stormControlResource struct {
	client *sdk.HRUIClient
}

// Constants for supported storm types.
const (
	stormTypeBroadcast        = "broadcast"
	stormTypeKnownMulticast   = "known multicast"
	stormTypeUnknownUnicast   = "unknown unicast"
	stormTypeUnknownMulticast = "unknown multicast"
)

var validStormTypes = []string{
	"Broadcast",
	"Known Multicast",
	"Unknown Unicast",
	"Unknown Multicast",
}

// NewResource initializes a new instance of the `stormControlResource`.
func NewResource() resource.Resource {
	return &stormControlResource{}
}

// Metadata sets the resource name/type.
func (r *stormControlResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storm_control"
}

// Schema defines the schema for the resource.
func (r *stormControlResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Resource for managing storm control settings on HRUI devices.",
		Attributes: map[string]schema.Attribute{
			"port": schema.Int64Attribute{
				Description: "The port number to enable storm control on. Changing this will recreate the resource.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"storm_type": schema.StringAttribute{
				Description: fmt.Sprintf("The type of traffic to control. Options: %v.", strings.Join(validStormTypes, ", ")),
				Required:    true,
			},
			"state": schema.BoolAttribute{
				Description: "Whether storm control is enabled (`true`) or disabled (`false`).",
				Required:    true,
			},
			"rate": schema.Int64Attribute{
				Description: "The maximum rate (in kbps) for storm control traffic. Valid values are greater than 0 and less than the maximum rate.",
				Optional:    true,
			},
		},
	}
}

// Configure assigns the SDK client from provider configuration.
func (r *stormControlResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		client, ok := req.ProviderData.(*sdk.HRUIClient)
		if !ok {
			resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *sdk.HRUIClient")
			return
		}
		r.client = client
	}
}

// Create a new storm control configuration.
func (r *stormControlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data stormControlModel
	diags := req.Plan.Get(ctx, &data)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	maxRate, err := r.client.GetPortMaxRate(int(data.Port.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch maximum rate", err.Error())
		return
	}

	if err := validateStormControlRate(data.Rate.ValueInt64(), maxRate); err != nil {
		resp.Diagnostics.AddError("Invalid Rate for Storm Control", err.Error())
		return
	}

	err = r.client.SetStormControlConfig(
		data.StormType.ValueString(),
		[]int{int(data.Port.ValueInt64())},
		data.State.ValueBool(),
		toIntPointer(data.Rate),
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure storm control", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read the current storm control configuration.
func (r *stormControlResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state stormControlModel
	diags := req.State.Get(ctx, &state)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	config, err := r.client.GetStormControlStatus()
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch storm control status", err.Error())
		return
	}

	maxRate, err := r.client.GetPortMaxRate(int(state.Port.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch max rate for port", err.Error())
		return
	}

	var matchingRate *int
	isEntryFound := false

	for _, entry := range config.Entries {
		if entry.Port == int(state.Port.ValueInt64()) {
			isEntryFound = true
			switch strings.ToLower(state.StormType.ValueString()) {
			case stormTypeBroadcast:
				matchingRate = entry.BroadcastRateKbps
			case stormTypeKnownMulticast:
				matchingRate = entry.KnownMulticastRateKbps
			case stormTypeUnknownUnicast:
				matchingRate = entry.UnknownUnicastRateKbps
			case stormTypeUnknownMulticast:
				matchingRate = entry.UnknownMulticastRateKbps
			}
			break
		}
	}

	// Check if we found a matching entry and handle nil matchingRate
	if !isEntryFound || matchingRate == nil || isStormControlDisabled(matchingRate, maxRate) {
		state.State = types.BoolValue(false)
		state.Rate = types.Int64Null()
	} else {
		state.State = types.BoolValue(true)
		state.Rate = types.Int64Value(int64(*matchingRate))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies an existing storm control configuration.
func (r *stormControlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan stormControlModel
	diags := req.Plan.Get(ctx, &plan)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	maxRate, err := r.client.GetPortMaxRate(int(plan.Port.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch maximum rate", err.Error())
		return
	}

	if err := validateStormControlRate(plan.Rate.ValueInt64(), maxRate); err != nil {
		resp.Diagnostics.AddError("Invalid Rate for Storm Control", err.Error())
		return
	}

	err = r.client.SetStormControlConfig(
		plan.StormType.ValueString(),
		[]int{int(plan.Port.ValueInt64())},
		plan.State.ValueBool(),
		toIntPointer(plan.Rate),
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update storm control", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete disables storm control for the given port and storm type.
func (r *stormControlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state stormControlModel
	diags := req.State.Get(ctx, &state)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	err := r.client.SetStormControlConfig(
		state.StormType.ValueString(),
		[]int{int(state.Port.ValueInt64())},
		false, // Disable storm control
		nil,   // Reset to default (rate not provided)
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed to reset storm control configuration", err.Error())
	}
}

// Validate Storm Control Rate.
func validateStormControlRate(rate int64, maxRate int64) error {
	if rate == 0 || rate == maxRate {
		return fmt.Errorf(
			"The provided rate (%d) is invalid as it is equivalent to 'disabled'. "+
				"Setting rate to 0 or the maximum rate (%d) effectively disables storm control. "+
				"Please provide a valid rate.",
			rate, maxRate,
		)
	}
	return nil
}

func isStormControlDisabled(rate *int, maxRate int64) bool {
	return rate == nil || int64(*rate) == 0 || int64(*rate) == maxRate
}

func toIntPointer(value types.Int64) *int64 {
	if !value.IsNull() {
		v := value.ValueInt64()
		return &v
	}
	return nil
}
