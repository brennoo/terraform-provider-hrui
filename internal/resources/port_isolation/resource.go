package port_isolation

import (
	"context"
	"fmt"

	"github.com/brennoo/terraform-provider-hrui/internal/sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// portIsolationResource defines the resource implementation.
type portIsolationResource struct {
	client *sdk.HRUIClient
}

// NewResource initializes and returns a new resource instance.
func NewResource() resource.Resource {
	return &portIsolationResource{}
}

// Metadata defines the resource metadata.
func (r *portIsolationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_port_isolation"
}

// Schema defines the schema for the resource.
func (r *portIsolationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"port": schema.StringAttribute{
				Required:    true,
				Description: "The port name for which isolation will be configured. Acts as an implicit identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"isolation_list": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "List of isolated ports for the specified port.",
			},
		},
	}
}

// Configure assigns the provider-configured client to the resource.
func (r *portIsolationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Read fetches the current port isolation and updates the state.
func (r *portIsolationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read the state into the model
	var state portIsolationModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch port isolation configuration from the SDK
	portIsolations, err := r.client.GetPortIsolation()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Port Isolation",
			fmt.Sprintf("Failed to fetch port isolation configurations: %s", err.Error()),
		)
		return
	}

	// Find the isolation configuration for the current port
	port := state.Port.ValueString()
	var isolationList []string
	for _, isolation := range portIsolations {
		if isolation.Port == port {
			// Use isolation list from the backend as-is
			isolationList = isolation.IsolationList
			break
		}
	}

	// Update the Terraform state
	state.IsolationList = convertToTerraformList(isolationList)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Create sets up the port isolation using the given plan values.
func (r *portIsolationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read the plan into the model
	var plan portIsolationModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call SDK to configure port isolation
	port := plan.Port.ValueString()
	isolationList := extractStrings(ctx, plan.IsolationList)

	err := r.client.ConfigurePortIsolation(port, isolationList)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Port Isolation",
			fmt.Sprintf("Failed to configure port isolation for port '%s': %s", port, err.Error()),
		)
		return
	}

	// Save the state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Update modifies the port isolation.
func (r *portIsolationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read the plan into the model
	var plan portIsolationModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call SDK to update port isolation
	port := plan.Port.ValueString()
	isolationList := extractStrings(ctx, plan.IsolationList)

	err := r.client.ConfigurePortIsolation(port, isolationList)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Port Isolation",
			fmt.Sprintf("Failed to update port isolation for port '%s': %s", port, err.Error()),
		)
		return
	}

	// Save the updated state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete removes the port isolation.
func (r *portIsolationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Read the state into the model
	var state portIsolationModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Clear the port isolation
	port := state.Port.ValueString()
	err := r.client.DeletePortIsolation(port)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Port Isolation",
			fmt.Sprintf("Failed to delete port isolation for port '%s': %s", port, err.Error()),
		)
		return
	}
}

// extractStrings extracts Go strings from a Terraform List attribute.
func extractStrings(ctx context.Context, list types.List) []string {
	if list.IsUnknown() || list.IsNull() {
		return nil
	}

	var result []string
	list.ElementsAs(ctx, &result, false)
	return result
}

// convertToTerraformList converts a list of strings into a Terraform List attribute.
func convertToTerraformList(data []string) types.List {
	elements := make([]attr.Value, len(data))
	for i, v := range data {
		elements[i] = types.StringValue(v)
	}

	list, diags := types.ListValue(types.StringType, elements)
	if diags.HasError() {
		return types.ListNull(types.StringType)
	}

	return list
}
