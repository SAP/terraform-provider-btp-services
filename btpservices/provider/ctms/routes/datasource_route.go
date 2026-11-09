package ctmsroutes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/utils"
	ctmsmodels "github.com/SAP/terraform-provider-btp-services/internal/ctms/models"
)

var _ datasource.DataSourceWithConfigure = &routeDataSource{}

type routeDataSource struct {
	utils.Connection
}

func NewRouteDataSource() datasource.DataSource {
	return &routeDataSource{}
}

func (d *routeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ctms_route"
}

func (d *routeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a CTMS route by decimal ID or exact name. Configure exactly one selector.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Decimal route ID. Configure exactly one of id or name.",
				Validators: []validator.String{
					utils.IDValidator(),
					stringvalidator.ExactlyOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Exact route name. Configure exactly one of id or name.",
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
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
	}
}

func (d *routeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(d.Connection.Configure(req.ProviderData)...)
}

func (d *routeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config routeModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result *ctmsmodels.RouteResponse
	var err error
	if !config.ID.IsNull() {
		var id int64
		id, err = utils.ParseID(config.ID)
		if err == nil {
			result, err = d.Client.Routes.Get(ctx, id)
		}
	} else {
		var items []ctmsmodels.RouteResponse
		items, err = d.Client.Routes.List(ctx)
		if err == nil {
			matches := 0
			for _, item := range items {
				if item.Name != nil && *item.Name == config.Name.ValueString() {
					matches++
					result = &item
				}
			}
			if matches != 1 {
				err = utils.MissingObject("route", matches)
			}
		}
	}

	if err != nil {
		resp.Diagnostics.AddError("Error Reading CTMS Route", err.Error())
		return
	}

	model, diags := routeValueFrom(ctx, *result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
