package ctmsroutes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/utils"
	ctmsmodels "github.com/SAP/terraform-provider-btp-services/internal/ctms/models"
)

type routeModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	SourceNodeID types.String `tfsdk:"source_node_id"`
	TargetNodeID types.String `tfsdk:"target_node_id"`
}

type routesModel struct {
	Values types.List `tfsdk:"values"`
}

var routeAttrTypes = map[string]attr.Type{
	"id":             types.StringType,
	"name":           types.StringType,
	"description":    types.StringType,
	"source_node_id": types.StringType,
	"target_node_id": types.StringType,
}

func routeValueFrom(ctx context.Context, v ctmsmodels.RouteResponse) (routeModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	// Represent absent descriptions as empty strings for consistent apply,
	// refresh, and import.
	description := ""
	if v.Description != nil {
		description = *v.Description
	}

	m := routeModel{
		ID:           utils.IDValue(v.Id),
		Name:         types.StringPointerValue(v.Name),
		Description:  types.StringValue(description),
		SourceNodeID: utils.IDValue(v.SourceNodeId),
		TargetNodeID: utils.IDValue(v.TargetNodeId),
	}

	if v.Id == nil || v.Name == nil {
		diags.AddError("Incomplete CTMS Response", "The API response must contain an ID and name.")
	}
	_ = ctx

	if v.SourceNodeId == nil || v.TargetNodeId == nil {
		diags.AddError("Incomplete CTMS Route", "The API response must contain sourceNodeId and targetNodeId.")
	}

	return m, diags
}

// Report an ignored reset while preserving the description returned by CTMS.
func routeDescriptionResetDiagnostics(requested, returned types.String) diag.Diagnostics {
	var diags diag.Diagnostics
	if requested.Equal(types.StringValue("")) && returned.ValueString() != "" {
		diags.AddAttributeError(path.Root("description"), "CTMS Route Description Reset Failed",
			"CTMS returned a non-empty description after receiving description: null. "+
				"The API did not confirm the reset. The returned description and route identity have been retained in state.")
	}
	return diags
}
