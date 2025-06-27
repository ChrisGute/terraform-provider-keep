// resource_mapping_rule.go - Resource implementation for KeepHQ mapping rules
package provider

import (
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/keephq/terraform-provider-keep/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &mappingRuleResource{}
	_ resource.ResourceWithConfigure   = &mappingRuleResource{}
	_ resource.ResourceWithImportState = &mappingRuleResource{}
)

// NewMappingRuleResource is a helper function to simplify the provider implementation.
func NewMappingRuleResource() resource.Resource {
	return &mappingRuleResource{}
}

// csvRow represents a single row in the CSV data
type csvRow map[string]string

// parseCSVData parses the CSV data string into a slice of rows
func parseCSVData(csvData string) ([]csvRow, error) {
	// Create a new CSV reader
	reader := csv.NewReader(strings.NewReader(csvData))
	
	// Read the header row
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV header: %w", err)
	}

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV records: %w", err)
	}

	// Convert records to CSV rows
	var rows []csvRow
	for _, record := range records {
		if len(record) != len(header) {
			return nil, fmt.Errorf("CSV record has wrong number of fields: expected %d, got %d", len(header), len(record))
		}

		row := make(csvRow)
		for i, key := range header {
			row[strings.TrimSpace(key)] = strings.TrimSpace(record[i])
		}
		rows = append(rows, row)
	}

	return rows, nil
}

// mappingRuleResource defines the resource implementation.
type mappingRuleResource struct {
	client *client.Client
}

// mappingRuleResourceModel maps the resource schema data.
type mappingRuleResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Priority    types.Int64  `tfsdk:"priority"`
	Disabled    types.Bool   `tfsdk:"disabled"`
	Matchers    types.Map    `tfsdk:"matchers"`
	CSVData     types.String `tfsdk:"csv_data"`
	LastUpdated types.String `tfsdk:"-"`
}

// matcher represents a single matcher in the API request
type matcher struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Metadata returns the resource type name.
func (r *mappingRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mapping_rule"
}

