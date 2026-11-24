package ctmsnodes

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

var _ datasource.DataSourceWithConfigure = &nodesDataSource{}

type nodesDataSource struct {
	utils.Connection
}

func NewNodesDataSource() datasource.DataSource {
	return &nodesDataSource{}
}

func (d *nodesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ctms_nodes"
}

func (d *nodesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads all CTMS nodes, ordered by numeric ID.",
		Attributes: map[string]schema.Attribute{
			"values": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Nodes in numeric ID order.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Decimal node ID assigned by CTMS.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Case-sensitive node name returned by CTMS.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Node description returned by CTMS.",
						},
						"upload_allowed": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether file uploads to this node are allowed.",
						},
						"forward_mode": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Forwarding mode: AUTO (pre-import), POST_IMPORT, ON_SUCCESS, or MANUAL.",
						},
						"notification_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether notifications are enabled for the node.",
						},
						"import_disabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether import is disabled for the node.",
						},
						"import_disabled_reason": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Reason import is disabled for the node.",
						},
						"virtual": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the node is virtual. Virtual nodes cannot have deployment targets.",
						},
						"auto_import": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether automatic import is enabled for the node.",
						},
						"controlled_by": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Application controlling the node, if returned by CTMS.",
						},
						"job_id": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Import scheduler job ID, if present.",
						},
						"targets": schema.ListNestedAttribute{
							Computed:            true,
							MarkdownDescription: "Deployment targets returned by CTMS, in API order.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"content_type": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Content type: MTA, XSCDU, APP, or BTP_ABAP.",
									},
									"destination_name": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Transport destination name returned by CTMS.",
									},
									"import_options": schema.SingleNestedAttribute{
										Computed:            true,
										MarkdownDescription: "Deployment import options returned by CTMS, if present.",
										Attributes: map[string]schema.Attribute{
											"strategy": schema.StringAttribute{
												Computed:            true,
												MarkdownDescription: "Deployment strategy returned by CTMS: default or blue-green.",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *nodesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(d.Connection.Configure(req.ProviderData)...)
}

func (d *nodesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	items, err := d.Client.Nodes.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Listing CTMS Nodes", err.Error())
		return
	}
	slices.SortFunc(items, func(a, b ctmsmodels.NodeResponseTarget) int {
		return cmp.Compare(*a.Id, *b.Id)
	})

	models := make([]nodeDataSourceModel, 0, len(items))
	for _, item := range items {
		model, diags := nodeDataSourceValueFrom(ctx, item)
		resp.Diagnostics.Append(diags...)
		models = append(models, model)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	values, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: nodeDataSourceAttrTypes}, models)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := nodesModel{
		Values: values,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
