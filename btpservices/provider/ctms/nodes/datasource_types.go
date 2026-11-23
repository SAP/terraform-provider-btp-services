package ctmsnodes

import (
	"context"
	"maps"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	ctmsmodels "github.com/SAP/terraform-provider-btp-services/internal/ctms/models"
)

// Import status is exposed by data sources without changing the managed resource model.
type nodeDataSourceModel struct {
	nodeModel
	ImportDisabled       types.Bool   `tfsdk:"import_disabled"`
	ImportDisabledReason types.String `tfsdk:"import_disabled_reason"`
}

var nodeDataSourceAttrTypes = func() map[string]attr.Type {
	attributes := maps.Clone(nodeAttrTypes)
	attributes["import_disabled"] = types.BoolType
	attributes["import_disabled_reason"] = types.StringType
	return attributes
}()

func nodeDataSourceValueFrom(ctx context.Context, v ctmsmodels.NodeResponseTarget) (nodeDataSourceModel, diag.Diagnostics) {
	model, diags := nodeValueFrom(ctx, v)
	return nodeDataSourceModel{
		nodeModel:            model,
		ImportDisabled:       types.BoolPointerValue(v.ImportDisabled),
		ImportDisabledReason: types.StringPointerValue(v.ImportDisabledReason),
	}, diags
}
