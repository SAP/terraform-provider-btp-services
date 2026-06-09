// btpservices/provider/cicd/jobs/action_run_build.go

package cicdjobs

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cicdclient "github.com/SAP/terraform-provider-btp-services/internal/cicd/client"
	cicdmodels "github.com/SAP/terraform-provider-btp-services/internal/cicd/models"
	"github.com/SAP/terraform-provider-btp-services/internal/shared"
)

var _ action.Action = &runBuildAction{}
var _ action.ActionWithConfigure = &runBuildAction{}

// NewRunBuildAction is the constructor for btpservice_cicd_run_build.
func NewRunBuildAction() action.Action {
	return &runBuildAction{}
}

type runBuildAction struct {
	cli *cicdclient.CicdClientFacade
}

// runBuildModel is the configuration model for the btpservice_cicd_run_build action.
type runBuildModel struct {
	Job             types.String      `tfsdk:"job"`
	JobETag         types.String      `tfsdk:"job_etag"`
	CommitToBeBuilt types.String      `tfsdk:"commit_to_be_built"`
	Parameters      []buildParamModel `tfsdk:"parameters"`
}

type buildParamModel struct {
	Name       types.String `tfsdk:"name"`
	Value      types.String `tfsdk:"value"`
	Visibility types.String `tfsdk:"visibility"`
}

func (a *runBuildAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_cicd_run_build", req.ProviderTypeName)
}

func (a *runBuildAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Triggers a CI/CD job build via the SAP BTP CI/CD service.",
		Attributes: map[string]schema.Attribute{
			"job": schema.StringAttribute{
				MarkdownDescription: "Name or ID of the CI/CD job to trigger.",
				Required:            true,
			},
			"job_etag": schema.StringAttribute{
				MarkdownDescription: "ETag of the job from `data.btpservice_cicd_job.*.etag`. " +
					"When set, the build is rejected with a 409 error if the job was modified " +
					"after the ETag was read. Omit to trigger the build unconditionally.",
				Optional: true,
			},
			"commit_to_be_built": schema.StringAttribute{
				MarkdownDescription: "Commit hash or branch name to build. " +
					"Required by the API if the job has an associated repository.",
				Optional: true,
			},
			"parameters": schema.ListNestedAttribute{
				MarkdownDescription: "Per-build runtime parameter overrides. Does not modify the job's stored pipeline parameters.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Parameter name (pattern: `[a-zA-Z0-9_-]*(\\.[a-zA-Z0-9_-]+)*`).",
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "Parameter value. Can be multi-line YAML content.",
							Required:            true,
						},
						"visibility": schema.StringAttribute{
							MarkdownDescription: "Controls visibility of this parameter in API responses. " +
								"`PUBLIC` (default) — value visible. `RESTRICTED` — value hidden from responses (use for secrets).",
							Optional: true,
							Validators: []validator.String{
								stringvalidator.OneOf("PUBLIC", "RESTRICTED"),
							},
						},
					},
				},
			},
		},
	}
}

func (a *runBuildAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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
			"A cicd{} block must be configured in the provider to use CI/CD actions.",
		)
		return
	}
	a.cli = clients.Cicd
}

func (a *runBuildAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config runBuildModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildReq := cicdmodels.BuildRequestDTO{}

	if !config.CommitToBeBuilt.IsNull() && !config.CommitToBeBuilt.IsUnknown() {
		buildReq.CommitToBeBuilt = config.CommitToBeBuilt.ValueString()
	}
	if !config.JobETag.IsNull() && !config.JobETag.IsUnknown() {
		buildReq.JobETag = config.JobETag.ValueString()
	}

	for _, p := range config.Parameters {
		param := cicdmodels.BuildParameter{
			Name:  p.Name.ValueString(),
			Value: p.Value.ValueString(),
		}
		if !p.Visibility.IsNull() && !p.Visibility.IsUnknown() {
			param.Visibility = p.Visibility.ValueString()
		}
		buildReq.Parameters = append(buildReq.Parameters, param)
	}

	if err := a.cli.Builds.Trigger(ctx, config.Job.ValueString(), buildReq); err != nil {
		if cicdmodels.IsConflict(err) {
			resp.Diagnostics.AddError(
				"Build Not Triggered — Job Configuration Changed",
				"The job was modified after the ETag was read. The build was not triggered.\n\n"+
					"To fix: re-run `terraform plan` to refresh the job data source and get the "+
					"current ETag, then apply again.",
			)
			return
		}
		resp.Diagnostics.AddError("Error Triggering Build", err.Error())
		return
	}
}
