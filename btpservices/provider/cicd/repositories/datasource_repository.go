// btpservices/provider/cicd/repositories/datasource_repository.go

package cicdrepositories

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	cicdclient "github.com/SAP/terraform-provider-btp-services/internal/cicd/client"
	cicdmodels "github.com/SAP/terraform-provider-btp-services/internal/cicd/models"
	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

var _ datasource.DataSource = &repositoryDataSource{}
var _ datasource.DataSourceWithConfigure = &repositoryDataSource{}

func NewRepositoryDataSource() datasource.DataSource {
	return &repositoryDataSource{}
}

type repositoryDataSource struct {
	cli *cicdclient.CicdClientFacade
}

func (d *repositoryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_repository", req.ProviderTypeName)
}

func (d *repositoryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a repository from the SAP BTP CI/CD service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the repository. Exactly one of `id` or `name` must be set.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("name"),
					}...),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the repository. Exactly one of `id` or `name` must be set.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("id"),
					}...),
				},
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
	}
}

func (d *repositoryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *repositoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config repositoryDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var reference string
	if !config.ID.IsNull() && !config.ID.IsUnknown() {
		reference = config.ID.ValueString()
	} else {
		reference = config.Name.ValueString()
	}

	result, err := d.cli.Repositories.Get(ctx, reference)
	if err != nil {
		if cicdmodels.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Repository Not Found",
				fmt.Sprintf("No repository found with reference %q.", reference),
			)
			return
		}
		resp.Diagnostics.AddError("Error Reading Repository", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, repositoryDSValueFrom(*result))...)
}
