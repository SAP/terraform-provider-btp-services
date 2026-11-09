package ctmsui

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/utils"
)

var _ datasource.DataSourceWithConfigure = &uiURLDataSource{}

type uiURLDataSource struct {
	utils.Connection
}

func NewUIURLDataSource() datasource.DataSource {
	return &uiURLDataSource{}
}

func (d *uiURLDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ctms_ui_url"
}

func (d *uiURLDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads the tenant-specific CTMS UI URL.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "URL returned by CTMS.",
			},
		},
	}
}

func (d *uiURLDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(d.Connection.Configure(req.ProviderData)...)
}

func (d *uiURLDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.Client.UI.GetURL(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading CTMS UI URL", err.Error())
		return
	}

	if result.TmsUiUrl == nil || *result.TmsUiUrl == "" {
		resp.Diagnostics.AddError("Incomplete CTMS URL Response", "The API did not return tmsUiUrl.")
		return
	}

	state := uiURLModel{
		URL: types.StringPointerValue(result.TmsUiUrl),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
