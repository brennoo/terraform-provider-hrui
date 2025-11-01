package jumbo_frame

import (
	"context"
	"fmt"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the required interfaces.
var (
	_ resource.Resource              = &jumboFrameResource{}
	_ resource.ResourceWithConfigure = &jumboFrameResource{}
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
func (r *jumboFrameResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*sdk.HRUIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.HRUIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Create sets the initial Jumbo Frame size.
func (r *jumboFrameResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Parse the plan (input configuration from the user)
	var plan jumboFrameModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call the SDK to set the Jumbo Frame size
	err := r.client.SetJumboFrame(ctx, int(plan.Size.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Failed to set Jumbo Frame size: %s", err),
		)
		return
	}

	// Set the state equal to the plan because the actual value is expected to be what was set.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read retrieves the current Jumbo Frame size from the device and updates the state.
func (r *jumboFrameResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
			"Client Error",
			fmt.Sprintf("Failed to read Jumbo Frame size: %s", err),
		)
		return
	}

	// Update the state with the current value
	state.Size = types.Int64Value(int64(jumboFrame.FrameSize))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update changes the Jumbo Frame size to the new value in the plan.
func (r *jumboFrameResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Parse the plan (new configuration)
	var plan jumboFrameModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call the SDK to set the new Jumbo Frame size
	err := r.client.SetJumboFrame(ctx, int(plan.Size.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Failed to update Jumbo Frame size: %s", err),
		)
		return
	}

	// Set the state equal to the plan because the actual value is expected to be what was set
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete resets the Jumbo Frame size to its default (if applicable).
func (r *jumboFrameResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defaultFrameSize := 16383

	// Call the SDK to reset the Jumbo Frame size to the default
	err := r.client.SetJumboFrame(ctx, defaultFrameSize)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Failed to reset Jumbo Frame size to default: %s", err),
		)
	}
}
