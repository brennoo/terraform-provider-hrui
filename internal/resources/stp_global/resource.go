package stp_global

import (
	"context"
	"fmt"
	"time"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure that the hruiSTPGlobal implements the Terraform resource interface.
var (
	_ resource.Resource              = &stpGlobalResource{}
	_ resource.ResourceWithConfigure = &stpGlobalResource{}
)

// stpGlobalResource implements the resource for STP Global configuration.
type stpGlobalResource struct {
	client *sdk.HRUIClient
}

// NewResource is a constructor for the STP Global resource.
func NewResource() resource.Resource {
	return &stpGlobalResource{}
}

// Metadata provides the resource name for registration.
func (r *stpGlobalResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stp_global"
}

// Schema defines the attributes for STP Global configuration.
func (r *stpGlobalResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages STP Global settings configuration.",
		Attributes: map[string]schema.Attribute{
			"stp_status": schema.StringAttribute{
				Description: "Specifies whether STP is enabled or disabled. This is read-only.",
				Computed:    true,
			},
			"force_version": schema.StringAttribute{
				Description: "Specifies whether to use STP ('STP') or RSTP ('RSTP').",
				Required:    true,
			},
			"priority": schema.Int64Attribute{
				Description: "The bridge priority for the STP instance.",
				Required:    true,
			},
			"max_age": schema.Int64Attribute{
				Description: "Maximum age for STP information before it’s discarded (in seconds).",
				Required:    true,
			},
			"hello_time": schema.Int64Attribute{
				Description: "Time interval (in seconds) between Hello messages.",
				Required:    true,
			},
			"forward_delay": schema.Int64Attribute{
				Description: "Forward delay (in seconds).",
				Required:    true,
			},
			"root_mac": schema.StringAttribute{
				Description: "Root bridge MAC address (read-only).",
				Computed:    true,
			},
			"root_path_cost": schema.Int64Attribute{
				Description: "Root path cost (read-only).",
				Computed:    true,
			},
			"root_port": schema.StringAttribute{
				Description: "Root port (read-only).",
				Computed:    true,
			},
		},
	}
}

