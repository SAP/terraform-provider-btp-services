package ctmsroutes

import (
	"cmp"
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/utils"
	ctmsmodels "github.com/SAP/terraform-provider-btp-services/internal/ctms/models"
)

var _ datasource.DataSourceWithConfigure = &routesDataSource{}

type routesDataSource struct {
	utils.Connection
}

func NewRoutesDataSource() datasource.DataSource {
	return &routesDataSource{}
}

func (d *routesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ctms_routes"
}

func (d *routesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads all CTMS routes, ordered by numeric ID.",
		Attributes: map[string]schema.Attribute{
			"values": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Routes in numeric ID order.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Decimal route ID assigned by CTMS.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Route name. Supports in-place updates.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Route description. Omit to leave it managed by CTMS.",
						},
						"source_node_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Decimal source node ID. A source can have multiple outgoing routes.",
						},
						"target_node_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Decimal target node ID. A node can be the target of only one route.",
						},
					},
				},
			},
		},
	}
}

func (d *routesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(d.Connection.Configure(req.ProviderData)...)
}

func (d *routesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	items, err := d.Client.Routes.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Listing CTMS Routes", err.Error())
		return
	}
	slices.SortFunc(items, func(a, b ctmsmodels.RouteResponse) int {
		return cmp.Compare(*a.Id, *b.Id)
	})

	models := make([]routeModel, 0, len(items))
	for _, item := range items {
		model, diags := routeValueFrom(ctx, item)
		resp.Diagnostics.Append(diags...)
		models = append(models, model)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	values, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: routeAttrTypes}, models)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := routesModel{
		Values: values,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
