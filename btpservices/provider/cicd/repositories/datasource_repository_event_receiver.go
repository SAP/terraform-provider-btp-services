package cicdrepositories

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	cicdclient "github.com/SAP/terraform-provider-sap-btp-services/internal/cicd/client"
	cicdmodels "github.com/SAP/terraform-provider-sap-btp-services/internal/cicd/models"
	"github.com/SAP/terraform-provider-sap-btp-services/internal/shared"
)

var _ datasource.DataSource = &repositoryEventReceiverDataSource{}
var _ datasource.DataSourceWithConfigure = &repositoryEventReceiverDataSource{}

func NewRepositoryEventReceiverDataSource() datasource.DataSource {
	return &repositoryEventReceiverDataSource{}
}

type repositoryEventReceiverDataSource struct {
	cli *cicdclient.CicdClientFacade
}

func (d *repositoryEventReceiverDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_repository_event_receiver", req.ProviderTypeName)
}

func (d *repositoryEventReceiverDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads the event receiver configuration for a repository in the SAP BTP CI/CD service. Event receivers allow the service to receive webhook events from your source code management system to trigger builds.",
		Attributes: map[string]schema.Attribute{
			"repository": schema.StringAttribute{
				MarkdownDescription: "Name or ID of the repository.",
				Required:            true,
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Whether event processing is enabled for this repository.",
				Computed:            true,
			},
			"scm_type": schema.StringAttribute{
				MarkdownDescription: "Source code manager type. One of: `GITHUB`, `BITBUCKET_CLOUD`, `BITBUCKET`, `GITLAB`, `AZUREREPOS`.",
				Computed:            true,
			},
			"webhook_id": schema.StringAttribute{
				MarkdownDescription: "Immutable technical ID of the webhook assigned by the API.",
				Computed:            true,
			},
			"webhook_token_credential_id": schema.StringAttribute{
				MarkdownDescription: "ID of the webhook secret credential used to authenticate incoming events.",
				Computed:            true,
			},
		},
	}
}

func (d *repositoryEventReceiverDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *repositoryEventReceiverDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config repositoryEventReceiverDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.cli.Repositories.GetEventReceiver(ctx, config.Repository.ValueString())
	if err != nil {
		if cicdmodels.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Repository Not Found",
				fmt.Sprintf("No repository found with reference %q.", config.Repository.ValueString()),
			)
			return
		}
		resp.Diagnostics.AddError("Error Reading Repository Event Receiver", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, repositoryEventReceiverDSValueFrom(config.Repository.ValueString(), *result))...)
}
