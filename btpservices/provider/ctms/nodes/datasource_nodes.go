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
							MarkdownDescription: "Case-sensitive node name. Changing it replaces the node.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Node description. Omit to leave it managed by CTMS.",
						},
						"upload_allowed": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Allow file uploads to this node.",
						},
						"forward_mode": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Forwarding mode: AUTO (pre-import), POST_IMPORT, ON_SUCCESS, or MANUAL.",
						},
						"notification_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Enable notifications. Requires an Alert Notification subscription and configuration in the subaccount.",
						},
						"import_disabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Disable imports, including scheduled and automatic imports; forwarding can still occur.",
						},
						"import_disabled_reason": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Optional reason for disabling imports.",
						},
						"virtual": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Virtual nodes cannot have deployment targets.",
						},
						"auto_import": schema.BoolAttribute{
							Computed: true,
							MarkdownDescription: "Automatically import eligible requests. Enabling this can immediately import queued requests and " +
								"deactivate existing schedules. It can coexist with import_disabled; imports resume when that flag is " +
								"cleared.",
						},
						"controlled_by": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Application controlling the node. Read-only until the API reset contract is confirmed.",
						},
						"job_id": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Import scheduler job ID, if present.",
						},
						"targets": schema.ListNestedAttribute{
							Computed: true,
							MarkdownDescription: "Deployment targets in API order. Omit to leave targets managed by CTMS; an explicit empty list " +
								"requests clearing them. Virtual nodes cannot have targets.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"content_type": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Content type.",
									},
									"destination_name": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Name of an existing transport destination.",
									},
									"import_options": schema.SingleNestedAttribute{
										Computed:            true,
										MarkdownDescription: "MTA deployment options. The destination must address Cloud Foundry.",
										Attributes: map[string]schema.Attribute{
											"strategy": schema.StringAttribute{
												Computed:            true,
												MarkdownDescription: "Deployment strategy: default or blue-green. Only for MTA deployment to Cloud Foundry.",
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

	models := make([]nodeModel, 0, len(items))
	for _, item := range items {
		model, diags := nodeValueFrom(ctx, item)
		resp.Diagnostics.Append(diags...)
		models = append(models, model)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	values, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: nodeAttrTypes}, models)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := nodesModel{
		Values: values,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
