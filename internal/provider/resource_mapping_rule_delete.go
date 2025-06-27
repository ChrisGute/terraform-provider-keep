package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Delete deletes the resource and removes the Terraform state on success.
func (r *mappingRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state mappingRuleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the mapping rule
	err := r.client.DeleteMappingRule(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting mapping rule",
			fmt.Sprintf("Could not delete mapping rule, unexpected error: %s", err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "Successfully deleted mapping rule", map[string]interface{}{
		"id": state.ID.ValueString(),
	})
}
