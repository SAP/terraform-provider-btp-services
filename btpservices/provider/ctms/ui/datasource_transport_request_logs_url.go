package ctmsui

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/utils"
)

var _ datasource.DataSourceWithConfigure = &transportRequestLogsURLDataSource{}

type transportRequestLogsURLDataSource struct {
	utils.Connection
}

func NewTransportRequestLogsURLDataSource() datasource.DataSource {
	return &transportRequestLogsURLDataSource{}
}

func (d *transportRequestLogsURLDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ctms_transport_request_logs_url"
}

func (d *transportRequestLogsURLDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads the CTMS UI link to transport request logs; it does not retrieve log contents.",
		Attributes: map[string]schema.Attribute{
			"transport_request_id": schema.StringAttribute{
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
				MarkdownDescription: "Transport request identifier. This endpoint accepts a string.",
			},
			"node_id": schema.StringAttribute{
				Optional:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
				MarkdownDescription: "Optional node identifier to link to the import queue logs view. This endpoint accepts a string.",
			},
			"url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "URL returned by CTMS.",
			},
		},
	}
}

func (d *transportRequestLogsURLDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(d.Connection.Configure(req.ProviderData)...)
}

func (d *transportRequestLogsURLDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config transportRequestLogsURLModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.Client.UI.GetTransportRequestLogsURL(ctx, config.TransportRequestID.ValueString(), utils.StringPointer[string](config.NodeID))
	if err != nil {
		resp.Diagnostics.AddError("Error Reading CTMS Logs URL", err.Error())
		return
	}

	if result.TmsUiLogUrl == nil || *result.TmsUiLogUrl == "" {
		resp.Diagnostics.AddError("Incomplete CTMS URL Response", "The API did not return tmsUiLogUrl.")
		return
	}

	config.URL = types.StringPointerValue(result.TmsUiLogUrl)
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
