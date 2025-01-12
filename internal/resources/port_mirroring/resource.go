package port_mirroring

import (
	"context"
	"fmt"
	"strings"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// portMirroringResource defines the resource implementation.
type portMirroringResource struct {
	client *sdk.HRUIClient
}

// NewResource initializes and returns a new resource instance.
func NewResource() resource.Resource {
	return &portMirroringResource{}
}

// Metadata defines the resource metadata, including the type name.
func (r *portMirroringResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_port_mirroring"
}

// Schema defines the schema for the resource, specifying required and optional attributes.
func (r *portMirroringResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"mirror_direction": schema.StringAttribute{
				Required:    true,
				Description: "The mirroring direction: 'Rx', 'Tx', or 'BOTH'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"mirroring_port": schema.StringAttribute{
				Required:    true,
				Description: "The port performing the mirroring (e.g., 'Port 1').",
			},
			"mirrored_port": schema.StringAttribute{
				Required:    true,
				Description: "The port being mirrored (e.g., 'Port 2').",
			},
		},
	}
}

// Configure assigns the provider-configured client to the resource for making API calls.
func (r *portMirroringResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*sdk.HRUIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *sdk.HRUIClient, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create sets up the port mirroring configuration using the given plan values.
func (r *portMirroringResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan portMirroringModel

	// Extract plan values into the model
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map plan values to the SDK PortMirror struct
	portMirror := &sdk.PortMirror{
		MirrorDirection: plan.MirrorDirection.ValueString(),
		MirroringPort:   plan.MirroringPort.ValueString(),
		MirroredPort:    plan.MirroredPort.ValueString(),
	}

	// Call the SDK to configure port mirroring
	if err := r.client.ConfigurePortMirror(portMirror); err != nil {
		resp.Diagnostics.AddError(
			"Error Configuring Port Mirroring",
			"An error occurred while configuring port mirroring: "+err.Error(),
		)
		return
	}

	// Set the plan as the initial state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read fetches and updates the resource state based on the actual configuration.
func (r *portMirroringResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state portMirroringModel

	// Retrieve the current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the current port mirroring configuration
	portMirror, err := r.client.GetPortMirror()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Port Mirroring",
			"An error occurred while getting the port mirroring configuration: "+err.Error(),
		)
		return
	}

	// If no configuration exists, remove the resource from state
	if portMirror == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state with the retrieved configuration
	state.MirrorDirection = types.StringValue(portMirror.MirrorDirection)
	state.MirroringPort = types.StringValue(portMirror.MirroringPort)
	state.MirroredPort = types.StringValue(portMirror.MirroredPort)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// Update modifies the port mirroring configuration based on the plan.
func (r *portMirroringResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan portMirroringModel

	// Extract plan values into the model
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map plan values to the SDK PortMirror struct
	portMirror := &sdk.PortMirror{
		MirrorDirection: plan.MirrorDirection.ValueString(),
		MirroringPort:   plan.MirroringPort.ValueString(),
		MirroredPort:    plan.MirroredPort.ValueString(),
	}

	// Call the SDK to update port mirroring
	if err := r.client.ConfigurePortMirror(portMirror); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Port Mirroring",
			"An error occurred while updating the port mirroring configuration: "+err.Error(),
		)
		return
	}

	// Update the state with the new configuration
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete removes the port mirroring configuration from the system.
func (r *portMirroringResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Call the SDK to delete the port mirroring configuration
	if err := r.client.DeletePortMirror(); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Port Mirroring",
			"An error occurred while deleting the port mirroring configuration: "+err.Error(),
		)
	}
}

// ImportState allows importing an existing port mirroring configuration by ID.
func (r *portMirroringResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Expect the import ID to be in the format `mirroring_port:mirrored_port`
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected ID in the format `mirroring_port:mirrored_port`.",
		)
		return
	}

	// Fetch the current port mirroring configuration
	portMirror, err := r.client.GetPortMirror()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Port Mirroring",
			"An error occurred while fetching the port mirroring configuration: "+err.Error(),
		)
		return
	}

	// Validate the fetched configuration with the import ID
	if portMirror.MirroringPort != parts[0] || portMirror.MirroredPort != parts[1] {
		resp.Diagnostics.AddError(
			"Port Mirroring Configuration Not Found",
			fmt.Sprintf("No port mirroring configuration found for ID: %s.", req.ID),
		)
		return
	}

	// Set the imported attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("mirror_direction"), portMirror.MirrorDirection)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("mirroring_port"), portMirror.MirroringPort)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("mirrored_port"), portMirror.MirroredPort)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