// Configure binds the provider's client to this resource.
func (r *stpGlobalResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*sdk.HRUIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.HRUIClient but got %T. Please report this issue.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create provisions the STP Global settings using the provider.
func (r *stpGlobalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Extract the desired configuration from the Terraform plan
	var plan stpGlobalModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map the Terraform model to the backend API model
	stpSettings := sdk.STPGlobalSettings{
		ForceVersion: plan.ForceVersion.ValueString(),
		Priority:     int(plan.Priority.ValueInt64()),
		MaxAge:       int(plan.MaxAge.ValueInt64()),
		HelloTime:    int(plan.HelloTime.ValueInt64()),
		ForwardDelay: int(plan.ForwardDelay.ValueInt64()),
	}

	// Call the backend API to create/configure the STP settings
	err := r.client.UpdateSTPSettings(&stpSettings)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Update STP Settings",
			fmt.Sprintf("An error occurred while updating the STP settings: %v", err),
		)
		return
	}

	// Wait for the backend to reflect the changes
	time.Sleep(5 * time.Second)

	// Fetch the updated state from the backend
	stpFromBackend, err := r.client.GetSTPSettings()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read STP Settings",
			fmt.Sprintf("An error occurred while reading the STP settings from the backend: %v", err),
		)
		return
	}

	// Map the backend data into the Terraform state
	var state stpGlobalModel
	state.ForceVersion = types.StringValue(stpFromBackend.ForceVersion)
	state.Priority = types.Int64Value(int64(stpFromBackend.Priority))
	state.MaxAge = types.Int64Value(int64(stpFromBackend.MaxAge))
	state.HelloTime = types.Int64Value(int64(stpFromBackend.HelloTime))
	state.ForwardDelay = types.Int64Value(int64(stpFromBackend.ForwardDelay))
	state.STPStatus = types.StringValue(stpFromBackend.STPStatus)

	// Map non-duplicate computed fields
	state.RootMAC = types.StringValue(stpFromBackend.RootMAC)
	state.RootPathCost = types.Int64Value(int64(stpFromBackend.RootPathCost))
	state.RootPort = types.StringValue(stpFromBackend.RootPort)

	// Save the state back to Terraform
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Read fetches the current state for hrui_stp_global
func (r *stpGlobalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state stpGlobalModel

	// Retrieve the current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch the STP settings from the backend
	stpFromBackend, err := r.client.GetSTPSettings()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading STP Settings",
			fmt.Sprintf("Could not retrieve STP settings from the backend: %v", err),
		)
		return
	}

	// Map the backend data into the Terraform state
	state.ForceVersion = types.StringValue(stpFromBackend.ForceVersion)
	state.Priority = types.Int64Value(int64(stpFromBackend.Priority))
	state.MaxAge = types.Int64Value(int64(stpFromBackend.MaxAge))
	state.HelloTime = types.Int64Value(int64(stpFromBackend.HelloTime))
	state.ForwardDelay = types.Int64Value(int64(stpFromBackend.ForwardDelay))
	state.STPStatus = types.StringValue(stpFromBackend.STPStatus)

	// Map computed fields from the backend
	state.RootMAC = types.StringValue(stpFromBackend.RootMAC)
	state.RootPathCost = types.Int64Value(int64(stpFromBackend.RootPathCost))
	state.RootPort = types.StringValue(stpFromBackend.RootPort)

	// Save the state back to Terraform
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update modifies the STP global settings.
func (r *stpGlobalResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Extract the desired plan from the Terraform configuration
	var plan stpGlobalModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map the Terraform model to the backend API model
	stpSettings := sdk.STPGlobalSettings{
		ForceVersion: plan.ForceVersion.ValueString(),
		Priority:     int(plan.Priority.ValueInt64()),
		MaxAge:       int(plan.MaxAge.ValueInt64()),
		HelloTime:    int(plan.HelloTime.ValueInt64()),
		ForwardDelay: int(plan.ForwardDelay.ValueInt64()),
	}

	// Call the backend API to update the STP settings
	err := r.client.UpdateSTPSettings(&stpSettings)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Update STP Settings",
			fmt.Sprintf("An error occurred while updating STP settings: %v", err),
		)
		return
	}

	// Wait for the backend to reflect the changes (optional delay)
	time.Sleep(5 * time.Second)

	// Fetch the updated state from the backend
	stpFromBackend, err := r.client.GetSTPSettings()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read Updated STP Settings",
			fmt.Sprintf("An error occurred while reading the updated STP settings from the backend: %v", err),
		)
		return
	}

	// Map the backend data to the Terraform state
	var state stpGlobalModel
	state.ForceVersion = types.StringValue(stpFromBackend.ForceVersion)
	state.Priority = types.Int64Value(int64(stpFromBackend.Priority))
	state.MaxAge = types.Int64Value(int64(stpFromBackend.MaxAge))
	state.HelloTime = types.Int64Value(int64(stpFromBackend.HelloTime))
	state.ForwardDelay = types.Int64Value(int64(stpFromBackend.ForwardDelay))
	state.STPStatus = types.StringValue(stpFromBackend.STPStatus)

	// Map computed fields from the backend
	state.RootMAC = types.StringValue(stpFromBackend.RootMAC)
	state.RootPathCost = types.Int64Value(int64(stpFromBackend.RootPathCost))
	state.RootPort = types.StringValue(stpFromBackend.RootPort)

	// Save the state back to Terraform
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Delete disables STP by setting STPStatus to "Disable" and clears state.
func (r *stpGlobalResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	stpSettings := sdk.STPGlobalSettings{
		STPStatus: "Disable",
	}

	err := r.client.UpdateSTPSettings(&stpSettings)
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting STP Global Resource", fmt.Sprintf("Unable to delete STP Global Resource: %s", err))
	}

	// Since there's no hard deletion, just remove the resource from state.
	resp.State.RemoveResource(ctx)
}
