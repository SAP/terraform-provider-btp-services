package ctmsdestinations

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	ctmsmodels "github.com/SAP/terraform-provider-btp-services/internal/ctms/models"
)

type destinationModel struct {
	Name          types.String `tfsdk:"name"`
	URL           types.String `tfsdk:"url"`
	Description   types.String `tfsdk:"description"`
	PlatformTypes types.List   `tfsdk:"platform_types"`
	Application   types.String `tfsdk:"application"`
	DataCenter    types.String `tfsdk:"data_center"`
}

type destinationDetailsModel struct {
	destinationModel
	TenantID        types.String `tfsdk:"tenant_id"`
	IncludeTenantID types.Bool   `tfsdk:"include_tenant_id"`
}

type destinationsModel struct {
	Values types.List `tfsdk:"values"`
}

var destinationAttrTypes = map[string]attr.Type{
	"name":           types.StringType,
	"url":            types.StringType,
	"description":    types.StringType,
	"platform_types": types.ListType{ElemType: types.StringType},
	"application":    types.StringType,
	"data_center":    types.StringType,
}

func destinationValueFrom(ctx context.Context, v ctmsmodels.DestinationBase) (destinationModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	platforms := types.ListNull(types.StringType)
	if v.PlatformTypes != nil {
		var d diag.Diagnostics
		platforms, d = types.ListValueFrom(ctx, types.StringType, *v.PlatformTypes)
		diags.Append(d...)
	}

	if v.Name == nil {
		diags.AddError("Incomplete CTMS Destination", "The API returned a destination without a name.")
	}

	return destinationModel{
		Name:          types.StringPointerValue(v.Name),
		URL:           types.StringPointerValue(v.Url),
		Description:   types.StringPointerValue(v.Description),
		PlatformTypes: platforms,
		Application:   types.StringPointerValue(v.Application),
		DataCenter:    types.StringPointerValue(v.DataCenter),
	}, diags
}

func destinationDetailsValueFrom(ctx context.Context, v ctmsmodels.DestinationExpanded, includeTenantID types.Bool) (destinationDetailsModel, diag.Diagnostics) {
	base := ctmsmodels.DestinationBase{
		Name:          v.Name,
		Url:           v.Url,
		Description:   v.Description,
		PlatformTypes: v.PlatformTypes,
		Application:   v.Application,
		DataCenter:    v.DataCenter,
	}
	model, diags := destinationValueFrom(ctx, base)

	return destinationDetailsModel{
		destinationModel: model,
		TenantID:         types.StringPointerValue(v.TenantId),
		IncludeTenantID:  includeTenantID,
	}, diags
}
