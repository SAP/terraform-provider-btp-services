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
	}
}

func (d *nodeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(d.Connection.Configure(req.ProviderData)...)
}

func (d *nodeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config nodeModel

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

	model, diags := nodeValueFrom(ctx, *result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
