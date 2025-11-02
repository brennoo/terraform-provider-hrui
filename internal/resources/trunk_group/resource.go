package trunk_group

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure implementation satisfies the resource.Resource interface.
var _ resource.Resource = &trunkGroupResource{}

// trunkGroupResource manages trunk groups on the HRUI switch.
type trunkGroupResource struct {
	client *sdk.HRUIClient
}

// NewResource creates a new instance of the trunk group resource.
func NewResource() resource.Resource {
	return &trunkGroupResource{}
}

// Metadata sets the resource type name.
func (r *trunkGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trunk_group"
}

// Schema defines the schema for the resource.
func (r *trunkGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages trunk group settings.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The trunk group ID. Must match one of the available trunk group IDs on the device.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Type of the trunk group ('static' or 'LACP').",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("static", "LACP"),
				},
			},
			"ports": schema.ListAttribute{
				Description: "List of ports in the trunk group (1-indexed: Port 1 = 1, Port 2 = 2, etc.).",
				ElementType: types.Int64Type,
				Required:    true,
			},
		},
	}
}

// Configure assigns the SDK client from provider configuration.
func (r *trunkGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		client, ok := req.ProviderData.(*sdk.HRUIClient)
		if !ok {
			resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *sdk.HRUIClient")
			return
		}
		r.client = client
	}
}

