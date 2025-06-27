// resource_provider.go - Provider resource implementation
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/keephq/terraform-provider-keep/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &providerResource{}
	_ resource.ResourceWithConfigure   = &providerResource{}
	_ resource.ResourceWithImportState = &providerResource{}
)

// NewProviderResource is a helper function to simplify the provider implementation.
func NewProviderResource() resource.Resource {
	return &providerResource{}
}

// providerResource is the resource implementation.
type providerResource struct {
	client *client.Client
}

// providerResourceModel maps the resource schema data.
type providerResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Type             types.String `tfsdk:"type"`
	Config           types.Map    `tfsdk:"config"`
	Installed        types.Bool   `tfsdk:"installed"`
	LastAlertReceived types.String `tfsdk:"last_alert_received"`
}

// toClientProvider converts the Terraform model to the API client model.
func (m *providerResourceModel) toClientProvider() (*client.Provider, error) {
	config := make(map[string]string)
	if !m.Config.IsNull() && !m.Config.IsUnknown() {
		diags := m.Config.ElementsAs(context.Background(), &config, false)
		if diags.HasError() {
			return nil, fmt.Errorf("error converting config: %v", diags)
		}
	}

	provider := &client.Provider{
		ID:        m.ID.ValueString(),
		Name:      m.Name.ValueString(),
		Type:      m.Type.ValueString(),
		Config:    config,
		Installed: m.Installed.ValueBool(),
	}

	if !m.LastAlertReceived.IsNull() && !m.LastAlertReceived.IsUnknown() {
		provider.LastAlertReceived = m.LastAlertReceived.ValueString()
	}

	return provider, nil
}

// fromClientProvider updates the Terraform model from the API client model.
func (m *providerResourceModel) fromClientProvider(provider *client.Provider) error {
	m.ID = types.StringValue(provider.ID)
	m.Name = types.StringValue(provider.Name)
	m.Type = types.StringValue(provider.Type)
	m.Installed = types.BoolValue(provider.Installed)

	// Convert config map to types.Map
	config, diags := types.MapValueFrom(context.Background(), types.StringType, provider.Config)
	if diags.HasError() {
		return fmt.Errorf("error converting config to map: %v", diags)
	}
	m.Config = config

	// Set last_alert_received if available
	if provider.LastAlertReceived != "" {
		m.LastAlertReceived = types.StringValue(provider.LastAlertReceived)
	} else {
		m.LastAlertReceived = types.StringNull()
	}

	return nil
}

// Metadata returns the resource type name.
func (r *providerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider"
}

// Schema defines the schema for the resource.
func (r *providerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a KeepHQ provider. This resource allows you to create, read, update, and delete providers in KeepHQ.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique ID of the provider.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The display name of the provider. Must be unique within the KeepHQ instance.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"type": schema.StringAttribute{
				Description: "The type of the provider (e.g., 'datadog', 'pagerduty'). This cannot be changed after creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"config": schema.MapAttribute{
				Description: "Provider-specific configuration. This is a map of key-value pairs that are specific to the provider type.",
				ElementType: types.StringType,
				Required:    true,
				Sensitive:   true,
			},
			"installed": schema.BoolAttribute{
				Description: "Whether the provider is installed and ready to use.",
				Computed:    true,
			},
			"last_alert_received": schema.StringAttribute{
				Description: "Timestamp of the last alert received from this provider.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *providerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *providerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan providerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating provider", map[string]interface{}{
		"name": plan.Name.ValueString(),
		"type": plan.Type.ValueString(),
	})

	// Convert plan to API model
	provider, err := plan.toClientProvider()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating provider",
			"Could not convert provider data: "+err.Error(),
		)
		return
	}

	// Create the provider via API
	createReq := client.CreateProviderRequest{
		Name:   provider.Name,
		Type:   provider.Type,
		Config: provider.Config,
	}

	createdProvider, err := r.client.CreateProvider(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating provider",
			"Could not create provider, unexpected error: "+err.Error(),
		)
		return
	}

	// Update the plan with the response
	if err := plan.fromClientProvider(createdProvider); err != nil {
		resp.Diagnostics.AddError(
			"Error creating provider",
			"Could not process API response: "+err.Error(),
		)
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Created provider", map[string]interface{}{
		"id":   plan.ID.ValueString(),
		"name": plan.Name.ValueString(),
	})
}

// Read refreshes the Terraform state with the latest data.
func (r *providerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state providerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	providerID := state.ID.ValueString()
	if providerID == "" {
		resp.Diagnostics.AddError(
			"Error reading provider",
			"Provider ID is missing from state",
		)
		return
	}

	tflog.Debug(ctx, "Reading provider", map[string]interface{}{
		"id": providerID,
	})

	// Get provider from API
	provider, err := r.client.GetProvider(ctx, providerID)
	if err != nil {
		// If the provider is not found, remove it from state
		resp.State.RemoveResource(ctx)
		tflog.Info(ctx, "Provider not found, removing from state", map[string]interface{}{
			"id": providerID,
		})
		return
	}

	// Update state with refreshed values
	if err := state.fromClientProvider(provider); err != nil {
		resp.Diagnostics.AddError(
			"Error reading provider",
			"Could not process API response: "+err.Error(),
		)
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Read provider", map[string]interface{}{
		"id":   providerID,
		"name": state.Name.ValueString(),
	})
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *providerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan providerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state providerResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	providerID := state.ID.ValueString()
	if providerID == "" {
		resp.Diagnostics.AddError(
			"Error updating provider",
		"Provider ID is missing from state",
		)
		return
	}

	tflog.Debug(ctx, "Updating provider", map[string]interface{}{
		"id": providerID,
	})

	// Convert plan to API model
	provider, err := plan.toClientProvider()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating provider",
			"Could not convert provider data: "+err.Error(),
		)
		return
	}

	// Update the provider via API
	updateReq := client.UpdateProviderRequest{
		Name:   provider.Name,
		Config: provider.Config,
	}

	updatedProvider, err := r.client.UpdateProvider(ctx, providerID, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating provider",
			"Could not update provider, unexpected error: "+err.Error(),
		)
		return
	}

	// Update the plan with the response
	if err := plan.fromClientProvider(updatedProvider); err != nil {
		resp.Diagnostics.AddError(
			"Error updating provider",
			"Could not process API response: "+err.Error(),
		)
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Updated provider", map[string]interface{}{
		"id":   providerID,
		"name": plan.Name.ValueString(),
	})
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *providerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state providerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	providerID := state.ID.ValueString()
	if providerID == "" {
		resp.Diagnostics.AddError(
			"Error deleting provider",
			"Provider ID is missing from state",
		)
		return
	}

	tflog.Debug(ctx, "Deleting provider", map[string]interface{}{
		"id":   providerID,
		"name": state.Name.ValueString(),
	})

	// Delete the provider via API
	err := r.client.DeleteProvider(ctx, providerID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting provider",
			"Could not delete provider, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "Deleted provider", map[string]interface{}{
		"id":   providerID,
		"name": state.Name.ValueString(),
	})
}

// ImportState handles resource import.
func (r *providerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
