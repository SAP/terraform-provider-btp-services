package ctmsdestinations

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

var _ datasource.DataSourceWithConfigure = &destinationsDataSource{}

type destinationsDataSource struct {
	utils.Connection
}

func NewDestinationsDataSource() datasource.DataSource {
	return &destinationsDataSource{}
}

func (d *destinationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ctms_destinations"
}

func (d *destinationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads all transport destination summaries in name order.",
		Attributes: map[string]schema.Attribute{
			"values": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Destination summaries.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Destination name.",
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
					},
				},
			},
		},
	}
}

func (d *destinationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(d.Connection.Configure(req.ProviderData)...)
}

func (d *destinationsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	items, err := d.Client.Destinations.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Listing CTMS Destinations", err.Error())
		return
	}

	slices.SortFunc(
		items,
		func(a, b ctmsmodels.DestinationBase) int {
			return cmp.Compare(types.StringPointerValue(a.Name).ValueString(), types.StringPointerValue(b.Name).ValueString())
		},
	)

	models := make([]destinationModel, 0, len(items))

	for _, item := range items {
		model, diags := destinationValueFrom(ctx, item)
		resp.Diagnostics.Append(diags...)
		models = append(models, model)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	values, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: destinationAttrTypes}, models)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := destinationsModel{
		Values: values,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
