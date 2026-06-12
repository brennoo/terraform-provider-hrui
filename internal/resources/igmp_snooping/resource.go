package igmp_snooping

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

// Ensure implementation satisfies resource interfaces.
var (
	_ resource.Resource                = &igmpSnoopingResource{}
	_ resource.ResourceWithConfigure   = &igmpSnoopingResource{}
	_ resource.ResourceWithImportState = &igmpSnoopingResource{}
)

// igmpSnoopingResource represents the global IGMP snooping configuration.
type igmpSnoopingResource struct {
	client *sdk.HRUIClient
}

// NewResource creates a new IGMP snooping resource.
func NewResource() resource.Resource {
	return &igmpSnoopingResource{}
}

// Metadata defines the resource type name.
func (r *igmpSnoopingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_igmp_snooping"
}

// Schema defines the configuration schema for the resource.
func (r *igmpSnoopingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages global IGMP Snooping configuration.",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Specifies whether IGMP snooping is enabled or disabled globally.",
				Required:    true,
			},
		},
	}
}

// Configure sets the client from the provider data.
func (r *igmpSnoopingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Create enables or disables global IGMP Snooping.
func (r *igmpSnoopingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Creating IGMP snooping settings")

	var plan igmpSnoopingModel

	// Retrieve the plan state
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the global IGMP snooping status
	if plan.Enabled.ValueBool() {
		if err := r.client.EnableIGMPSnooping(ctx); err != nil {
			resp.Diagnostics.AddError(
				"Error Creating IGMP Snooping",
				fmt.Sprintf("Failed to enable global IGMP snooping: %s", err.Error()),
			)
			return
		}
	} else {
		if err := r.client.DisableIGMPSnooping(ctx); err != nil {
			resp.Diagnostics.AddError(
				"Error Creating IGMP Snooping",
				fmt.Sprintf("Failed to disable global IGMP snooping: %s", err.Error()),
			)
			return
		}
	}

	// Set the resource state in Terraform
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	tflog.Debug(ctx, "IGMP snooping settings created")
}

// Read synchronizes the Terraform state with the current global IGMP snooping configuration.
func (r *igmpSnoopingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Reading IGMP snooping settings")

	var state igmpSnoopingModel

	// Retrieve the current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Query the global IGMP snooping configuration using the SDK
	config, err := r.client.FetchIGMPConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading IGMP Snooping",
			fmt.Sprintf("Failed to fetch global IGMP snooping configuration: %s", err.Error()),
		)
		return
	}

	// Update the resource state with the current configuration
	state.Enabled = types.BoolValue(config.Enabled)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	tflog.Debug(ctx, "IGMP snooping settings read")
}

// Update modifies the global IGMP snooping configuration.
func (r *igmpSnoopingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Updating IGMP snooping settings")

	var plan igmpSnoopingModel

	// Retrieve the updated plan
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the configuration based on `enabled` state
	if plan.Enabled.ValueBool() {
		if err := r.client.EnableIGMPSnooping(ctx); err != nil {
			resp.Diagnostics.AddError(
				"Error Updating IGMP Snooping",
				fmt.Sprintf("Failed to enable global IGMP snooping: %s", err.Error()),
			)
			return
		}
	} else {
		if err := r.client.DisableIGMPSnooping(ctx); err != nil {
			resp.Diagnostics.AddError(
				"Error Updating IGMP Snooping",
				fmt.Sprintf("Failed to disable global IGMP snooping: %s", err.Error()),
			)
			return
		}
	}

	// Update the resource state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	tflog.Debug(ctx, "IGMP snooping settings updated")
}

// Delete disables global IGMP Snooping.
func (r *igmpSnoopingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Deleting IGMP snooping settings")

	// Disable global IGMP snooping
	if err := r.client.DisableIGMPSnooping(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting IGMP Snooping",
			fmt.Sprintf("Failed to disable global IGMP snooping: %s", err.Error()),
		)
		return
	}

	// Remove the resource from the Terraform state
	resp.State.RemoveResource(ctx)
}

// ImportState imports an existing IGMP Snooping resource by fetching the current state.
func (r *igmpSnoopingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Importing IGMP snooping settings", map[string]any{"id": req.ID})

	config, err := r.client.FetchIGMPConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing IGMP Snooping Settings", fmt.Sprintf("Unable to import IGMP snooping settings: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &igmpSnoopingModel{Enabled: types.BoolValue(config.Enabled)})...)
}
