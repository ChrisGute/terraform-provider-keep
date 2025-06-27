package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/keephq/terraform-provider-keep/internal/client"
)

type AlertResource struct {
	client *client.Client
}

type AlertResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Status      types.String `tfsdk:"status"`
	Severity    types.String `tfsdk:"severity"`
	Environment types.String `tfsdk:"environment"`
	Service     types.String `tfsdk:"service"`
	Source      types.List   `tfsdk:"source"`
	Message     types.String `tfsdk:"message"`
	Description types.String `tfsdk:"description"`
	URL         types.String `tfsdk:"url"`
	ImageURL    types.String `tfsdk:"image_url"`
	Labels      types.Map    `tfsdk:"labels"`
	Fingerprint types.String `tfsdk:"fingerprint"`
	LastReceived types.String `tfsdk:"last_received"`
}

// toClientAlert converts the Terraform model to a client.Alert
func (m *AlertResourceModel) toClientAlert(ctx context.Context) (*client.Alert, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Convert source list to []string
	sources := make([]string, 0, len(m.Source.Elements()))
	diags.Append(m.Source.ElementsAs(ctx, &sources, false)...)
	if diags.HasError() {
		return nil, diags
	}

	// Convert labels map to map[string]string
	labels := make(map[string]string)
	diags.Append(m.Labels.ElementsAs(ctx, &labels, false)...)
	if diags.HasError() {
		return nil, diags
	}

	// Set default status if not provided
	status := m.Status.ValueString()
	if status == "" {
		status = "firing"
	}

	// Set default severity if not provided
	severity := m.Severity.ValueString()
	if severity == "" {
		severity = "critical"
	}

	// Set last received time
	lastReceived := m.LastReceived.ValueString()
	if lastReceived == "" {
		lastReceived = time.Now().UTC().Format(time.RFC3339)
	}

	return &client.Alert{
		ID:           m.ID.ValueString(),
		Fingerprint:  m.Fingerprint.ValueString(),
		Name:         m.Name.ValueString(),
		Status:       status,
		Severity:     severity,
		Environment:  m.Environment.ValueString(),
		Service:      m.Service.ValueString(),
		Source:       sources,
		Message:      m.Message.ValueString(),
		Description:  m.Description.ValueString(),
		URL:          m.URL.ValueString(),
		ImageURL:     m.ImageURL.ValueString(),
		Labels:       labels,
		LastReceived: lastReceived,
	}, diags
}

// fromClientAlert converts a client.Alert to the Terraform model
func (m *AlertResourceModel) fromClientAlert(ctx context.Context, alert map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Set basic fields
	m.ID = types.StringValue(alert["id"].(string))
	m.Fingerprint = types.StringValue(alert["fingerprint"].(string))
	m.Name = types.StringValue(alert["name"].(string))
	m.Status = types.StringValue(alert["status"].(string))
	m.Severity = types.StringValue(alert["severity"].(string))
	m.Environment = types.StringValue(alert["environment"].(string))
	m.Service = types.StringValue(alert["service"].(string))
	m.Message = types.StringValue(alert["message"].(string))
	m.Description = types.StringValue(alert["description"].(string))
	m.URL = types.StringValue(alert["url"].(string))
	m.ImageURL = types.StringValue(alert["image_url"].(string))
	m.LastReceived = types.StringValue(alert["lastReceived"].(string))

	// Convert source slice to List
	sources := make([]string, 0)
	if src, ok := alert["source"].([]interface{}); ok {
		for _, s := range src {
			sources = append(sources, s.(string))
		}
	}
	sourceList, d := types.ListValueFrom(ctx, types.StringType, sources)
	diags.Append(d...)
	m.Source = sourceList

	// Convert labels map to Map
	labels := make(map[string]string)
	if l, ok := alert["labels"].(map[string]interface{}); ok {
		for k, v := range l {
			if s, ok := v.(string); ok {
				labels[k] = s
			}
		}
	}
	labelMap, d := types.MapValueFrom(ctx, types.StringType, labels)
	diags.Append(d...)
	m.Labels = labelMap

	return diags
}

func NewAlertResource() resource.Resource {
	return &AlertResource{}
}

func (r *AlertResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert"
}

func (r *AlertResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Alert resource for managing alerts in KeepHQ",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the alert",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the alert",
			},
			"status": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The status of the alert (firing, resolved, acknowledged, suppressed, pending)",
			},
			"severity": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The severity of the alert (critical, high, warning, info, low)",
			},
			"environment": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The environment of the alert",
			},
			"service": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The service associated with the alert",
			},
			"source": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "The source(s) of the alert",
			},
			"message": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The message of the alert",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The description of the alert",
			},
			"url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The URL associated with the alert",
			},
			"image_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The URL of an image associated with the alert",
			},
			"labels": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Labels to attach to the alert",
			},
			"fingerprint": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The fingerprint of the alert (used for deduplication)",
			},
			"last_received": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The timestamp when the alert was last received",
			},
		},
	}
}

func (r *AlertResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AlertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AlertResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to client alert
	alert, diags := data.toClientAlert(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call the API to create the alert
	result, err := r.client.CreateAlert(ctx, *alert)
	if err != nil {
		resp.Diagnostics.AddError("Error creating alert", err.Error())
		return
	}

	// Set the ID and fingerprint from the response
	if id, ok := result["id"].(string); ok {
		data.ID = types.StringValue(id)
	}

	if fingerprint, ok := result["fingerprint"].(string); ok {
		data.Fingerprint = types.StringValue(fingerprint)
	}

	// Set the state with the populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AlertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AlertResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the fingerprint from the state
	fingerprint := data.Fingerprint.ValueString()
	if fingerprint == "" {
		resp.Diagnostics.AddError("Alert fingerprint is required", "Alert fingerprint is required to read the alert")
		return
	}

	// Call the API to get the alert
	alert, err := r.client.GetAlert(ctx, fingerprint)
	if err != nil {
		resp.Diagnostics.AddError("Error reading alert", err.Error())
		return
	}

	// Update the model with the response data
	diags := data.fromClientAlert(ctx, alert)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AlertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AlertResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to client alert
	alert, diags := data.toClientAlert(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare the enrich data
	enrichData := map[string]interface{}{
		"fingerprint": alert.Fingerprint,
		"status":      alert.Status,
		"severity":    alert.Severity,
		"message":     alert.Message,
		"description": alert.Description,
		"labels":      alert.Labels,
	}

	// Call the API to update the alert
	_, err := r.client.EnrichAlert(ctx, alert.Fingerprint, enrichData)
	if err != nil {
		resp.Diagnostics.AddError("Error updating alert", err.Error())
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AlertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AlertResourceModel

	// Read Terraform prior state data into the model
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the fingerprint from the state
	fingerprint := data.Fingerprint.ValueString()
	if fingerprint == "" {
		resp.Diagnostics.AddError("Alert fingerprint is required", "Alert fingerprint is required to delete the alert")
		return
	}

	// Call the API to delete the alert
	err := r.client.DeleteAlert(ctx, fingerprint)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting alert", err.Error())
		return
	}
}

func (r *AlertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The import ID is expected to be the alert's fingerprint
	resource.ImportStatePassthroughID(ctx, path.Root("fingerprint"), req, resp)
}
