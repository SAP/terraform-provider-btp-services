package ctmsnodes

import (
	"cmp"
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/utils"
	ctmsmodels "github.com/SAP/terraform-provider-btp-services/internal/ctms/models"
)

var _ list.ListResourceWithConfigure = &nodeListResource{}

type nodeListResource struct{ utils.Connection }

func NewNodeListResource() list.ListResource {
	return &nodeListResource{}
}

func (r *nodeListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ctms_node"
}

func (r *nodeListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(r.Connection.Configure(req.ProviderData)...)
}

func (r *nodeListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		MarkdownDescription: "Discovers CTMS nodes for import. Filters use exact matches and are evaluated locally with AND semantics.",
		Attributes: map[string]listschema.Attribute{
			"name": listschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Exact node name.",
			},
		},
	}
}

func (r *nodeListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var config struct {
		Name types.String `tfsdk:"name"`
	}
	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}
	items, err := r.Client.Nodes.List(ctx)
	if err != nil {
		diags.AddError("Error Listing CTMS Nodes", err.Error())
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}
	slices.SortFunc(items, func(a, b ctmsmodels.NodeResponseTarget) int {
		return cmp.Compare(*a.Id, *b.Id)
	})
	stream.Results = func(push func(list.ListResult) bool) {
		var count int64
		for _, item := range items {
			if err := ctx.Err(); err != nil {
				var d diag.Diagnostics
				d.AddError("CTMS Discovery Cancelled", err.Error())
				push(list.ListResult{Diagnostics: d})
				return
			}
			if !config.Name.IsNull() && (item.Name == nil || *item.Name != config.Name.ValueString()) {
				continue
			}

			if req.Limit > 0 && count >= req.Limit {
				return
			}
			result := req.NewListResult(ctx)
			if item.Name != nil {
				result.DisplayName = *item.Name
			} else {
				result.DisplayName = utils.IDValue(item.Id).ValueString()
			}
			result.Diagnostics.Append(result.Identity.Set(ctx, r.Identity(utils.IDValue(item.Id)))...)
			if req.IncludeResource {
				model, d := nodeValueFrom(ctx, item)
				result.Diagnostics.Append(d...)
				result.Diagnostics.Append(result.Resource.Set(ctx, model)...)
			}
			count++
			if !push(result) || result.Diagnostics.HasError() {
				return
			}
		}
	}
}
