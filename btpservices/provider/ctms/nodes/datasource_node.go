package ctmsnodes

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

var _ datasource.DataSourceWithConfigure = &nodeDataSource{}

type nodeDataSource struct {
	utils.Connection
}

func NewNodeDataSource() datasource.DataSource {
	return &nodeDataSource{}
}

func (d *nodeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ctms_node"
}

func (d *nodeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a CTMS node by decimal ID or exact name. Configure exactly one selector.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Decimal node ID. Configure exactly one of id or name.",
				Validators: []validator.String{
					utils.IDValidator(),
					stringvalidator.ExactlyOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Exact node name. Configure exactly one of id or name.",
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
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
	}
}

func (d *nodeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(d.Connection.Configure(req.ProviderData)...)
}

func (d *nodeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config nodeDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result *ctmsmodels.NodeResponseTarget
	var err error
	if !config.ID.IsNull() {
		var id int64
		id, err = utils.ParseID(config.ID)
		if err == nil {
			result, err = d.Client.Nodes.Get(ctx, id)
		}
	} else {
		var items []ctmsmodels.NodeResponseTarget
		items, err = d.Client.Nodes.List(ctx)
		if err == nil {
			matches := 0
			for _, item := range items {
				if item.Name != nil && *item.Name == config.Name.ValueString() {
					matches++
					result = &item
				}
			}
			if matches != 1 {
				err = utils.MissingObject("node", matches)
			}
		}
	}

	if err != nil {
		resp.Diagnostics.AddError("Error Reading CTMS Node", err.Error())
		return
	}

	model, diags := nodeDataSourceValueFrom(ctx, *result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