// Create a new trunk group.
func (r *trunkGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data trunkGroupModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate if the trunk ID is part of the available trunks.
	availableTrunks, err := r.client.ListAvailableTrunks(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list available trunks", err.Error())
		return
	}

	isValidID := false
	for _, trunk := range availableTrunks {
		if trunk.ID == int(data.ID.ValueInt64()) {
			isValidID = true
			break
		}
	}

	if !isValidID {
		resp.Diagnostics.AddError(
			"Invalid Trunk ID",
			fmt.Sprintf("Trunk ID %d is not available. Please choose from the available Trunk IDs.", data.ID.ValueInt64()),
		)
		return
	}

	// Get the list of ports from the input
	var ports []int64
	diags = data.Ports.ElementsAs(ctx, &ports, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check for port mirroring conflicts and remove port mirroring if needed
	portMirror, err := r.client.GetPortMirror(ctx)
	if err == nil && portMirror != nil {
		// Check if any trunk port is being used in port mirroring
		needsCleanup := false
		mirroringPortID, _ := r.client.GetPortByName(ctx, portMirror.MirroringPort)

		for _, port := range ports {
			portID := int(port)
			// Check if this port is the mirroring port
			if portID == mirroringPortID {
				needsCleanup = true
				break
			}
			// Check if this port is in the mirrored port list
			if strings.Contains(portMirror.MirroredPort, fmt.Sprintf("%d", portID)) ||
				strings.Contains(portMirror.MirroredPort, fmt.Sprintf("Port %d", portID)) {
				needsCleanup = true
				break
			}
		}

		if needsCleanup {
			// Delete port mirroring configuration before creating trunk
			if err := r.client.DeletePortMirror(ctx); err != nil {
				resp.Diagnostics.AddError(
					"Failed to remove port mirroring conflict",
					fmt.Sprintf("Port mirroring is configured and conflicts with trunk ports. Failed to remove: %s", err),
				)
				return
			}
		}
	}

	// Convert to SDK-compatible types
	sdkPorts := make([]int, len(ports))
	for i, port := range ports {
		sdkPorts[i] = int(port)
	}

	// Call the SDK to create the trunk group
	err = r.client.ConfigureTrunk(ctx, &sdk.TrunkConfig{
		ID:    int(data.ID.ValueInt64()),
		Type:  data.Type.ValueString(),
		Ports: sdkPorts,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Trunk Group", err.Error())
		return
	}

	// Wait for the device to reflect the changes
	time.Sleep(2 * time.Second)

	// Read back from the device to ensure state matches what was actually applied
	trunkGroup, err := r.client.GetTrunk(ctx, int(data.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch Trunk Group details after creation", err.Error())
		return
	}

	// Update state with values from device
	state := trunkGroupModel{
		ID:   data.ID,
		Type: types.StringValue(trunkGroup.Type),
	}

	// Convert ports from SDK (1-indexed) to Terraform list
	trunkPorts := make([]int64, len(trunkGroup.Ports))
	for i, port := range trunkGroup.Ports {
		trunkPorts[i] = int64(port)
	}
	state.Ports, diags = types.ListValueFrom(ctx, types.Int64Type, trunkPorts)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save the state based on what was read from the device
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Read retrieves the trunk group details.
func (r *trunkGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state trunkGroupModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch trunk group details from the SDK
	trunkGroup, err := r.client.GetTrunk(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch Trunk Group details", err.Error())
		return
	}

	// Update state with fetched data
	state.Type = types.StringValue(trunkGroup.Type)
	state.Ports, diags = types.ListValueFrom(ctx, types.Int64Type, trunkGroup.Ports)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save the updated state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update modifies an existing trunk group.
func (r *trunkGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan trunkGroupModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract ports from the plan
	var ports []int64
	diags = plan.Ports.ElementsAs(ctx, &ports, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check for port mirroring conflicts and remove port mirroring if needed
	portMirror, err := r.client.GetPortMirror(ctx)
	if err == nil && portMirror != nil {
		// Check if any trunk port is being used in port mirroring
		needsCleanup := false
		mirroringPortID, _ := r.client.GetPortByName(ctx, portMirror.MirroringPort)

		for _, port := range ports {
			portID := int(port)
			// Check if this port is the mirroring port
			if portID == mirroringPortID {
				needsCleanup = true
				break
			}
			// Check if this port is in the mirrored port list
			if strings.Contains(portMirror.MirroredPort, fmt.Sprintf("%d", portID)) ||
				strings.Contains(portMirror.MirroredPort, fmt.Sprintf("Port %d", portID)) {
				needsCleanup = true
				break
			}
		}

		if needsCleanup {
			// Delete port mirroring configuration before updating trunk
			if err := r.client.DeletePortMirror(ctx); err != nil {
				resp.Diagnostics.AddError(
					"Failed to remove port mirroring conflict",
					fmt.Sprintf("Port mirroring is configured and conflicts with trunk ports. Failed to remove: %s", err),
				)
				return
			}
		}
	}

	// Convert to SDK-compatible types
	sdkPorts := make([]int, len(ports))
	for i, port := range ports {
		sdkPorts[i] = int(port)
	}

	// Call the SDK to update the trunk group
	err = r.client.ConfigureTrunk(ctx, &sdk.TrunkConfig{
		ID:    int(plan.ID.ValueInt64()),
		Type:  plan.Type.ValueString(),
		Ports: sdkPorts,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to update Trunk Group", err.Error())
		return
	}

	// Wait for the device to reflect the changes
	time.Sleep(2 * time.Second)

	// Read back from the device to ensure state matches what was actually applied
	trunkGroup, err := r.client.GetTrunk(ctx, int(plan.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch Trunk Group details after update", err.Error())
		return
	}

	// Update state with values from device
	state := trunkGroupModel{
		ID:   plan.ID,
		Type: types.StringValue(trunkGroup.Type),
	}

	// Convert ports from SDK (1-indexed) to Terraform list
	trunkPorts := make([]int64, len(trunkGroup.Ports))
	for i, port := range trunkGroup.Ports {
		trunkPorts[i] = int64(port)
	}
	state.Ports, diags = types.ListValueFrom(ctx, types.Int64Type, trunkPorts)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save the updated state based on what was read from the device
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Delete removes a trunk group.
func (r *trunkGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state trunkGroupModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// SDK Call to delete the trunk group
	err := r.client.DeleteTrunk(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete Trunk Group", err.Error())
	}
}
