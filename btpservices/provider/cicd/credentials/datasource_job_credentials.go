package cicdcredentials

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cicdclient "github.com/SAP/terraform-provider-btp-services/internal/cicd/client"
	cicdmodels "github.com/SAP/terraform-provider-btp-services/internal/cicd/models"
	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

var _ datasource.DataSource = &jobCredentialsDataSource{}
var _ datasource.DataSourceWithConfigure = &jobCredentialsDataSource{}

func NewJobCredentialsDataSource() datasource.DataSource {
	return &jobCredentialsDataSource{}
}

type jobCredentialsDataSource struct {
	cli *cicdclient.CicdClientFacade
}

func (d *jobCredentialsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_job_credentials", req.ProviderTypeName)
}

func (d *jobCredentialsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists the credential IDs that a build of a given job is allowed to use in the SAP BTP CI/CD service.",
		Attributes: map[string]schema.Attribute{
			"job": schema.StringAttribute{
				MarkdownDescription: "Name or ID of the job.",
				Required:            true,
			},
			"credential_ids": schema.ListAttribute{
				MarkdownDescription: "The list of credential IDs configured for the job.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *jobCredentialsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *jobCredentialsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config jobCredentialsDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ids, err := d.cli.Credentials.GetJobCredentials(ctx, config.Job.ValueString())
	if err != nil {
		if cicdmodels.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Job Not Found",
				fmt.Sprintf("No job found with reference %q.", config.Job.ValueString()),
			)
			return
		}
		resp.Diagnostics.AddError("Error Reading Job Credentials", err.Error())
		return
	}

	credIDs, diags := types.ListValueFrom(ctx, types.StringType, ids)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, jobCredentialsDSModel{
		Job:           config.Job,
		CredentialIDs: credIDs,
	})...)
}