// Schema defines the schema for the resource.
func (r *mappingRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a mapping rule in Keep. Mapping rules define how to enrich alerts with additional data from CSV files or topology data.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the mapping rule.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the mapping rule.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of what the mapping rule does.",
				Optional:    true,
			},
			"priority": schema.Int64Attribute{
				Description: "The priority of the mapping rule. Lower numbers have higher priority.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
			"disabled": schema.BoolAttribute{
				Description: "Whether the mapping rule is disabled. Note: This field is currently not supported by the API and will be ignored.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					// Always set to false since the API doesn't support this field
					boolplanmodifier.UseStateForUnknown(),
				},
				Default: booldefault.StaticBool(false),
			},
			"matchers": schema.MapAttribute{
				Description: "A map of matchers that determine when this rule should be applied.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"csv_data": schema.StringAttribute{
				Description: "The CSV data to use for mapping. Each row should contain the matcher values and the fields to add to matching alerts.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *mappingRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *mappingRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan mappingRuleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert matchers to the expected format (list of key-value pairs)
	var matchersMap map[string]string
	diags = plan.Matchers.ElementsAs(ctx, &matchersMap, false)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Convert map to list of matcher lists (e.g., [["key", "value"]])
	matchersList := make([][]string, 0, len(matchersMap))
	for key, value := range matchersMap {
		matchersList = append(matchersList, []string{key, value})
	}

	// Initialize the rule with common fields
	// Note: 'disabled' field is intentionally omitted as it's not supported by the API
	rule := map[string]interface{}{
		"name":        plan.Name.ValueString(),
		"description": plan.Description.ValueString(),
		"priority":    plan.Priority.ValueInt64(),
		"matchers":    matchersList,
	}

	// Parse the CSV data if provided
	var csvDataStr string
	if !plan.CSVData.IsNull() && !plan.CSVData.IsUnknown() {
		csvDataStr = plan.CSVData.ValueString()
		tflog.Debug(ctx, "CSV data from plan", map[string]interface{}{
			"raw_length": len(csvDataStr),
			"raw_value":  getStringPreview(csvDataStr, 100),
		})
		
		// Validate the CSV data
		_, err := parseCSVData(csvDataStr)
		if err != nil {
			tflog.Error(ctx, "Failed to parse CSV data", map[string]interface{}{
				"error": err.Error(),
			})
			resp.Diagnostics.AddError(
				"Invalid CSV Data",
				fmt.Sprintf("Failed to parse CSV data: %s", err.Error()),
			)
			return
		}
	} else {
		tflog.Debug(ctx, "No CSV data provided in plan")
	}

	// Handle CSV data if provided
	if !plan.CSVData.IsNull() && !plan.CSVData.IsUnknown() {
		csvData := plan.CSVData.ValueString()
		rule["csv_data"] = csvData
		rule["type"] = "csv"

		// Parse CSV data to create rows
		rows, err := parseCSVData(csvData)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error parsing CSV data",
				"Could not parse CSV data: "+err.Error(),
			)
			return
		}
		rule["rows"] = rows
	}

	// Log the full request payload for debugging (without the potentially large CSV data for brevity)
	loggableRule := make(map[string]interface{})
	for k, v := range rule {
		if k == "csv_data" {
			csvStr, _ := v.(string)
			loggableRule["csv_data"] = fmt.Sprintf("[CSV data, length: %d, preview: %s]", len(csvStr), getStringPreview(csvStr, 50))
		} else if k == "rows" {
			loggableRule["rows"] = "[CSV rows]"
		} else {
			loggableRule[k] = v
		}
	}
	tflog.Debug(ctx, "Creating mapping rule - Request payload", map[string]interface{}{
		"request": loggableRule,
	})

	// Log the mapping rule being created with field details
	hasCSV := !plan.CSVData.IsNull() && !plan.CSVData.IsUnknown()
	csvPreview := ""
	if hasCSV {
		csvPreview = getStringPreview(plan.CSVData.ValueString(), 50)
	}

	tflog.Debug(ctx, "Creating mapping rule - Field details", map[string]interface{}{
		"id":           plan.ID.ValueString(),
		"name":         plan.Name.ValueString(),
		"description":  plan.Description.ValueString(),
		"priority":     plan.Priority.ValueInt64(),
		"disabled":     plan.Disabled.ValueBool(),
		"matchers":     matchersList,
		"has_csv":      hasCSV,
		"csv_preview":  csvPreview,
		"matchers_len": len(matchersList),
	})

	// Create the mapping rule via API
	createdRule, err := r.client.CreateMappingRule(ctx, rule)
	if err != nil {
		tflog.Error(ctx, "Failed to create mapping rule", map[string]interface{}{
			"error": err.Error(),
		})
		resp.Diagnostics.AddError(
			"Error creating mapping rule",
			"Could not create mapping rule, unexpected error: "+err.Error(),
		)
		return
	}

	// Log the API response with type information
	responseTypes := make(map[string]string)
	for k, v := range createdRule {
		responseTypes[k] = fmt.Sprintf("%T", v)
	}

	// Create a loggable version of the response (without large data)
	loggableResponse := make(map[string]interface{})
	for k, v := range createdRule {
		if k == "csv_data" {
			if csvStr, ok := v.(string); ok {
				loggableResponse["csv_data"] = fmt.Sprintf("[CSV data, length: %d, preview: %s]", len(csvStr), getStringPreview(csvStr, 50))
			} else {
				loggableResponse["csv_data"] = fmt.Sprintf("[%T, value: %v]", v, v)
			}
		} else {
			loggableResponse[k] = v
		}
	}

	tflog.Debug(ctx, "Mapping rule created - API response", map[string]interface{}{
		"response":      loggableResponse,
		"response_types": responseTypes,
	})

	// Update the plan with the response
	if id, ok := createdRule["id"]; ok {
		plan.ID = types.StringValue(fmt.Sprint(id))
		tflog.Debug(ctx, "Set ID from API response", map[string]interface{}{"id": id})
	}

	if name, ok := createdRule["name"].(string); ok {
		plan.Name = types.StringValue(name)
		tflog.Debug(ctx, "Set name from API response", map[string]interface{}{"name": name})
	}

	if desc, ok := createdRule["description"].(string); ok {
		plan.Description = types.StringValue(desc)
		tflog.Debug(ctx, "Set description from API response", map[string]interface{}{"description": desc})
	}

	if prio, ok := createdRule["priority"].(float64); ok {
		plan.Priority = types.Int64Value(int64(prio))
		tflog.Debug(ctx, "Set priority from API response", map[string]interface{}{"priority": prio})
	}

	if disabled, ok := createdRule["disabled"].(bool); ok {
		plan.Disabled = types.BoolValue(disabled)
		tflog.Debug(ctx, "Set disabled from API response", map[string]interface{}{"disabled": disabled})
	}

	// Handle CSV data from API response
	if csvData, exists := createdRule["csv_data"]; exists && csvData != nil {
		if v, ok := csvData.(string); ok && v != "" {
			normalizedCSV := strings.ReplaceAll(strings.TrimSpace(v), "\r\n", "\n")
			plan.CSVData = types.StringValue(normalizedCSV)
			tflog.Debug(ctx, "Set csv_data from API response in Create", map[string]interface{}{
				"csv_data_length":  len(v),
				"csv_data_preview": getStringPreview(v, 50),
			})
		} else {
			// If API returns empty string, set to null
			plan.CSVData = types.StringNull()
			tflog.Debug(ctx, "Empty csv_data in API response in Create")
		}
	} else {
		// If API omits csv_data, preserve value from plan
		plan.CSVData = plan.CSVData
		tflog.Debug(ctx, "No csv_data in API response, preserving value from plan")
	}

	// Handle matchers from API response
	switch matchers := createdRule["matchers"].(type) {
	case map[string]interface{}:
		if len(matchers) > 0 {
			matcherMap := make(map[string]attr.Value)
			for k, v := range matchers {
				switch v := v.(type) {
				case string:
					matcherMap[k] = types.StringValue(v)
				case bool, float64:
					matcherMap[k] = types.StringValue(fmt.Sprintf("%v", v))
				default:
					tflog.Warn(ctx, "Unsupported matcher value type", map[string]interface{}{
						"key":   k,
						"type":  fmt.Sprintf("%T", v),
						"value": v,
					})
				}
			}
			mapValue, diags := types.MapValue(types.StringType, matcherMap)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				plan.Matchers = mapValue
			}
		}
	case []interface{}:
		matcherMap := make(map[string]attr.Value)
		for _, m := range matchers {
			if pair, ok := m.([]interface{}); ok && len(pair) == 2 {
				if key, ok := pair[0].(string); ok {
					matcherMap[key] = types.StringValue(fmt.Sprintf("%v", pair[1]))
				}
			}
		}
		mapValue, diags := types.MapValue(types.StringType, matcherMap)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			plan.Matchers = mapValue
			tflog.Debug(ctx, "Set matchers from API response (list format)", map[string]interface{}{
				"matchers_count": len(matcherMap),
				"matchers_keys":  getMapKeys(matcherMap),
			})
		}
	default:
		plan.Matchers = types.MapNull(types.StringType)
		tflog.Debug(ctx, "Unexpected matchers type in API response", map[string]interface{}{
			"type":  fmt.Sprintf("%T", createdRule["matchers"]),
			"value": createdRule["matchers"],
		})
	}

	// Defensive: If matchers is still unknown or null, set as types.MapNull(types.StringType)
	if plan.Matchers.IsUnknown() || plan.Matchers.IsNull() {
		plan.Matchers = types.MapNull(types.StringType)
	}

	// Map response body to model with proper type assertions
	if id, ok := createdRule["id"]; ok {
		switch v := id.(type) {
		case float64:
			plan.ID = types.StringValue(fmt.Sprintf("%.0f", v))
		case string:
			plan.ID = types.StringValue(v)
		default:
			tflog.Warn(ctx, "Unexpected type for ID in API response", map[string]interface{}{
				"type": fmt.Sprintf("%T", id),
			})
		}
	}

	if name, ok := createdRule["name"]; ok {
		if nameStr, ok := name.(string); ok {
			plan.Name = types.StringValue(nameStr)
		}
	}

	if desc, ok := createdRule["description"]; ok {
		if descStr, ok := desc.(string); ok {
			plan.Description = types.StringValue(descStr)
		}
	}

	if prio, ok := createdRule["priority"]; ok {
		switch v := prio.(type) {
		case float64:
			plan.Priority = types.Int64Value(int64(v))
		case int:
			plan.Priority = types.Int64Value(int64(v))
		case int64:
			plan.Priority = types.Int64Value(v)
		}
	}

	if disabled, ok := createdRule["disabled"]; ok {
		if disabledBool, ok := disabled.(bool); ok {
			plan.Disabled = types.BoolValue(disabledBool)
		}
	}

	// Handle CSV data from API response
	if csvData, exists := createdRule["csv_data"]; exists && csvData != nil {
		switch v := csvData.(type) {
		case string:
			if v != "" {
				// Normalize line endings and trim whitespace
				normalizedCSV := strings.ReplaceAll(strings.TrimSpace(v), "\r\n", "\n")
				plan.CSVData = types.StringValue(normalizedCSV)
				tflog.Debug(ctx, "Set csv_data from API response in Create", map[string]interface{}{
					"csv_data_length":  len(v),
					"csv_data_preview": getStringPreview(v, 50),
				})
			} else {
				plan.CSVData = types.StringNull()
				tflog.Debug(ctx, "Empty csv_data in API response in Create")
			}
		default:
			tflog.Debug(ctx, "Unexpected type for csv_data in API response", map[string]interface{}{
				"type":  fmt.Sprintf("%T", csvData),
				"value": fmt.Sprintf("%v", csvData),
			})
		}
	} else {
		tflog.Debug(ctx, "No csv_data in API response, preserving existing value")
	}

	// Set the last updated time
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Successfully created mapping rule", map[string]interface{}{
		"id":          plan.ID.ValueString(),
		"name":        plan.Name.ValueString(),
		"has_csv":     !plan.CSVData.IsNull() && !plan.CSVData.IsUnknown(),
	})
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *mappingRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Starting Update method", map[string]interface{}{
		"state_id": req.State.Raw.String(),
	})

	// Retrieve values from plan
	var plan mappingRuleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to retrieve values from plan", map[string]interface{}{
			"diagnostics": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Debug(ctx, "Retrieved plan values", map[string]interface{}{
		"name":        plan.Name.ValueString(),
		"description": plan.Description.ValueString(),
		"priority":    plan.Priority.ValueInt64(),
		"disabled":    plan.Disabled.ValueBool(),
	})

	// Get current state to get the ID
	var state mappingRuleResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to retrieve current state", map[string]interface{}{
			"diagnostics": resp.Diagnostics.Errors(),
		})
		return
	}

	if state.ID.IsNull() || state.ID.IsUnknown() {
		resp.Diagnostics.AddError(
			"Invalid state",
			"Cannot update mapping rule with empty ID",
		)
		return
	}

	tflog.Debug(ctx, "Retrieved current state", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	// Extract matchers from plan as a list of [key, value] pairs
	var matchersList [][]string
	if !plan.Matchers.IsNull() && !plan.Matchers.IsUnknown() {
		var matchersMap map[string]string
		diags = plan.Matchers.ElementsAs(ctx, &matchersMap, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		for k, v := range matchersMap {
			matchersList = append(matchersList, []string{k, v})
		}
	}

	// Prepare the update payload
	// Note: 'disabled' field is intentionally omitted as it's not supported by the API
	updateData := map[string]interface{}{
		"name":        plan.Name.ValueString(),
		"description": plan.Description.ValueString(),
		"priority":    plan.Priority.ValueInt64(),
		"matchers":    matchersList,
	}

	// Add CSV data and rows if provided
	if !plan.CSVData.IsNull() && !plan.CSVData.IsUnknown() {
		csvData := plan.CSVData.ValueString()
		updateData["csv_data"] = csvData
		
		// Parse CSV data to rows for the API
		if csvData != "" {
			rows, err := parseCSVData(csvData)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error parsing CSV data",
					fmt.Sprintf("Could not parse CSV data: %s", err.Error()),
				)
				return
			}
			updateData["rows"] = rows
		}
	}

	// Log the update data being sent to the API
	tflog.Debug(ctx, "Sending update request to API", map[string]interface{}{
		"id":          state.ID.ValueString(),
		"update_data": updateData,
		"disabled_in_plan": plan.Disabled.ValueBool(),
	})

	// Update the mapping rule
	updatedRule, err := r.client.UpdateMappingRule(ctx, state.ID.ValueString(), updateData)
	if err != nil {
		errMsg := fmt.Sprintf("Could not update mapping rule, unexpected error: %s", err.Error())
		tflog.Error(ctx, errMsg, map[string]interface{}{
			"error": err,
		})
		resp.Diagnostics.AddError(
			"Error updating mapping rule",
			errMsg,
		)
		return
	}

	// Log the entire API response for debugging
	tflog.Debug(ctx, "API response from UpdateMappingRule", map[string]interface{}{
		"response": updatedRule,
	})

	tflog.Debug(ctx, "Received update response from API", map[string]interface{}{
		"response": updatedRule,
	})

	// Update the plan with the response
	if id, ok := updatedRule["id"].(string); ok {
		plan.ID = types.StringValue(id)
	} else {
		tflog.Error(ctx, "Missing or invalid 'id' in API response", map[string]interface{}{
			"type":  fmt.Sprintf("%T", updatedRule["id"]),
			"value": updatedRule["id"],
		})
	}

	if name, ok := updatedRule["name"].(string); ok {
		plan.Name = types.StringValue(name)
	} else {
		tflog.Error(ctx, "Missing or invalid 'name' in API response", map[string]interface{}{
			"type":  fmt.Sprintf("%T", updatedRule["name"]),
			"value": updatedRule["name"],
		})
	}

	if description, ok := updatedRule["description"].(string); ok {
		plan.Description = types.StringValue(description)
	} else {
		tflog.Debug(ctx, "Missing or invalid 'description' in API response, using existing value")
	}

	if priority, ok := updatedRule["priority"].(float64); ok {
		plan.Priority = types.Int64Value(int64(priority))
	} else {
		tflog.Error(ctx, "Missing or invalid 'priority' in API response", map[string]interface{}{
			"type":  fmt.Sprintf("%T", updatedRule["priority"]),
			"value": updatedRule["priority"],
		})
	}

	// Save the disabled value from the plan before any updates
	planDisabled := plan.Disabled
	tflog.Debug(ctx, "Disabled value from plan before update", map[string]interface{}{
		"disabled":      planDisabled.ValueBool(),
		"disabled_type": fmt.Sprintf("%T", planDisabled),
		"is_null":       planDisabled.IsNull(),
		"is_unknown":    planDisabled.IsUnknown(),
	})

	// Handle the disabled field from the API response
	if disabledVal, exists := updatedRule["disabled"]; exists {
		tflog.Debug(ctx, "Disabled value from API response", map[string]interface{}{
			"type":  fmt.Sprintf("%T", disabledVal),
			"value": disabledVal,
		})

		switch v := disabledVal.(type) {
		case bool:
			plan.Disabled = types.BoolValue(v)
			tflog.Debug(ctx, "Set disabled from API response (bool)", map[string]interface{}{
				"disabled": v,
			})
		case string:
			if v == "true" || v == "false" {
				disabled, _ := strconv.ParseBool(v)
				plan.Disabled = types.BoolValue(disabled)
				tflog.Debug(ctx, "Set disabled from API response (string)", map[string]interface{}{
					"disabled": disabled,
				})
			} else {
				tflog.Debug(ctx, "Invalid 'disabled' string value in API response, using plan value", map[string]interface{}{
					"value": v,
				})
				plan.Disabled = planDisabled
			}
		default:
			tflog.Debug(ctx, "Unexpected 'disabled' type in API response, using plan value", map[string]interface{}{
				"type":  fmt.Sprintf("%T", v),
				"value": v,
			})
			plan.Disabled = planDisabled
		}
	} else {
		tflog.Debug(ctx, "'disabled' field not in API response, using plan value", map[string]interface{}{
			"plan_disabled": planDisabled.ValueBool(),
		})
		plan.Disabled = planDisabled
	}

	// Log the final disabled value that will be set in the state
	tflog.Debug(ctx, "Final disabled value for state", map[string]interface{}{
		"disabled": plan.Disabled.ValueBool(),
		"disabled_type": fmt.Sprintf("%T", plan.Disabled),
	})

	// Log the entire plan before setting state
	tflog.Debug(ctx, "Plan before setting state", map[string]interface{}{
		"id":          plan.ID.ValueString(),
		"name":        plan.Name.ValueString(),
		"description": plan.Description.ValueString(),
		"priority":    plan.Priority.ValueInt64(),
		"disabled":    plan.Disabled.ValueBool(),
		"has_csv":     !plan.CSVData.IsNull() && !plan.CSVData.IsUnknown(),
		"has_matchers": !plan.Matchers.IsNull() && !plan.Matchers.IsUnknown(),
	})

	// Handle matchers from response
	matcherMap := make(map[string]attr.Value)

	switch matchers := updatedRule["matchers"].(type) {
	case map[string]interface{}:
		// Handle map format
		for k, v := range matchers {
			switch v := v.(type) {
			case string:
				matcherMap[k] = types.StringValue(v)
			case float64:
				// Convert float64 to string, handling potential integer values
				if v == float64(int64(v)) {
					matcherMap[k] = types.StringValue(fmt.Sprintf("%d", int64(v)))
				} else {
					matcherMap[k] = types.StringValue(fmt.Sprintf("%v", v))
				}
			default:
				// Fallback for any other type
				matcherMap[k] = types.StringValue(fmt.Sprintf("%v", v))
			}
		}

	case []interface{}:
		// Handle list of [key, value] pairs format
		for _, m := range matchers {
			if pair, ok := m.([]interface{}); ok && len(pair) == 2 {
				key, keyOk := pair[0].(string)
				if !keyOk {
					continue
				}

				switch v := pair[1].(type) {
				case string:
					matcherMap[key] = types.StringValue(v)
				case float64:
					// Convert float64 to string, handling potential integer values
					if v == float64(int64(v)) {
						matcherMap[key] = types.StringValue(fmt.Sprintf("%d", int64(v)))
					} else {
						matcherMap[key] = types.StringValue(fmt.Sprintf("%v", v))
					}
				default:
					// Fallback for any other type
					matcherMap[key] = types.StringValue(fmt.Sprintf("%v", v))
				}
			}
		}

	default:
		tflog.Debug(ctx, "Unexpected matchers type in API response", map[string]interface{}{
			"type":  fmt.Sprintf("%T", updatedRule["matchers"]),
			"value": updatedRule["matchers"],
		})
	}

	// Set the matchers in the plan
	if len(matcherMap) > 0 {
		mapValue, diags := types.MapValue(types.StringType, matcherMap)
		if !diags.HasError() {
			plan.Matchers = mapValue
		} else {
			tflog.Debug(ctx, "Failed to create map value for matchers", map[string]interface{}{
				"error": diags.Errors(),
			})
			plan.Matchers = types.MapNull(types.StringType)
		}
	} else {
		plan.Matchers = types.MapNull(types.StringType)
	}

	// Defensive: If matchers is still unknown or null, set as types.MapNull(types.StringType)
	if plan.Matchers.IsUnknown() || plan.Matchers.IsNull() {
		plan.Matchers = types.MapNull(types.StringType)
	}

	// Update CSV data from response if present
	if csvData, ok := updatedRule["csv_data"].(string); ok && csvData != "" {
		// Normalize line endings and trim whitespace
		normalizedCSV := strings.ReplaceAll(strings.TrimSpace(csvData), "\r\n", "\n")
		plan.CSVData = types.StringValue(normalizedCSV)
	}

	// Set the last updated time
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updated mapping rule", map[string]interface{}{
		"id":          plan.ID.ValueString(),
		"name":        plan.Name.ValueString(),
		"has_csv":     !plan.CSVData.IsNull() && !plan.CSVData.IsUnknown(),
	})
}

