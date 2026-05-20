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

var _ datasource.DataSource = &repositoryWebhookConfigDataSource{}
var _ datasource.DataSourceWithConfigure = &repositoryWebhookConfigDataSource{}

func NewRepositoryWebhookConfigDataSource() datasource.DataSource {
	return &repositoryWebhookConfigDataSource{}
}

type repositoryWebhookConfigDataSource struct {
	cli *cicdclient.CicdClientFacade
}

func (d *repositoryWebhookConfigDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_repository_webhook_config", req.ProviderTypeName)
}

func (d *repositoryWebhookConfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads the webhook configuration for a repository in the SAP BTP CI/CD service. The webhook URI is the target address to configure in your source code management system to send push events that trigger builds.",
		Attributes: map[string]schema.Attribute{
			"repository": schema.StringAttribute{
				MarkdownDescription: "Name or ID of the repository.",
				Required:            true,
			},
			"webhook_uri": schema.StringAttribute{
				MarkdownDescription: "The URI to configure in your SCM webhook to send events to the CI/CD service.",
				Computed:            true,
			},
		},
	}
}

func (d *repositoryWebhookConfigDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *repositoryWebhookConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config repositoryWebhookConfigDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if d.cli == nil {
		resp.Diagnostics.AddError(
			"Missing CI/CD Configuration",
			"A cicd{} block must be configured in the provider to use CI/CD data sources.",
		)
		return
	}

	result, err := d.cli.Repositories.GetWebhookConfig(ctx, config.Repository.ValueString())
	if err != nil {
		if cicdmodels.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Repository Not Found",
				fmt.Sprintf("No repository found with reference %q.", config.Repository.ValueString()),
			)
			return
		}
		resp.Diagnostics.AddError("Error Reading Repository Webhook Config", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, repositoryWebhookConfigDSValueFrom(config.Repository.ValueString(), *result))...)
}
