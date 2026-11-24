package ctmsdestinations

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/SAP/terraform-provider-btp-services/btpservices/provider/ctms/utils"
)

var _ datasource.DataSourceWithConfigure = &destinationDataSource{}

type destinationDataSource struct {
	utils.Connection
}

func NewDestinationDataSource() datasource.DataSource {
	return &destinationDataSource{}
}

func (d *destinationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ctms_destination"
}

func (d *destinationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing transport destination. CTMS does not manage destination credentials through this API.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
				MarkdownDescription: "Exact destination name.",
			},
			"url": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Destination URL. Redacted because URLs may contain credentials; sensitive output is still stored in state.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Destination description, if returned.",
			},
			"platform_types": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Platform types in API order.",
			},
			"application": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Application, if returned.",
			},
			"data_center": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Data center, if returned.",
			},
			"include_tenant_id": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Request the destination tenant ID. Defaults to false when omitted.",
			},
			"tenant_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Destination tenant ID, if requested and returned. This is not the CTMS provider's subaccount scope.",
			},
		},
	}
}

func (d *destinationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(d.Connection.Configure(req.ProviderData)...)
}

func (d *destinationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config destinationDetailsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.Client.Destinations.Get(ctx, config.Name.ValueString(), config.IncludeTenantID.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading CTMS Destination", err.Error())
		return
	}

	if result.Name == nil || *result.Name != config.Name.ValueString() {
		resp.Diagnostics.AddError("Unexpected CTMS Destination", "The API response does not match the requested destination name.")
		return
	}

	state, diags := destinationDetailsValueFrom(ctx, *result, config.IncludeTenantID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