// Read refreshes the Terraform state with the latest data.
func (r *mappingRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state mappingRuleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed mapping rule from API
	rule, err := r.client.GetMappingRule(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading mapping rule",
			fmt.Sprintf("Could not read mapping rule ID %s: %s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	// Overwrite state with refreshed values
	if name, ok := rule["name"].(string); ok {
		state.Name = types.StringValue(name)
	}

	if desc, ok := rule["description"].(string); ok {
		state.Description = types.StringValue(desc)
	}

	if prio, ok := rule["priority"].(float64); ok {
		state.Priority = types.Int64Value(int64(prio))
	}

	if disabled, ok := rule["disabled"].(bool); ok {
		state.Disabled = types.BoolValue(disabled)
	}

	// Handle matchers
	switch matchers := rule["matchers"].(type) {
	case map[string]interface{}:
		matcherMap := make(map[string]attr.Value)
		for k, v := range matchers {
			matcherMap[k] = types.StringValue(fmt.Sprintf("%v", v))
		}
		mapValue, diags := types.MapValue(types.StringType, matcherMap)
		if !diags.HasError() {
			state.Matchers = mapValue
		}
	case []interface{}:
		matcherMap := make(map[string]attr.Value)
		for _, m := range matchers {
			if pair, ok := m.([]interface{}); ok && len(pair) == 2 {
				if key, ok := pair[0].(string); ok {
					matcherMap[key] = types.StringValue(fmt.Sprintf("%v", pair[1]))
				}
			}
		}
		mapValue, diags := types.MapValue(types.StringType, matcherMap)
		if !diags.HasError() {
			state.Matchers = mapValue
		}
	default:
		state.Matchers = types.MapNull(types.StringType)
		tflog.Debug(ctx, "Unexpected matchers type in API response", map[string]interface{}{
			"type":  fmt.Sprintf("%T", rule["matchers"]),
			"value": rule["matchers"],
		})
	}

	// Defensive: If matchers is still unknown or null, set as types.MapNull(types.StringType)
	if state.Matchers.IsUnknown() || state.Matchers.IsNull() {
		state.Matchers = types.MapNull(types.StringType)
	}

	// Handle CSV data
	if csvData, ok := rule["csv_data"].(string); ok && csvData != "" {
		// Normalize line endings and trim whitespace
		normalizedCSV := strings.ReplaceAll(strings.TrimSpace(csvData), "\r\n", "\n")
		state.CSVData = types.StringValue(normalizedCSV)
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Refreshed mapping rule state", map[string]interface{}{
		"id":          state.ID.ValueString(),
		"name":        state.Name.ValueString(),
		"has_csv":     !state.CSVData.IsNull() && !state.CSVData.IsUnknown(),
	})
}

// ImportState implements resource.ResourceWithImportState.
func (r *mappingRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Log the import request
	tflog.Debug(ctx, "Importing mapping rule", map[string]interface{}{
		"import_id": req.ID,
	})

	// Get the mapping rule from the API
	rule, err := r.client.GetMappingRule(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing mapping rule",
			fmt.Sprintf("Could not read mapping rule during import: %s", err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "Retrieved mapping rule from API", map[string]interface{}{
		"rule_data": rule,
	})

	// Create a new state model
	var state mappingRuleResourceModel

	// Set the ID from the import ID
	state.ID = types.StringValue(req.ID)

	// Set other fields from the API response
	if name, ok := rule["name"].(string); ok {
		state.Name = types.StringValue(name)
	}

	if desc, ok := rule["description"].(string); ok {
		state.Description = types.StringValue(desc)
	}

	if prio, ok := rule["priority"].(float64); ok {
		state.Priority = types.Int64Value(int64(prio))
	}

	// Handle disabled field - default to false if not set
	disabled := false
	if d, ok := rule["disabled"].(bool); ok {
		disabled = d
	}
	state.Disabled = types.BoolValue(disabled)

	// Handle matchers
	matcherMap := make(map[string]attr.Value)
	if matchers, exists := rule["matchers"]; exists && matchers != nil {
		switch m := matchers.(type) {
		case map[string]interface{}:
			for k, v := range m {
				matcherMap[k] = types.StringValue(fmt.Sprintf("%v", v))
			}
		case []interface{}:
			for _, item := range m {
				if pair, ok := item.([]interface{}); ok && len(pair) == 2 {
					if key, ok := pair[0].(string); ok {
						matcherMap[key] = types.StringValue(fmt.Sprintf("%v", pair[1]))
					}
				}
			}
		}
	}

	// Set matchers in state, or null if empty
	if len(matcherMap) > 0 {
		mapValue, diags := types.MapValue(types.StringType, matcherMap)
		if !diags.HasError() {
			state.Matchers = mapValue
		} else {
			state.Matchers = types.MapNull(types.StringType)
		}
	} else {
		state.Matchers = types.MapNull(types.StringType)
	}

	// Handle CSV data - ensure consistent formatting
	if csvData, ok := rule["csv_data"].(string); ok && csvData != "" {
		// Normalize line endings, trim whitespace, and ensure consistent line endings
		normalizedCSV := strings.TrimSpace(csvData)
		normalizedCSV = strings.ReplaceAll(normalizedCSV, "\r\n", "\n")
		normalizedCSV = strings.ReplaceAll(normalizedCSV, "\r", "\n")
		state.CSVData = types.StringValue(normalizedCSV)
	} else {
		// Explicitly set to null if not present
		state.CSVData = types.StringNull()
	}

	// Set the last updated time
	state.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))

	tflog.Debug(ctx, "Setting state for imported mapping rule", map[string]interface{}{
		"id":          state.ID.ValueString(),
		"name":        state.Name.ValueString(),
		"disabled":    state.Disabled.ValueBool(),
		"has_csv":     !state.CSVData.IsNull() && !state.CSVData.IsUnknown(),
		"has_matchers": !state.Matchers.IsNull() && !state.Matchers.IsUnknown(),
	})

	// Set the state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to set state during import", map[string]interface{}{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Info(ctx, "Successfully imported mapping rule", map[string]interface{}{
		"id": state.ID.ValueString(),
	})
}
