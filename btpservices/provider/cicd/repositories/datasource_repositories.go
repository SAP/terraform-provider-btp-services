// btpservices/provider/cicd/repositories/datasource_repositories.go

package cicdrepositories

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cicdclient "github.com/SAP/terraform-provider-btp-services/internal/cicd/client"
	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

var _ datasource.DataSource = &repositoriesDataSource{}
var _ datasource.DataSourceWithConfigure = &repositoriesDataSource{}

func NewRepositoriesDataSource() datasource.DataSource {
	return &repositoriesDataSource{}
}

type repositoriesDataSource struct {
	cli *cicdclient.CicdClientFacade
}

func (d *repositoriesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_repositories", req.ProviderTypeName)
}

func (d *repositoriesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all repositories in the SAP BTP CI/CD service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the data source.",
				Computed:            true,
			},
			"values": schema.ListNestedAttribute{
				MarkdownDescription: "The list of repositories.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Unique identifier of the repository.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the repository.",
							Computed:            true,
						},
						"clone_url": schema.StringAttribute{
							MarkdownDescription: "URL used to clone the repository.",
							Computed:            true,
						},
						"clone_credential_id": schema.StringAttribute{
							MarkdownDescription: "ID of the credential used to authenticate against the Git server when cloning.",
							Computed:            true,
						},
						"cloud_connector_credential_id": schema.StringAttribute{
							MarkdownDescription: "ID of the credential containing the Cloud Connector Location ID used for cloning.",
							Computed:            true,
						},
						"event_receiver": schema.SingleNestedAttribute{
							MarkdownDescription: "Event receiver configuration for triggering builds via webhooks.",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"active": schema.BoolAttribute{
									MarkdownDescription: "Whether event processing is enabled.",
									Computed:            true,
								},
								"scm_type": schema.StringAttribute{
									MarkdownDescription: "Source code manager type.",
									Computed:            true,
								},
								"webhook_id": schema.StringAttribute{
									MarkdownDescription: "Immutable technical ID of the webhook.",
									Computed:            true,
								},
								"webhook_token_credential_id": schema.StringAttribute{
									MarkdownDescription: "ID of the webhook secret credential.",
									Computed:            true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *repositoriesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	clients, ok := req.ProviderData.(*shared.ProviderClients)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *shared.ProviderClients, got: %T", req.ProviderData),
		)
		return
	}
	if clients.Cicd == nil {
		resp.Diagnostics.AddError(
			"Missing CI/CD Configuration",
			"A cicd{} block must be configured in the provider to use CI/CD data sources.",
		)
		return
	}
	d.cli = clients.Cicd
}

func (d *repositoriesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	list, err := d.cli.Repositories.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Listing Repositories", err.Error())
		return
	}

	state := repositoriesDSModel{
		ID: types.StringValue("repositories"),
	}
	values, diags := repositoriesDSItemsFrom(list)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Values = values

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
