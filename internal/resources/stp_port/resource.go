package stp_port

import (
	"context"
	"fmt"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the resource implements the Terraform interfaces.
var (
	_ resource.Resource              = &stpPortResource{}
	_ resource.ResourceWithConfigure = &stpPortResource{}
)

type stpPortResource struct {
	client *sdk.HRUIClient
}

// NewResource creates a new instance of the resource.
func NewResource() resource.Resource {
	return &stpPortResource{}
}

// Metadata sets the resource type name in the provider.
func (r *stpPortResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stp_port"
}

// Schema returns the schema definition for the resource.
func (r *stpPortResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages STP port settings.",
		Attributes: map[string]schema.Attribute{
			"port": schema.StringAttribute{
				Description: "The port name to configure STP. Changing this will recreate the resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"path_cost": schema.Int64Attribute{
				Description: "The desired STP path cost for the port.",
				Required:    true,
			},
			"priority": schema.Int64Attribute{
				Description: "The STP port priority, affecting the port's contribution to the spanning-tree root bridge decision.",
				Required:    true,
			},
			"p2p": schema.StringAttribute{
				Description: `Point-to-point (P2P) configuration:

- **Valid values:** 'Auto', 'True', 'False'.
- When set to 'Auto', the system automatically determines the P2P configuration based on the port's operation.
- **Note:** Due to a known limitation in firmware version 1.9, changes to this attribute do not take effect.`,
				Computed: true,
			},
			"edge": schema.StringAttribute{
				Description: "Edge port setting, used to designate whether the port connects to an end device ('True') or another switch ('False').",
				Required:    true,
			},
			"state": schema.StringAttribute{
				Description: "Reflects the current spanning-tree protocol (STP) state of the port (e.g., 'Forwarding', 'Blocked').",
				Computed:    true,
			},
			"role": schema.StringAttribute{
				Description: "Current role of the port in the STP topology (e.g., 'Designated', 'Root').",
				Computed:    true,
			},
		},
	}
}

// Configure sets up the client for the resource.
func (r *stpPortResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*sdk.HRUIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.HRUIClient but got %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create provisions the STP port settings and synchronizes the Terraform state.
func (r *stpPortResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan stpPortModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.SetSTPPortSettings(ctx,
		plan.Port.ValueString(),
		int(plan.PathCost.ValueInt64()),
		int(plan.Priority.ValueInt64()),
		plan.P2P.ValueString(),
		plan.Edge.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating STP Port",
			fmt.Sprintf("Failed to configure STP settings for port %s: %v", plan.Port.ValueString(), err),
		)
		return
	}

	// Fetch and set the state
	r.readPortState(ctx, plan.Port.ValueString(), &plan, &resp.Diagnostics)

	// Update the state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Read fetches the current state of the STP port from the backend.
func (r *stpPortResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state stpPortModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch and set the updated state
	r.readPortState(ctx, state.Port.ValueString(), &state, &resp.Diagnostics)

	// Write the updated state back to Terraform
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update modifies the STP port settings in the backend.
func (r *stpPortResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan stpPortModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.SetSTPPortSettings(ctx,
		plan.Port.ValueString(),
		int(plan.PathCost.ValueInt64()),
		int(plan.Priority.ValueInt64()),
		plan.P2P.ValueString(),
		plan.Edge.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating STP Port",
			fmt.Sprintf("Failed to update STP settings for port %s: %v", plan.Port.ValueString(), err),
		)
		return
	}

	// Fetch and set the updated state
	r.readPortState(ctx, plan.Port.ValueString(), &plan, &resp.Diagnostics)

	// Update the state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete resets the STP port settings and removes the resource from the state.
func (r *stpPortResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state stpPortModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.SetSTPPortSettings(ctx,
		state.Port.ValueString(),
		20000, // Default path cost
		128,   // Default priority
		"Auto",
		"False",
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Resetting STP Port",
			fmt.Sprintf("Failed to reset STP settings for port %s: %v", state.Port.ValueString(), err),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}

// Helper function to fetch the port state and set it in the model.
func (r *stpPortResource) readPortState(ctx context.Context, portName string, model *stpPortModel, diagnostics *diag.Diagnostics) {
	stpPort, err := r.client.GetSTPPort(ctx, portName)
	if err != nil {
		diagnostics.AddError(
			"Error Reading STP Port",
			fmt.Sprintf("Failed to retrieve STP settings for port %s: %v", portName, err),
		)
		return
	}

	model.State = types.StringValue(stpPort.State)
	model.Role = types.StringValue(stpPort.Role)
	model.PathCost = types.Int64Value(int64(stpPort.PathCostConfig))
	model.Priority = types.Int64Value(int64(stpPort.Priority))
	model.P2P = types.StringValue(stpPort.P2PConfig)
	model.Edge = types.StringValue(stpPort.EdgeConfig)
}
