// resource_extraction_rule.go - Resource implementation for KeepHQ extraction rules
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/keephq/terraform-provider-keep/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &extractionRuleResource{}
	_ resource.ResourceWithConfigure   = &extractionRuleResource{}
	_ resource.ResourceWithImportState = &extractionRuleResource{}
)

// NewExtractionRuleResource is a helper function to simplify the provider implementation.
func NewExtractionRuleResource() resource.Resource {
	return &extractionRuleResource{}
}

// extractionRuleResource defines the resource implementation.
type extractionRuleResource struct {
	client *client.Client
}

// extractionRuleResourceModel maps the resource schema data.
type extractionRuleResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Priority    types.Int64  `tfsdk:"priority"`
	Disabled    types.Bool   `tfsdk:"disabled"`
	Pre         types.Bool   `tfsdk:"pre"`
	Condition   types.String `tfsdk:"condition"`
	Attribute   types.String `tfsdk:"attribute"`
	Regex       types.String `tfsdk:"regex"`
}

// Metadata returns the resource type name.
func (r *extractionRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_extraction_rule"
}

// Schema defines the schema for the resource.
func (r *extractionRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an extraction rule in Keep. Extraction rules define how to extract and transform data from incoming alerts.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the extraction rule.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the extraction rule.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of what the extraction rule does.",
				Optional:    true,
				Computed:    true,
			},
			"priority": schema.Int64Attribute{
				Description: "The priority of the extraction rule (lower number = higher priority). Defaults to 0.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
			"disabled": schema.BoolAttribute{
				Description: "Whether the extraction rule is disabled. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"pre": schema.BoolAttribute{
				Description: "Whether this is a pre-processing rule that runs before other rules. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"condition": schema.StringAttribute{
				Description: "CEL expression that determines when this rule should be applied. If not specified, the rule will always be applied.",
				Optional:    true,
			},
			"attribute": schema.StringAttribute{
				Description: "The attribute to extract from the alert.",
				Required:    true,
			},
			"regex": schema.StringAttribute{
				Description: "The regex pattern to use for extraction.",
				Required:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *extractionRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *extractionRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan extractionRuleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new extraction rule
	extractionRule := map[string]interface{}{
		"name":        plan.Name.ValueString(),
		"description": plan.Description.ValueString(),
		"priority":    int(plan.Priority.ValueInt64()),
		"disabled":    plan.Disabled.ValueBool(),
		"pre":         plan.Pre.ValueBool(),
		"condition":   plan.Condition.ValueString(),
		"attribute":   plan.Attribute.ValueString(),
		"regex":       plan.Regex.ValueString(),
	}
	
	// Create the extraction rule via API
	createdRule, err := r.client.CreateExtractionRule(ctx, extractionRule)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating extraction rule",
			"Could not create extraction rule, unexpected error: "+err.Error(),
		)
		return
	}
	
	// Map response back to the plan
	plan.ID = types.StringValue(fmt.Sprintf("%d", int(createdRule["id"].(float64))))
	plan.Name = types.StringValue(createdRule["name"].(string))
	if desc, ok := createdRule["description"].(string); ok {
		plan.Description = types.StringValue(desc)
	}
	if priority, ok := createdRule["priority"].(float64); ok {
		plan.Priority = types.Int64Value(int64(priority))
	}
	if disabled, ok := createdRule["disabled"].(bool); ok {
		plan.Disabled = types.BoolValue(disabled)
	}
	if pre, ok := createdRule["pre"].(bool); ok {
		plan.Pre = types.BoolValue(pre)
	}
	if condition, ok := createdRule["condition"].(string); ok && condition != "" {
		plan.Condition = types.StringValue(condition)
	}
	if attribute, ok := createdRule["attribute"].(string); ok {
		plan.Attribute = types.StringValue(attribute)
	}
	if regex, ok := createdRule["regex"].(string); ok {
		plan.Regex = types.StringValue(regex)
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *extractionRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state extractionRuleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get extraction rule from API
	ruleID := state.ID.ValueString()
	extractionRule, err := r.client.GetExtractionRule(ctx, ruleID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading extraction rule",
			"Could not read extraction rule ID "+ruleID+": "+err.Error(),
		)
		return
	}
	
	// Update the state with the response
	state.ID = types.StringValue(fmt.Sprintf("%d", int(extractionRule["id"].(float64))))
	state.Name = types.StringValue(extractionRule["name"].(string))
	if desc, ok := extractionRule["description"].(string); ok {
		state.Description = types.StringValue(desc)
	}
	if priority, ok := extractionRule["priority"].(float64); ok {
		state.Priority = types.Int64Value(int64(priority))
	}
	if disabled, ok := extractionRule["disabled"].(bool); ok {
		state.Disabled = types.BoolValue(disabled)
	}
	if pre, ok := extractionRule["pre"].(bool); ok {
		state.Pre = types.BoolValue(pre)
	}
	if condition, ok := extractionRule["condition"].(string); ok && condition != "" {
		state.Condition = types.StringValue(condition)
	}
	if attribute, ok := extractionRule["attribute"].(string); ok {
		state.Attribute = types.StringValue(attribute)
	}
	if regex, ok := extractionRule["regex"].(string); ok {
		state.Regex = types.StringValue(regex)
	}


	// Map response to state
	// state.ID = types.StringValue(fmt.Sprintf("%d", int(extractionRule["id"].(float64))))
	// state.Name = types.StringValue(extractionRule["name"].(string))
	// state.Description = types.StringValue(extractionRule["description"].(string))
	// state.Priority = types.Int64Value(int64(extractionRule["priority"].(float64)))
	// state.Disabled = types.BoolValue(extractionRule["disabled"].(bool))
	// state.Pre = types.BoolValue(extractionRule["pre"].(bool))
	// state.Condition = types.StringValue(extractionRule["condition"].(string))
	// state.Attribute = types.StringValue(extractionRule["attribute"].(string))
	// state.Regex = types.StringValue(extractionRule["regex"].(string))

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *extractionRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan extractionRuleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state extractionRuleResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update extraction rule via API
	updatedRule, err := r.client.UpdateExtractionRule(ctx, state.ID.ValueString(), map[string]interface{}{
		"name":        plan.Name.ValueString(),
		"description": plan.Description.ValueString(),
		"priority":    int(plan.Priority.ValueInt64()),
		"disabled":    plan.Disabled.ValueBool(),
		"pre":         plan.Pre.ValueBool(),
		"condition":   plan.Condition.ValueString(),
		"attribute":   plan.Attribute.ValueString(),
		"regex":       plan.Regex.ValueString(),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating extraction rule",
			"Could not update extraction rule, unexpected error: "+err.Error(),
		)
		return
	}

	// Update the plan with the response data
	if name, ok := updatedRule["name"].(string); ok {
		plan.Name = types.StringValue(name)
	}
	if desc, ok := updatedRule["description"].(string); ok {
		plan.Description = types.StringValue(desc)
	}
	if priority, ok := updatedRule["priority"].(float64); ok {
		plan.Priority = types.Int64Value(int64(priority))
	}
	if disabled, ok := updatedRule["disabled"].(bool); ok {
		plan.Disabled = types.BoolValue(disabled)
	}
	if pre, ok := updatedRule["pre"].(bool); ok {
		plan.Pre = types.BoolValue(pre)
	}
	if condition, ok := updatedRule["condition"].(string); ok && condition != "" {
		plan.Condition = types.StringValue(condition)
	}
	if attribute, ok := updatedRule["attribute"].(string); ok {
		plan.Attribute = types.StringValue(attribute)
	}
	if regex, ok := updatedRule["regex"].(string); ok {
		plan.Regex = types.StringValue(regex)
	}


	// Map response to schema
	// plan.ID = types.StringValue(fmt.Sprintf("%d", int(updatedRule["id"].(float64))))
	// plan.Name = types.StringValue(updatedRule["name"].(string))
	// plan.Description = types.StringValue(updatedRule["description"].(string))
	// plan.Priority = types.Int64Value(int64(updatedRule["priority"].(float64)))
	// plan.Disabled = types.BoolValue(updatedRule["disabled"].(bool))
	// plan.Pre = types.BoolValue(updatedRule["pre"].(bool))
	// plan.Condition = types.StringValue(updatedRule["condition"].(string))
	// plan.Attribute = types.StringValue(updatedRule["attribute"].(string))
	// plan.Regex = types.StringValue(updatedRule["regex"].(string))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *extractionRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state extractionRuleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete extraction rule via API
	err := r.client.DeleteExtractionRule(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting extraction rule",
			"Could not delete extraction rule, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState implements resource.ResourceWithImportState.
func (r *extractionRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
