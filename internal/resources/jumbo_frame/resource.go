package jumbo_frame

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
	_ resource.Resource                = &jumboFrameResource{}
	_ resource.ResourceWithConfigure   = &jumboFrameResource{}
	_ resource.ResourceWithImportState = &jumboFrameResource{}
)

// jumboFrameResource is the implementation of the Jumbo Frame Terraform resource.
type jumboFrameResource struct {
	client *sdk.HRUIClient
}

// NewResource creates a new instance of the Jumbo Frame resource.
func NewResource() resource.Resource {
	return &jumboFrameResource{}
}

// Metadata sets the resource name.
func (r *jumboFrameResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jumbo_frame"
}

// Schema defines the schema for the Jumbo Frame resource.
func (r *jumboFrameResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Configures jumbo frame settings.",
		Attributes: map[string]schema.Attribute{
			"size": schema.Int64Attribute{
				Required:    true,
				Description: "Size of the Jumbo Frame in bytes. Valid options are 1522, 1536, 1552, 9216, and 16383.",
			},
		},
	}
}

// Configure assigns the provider-configured client to the resource.
func (r *jumboFrameResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Create sets the initial Jumbo Frame size.
func (r *jumboFrameResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Creating jumbo frame settings")

	// Parse the plan (input configuration from the user)
	var plan jumboFrameModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appliedSize, err := r.client.SetJumboFrame(ctx, int(plan.Size.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Jumbo Frame",
			fmt.Sprintf("Failed to set Jumbo Frame size: %s", err),
		)
		return
	}

	var state jumboFrameModel
	state.Size = types.Int64Value(int64(appliedSize))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	tflog.Debug(ctx, "Jumbo frame settings created")
}

// Read retrieves the current Jumbo Frame size from the device and updates the state.
func (r *jumboFrameResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Reading jumbo frame settings")

	// Parse the state (current resource state in Terraform)
	var state jumboFrameModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call the SDK to get the current Jumbo Frame size
	jumboFrame, err := r.client.GetJumboFrame(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Jumbo Frame",
			fmt.Sprintf("Failed to read Jumbo Frame size: %s", err),
		)
		return
	}

	// Update the state with the current value
	state.Size = types.Int64Value(int64(jumboFrame.FrameSize))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	tflog.Debug(ctx, "Jumbo frame settings read")
}

// Update changes the Jumbo Frame size to the new value in the plan.
func (r *jumboFrameResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Updating jumbo frame settings")

	// Parse the plan (new configuration)
	var plan jumboFrameModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appliedSize, err := r.client.SetJumboFrame(ctx, int(plan.Size.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Jumbo Frame",
			fmt.Sprintf("Failed to update Jumbo Frame size: %s", err),
		)
		return
	}

	var state jumboFrameModel
	state.Size = types.Int64Value(int64(appliedSize))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	tflog.Debug(ctx, "Jumbo frame settings updated")
}

// Delete resets the Jumbo Frame size to its default (if applicable).
func (r *jumboFrameResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Deleting jumbo frame settings")

	defaultFrameSize := 16383

	// Call the SDK to reset the Jumbo Frame size to the default
	_, err := r.client.SetJumboFrame(ctx, defaultFrameSize)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Jumbo Frame",
			fmt.Sprintf("Failed to reset Jumbo Frame size to default: %s", err),
		)
	}
}

// ImportState imports an existing Jumbo Frame resource by fetching the current state.
func (r *jumboFrameResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Importing jumbo frame settings", map[string]any{"id": req.ID})

	jumboFrame, err := r.client.GetJumboFrame(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing Jumbo Frame Settings", fmt.Sprintf("Unable to import jumbo frame settings: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &jumboFrameModel{Size: types.Int64Value(int64(jumboFrame.FrameSize))})...)
}
