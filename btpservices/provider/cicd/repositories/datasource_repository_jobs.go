// btpservices/provider/cicd/repositories/datasource_repository_jobs.go

package cicdrepositories

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

var _ datasource.DataSource = &repositoryJobsDataSource{}
var _ datasource.DataSourceWithConfigure = &repositoryJobsDataSource{}

func NewRepositoryJobsDataSource() datasource.DataSource {
	return &repositoryJobsDataSource{}
}

type repositoryJobsDataSource struct {
	cli *cicdclient.CicdClientFacade
}

func (d *repositoryJobsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_repository_jobs", req.ProviderTypeName)
}

func (d *repositoryJobsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all jobs configured for a specific repository in the SAP BTP CI/CD service.",
		Attributes: map[string]schema.Attribute{
			"repository": schema.StringAttribute{
				MarkdownDescription: "The name or ID of the repository whose jobs are listed.",
				Required:            true,
			},
			"values": schema.ListNestedAttribute{
				MarkdownDescription: "The list of jobs configured for the repository.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: repositoryJobsDSItemAttributes(),
				},
			},
		},
	}
}

// repositoryJobsDSItemAttributes returns the shared attribute map used in both the schema and the
// nested object type definition.
func repositoryJobsDSItemAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Immutable technical ID of the job.",
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the job.",
			Computed:            true,
		},
		"active": schema.BoolAttribute{
			MarkdownDescription: "Whether the job is active. Inactive jobs cannot be executed.",
			Computed:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "Description of the job.",
			Computed:            true,
		},
		"pipeline": schema.StringAttribute{
			MarkdownDescription: "Pipeline type of the job (e.g. `sap-cloud-sdk`, `cpi`).",
			Computed:            true,
		},
		"pipeline_version": schema.StringAttribute{
			MarkdownDescription: "Version of the pipeline type.",
			Computed:            true,
		},
		"pipeline_parameters": schema.MapAttribute{
			MarkdownDescription: "Key-value parameters for the pipeline.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"build_retention_days": schema.Int64Attribute{
			MarkdownDescription: "Number of days builds of this job are retained.",
			Computed:            true,
		},
		"max_builds_to_keep": schema.Int64Attribute{
			MarkdownDescription: "Maximum number of builds retained for this job.",
			Computed:            true,
		},
		"branch": schema.StringAttribute{
			MarkdownDescription: "Branch pattern the job is executed for.",
			Computed:            true,
		},
		"repository_id": schema.StringAttribute{
			MarkdownDescription: "ID of the source repository used by the job.",
			Computed:            true,
		},
		"notification_configuration": schema.SingleNestedAttribute{
			MarkdownDescription: "Notification configuration for the job.",
			Computed:            true,
			Attributes: map[string]schema.Attribute{
				"ans": schema.SingleNestedAttribute{
					MarkdownDescription: "SAP Alert Notification Service configuration.",
					Computed:            true,
					Attributes: map[string]schema.Attribute{
						"active": schema.BoolAttribute{
							MarkdownDescription: "Whether alert notification is active.",
							Computed:            true,
						},
						"credential_id": schema.StringAttribute{
							MarkdownDescription: "ID of the credential containing the ANS service key.",
							Computed:            true,
						},
						"custom_tag": schema.StringAttribute{
							MarkdownDescription: "Custom tag value for the SAP Alert Notification Service.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *repositoryJobsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *repositoryJobsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config repositoryJobsDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repositoryReference := config.Repository.ValueString()

	list, err := d.cli.Jobs.ListByRepository(ctx, repositoryReference)
	if err != nil {
		if cicdmodels.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Repository Not Found",
				fmt.Sprintf("No repository found with reference %q.", repositoryReference),
			)
			return
		}
		resp.Diagnostics.AddError("Error Listing Repository Jobs", err.Error())
		return
	}

	state := repositoryJobsDSModel{
		Repository: config.Repository,
	}

	values, diags := repositoryJobsDSItemsFrom(ctx, list)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Values = values

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
