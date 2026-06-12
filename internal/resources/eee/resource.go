package eee

import (
	"context"
	"fmt"

	"github.com/brennoo/terraform-provider-hrui/internal/providerutil"
	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the required interfaces.
var (
	_ resource.Resource                = &eeeResource{}
	_ resource.ResourceWithConfigure   = &eeeResource{}
	_ resource.ResourceWithImportState = &eeeResource{}
)

// eeeResource is the implementation of the EEE Terraform resource.
type eeeResource struct {
	client *sdk.HRUIClient
}

// NewResource creates a new instance of the EEE resource.
func NewResource() resource.Resource {
	return &eeeResource{}
}

// Metadata sets the resource name for Terraform.
func (r *eeeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eee"
}

// Schema defines the schema for the EEE resource.
func (r *eeeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the Energy Efficient Ethernet (EEE) settings.",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether EEE is enabled (`true`) or disabled (`false`).",
			},
		},
	}
}

// Configure assigns the provider-configured client to the resource.
func (r *eeeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Create sets the initial EEE state.
func (r *eeeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Creating EEE settings")

	// Parse the plan (input configuration from the user)
	var plan eeeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call the SDK to set the EEE status
	err := r.client.SetEEE(ctx, plan.Enabled.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating EEE Settings",
			fmt.Sprintf("Failed to set EEE status: %s", err),
		)
		return
	}

	// Set the state equal to the plan, as the actual value is expected to match the user's input
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	tflog.Debug(ctx, "EEE settings created")
}

// Read retrieves the current EEE status from the device and updates the state.
func (r *eeeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Reading EEE settings")

	// Parse the current state
	var state eeeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call the SDK to get the current EEE status
	enabled, err := r.client.GetEEE(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading EEE Settings",
			fmt.Sprintf("Failed to read EEE status: %s", err),
		)
		return
	}

	// Update the state with the current value
	state.Enabled = types.BoolValue(enabled)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	tflog.Debug(ctx, "EEE settings read")
}

// Update changes the EEE status to the new value in the plan.
func (r *eeeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Updating EEE settings")

	// Parse the plan (new configuration)
	var plan eeeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call the SDK to set the new EEE status
	err := r.client.SetEEE(ctx, plan.Enabled.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating EEE Settings",
			fmt.Sprintf("Failed to update EEE status: %s", err),
		)
		return
	}

	// Set the state equal to the plan, as the actual value is expected to match the user's input
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	tflog.Debug(ctx, "EEE settings updated")
}

// Delete disables EEE by setting the property to its default value (`false`).
func (r *eeeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Deleting EEE settings")

	// Call the SDK to disable EEE (default is off)
	err := r.client.SetEEE(ctx, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting EEE Settings",
			fmt.Sprintf("Failed to disable EEE: %s", err),
		)
	}
}

// ImportState imports an existing EEE resource by fetching the current state.
func (r *eeeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Importing EEE settings", map[string]any{"id": req.ID})

	enabled, err := r.client.GetEEE(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing EEE Settings", fmt.Sprintf("Unable to import EEE settings: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &eeeModel{Enabled: types.BoolValue(enabled)})...)
}
