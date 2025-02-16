package port_settings

import (
	"context"
	"fmt"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &portSettingResource{}
	_ resource.ResourceWithConfigure   = &portSettingResource{}
	_ resource.ResourceWithImportState = &portSettingResource{}
)

// portSettingResource is the resource implementation.
type portSettingResource struct {
	client *sdk.HRUIClient
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &portSettingResource{}
}

func (r *portSettingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages port settings.",
		Attributes: map[string]schema.Attribute{
			"port": schema.StringAttribute{
				Required:    true,
				Description: "The port name or ID (e.g., 'Port 1', 'Trunk1').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the port is enabled.",
			},
			"speed": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Speed and duplex settings of the port.",
				Attributes: map[string]schema.Attribute{
					"config": schema.StringAttribute{
						Optional:    true,
						Description: "Configured speed and duplex mode.",
					},
					"actual": schema.StringAttribute{
						Computed:    true,
						Description: "Actual speed and duplex mode returned by the system.",
					},
				},
			},
			"flow_control": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Flow control configuration of the port.",
				Attributes: map[string]schema.Attribute{
					"config": schema.StringAttribute{
						Optional:    true,
						Description: "Configured flow control setting.",
					},
					"actual": schema.StringAttribute{
						Computed:    true,
						Description: "Actual flow control setting returned by the system.",
					},
				},
			},
		},
	}
}

// Configure assigns the provider-configured client to the resource.
func (r *portSettingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Metadata sets the resource name.
func (r *portSettingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_port_settings"
}

// Create creates the port settings in the HRUI system.
func (r *portSettingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan portSettingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Apply configuration using the SDK
	port := &sdk.Port{
		ID:                plan.Port.ValueString(),
		State:             boolToInt(plan.Enabled.ValueBool()),
		SpeedDuplexConfig: plan.Speed.Config.ValueString(),
		FlowControlConfig: plan.FlowControl.Config.ValueString(),
	}

	// Call API to configure the port
	finalPort, err := r.client.ConfigurePort(port)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to create HRUI port settings: %s", err))
		return
	}

	// Populate the plan with actual values fetched after applying
	plan.Speed = &portSettingSpeed{
		Config: plan.Speed.Config,
		Actual: types.StringValue(finalPort.SpeedDuplexActual),
	}
	plan.FlowControl = &portSettingFlowControl{
		Config: plan.FlowControl.Config,
		Actual: types.StringValue(finalPort.FlowControlActual),
	}
	plan.Enabled = types.BoolValue(finalPort.State == 1)

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read fetches the current port settings and updates state.
func (r *portSettingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state portSettingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch the current data for the port from the switch
	port, err := r.client.GetPort(state.Port.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to read HRUI port settings: %s", err))
		return
	}

	// Extract values returned from the device
	deviceSpeedConfig := port.SpeedDuplexConfig
	deviceFlowControlConfig := port.FlowControlConfig

	// Update the state with both the actual and configured values from the device
	state.Enabled = types.BoolValue(port.State == 1)
	state.Speed = &portSettingSpeed{
		Config: types.StringValue(deviceSpeedConfig),
		Actual: types.StringValue(port.SpeedDuplexActual),
	}
	state.FlowControl = &portSettingFlowControl{
		Config: types.StringValue(deviceFlowControlConfig),
		Actual: types.StringValue(port.FlowControlActual),
	}

	// Save the updated state back to Terraform
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update applies changes to the port settings.
func (r *portSettingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan portSettingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Apply updated configuration using `ConfigurePort`
	port := &sdk.Port{
		ID:                plan.Port.ValueString(),
		State:             boolToInt(plan.Enabled.ValueBool()),
		SpeedDuplexConfig: plan.Speed.Config.ValueString(),
		FlowControlConfig: plan.FlowControl.Config.ValueString(),
	}

	finalPort, err := r.client.ConfigurePort(port)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update HRUI port settings, got error: %s", err))
		return
	}

	// Update plan with actual values fetched after applying
	plan.Speed = &portSettingSpeed{
		Config: plan.Speed.Config,
		Actual: types.StringValue(finalPort.SpeedDuplexActual),
	}
	plan.FlowControl = &portSettingFlowControl{
		Config: plan.FlowControl.Config,
		Actual: types.StringValue(finalPort.FlowControlActual),
	}
	plan.Enabled = types.BoolValue(finalPort.State == 1)

	// Save updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *portSettingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state portSettingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Default settings to restore
	defaultPort := &sdk.Port{
		ID:                state.Port.ValueString(),
		State:             1,
		SpeedDuplexConfig: "Auto",
		FlowControlConfig: "Off",
	}

	// Reset the port to defaults
	_, err := r.client.ConfigurePort(defaultPort)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Reset Port to Default",
			fmt.Sprintf("Unable to reset port '%s' to its default settings: %s", state.Port.ValueString(), err),
		)
		return
	}

	// Remove the resource from Terraform's state
	resp.State.RemoveResource(ctx)
}

// ImportState maps the imported ID to the `port` field.
func (r *portSettingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("port"), req.ID)...) // Use `port` as the unique ID
}

// Helper function to convert a bool to an int.
func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
