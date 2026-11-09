package ctmsnodes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/utils"
	ctmsmodels "github.com/SAP/terraform-provider-btp-services/internal/ctms/models"
)

type nodeModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Description          types.String `tfsdk:"description"`
	UploadAllowed        types.Bool   `tfsdk:"upload_allowed"`
	ForwardMode          types.String `tfsdk:"forward_mode"`
	NotificationEnabled  types.Bool   `tfsdk:"notification_enabled"`
	ImportDisabled       types.Bool   `tfsdk:"import_disabled"`
	ImportDisabledReason types.String `tfsdk:"import_disabled_reason"`
	Virtual              types.Bool   `tfsdk:"virtual"`
	AutoImport           types.Bool   `tfsdk:"auto_import"`
	ControlledBy         types.String `tfsdk:"controlled_by"`
	JobID                types.Int64  `tfsdk:"job_id"`
	Targets              types.List   `tfsdk:"targets"`
}

type nodesModel struct {
	Values types.List `tfsdk:"values"`
}

var nodeAttrTypes = map[string]attr.Type{
	"id":                     types.StringType,
	"name":                   types.StringType,
	"description":            types.StringType,
	"upload_allowed":         types.BoolType,
	"forward_mode":           types.StringType,
	"notification_enabled":   types.BoolType,
	"import_disabled":        types.BoolType,
	"import_disabled_reason": types.StringType,
	"virtual":                types.BoolType,
	"auto_import":            types.BoolType,
	"controlled_by":          types.StringType,
	"job_id":                 types.Int64Type,
	"targets":                types.ListType{ElemType: targetType},
}

func nodeValueFrom(ctx context.Context, v ctmsmodels.NodeResponseTarget) (nodeModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	m := nodeModel{
		ID:                   utils.IDValue(v.Id),
		Name:                 types.StringPointerValue(v.Name),
		Description:          types.StringPointerValue(v.Description),
		UploadAllowed:        types.BoolPointerValue(v.UploadAllowed),
		ForwardMode:          utils.EnumValue(v.ForwardMode),
		NotificationEnabled:  types.BoolPointerValue(v.NotificationEnabled),
		ImportDisabled:       types.BoolPointerValue(v.ImportDisabled),
		ImportDisabledReason: types.StringPointerValue(v.ImportDisabledReason),
		Virtual:              types.BoolPointerValue(v.Virtual),
		AutoImport:           types.BoolPointerValue(v.AutoImport),
		ControlledBy:         utils.EnumValue(v.ControlledBy),
		JobID:                types.Int64PointerValue(v.JobId),
	}
	if v.Id == nil || v.Name == nil {
		diags.AddError("Incomplete CTMS Response", "The API response must contain an ID and name.")
	}
	var d diag.Diagnostics
	m.Targets, d = targetsValueFrom(ctx, v.Targets)
	diags.Append(d...)
	return m, diags
}

var importOptionsType = types.ObjectType{AttrTypes: map[string]attr.Type{"strategy": types.StringType}}
var targetType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"content_type":     types.StringType,
		"destination_name": types.StringType,
		"import_options":   importOptionsType,
	},
}

type targetModel struct {
	ContentType     types.String `tfsdk:"content_type"`
	DestinationName types.String `tfsdk:"destination_name"`
	ImportOptions   types.Object `tfsdk:"import_options"`
}

func targetsValueFrom(ctx context.Context, targets *[]ctmsmodels.DeployTarget) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	if targets == nil {
		return types.ListNull(targetType), diags
	}
	values := make([]targetModel, 0, len(*targets))
	for _, target := range *targets {
		options := types.ObjectNull(importOptionsType.AttrTypes)
		if target.ImportOptions != nil {
			var d diag.Diagnostics
			options, d = types.ObjectValue(
				importOptionsType.AttrTypes,
				map[string]attr.Value{"strategy": utils.EnumValue(target.ImportOptions.Strategy)},
			)
			diags.Append(d...)
		}
		if target.ContentType == nil {
			diags.AddError("Incomplete CTMS Target", "A returned deployment target is missing contentType.")
		}
		values = append(
			values,
			targetModel{
				ContentType:     utils.EnumValue(target.ContentType),
				DestinationName: types.StringValue(target.DestinationName),
				ImportOptions:   options,
			},
		)
	}
	result, d := types.ListValueFrom(ctx, targetType, values)
	diags.Append(d...)
	return result, diags
}

func targetsRequest(ctx context.Context, value types.List) (*[]ctmsmodels.DeployTarget, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() || value.IsUnknown() {
		return nil, diags
	}
	var targets []targetModel
	diags.Append(value.ElementsAs(ctx, &targets, false)...)
	result := make([]ctmsmodels.DeployTarget, 0, len(targets))
	for _, target := range targets {
		entry := ctmsmodels.DeployTarget{
			ContentType:     utils.StringPointer[ctmsmodels.ContentType](target.ContentType),
			DestinationName: target.DestinationName.ValueString(),
		}
		if !target.ImportOptions.IsNull() && !target.ImportOptions.IsUnknown() {
			strategy, ok := target.ImportOptions.Attributes()["strategy"].(types.String)
			if !ok {
				diags.AddError("Invalid CTMS Import Options", "Expected a string strategy.")
				continue
			}
			entry.ImportOptions = &ctmsmodels.ImportOptions{Strategy: utils.StringPointer[ctmsmodels.Strategy](strategy)}
		}
		result = append(result, entry)
	}
	return &result, diags
}
