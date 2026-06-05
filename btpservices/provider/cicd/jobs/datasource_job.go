// btpservices/provider/cicd/jobs/datasource_job.go

package cicdjobs

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"

	cicdclient "github.com/SAP/terraform-provider-btp-services/internal/cicd/client"
	cicdmodels "github.com/SAP/terraform-provider-btp-services/internal/cicd/models"
	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

var _ datasource.DataSource = &jobDataSource{}
var _ datasource.DataSourceWithConfigure = &jobDataSource{}

func NewJobDataSource() datasource.DataSource {
	return &jobDataSource{}
}

type jobDataSource struct {
	cli *cicdclient.CicdClientFacade
}

func (d *jobDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_job", req.ProviderTypeName)
}

func (d *jobDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a CI/CD job from the SAP BTP CI/CD service. " +
			"Exactly one of `id` or `name` must be set.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the job. Exactly one of `id` or `name` must be set.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("name"),
					}...),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the job. Exactly one of `id` or `name` must be set.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("id"),
					}...),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Human-readable description of the job.",
				Computed:            true,
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Whether the job is active.",
				Computed:            true,
			},
			"pipeline": schema.StringAttribute{
				MarkdownDescription: "Pipeline type of the job (e.g. `cf-env`, `cpi`).",
				Computed:            true,
			},
			"pipeline_version": schema.StringAttribute{
				MarkdownDescription: "Version of the pipeline type.",
				Computed:            true,
			},
			"pipeline_parameters": schema.StringAttribute{
				MarkdownDescription: "Pipeline parameters serialised as a canonical YAML string.",
				Computed:            true,
			},
			"build_retention_days": schema.Int64Attribute{
				MarkdownDescription: "Number of days build artifacts are retained.",
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
			"etag": schema.StringAttribute{
				MarkdownDescription: "Current ETag of the job. Pass this as `job_etag` in " +
					"`btpservice_cicd_run_build` to guard against triggering a build if the " +
					"job configuration was changed since the last plan.",
				Computed: true,
			},
			"notification_configuration": schema.SingleNestedAttribute{
				MarkdownDescription: "Notification settings for the job.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"ans": schema.SingleNestedAttribute{
						MarkdownDescription: "SAP Alert Notification Service (ANS) settings.",
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"active": schema.BoolAttribute{
								MarkdownDescription: "Whether ANS notifications are active for this job.",
								Computed:            true,
							},
							"credential_id": schema.StringAttribute{
								MarkdownDescription: "ID of the ANS credential to use.",
								Computed:            true,
							},
							"custom_tag": schema.StringAttribute{
								MarkdownDescription: "Optional custom tag added to ANS notifications.",
								Computed:            true,
							},
						},
					},
				},
			},
		},
	}
}

func (d *jobDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *jobDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config jobDSModel
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

	result, etag, err := d.cli.Builds.GetJobWithETag(ctx, reference)
	if err != nil {
		if cicdmodels.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Job Not Found",
				fmt.Sprintf("No job found with reference %q.", reference),
			)
			return
		}
		resp.Diagnostics.AddError("Error Reading Job", err.Error())
		return
	}

	state, err := jobDSValueFrom(*result, etag)
	if err != nil {
		resp.Diagnostics.AddError("Error Mapping Job Response", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
